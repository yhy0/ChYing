package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/panjf2000/ants/v2"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/spf13/cobra"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/mitmproxy"
	JieConf "github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/mode"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/mcpserver"
	"github.com/yhy0/logging"
)

var (
	proxyPort int
	mcpPort   int
	bindAddr  string
	project   string
	quiet     bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动被动扫描服务",
	Long:  "启动 HTTP 代理 + 被动扫描 + MCP SSE 服务",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntVar(&proxyPort, "proxy-port", 9080, "代理监听端口")
	serveCmd.Flags().IntVar(&mcpPort, "mcp-port", 9090, "MCP SSE 服务端口")
	serveCmd.Flags().StringVar(&bindAddr, "bind", "127.0.0.1", "监听地址 (Docker 场景用 0.0.0.0)")
	serveCmd.Flags().StringVar(&project, "project", "default", "项目名称")
	serveCmd.Flags().BoolVar(&quiet, "quiet", false, "静默模式，不输出流量和漏洞到终端")
}

func runServe(cmd *cobra.Command, args []string) error {
	// 1. 日志初始化
	logging.Logger = logging.New(true, file.ChyingDir, "ChYing-CLI", true)
	logging.Logger.Infoln("Starting ChYing CLI...")

	// 2. 文件系统
	file.New()

	// 3. 协程池
	pool, err := ants.NewPool(conf.Parallelism)
	if err != nil {
		return fmt.Errorf("创建协程池失败: %w", err)
	}
	defer pool.Release()

	// 4. 指纹识别引擎
	JieConf.Wappalyzer, _ = wappalyzer.New()

	// 5. 配置加载
	conf.InitConfig()
	conf.HotConf()
	conf.SyncJieConfig()

	// 关闭所有插件
	for k := range JieConf.Plugin {
		JieConf.Plugin[k] = false
	}

	// 6. 数据库
	db.Init(project, "sqlite")

	// 7. 被动模式
	mode.Passive()

	// 8. 覆盖代理配置
	conf.ProxyHost = bindAddr
	conf.ProxyPort = proxyPort
	conf.AppConf.Proxy.Host = bindAddr
	conf.AppConf.Proxy.Port = proxyPort

	// 9. 启动代理
	go func() {
		logging.Logger.Infof("Starting proxy on %s:%d", bindAddr, proxyPort)
		mitmproxy.Proxify()
		logging.Logger.Errorln("Proxy server has stopped.")
	}()
	time.Sleep(2 * time.Second)

	printStatus("Proxy listening on %s:%d", bindAddr, proxyPort)

	// 10. 启动 MCP Server
	mcpAddr, mcpErr := mcpserver.StartHTTPServer(mcpPort, bindAddr)
	if mcpErr != nil {
		return fmt.Errorf("MCP server 启动失败: %w", mcpErr)
	}
	printStatus("MCP server on %s/mcp", mcpAddr)

	// 11. 事件循环
	go cliEventNotification(pool)
	go cliVulnLoop()

	printStatus("ChYing CLI ready. Press Ctrl+C to stop.")

	// 12. 优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	printStatus("Shutting down...")
	return nil
}

// cliEventNotification 消费代理事件，入库
func cliEventNotification(pool *ants.Pool) {
	for _data := range mitmproxy.EventDataChan {
		if _data.Name == "HttpHistory" {
			_http, ok := _data.Data.(*mitmproxy.HTTPHistory)
			if !ok {
				continue
			}

			err := pool.Submit(func() {
				historyData := &db.HTTPHistory{
					Hid:         _http.Id,
					Host:        _http.Host,
					Method:      _http.Method,
					FullUrl:     _http.FullUrl,
					Path:        _http.Path,
					Status:      _http.Status,
					Length:      _http.Length,
					ContentType: _http.ContentType,
					MIMEType:    _http.MIMEType,
					Extension:   _http.Extension,
					Title:       _http.Title,
					IP:          _http.IP,
					Color:       _http.Color,
					Note:        _http.Note,
					Source:      "local",
					SourceID:    "localhost",
					NodeName:    "CLI",
					SessionID:   _http.SessionID,
				}
				db.AddHistory(historyData)

				if _http.MIMEType == "image" {
					return
				}

				// 存储请求/响应体
				traffic, loaded := mitmproxy.HTTPBodyMap.Load(_http.Id)
				if loaded {
					httpBody, typeOk := traffic.(*mitmproxy.HTTPBody)
					if typeOk {
						req := &db.Request{
							RequestId:  uint(_http.Id),
							Url:        _http.FullUrl,
							Path:       _http.Path,
							Host:       _http.Host,
							RequestRaw: httpBody.RequestRaw,
						}
						resp := &db.Response{
							RequestId:   uint(_http.Id),
							Url:         _http.FullUrl,
							Host:        _http.Host,
							Path:        _http.Path,
							ContentType: _http.ContentType,
							ResponseRaw: httpBody.ResponseRaw,
						}
						db.AddRequest(req, resp)
					}
				}

				if !quiet {
					printTraffic(_http.Method, _http.FullUrl, _http.Status, _http.Length)
				}
			})
			if err != nil {
				logging.Logger.Errorln("submit event task err:", err)
			}
		}
	}
}

// cliVulnLoop 消费漏洞通道，入库 + 终端输出
func cliVulnLoop() {
	for vuln := range output.OutChannel {
		vulnData := &db.Vulnerability{
			VulnID:      fmt.Sprintf("%s-%s-%d", vuln.VulnData.VulnType, vuln.VulnData.Target, time.Now().UnixNano()),
			VulnType:    vuln.VulnData.VulnType,
			Target:      vuln.VulnData.Target,
			Host:        vuln.VulnData.Target,
			Method:      vuln.VulnData.Method,
			Plugin:      vuln.Plugin,
			Level:       vuln.Level,
			IP:          vuln.VulnData.Ip,
			Param:       vuln.VulnData.Param,
			Payload:     vuln.VulnData.Payload,
			Description: vuln.VulnData.Description,
			CurlCommand: vuln.VulnData.CURLCommand,
			Request:     vuln.VulnData.Request,
			Response:    vuln.VulnData.Response,
			Source:      "local",
			SourceID:    "localhost",
			NodeName:    "CLI",
			ProjectID:   project,
		}
		_ = db.AddVulnerability(vulnData)

		if !quiet {
			printVuln(vuln.Level, vuln.VulnData.VulnType, vuln.VulnData.Target, vuln.VulnData.Param)
		}
	}
}

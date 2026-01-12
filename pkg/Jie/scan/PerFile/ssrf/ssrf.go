package ssrf

import (
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/reverse"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
	"runtime"
	"strings"
	"sync"
	"time"
)

/**
  @author: yhy
  @since: 2023/1/30
  @desc:
**/

// 参数中包含以下关键字的，进行 ssrf 测试
var sensitiveWords = []string{"url", "u", "uri", "api", "a", "target", "host", "h", "domain", "ip", "file", "f", "fi", "page", "path", "p"}

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if p.IsScanned(in.UniqueId) {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			logging.Logger.Errorln("recover from:", err)
			debugStack := make([]byte, 1024)
			runtime.Stack(debugStack, false)
			logging.Logger.Errorf("Stack Trace:%v", string(debugStack))

		}
	}()

	variations, err := httpx.ParseUri(in.Url, []byte(in.RequestBody), in.Method, in.ContentType, in.Headers)
	if err != nil {
		if strings.Contains(err.Error(), "data is empty") {
			logging.Logger.Debugln(err.Error())
		} else {
			logging.Logger.Errorln(err.Error())
		}
		return
	}

	var ssrfHost string
	var dnslog *reverse.Dig
	if conf.GlobalConfig.Reverse.Host != "" {
		dnslog = reverse.GetSubDomain()
		if dnslog == nil {
			ssrfHost = "https://www.baidu.com/"
		} else {
			ssrfHost = dnslog.Domain
		}
	} else {
		ssrfHost = "https://www.baidu.com/"
	}

	// 1. 这里先对一些可以参数进行测试，如果找到了，就不再进行下面的测试
	if ssrf(in, variations, ssrfHost, dnslog, client) {
		return
	}

	payloads := []string{"/etc/passwd", "c:/windows/win.ini"}

	if readFile(in, variations, payloads, client) {
		return
	}

	// 2. 如果没有找到，就对一些特殊的请求头进行测试，这里依赖于 dnslog 这里分两步: 这主要是防止请求头过多被拦截
	//     step1: 如果有特殊的请求头，就对特殊的请求头进行测试
	//    step2: 如果没有特殊的请求头，或者第一步没有测试出来，就对所有的危险请求头进行测试
	if dangerHeader(in, client) {
		return
	}

	logging.Logger.Debugln(in.Url, "ssrf vulnerability not found")
}

func (p *Plugin) IsScanned(key string) bool {
	if key == "" {
		return false
	}
	if _, ok := p.SeenRequests.Load(key); ok {
		return true
	}
	p.SeenRequests.Store(key, true)
	return false
}

func (p *Plugin) Name() string {
	return "ssrf"
}

// ssrf
func ssrf(in *input.CrawlResult, variations *httpx.Variations, payload string, dnslog *reverse.Dig, client *httpx.Client) bool {
	for _, p := range variations.Params {
		if !util.SliceInCaseFold(p.Name, sensitiveWords) {
			continue
		}

		payload = variations.SetPayloadByIndex(p.Index, in.Url, payload, in.Method)
		if payload == "" {
			continue
		}
		logging.Logger.Debugln("ssrf ", payload)
		var res *httpx.Response
		var err error
		if in.Method == "GET" {
			res, err = client.Request(payload, in.Method, "", in.Headers, "SSRF")
		} else {
			res, err = client.Request(in.Url, in.Method, payload, in.Headers, "SSRF")
		}

		if err != nil {
			logging.Logger.Errorln(err)
			continue
		}

		isVul := false
		var desc = ""
		if dnslog != nil {
			isVul = reverse.PullLogs(dnslog)
			desc = dnslog.Key + " key: " + dnslog.Token
		} else {
			isVul = funk.Contains(res.Body, "<title>百度一下，你就知道</title>")
			desc = "<title>百度一下，你就知道</title>"
		}

		if isVul {
			output.OutChannel <- output.VulMessage{
				DataType: "web_vul",
				Plugin:   "SSRF",
				VulnData: output.VulnData{
					CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
					Target:      in.Url,
					Method:      in.Method,
					Ip:          in.Ip,
					Param:       in.Kv,
					Request:     res.RequestDump,
					Response:    res.ResponseDump,
					Payload:     payload,
					Description: desc,
				},
				Level: output.Critical,
			}
			return true
		}
	}
	return false
}

// readFile 任意文件读取
func readFile(in *input.CrawlResult, variations *httpx.Variations, payloads []string, client *httpx.Client) bool {
	for _, p := range variations.Params {
		if !util.SliceInCaseFold(p.Name, sensitiveWords) {
			continue
		}
		for _, payload := range payloads {
			payload = variations.SetPayloadByIndex(p.Index, in.Url, payload, in.Method)
			if payload == "" {
				continue
			}
			logging.Logger.Debugln("ssrf ", payload)
			var res *httpx.Response
			var err error

			if in.Method == "GET" {
				res, err = client.Request(payload, in.Method, "", in.Headers, "SSRF")
			} else {
				res, err = client.Request(in.Url, in.Method, payload, in.Headers, "SSRF")
			}

			if err != nil {
				logging.Logger.Errorln(err)
				continue
			}

			if funk.Contains(res.Body, "root:x:0:0:root:/root:") || funk.Contains(res.Body, "root:[x*]:0:0:") || funk.Contains(res.Body, "; for 16-bit app support") {
				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "READ-FILE",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     in.Url,
						Method:     in.Method,
						Ip:         in.Ip,
						Param:      in.Kv,
						Request:    res.RequestDump,
						Response:   res.ResponseDump,
						Payload:    payload,
					},
					Level: output.Critical,
				}
				return true
			}
		}

	}
	return false
}

func dangerHeader(in *input.CrawlResult, client *httpx.Client) bool {
	if conf.GlobalConfig.Reverse.Host != "" {
		dnslog := reverse.GetSubDomain()
		if dnslog == nil {
			return false
		}

		// step1: 如果有特殊的请求头，就对特殊的请求头进行测试
		for _, h := range conf.DangerHeaders {
			if in.Headers[h] != "" {
				in.Headers[h] = h + "." + dnslog.Domain
			}
		}
		if header(in, dnslog, client) {
			return true
		}

		//    step2: 如果没有特殊的请求头，或者第一步没有测试出来，就对所有的危险请求头进行测试
		for _, h := range conf.DangerHeaders {
			in.Headers[h] = h + "." + dnslog.Domain
		}

		if header(in, dnslog, client) {
			return true
		}
	}

	return false
}

func header(in *input.CrawlResult, dnslog *reverse.Dig, client *httpx.Client) bool {
	res, err := client.Request(in.Url, in.Method, in.RequestBody, in.Headers, "SSRF")
	if err != nil {
		logging.Logger.Errorln(err)
		return false
	}

	isVul := reverse.PullLogs(dnslog)

	if isVul {
		output.OutChannel <- output.VulMessage{
			DataType: "web_vul",
			Plugin:   "SSRF",
			VulnData: output.VulnData{
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				Target:     in.Url,
				Method:     in.Method,
				Ip:         in.Ip,
				Param:      in.Kv,
				Request:    res.RequestDump,
				Response:   res.RequestDump,
				Payload:    "token: " + dnslog.Token + " msg: " + dnslog.Msg,
			},
			Level: output.Critical,
		}
		return true
	}
	return false
}

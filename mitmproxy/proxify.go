package mitmproxy

import (
	"bytes" // 新增，用于 GetHTTPBody 返回错误
	"fmt"
	"io"
	"math" // 需要导入 math 包以使用 math.MaxInt
	"net"
	"net/http"
	"net/http/httputil"
	"os"            // 新增 os 包用于获取临时目录
	"path/filepath" // 新增 filepath 包用于路径操作
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yhy0/logging"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/martian/v3"
	"github.com/projectdiscovery/proxify"
	"github.com/projectdiscovery/proxify/pkg/certs"
	"github.com/projectdiscovery/proxify/pkg/logger/elastic"
	"github.com/projectdiscovery/proxify/pkg/logger/kafka"
	proxifyTypes "github.com/projectdiscovery/proxify/pkg/types"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/conf/file"
)

/**
   @author yhy
   @since 2025/5/20
   @desc 实现Proxify代理的核心逻辑
**/

// 移除 ModifiedData，已迁移到 intercept.go

var (
	lock sync.Mutex

	// currentProxy 当前运行的代理实例
	currentProxy     *proxify.Proxy
	currentProxyLock sync.RWMutex

	// TempHistoryCache 用于在请求到达和响应到达之间临时存储HTTP事务的摘要信息。
	// 键是 flowID (来自martian.Context，用于关联请求和响应)，值是 *data.HTTPHistory (不包含原始请求/响应报文)。
	// 当请求到达时(onRequestCallback)，会创建条目；当相应响应到达并处理完毕后(onResponseCallback)，条目会被取出并删除。
	// 主要服务于代理内部的请求-响应匹配和数据组装，不直接服务前端。
	TempHistoryCache sync.Map

	// TempRequestRawCache 用于在请求到达和响应到达之间临时存储原始的请求报文字符串。
	// 键是 flowID，值是 string (原始请求报文)。
	// 之所以需要它，是因为 data.HTTPHistory 不再存储RequestRaw，但我们在响应到达时仍然需要原始请求报文
	// (可能是被拦截修改后的版本) 来填充最终的 HTTPBodyMap。
	// 当请求到达时创建条目；当相应响应到达并用于填充HTTPBodyMap后，条目被取出并删除。
	TempRequestRawCache sync.Map

	// HTTPBodyMap 是最终存储每个HTTP事务完整原始请求和响应报文的地方。
	// 键是 historyId (即 data.HTTPHistory.Id, int64)，值是 *data.HTTPBody (包含RequestRaw和ResponseRaw)。
	// 当前端需要显示某个历史记录的详情时，会根据 historyId 从这个Map中获取对应的原始报文。
	// 在 onResponseCallback 的末尾填充，当事务完全结束后。

	// HTTPBodyMap 存储 mitmproxy 的响应信息, 为什么不直接放到HttpHistory，是为了防止太多的请求/响应数据加载到前端，这样做只有前端点击每行数据时才会加载对应的数据到前端显示

	HTTPBodyMap sync.Map

	HistoryItemIDGenerator atomic.Int64

	// HTTPUrlMap url 和 HTTPBodyMap 中的id 映射, 首页点击时使用
	// 使用 sync.Map 保证并发安全
	HTTPUrlMap sync.Map

	// IntruderMap 使用 sync.Map 保证并发安全
	IntruderMap sync.Map
)

const defaultCertCacheSize = 256

func init() {
	// HTTPUrlMap 和 IntruderMap 现在是 sync.Map，不需要初始化
	// 启动缓存清理器
	go startCacheCleanup()
}

// InitProxifyState 初始化代理状态，应该在应用启动时调用一次
func InitProxifyState() {
	// 初始化拦截系统
	InitInterceptSystem()

	// 清除所有处理器，确保重新初始化时没有残留
	ClearAllProcessors()

	// 初始化匹配替换规则
	if err := InitMatchReplaceRules(); err != nil {
		logging.Logger.Infof("Failed to initialize match replace rules: %v\n", err)
	}

	// 初始化越权检测模块
	if err := InitAuthorizationChecker(); err != nil {
		logging.Logger.Infof("Failed to initialize authorization checker: %v\n", err)
	}

	// 初始化被动扫描插件
	InitPassiveScanPlugin()

	fmt.Println("Proxify state initialized.")
}

func getDefaultProxifyOptions() *proxify.Options {
	// 将所有数据统一放到 /Users/yhy/.config/ChYing/proxify_data/ 目录下
	baseAppConfigDir := filepath.Join(file.ChyingDir)
	proxifyDataDir := filepath.Join(baseAppConfigDir, "proxify_data")

	// 确保 proxify 数据目录存在
	if err := os.MkdirAll(proxifyDataDir, 0755); err != nil {
		logging.Logger.Infof("CRITICAL: Could not create proxify data directory %s: %v. Proxify may fail to start.\n", proxifyDataDir, err)
		// 如果无法创建关键目录，可能需要更强的错误处理或回退到临时目录
		// 为了简单起见，这里仅打印严重错误
	}
	logging.Logger.Infof("Proxify data directory set to: %s\n", proxifyDataDir)

	// 从配置文件获取代理监听地址
	proxyHost := "127.0.0.1"
	proxyPort := conf.ProxyPort

	// 优先使用配置文件中的设置
	if conf.AppConf.Proxy.Host != "" {
		proxyHost = conf.AppConf.Proxy.Host
	}
	if conf.AppConf.Proxy.Port > 0 {
		proxyPort = conf.AppConf.Proxy.Port
	}

	// 如果有启用的监听器，使用第一个启用的监听器配置
	for _, listener := range conf.AppConf.Proxy.Listeners {
		if listener.Enabled {
			proxyHost = listener.Host
			proxyPort = listener.Port
			break
		}
	}

	// 更新全局 ProxyPort 和 ProxyHost 变量以保持一致性
	conf.ProxyPort = proxyPort
	conf.ProxyHost = proxyHost

	return &proxify.Options{
		// Output
		OutputDirectory: proxifyDataDir,                                         // 日志和潜在的dump文件存储在这里
		OutputFile:      filepath.Join(proxifyDataDir, "proxify_traffic.jsonl"), // 默认日志输出文件
		OutputFormat:    "jsonl",                                                // 默认输出格式
		DumpRequest:     false,                                                  // 是否dump请求体到单独文件
		DumpResponse:    false,                                                  // 是否dump响应体到单独文件

		// Certificate related
		Directory:     proxifyDataDir, // ****** 重要：proxify加载/生成CA证书(ca.crt, ca.key)的目录 ******
		CertCacheSize: defaultCertCacheSize,
		// OutCAFile:     "", // ****** 移除：此字段不存在，proxify依赖Directory管理CA ******

		// Filter (Defaults to nil, meaning no filters active)
		RequestDSL:              nil,
		ResponseDSL:             nil,
		RequestMatchReplaceDSL:  nil,
		ResponseMatchReplaceDSL: nil,

		// Network
		ListenAddrHTTP:      fmt.Sprintf("%s:%d", proxyHost, proxyPort), // HTTP代理监听地址，使用配置的端口
		ListenAddrSocks5:    "",               // SOCKS5代理监听地址 (如果不需要则为空)
		ListenDNSAddr:       "",               // DNS服务器监听地址 (如果不需要则为空)
		DNSMapping:          "",               // DNS映射
		DNSFallbackResolver: "",               // 后备DNS解析器

		// Proxy
		UpstreamHTTPProxies:         nil,
		UpstreamSock5Proxies:        nil, // 修正：根据定义，字段名为 UpstreamSock5Proxies (小写s)
		UpstreamProxyRequestsNumber: 1,   // 切换上游代理的请求数

		// Export
		MaxSize: math.MaxInt, // 默认不限制导出大小

		// Configuration
		// ConfigDir: proxifyDataDir, // 移除：此字段不存在

		Allow:       nil,
		Deny:        nil,
		PassThrough: nil,

		// Debug / Verbosity related fields
		Verbosity: proxifyTypes.VerbosityDefault, // 使用 Verbosity 控制日志级别
		// Silent:    true, // 移除：通过 Verbosity 控制
		// NoColor:   true, // 移除：通过 Verbosity 或底层日志库控制

		// Specific Exporters Options - ensure these are pointers to structs
		Elastic: &elastic.Options{},
		Kafka:   &kafka.Options{},

		OutputJsonl: true,
	}
}

// ProxifyWithOptions 启动 Proxify 代理服务，接受一个 Options 参数
// 这是实际的启动函数，Proxify() 可以是一个简化的调用此函数的包装
func ProxifyWithOptions(options *proxify.Options) (*proxify.Proxy, error) {
	if options == nil {
		fmt.Println("ProxifyWithOptions called with nil options, using defaults.")
		options = getDefaultProxifyOptions()
	} else {
		// 如果外部传入了 options，我们仍然可以检查并补充关键目录（如果为空）
		// 但通常期望调用者如果提供了 options，则已正确配置
		defaultOptsForFallback := getDefaultProxifyOptions()

		if options.OutputDirectory == "" {
			options.OutputDirectory = defaultOptsForFallback.OutputDirectory
			logging.Logger.Infof("OutputDirectory was empty, set to default: %s\n", options.OutputDirectory)
		}
		if options.Directory == "" {
			options.Directory = defaultOptsForFallback.Directory
			logging.Logger.Infof("Certificate/Config Directory was empty, set to default: %s\n", options.Directory)
		}
		if options.OutputFile == "" {
			options.OutputFile = defaultOptsForFallback.OutputFile
		}
		if options.OutputFormat == "" {
			options.OutputFormat = defaultOptsForFallback.OutputFormat
		}
		// if options.ConfigDir == "" { // 移除：此字段不存在
		// 	options.ConfigDir = defaultOptsForFallback.ConfigDir
		// }
		if options.Elastic == nil {
			options.Elastic = &elastic.Options{}
		}
		if options.Kafka == nil {
			options.Kafka = &kafka.Options{}
		}
	}

	logging.Logger.Infof("ProxifyWithOptions effective options: ListenAddrHTTP=%s, OutputDir=%s, CertDir=%s\n",
		options.ListenAddrHTTP, options.OutputDirectory, options.Directory)

	// ++ 在调用 NewProxy 之前加载/生成证书 ++
	if options.Directory == "" {
		// 确保 options.Directory 有一个有效值，如果 getDefaultProxifyOptions 没有设置的话
		// 这段逻辑理论上不应该执行，因为 getDefaultProxifyOptions 会设置它
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			tmpDir := os.TempDir()
			gologger.Warning().Msgf("Could not get user config directory (%v), using temp directory for certs: %s", err, tmpDir)
			options.Directory = filepath.Join(tmpDir, "chying_proxify_certs") // 使用特定于ChYing的子目录
		} else {
			options.Directory = filepath.Join(userConfigDir, "ChYing", "proxify_certs")
		}
		if err := os.MkdirAll(options.Directory, 0700); err != nil {
			return nil, fmt.Errorf("could not create certificate directory %s for ChYing: %w", options.Directory, err)
		}
		gologger.Info().Msgf("ChYing: Certs directory was empty, resolved to: %s", options.Directory)
	}

	if err := certs.LoadCerts(options.Directory); err != nil {
		return nil, fmt.Errorf("ChYing: failed to load/generate certificates in %s: %w", options.Directory, err)
	}
	gologger.Info().Msgf("ChYing: Certificates loaded/generated successfully from/to directory: %s", options.Directory)
	// -- 证书处理完毕 --

	// 设置核心回调
	options.OnRequestCallback = onRequestCallback
	options.OnResponseCallback = onResponseCallback

	proxyInstance, err := proxify.NewProxy(options)
	if err != nil {
		// Provide more analysis in error if possible
		return nil, fmt.Errorf("failed to create proxify instance: %w (using ListenAddr: %s, OutputDir: %s, CertDir: %s)",
			err, options.ListenAddrHTTP, options.OutputDirectory, options.Directory)
	}

	fmt.Println("Proxify instance created. Starting proxy server...")
	return proxyInstance, nil
}

// Proxify 是一个简化版本，可以使用默认配置或从全局配置中读取
func Proxify() {
	InitProxifyState()

	defaultOptions := getDefaultProxifyOptions()

	fmt.Println("Attempting to start Proxify with default options...")
	proxy, err := ProxifyWithOptions(defaultOptions)
	if err != nil {
		logging.Logger.Infof("Error starting Proxify with default options: %v\n", err)
		return
	}
	logging.Logger.Infof("Proxify proxy server starting on %s\n", defaultOptions.ListenAddrHTTP)

	// 存储当前代理实例
	currentProxyLock.Lock()
	currentProxy = proxy
	fmt.Printf("[Proxify] 代理实例已存储: %p\n", currentProxy)
	currentProxyLock.Unlock()

	// 解析监听地址获取 host 和 port
	listenAddr := defaultOptions.ListenAddrHTTP
	host := "127.0.0.1"
	port := "9080"
	if colonIdx := strings.LastIndex(listenAddr, ":"); colonIdx != -1 {
		host = listenAddr[:colonIdx]
		port = listenAddr[colonIdx+1:]
	}

	// 更新代理状态为运行中
	StartMitmproxy(host, port, "passive")

	// Run() 是阻塞的
	if err := proxy.Run(); err != nil {
		logging.Logger.Infof("Proxify proxy server run error: %v\n", err)
		// 代理停止时更新状态
		StopMitmproxy()
	}

	// 清理代理实例
	currentProxyLock.Lock()
	currentProxy = nil
	currentProxyLock.Unlock()
}

// StopCurrentProxy 停止当前运行的代理
func StopCurrentProxy() {
	currentProxyLock.Lock()
	defer currentProxyLock.Unlock()

	if currentProxy != nil {
		logging.Logger.Info("正在停止代理...")
		fmt.Printf("[StopCurrentProxy] 正在停止代理实例: %p\n", currentProxy)
		currentProxy.Stop()
		// 给代理一些时间来清理资源
		time.Sleep(100 * time.Millisecond)
		currentProxy = nil
		logging.Logger.Info("代理已停止")
	} else {
		fmt.Println("[StopCurrentProxy] currentProxy 为 nil，无需停止")
	}
}

// RestartProxy 重启代理服务
// 注意：由于 proxify 的 Run() 是阻塞的，重启需要在新的 goroutine 中进行
func RestartProxy() error {
	logging.Logger.Info("正在重启代理服务...")

	// 停止当前代理
	StopCurrentProxy()
	StopMitmproxy()

	// 获取当前配置的端口和主机
	host := conf.AppConf.Proxy.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := conf.AppConf.Proxy.Port
	if port == 0 {
		port = 9080
	}

	// 等待端口释放，最多等待 10 秒
	// 使用尝试监听的方式来检查端口是否可用，这比尝试连接更可靠
	maxRetries := 20
	addr := fmt.Sprintf("%s:%d", host, port)
	portReleased := false

	for i := 0; i < maxRetries; i++ {
		// 尝试在该端口上创建监听器来检查端口是否可用
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			// 端口可用，立即关闭监听器
			listener.Close()
			logging.Logger.Infof("端口 %d 已释放，准备重启代理", port)
			portReleased = true
			break
		}

		// 端口仍被占用，等待后重试
		if i < maxRetries-1 {
			logging.Logger.Infof("等待端口 %d 释放... (%d/%d)", port, i+1, maxRetries)
			time.Sleep(500 * time.Millisecond)
		}
	}

	if !portReleased {
		return fmt.Errorf("端口 %d 仍被占用，无法重启代理（等待超时）", port)
	}

	// 在新的 goroutine 中启动代理
	go func() {
		logging.Logger.Info("正在启动新的代理实例...")
		Proxify()
	}()

	return nil
}

// onRequestCallback 是 martian 的请求修饰回调
func onRequestCallback(req *http.Request, ctx *martian.Context) error {
	// 在任何处理前先应用过滤
	if req != nil && req.URL != nil {
		if Filter(req.URL.Host) {
			logging.Logger.Debugln("过滤了", req.URL.Host)
			// 返回 nil 表示请求继续处理，但后续的插件将忽略它
			// 这里不返回错误，因为这只是过滤，不是错误情况
			return nil
		}
		if req.Method == "CONNECT" || req.Method == "OPTIONS" {
			return nil
		}
	}

	flowID := ctx.ID()

	// 1. 统一读取和缓存请求Body，避免后续重复读取导致的问题
	var requestBodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		requestBodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			logging.Logger.Errorf("onRequest: Failed to read request body for flow %s: %v", flowID, err)
			requestBodyBytes = []byte{} // 使用空字节数组作为回退
		}
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
	}

	// Body 已完整读取到内存，移除 chunked 编码标记，避免 DumpRequestOut 输出 chunk 格式
	if req.TransferEncoding != nil {
		for _, te := range req.TransferEncoding {
			if strings.EqualFold(te, "chunked") {
				req.TransferEncoding = nil
				req.ContentLength = int64(len(requestBodyBytes))
				// 同时更新 Header
				req.Header.Del("Transfer-Encoding")
				if len(requestBodyBytes) > 0 {
					req.Header.Set("Content-Length", strconv.Itoa(len(requestBodyBytes)))
				}
				break
			}
		}
	}

	// 2. 生成初始请求dump
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logging.Logger.Infof("onRequest: Failed to dump request: %v\n", err)
	}
	requestRawString := string(reqDump)

	tempHistoryID := HistoryItemIDGenerator.Add(1)
	tempEntry := &HTTPHistory{
		Id:        tempHistoryID,
		FlowID:    flowID,
		Host:      req.URL.Host,
		Method:    req.Method,
		FullUrl:   req.URL.String(),
		Path:      req.URL.RequestURI(),
		Timestamp: time.Now().Format(time.RFC3339Nano),
		// RequestRaw 字段已在 data.HTTPHistory 中移除
	}
	TempHistoryCache.Store(flowID, tempEntry)
	TempRequestRawCache.Store(flowID, requestRawString) // 单独存储 RequestRaw

	// 3. 用户拦截修改（优先执行，获取最终要发送的请求）
	if err := HandleRequestIntercept(req, ctx, flowID); err != nil {
		// 如果是用户丢弃或其他拦截错误，清理缓存并返回错误
		TempHistoryCache.Delete(flowID)
		TempRequestRawCache.Delete(flowID)
		return err
	}

	// 4. 应用修改类插件处理器（如 matchreplace）
	pluginsModified := ProcessModifyingRequest(req)
	if pluginsModified {
		// 如果请求被插件处理器修改，更新缓存的Body和RequestRaw
		if req.Body != nil && req.Body != http.NoBody {
			newBodyBytes, readErr := io.ReadAll(req.Body)
			if readErr == nil {
				requestBodyBytes = newBodyBytes
				req.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
			}
		}
	}

	// 5. 更新最终的RequestRaw到缓存
	finalReqDump, dumpErr := httputil.DumpRequestOut(req, true)
	if dumpErr == nil {
		TempRequestRawCache.Store(flowID, string(finalReqDump))
	}

	// 6. 应用只读类插件处理器（如 authcheck）- 基于最终请求进行分析
	ProcessReadOnlyRequest(req, requestBodyBytes)

	return nil
}

// onResponseCallback 是 martian 的响应修饰回调
func onResponseCallback(resp *http.Response, ctx *martian.Context) error {
	flowID := ctx.ID()

	// 1. 统一读取和缓存响应Body，避免后续重复读取导致的问题
	var responseBodyBytes []byte
	if resp.Body != nil && resp.Body != http.NoBody {
		var err error
		responseBodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			logging.Logger.Errorf("onResponse: Failed to read response body for flow %s: %v", flowID, err)
			responseBodyBytes = []byte{} // 使用空字节数组作为回退
		}
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBodyBytes))
	}

	var respModifiedByUserOrPlugin bool // 用于跟踪响应是否被用户或插件修改

	// 2. 用户拦截修改（优先执行）
	respModifiedByIntercept, interceptErr := HandleResponseIntercept(resp, ctx, flowID)
	if interceptErr != nil {
		// 如果是用户丢弃或其他拦截错误，清理缓存并返回错误
		TempHistoryCache.Delete(flowID)
		TempRequestRawCache.Delete(flowID)
		return interceptErr
	}
	if respModifiedByIntercept {
		respModifiedByUserOrPlugin = true
		// 更新缓存的Body数据
		if resp.Body != nil && resp.Body != http.NoBody {
			newBodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				responseBodyBytes = newBodyBytes
				resp.Body = io.NopCloser(bytes.NewBuffer(responseBodyBytes))
			}
		}
	}

	// 3. 应用修改类插件处理器（如 matchreplace）
	pluginsModified := ProcessModifyingResponse(resp)
	if pluginsModified {
		respModifiedByUserOrPlugin = true
		// 如果响应被插件处理器修改，更新缓存的Body数据
		if resp.Body != nil && resp.Body != http.NoBody {
			newBodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				responseBodyBytes = newBodyBytes
				resp.Body = io.NopCloser(bytes.NewBuffer(responseBodyBytes))
			}
		}
	}

	// 4. 统一处理响应体 (解码、获取原始编码等) - 基于最终响应
	processedInfo, processErr := ProcessResponseBody(resp)
	if processErr != nil {
		logging.Logger.Errorf("onResponse: ProcessResponseBody failed for flow %s: %v.", flowID, processErr)
		// 如果处理失败，创建一个回退的 processedInfo
		processedInfo = &ProcessedResponseBody{
			OriginalEncoding: resp.Header.Get("Content-Encoding"),
			ContentType:      resp.Header.Get("Content-Type"),
			IsText:           HttpIsTextContent(resp.Header.Get("Content-Type")),
			Content:          responseBodyBytes, // 使用缓存的Body数据作为回退
		}
	}

	// 5. 应用只读类插件处理器（如 passiveScan）- 基于最终响应进行分析
	ProcessReadOnlyResponse(resp, responseBodyBytes)
	// 确保 processedInfo 不是 nil (即使所有处理都失败了)
	if processedInfo == nil { // 这理论上不应发生，因为上面有回退逻辑
		processedInfo = &ProcessedResponseBody{Content: []byte{}}
	}
	if respModifiedByUserOrPlugin {
		SetContentModified(processedInfo) // 使用 httputil.SetContentModified
	}

	// 3. 生成用于 HTTPBodyMap 存储的 responseRawString (确保是解压且可读的)
	responseRawStringForMap := dumpResponseFromProcessedInfo(resp, processedInfo)

	// 4. 处理缓存和历史记录存储
	cachedTempEntry, loaded := TempHistoryCache.LoadAndDelete(flowID)
	if !loaded {
		TempRequestRawCache.Delete(flowID)
		return nil
	}
	tempEntry, ok := cachedTempEntry.(*HTTPHistory)
	if !ok {
		logging.Logger.Infof("onResponse: Cached temp history entry for flow %s is not of type *HTTPHistory.\n", flowID)
		TempRequestRawCache.Delete(flowID)
		return nil
	}
	rawReqData, rawReqLoaded := TempRequestRawCache.LoadAndDelete(flowID)
	var requestRawForBody string
	if rawReqLoaded {
		if reqStr, okCast := rawReqData.(string); okCast {
			requestRawForBody = reqStr
		} else {
			logging.Logger.Infof("onResponse: Warning - requestRaw data for flow %s is not a string.\n", flowID)
		}
	} else {
		logging.Logger.Infof("onResponse: Warning - No requestRaw data found in TempRequestRawCache for flow %s.\n", flowID)
	}

	// 5. 使用处理后的信息（可能被插件修改过）更新原始 resp，准备发送给客户端
	// UpdateResponseWithProcessedBody 会处理压缩并更新 resp.Body, resp.ContentLength, resp.Header["Content-Encoding"]
	updateErr := UpdateResponseWithProcessedBody(resp, processedInfo, true) // true表示尝试恢复压缩
	if updateErr != nil {
		logging.Logger.Errorf("onResponse: UpdateResponseWithProcessedBody failed for flow %s: %v. Client might receive incorrect response.", flowID, updateErr)
	}

	// 6. 创建用于事件和存储的 HTTPHistory 对象
	// 提取关键字段信息
	contentType := resp.Header.Get("Content-Type")
	mimeType := extractMIMEType(contentType)
	extension := extractExtension(tempEntry.Path)
	title := extractTitle(string(responseBodyBytes), contentType)
	ip := resolveHostIP(tempEntry.Host)

	historyForEvent := &HTTPHistory{
		Id:                tempEntry.Id,
		FlowID:            tempEntry.FlowID,
		Host:              tempEntry.Host,
		Method:            tempEntry.Method,
		FullUrl:           tempEntry.FullUrl,
		Path:              tempEntry.Path,
		Status:            strconv.Itoa(resp.StatusCode),
		Length:            strconv.FormatInt(resp.ContentLength, 10), // 使用最终更新后的ContentLength
		ContentType:       contentType,
		MIMEType:          mimeType,  // ✅ 新增：MIME类型
		Extension:         extension, // ✅ 新增：文件扩展名
		Title:             title,     // ✅ 新增：页面标题
		IP:                ip,        // ✅ 新增：IP地址
		Note:              "",        // ✅ 新增：备注（默认为空）
		Timestamp:         tempEntry.Timestamp,
		ResponseTimestamp: time.Now().Format(time.RFC3339Nano),
	}

	httpBody := &HTTPBody{
		Id:          tempEntry.Id,
		FlowID:      tempEntry.FlowID,
		RequestRaw:  requestRawForBody,
		ResponseRaw: responseRawStringForMap, // 使用从processedInfo生成的解压后dump
	}
	HTTPBodyMap.Store(tempEntry.Id, httpBody)

	if EventDataChan != nil {
		httpHistoryEvent := &EventData{
			Name: "HttpHistory",
			Data: historyForEvent,
		}
		HTTPUrlMap.Store(tempEntry.FullUrl, historyForEvent.Id)
		EventDataChan <- httpHistoryEvent
	} else {
		fmt.Println("Warning: EventDataChan is nil. Cannot send 'HttpHistory' event.")
	}

	// martian 框架将使用这个 resp 对象发送响应给客户端。
	// resp.Body 和相关头部 (Content-Length, Content-Encoding) 已被 UpdateResponseWithProcessedBody 更新。
	return nil
}

// dumpResponseFromProcessedInfo 根据处理后的响应信息（主要是解码后的内容和原始头部）生成一个用于显示的字符串 dump。
// 这个 dump 的 Body 部分是解码后的文本，Content-Encoding 头被移除。
func dumpResponseFromProcessedInfo(originalResp *http.Response, info *ProcessedResponseBody) string {
	var sb strings.Builder

	// 构造 Status-Line (从原始响应获取，如果可用)
	statusLine := fmt.Sprintf("%s %s", originalResp.Proto, originalResp.Status)
	if originalResp.Status == "" { // 回退，如果 Status 为空
		statusLine = fmt.Sprintf("HTTP/%d.%d %03d %s", originalResp.ProtoMajor, originalResp.ProtoMinor, originalResp.StatusCode, http.StatusText(originalResp.StatusCode))
	}
	sb.WriteString(statusLine)
	sb.WriteString("\r\n")

	// 准备头部进行写入
	headersToDump := originalResp.Header.Clone()                         // 从原始响应的当前头部开始克隆
	headersToDump.Del("Content-Encoding")                                // 因为内容是解码的
	headersToDump.Set("Content-Length", strconv.Itoa(len(info.Content))) // 设置为解码后内容的长度
	if info.OriginalEncoding != "" && info.OriginalEncoding != "identity" {
		headersToDump.Set("X-ChYing-Decoded-Original-Encoding", info.OriginalEncoding)
	}

	// 写入头部
	err := headersToDump.Write(&sb)
	if err != nil {
		logging.Logger.Errorf("dumpResponseFromProcessedInfo: Error writing headers: %v", err)
		// 即使头部写入失败，也尝试继续dump body
	}
	sb.WriteString("\r\n")

	// 写入解码后的Body内容
	sb.Write(info.Content)

	return sb.String()
}

// cloneRequest 创建请求的副本，特别处理 Body 的复制和恢复
func cloneRequest(r *http.Request) *http.Request {
	if r == nil {
		return nil
	}
	r2 := r.Clone(r.Context())
	if r.Body != nil && r.Body != http.NoBody {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logging.Logger.Infof("cloneRequest: Error reading body: %v\n", err)
			r.Body = io.NopCloser(bytes.NewBuffer(nil))
			r2.Body = io.NopCloser(bytes.NewBuffer(nil))
			return r2
		}
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		r2.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	return r2
}

// cloneRequestWithBody 使用预读的body数据创建请求的副本，避免重复读取
func cloneRequestWithBody(r *http.Request, bodyBytes []byte) *http.Request {
	if r == nil {
		return nil
	}
	r2 := r.Clone(r.Context())

	// 使用提供的bodyBytes而不是重新读取
	if bodyBytes != nil {
		r2.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else if r.Body != nil && r.Body != http.NoBody {
		// 如果没有提供bodyBytes，回退到原始的读取方式
		newBodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logging.Logger.Infof("cloneRequestWithBody: Error reading body: %v\n", err)
			r.Body = io.NopCloser(bytes.NewBuffer(nil))
			r2.Body = io.NopCloser(bytes.NewBuffer(nil))
			return r2
		}
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
		r2.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
	}
	return r2
}

// cloneResponseWithBody 创建响应的副本，并使用预读的 body
func cloneResponseWithBody(r *http.Response, bodyBytes []byte) *http.Response {
	if r == nil {
		return nil
	}
	r2 := *r // 浅拷贝结构体本身

	// 深拷贝 Header
	r2.Header = make(http.Header)
	for k, v := range r.Header {
		r2.Header[k] = append([]string(nil), v...)
	}

	// 处理 Body
	if bodyBytes != nil {
		// 如果提供了 bodyBytes (例如，在拦截时预读了)，则使用它
		r2.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		// 如果 bodyBytes 为 nil，则需要从原始响应 r.Body 读取
		if r.Body != nil && r.Body != http.NoBody {
			originalBodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logging.Logger.Infof("cloneResponseWithBody: Error reading original body: %v\n", err)
				// 出错则两个body都设为nil或空的NopCloser
				r.Body = http.NoBody  // 或 io.NopCloser(bytes.NewBuffer(nil))
				r2.Body = http.NoBody // 或 io.NopCloser(bytes.NewBuffer(nil))
			} else {
				r.Body.Close() // 关闭原始的reader
				// 恢复原始响应的 Body，使其可以被再次读取
				r.Body = io.NopCloser(bytes.NewBuffer(originalBodyBytes))
				// 设置克隆响应的 Body
				r2.Body = io.NopCloser(bytes.NewBuffer(originalBodyBytes))
			}
		} else {
			// 原始 Body 就是 nil 或 NoBody
			r2.Body = r.Body
		}
	}

	if r.Request != nil {
		r2.Request = cloneRequest(r.Request) // 假设 cloneRequest 也正确处理 Body
	}

	// 深拷贝 Trailer
	if r.Trailer != nil {
		r2.Trailer = make(http.Header)
		for k, v := range r.Trailer {
			r2.Trailer[k] = append([]string(nil), v...)
		}
	}
	return &r2
}

// 移除重复的函数，已迁移到 intercept.go

// ExportQueryHistoryByDSL 导出 dsl.go 中的 QueryHistoryByDSL 函数
// 此函数是为了在 app.go 中可以调用 mitmproxy 包中的 DSL 查询功能
func ExportQueryHistoryByDSL(dslQuery string) ([]HTTPHistory, error) {
	// 直接调用 dsl.go 中实现的 QueryHistoryByDSL 函数
	return QueryHistoryByDSL(dslQuery)
}

// startCacheCleanup 启动缓存清理器，定期清理超时的缓存条目
func startCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		cleanupExpiredCacheEntries()
	}
}

// cleanupExpiredCacheEntries 清理超时的缓存条目
func cleanupExpiredCacheEntries() {
	cutoffTime := time.Now().Add(-10 * time.Minute) // 清理10分钟以前的条目

	// 清理TempHistoryCache中的超时条目
	TempHistoryCache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*HTTPHistory); ok {
			if timestamp, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
				if timestamp.Before(cutoffTime) {
					logging.Logger.Warnf("Cleaning up expired temp history cache entry: %s", key)
					TempHistoryCache.Delete(key)
				}
			}
		}
		return true
	})

	// 清理TempRequestRawCache中的超时条目
	// 由于这个缓存没有时间戳信息，我们采用更简单的策略：
	// 如果对应的flowID在TempHistoryCache中已经不存在，则删除
	TempRequestRawCache.Range(func(key, value interface{}) bool {
		if _, exists := TempHistoryCache.Load(key); !exists {
			logging.Logger.Debugf("Cleaning up orphaned temp request raw cache entry: %s", key)
			TempRequestRawCache.Delete(key)
		}
		return true
	})
}

// ===== 字段提取辅助函数 =====

// extractMIMEType 从Content-Type中提取主要的MIME类型
func extractMIMEType(contentType string) string {
	if contentType == "" {
		return ""
	}

	// 移除参数部分，只保留主要类型
	parts := strings.Split(contentType, ";")
	if len(parts) > 0 {
		mainType := strings.TrimSpace(parts[0])

		// 进一步简化为主要类别
		switch {
		case strings.HasPrefix(mainType, "text/"):
			return "text"
		case strings.HasPrefix(mainType, "image/"):
			return "image"
		case strings.HasPrefix(mainType, "application/json"):
			return "json"
		case strings.HasPrefix(mainType, "application/xml") || strings.HasPrefix(mainType, "text/xml"):
			return "xml"
		case strings.HasPrefix(mainType, "application/javascript") || strings.HasPrefix(mainType, "text/javascript"):
			return "javascript"
		case strings.HasPrefix(mainType, "text/css"):
			return "css"
		case strings.HasPrefix(mainType, "text/html"):
			return "html"
		case strings.HasPrefix(mainType, "application/"):
			return "application"
		case strings.HasPrefix(mainType, "video/"):
			return "video"
		case strings.HasPrefix(mainType, "audio/"):
			return "audio"
		default:
			return mainType
		}
	}
	return ""
}

// extractExtension 从URL路径中提取文件扩展名
func extractExtension(path string) string {
	if path == "" {
		return ""
	}

	// 移除查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// 移除锚点
	if idx := strings.Index(path, "#"); idx != -1 {
		path = path[:idx]
	}

	// 获取最后一个路径段
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	filename := parts[len(parts)-1]
	if filename == "" {
		return ""
	}

	// 提取扩展名
	if idx := strings.LastIndex(filename, "."); idx != -1 && idx < len(filename)-1 {
		ext := strings.ToLower(filename[idx+1:])
		// 只返回常见的扩展名，避免过长的字符串
		if len(ext) <= 10 {
			return ext
		}
	}

	return ""
}

// extractTitle 从HTML响应体中提取页面标题
func extractTitle(responseBody, contentType string) string {
	// 只处理HTML内容
	if !strings.Contains(strings.ToLower(contentType), "html") {
		return ""
	}

	if responseBody == "" {
		return ""
	}

	// 查找title标签的开始和结束位置
	startTag := strings.Index(strings.ToLower(responseBody), "<title")
	if startTag == -1 {
		return ""
	}

	// 找到title标签的结束>
	startContent := strings.Index(responseBody[startTag:], ">")
	if startContent == -1 {
		return ""
	}
	startContent += startTag + 1

	// 找到</title>标签
	endTag := strings.Index(strings.ToLower(responseBody[startContent:]), "</title>")
	if endTag == -1 {
		return ""
	}

	title := strings.TrimSpace(responseBody[startContent : startContent+endTag])

	// 清理HTML实体和多余的空白字符
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	// 压缩多个空格为单个空格
	for strings.Contains(title, "  ") {
		title = strings.ReplaceAll(title, "  ", " ")
	}

	// 限制标题长度，避免过长
	if len(title) > 200 {
		title = title[:200] + "..."
	}

	return title
}

// resolveHostIP 解析主机名对应的IP地址
func resolveHostIP(host string) string {
	if host == "" {
		return ""
	}

	// 移除端口号
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		// 检查是否是IPv6地址
		if !strings.Contains(host[:idx], ":") {
			host = host[:idx]
		}
	}

	// 如果已经是IP地址，直接返回
	if net.ParseIP(host) != nil {
		return host
	}

	// 尝试解析域名
	ips, err := net.LookupIP(host)
	if err != nil {
		logging.Logger.Debugf("Failed to resolve IP for host %s: %v", host, err)
		return ""
	}

	// 优先返回IPv4地址
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String()
		}
	}

	// 如果没有IPv4，返回第一个IPv6地址
	if len(ips) > 0 {
		return ips[0].String()
	}

	return ""
}

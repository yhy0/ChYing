package httpx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imroc/req/v3"

	"github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/sensitive"
	"github.com/yhy0/logging"
	"go.uber.org/ratelimit"
)

/**
   @author yhy
   @since 2023/11/22
   @desc //TODO
**/

var (
	HTTPBodyMap           sync.Map
	RequestScanMsgChannel = make(chan RequestScanMsg)

	HistoryItemIDGenerator atomic.Int64
)

func NewClient(o *Options) *Client {
	if o == nil {
		o = &Options{
			Timeout:         conf.GlobalConfig.Http.Timeout,
			VerifySSL:       conf.GlobalConfig.Http.VerifySSL,
			RetryTimes:      conf.GlobalConfig.Http.RetryTimes,
			AllowRedirect:   conf.GlobalConfig.Http.AllowRedirect,
			Proxy:           conf.GlobalConfig.Http.Proxy,
			QPS:             conf.GlobalConfig.Http.MaxQps,
			MaxConnsPerHost: conf.GlobalConfig.Http.MaxConnsPerHost,
			Headers:         conf.GlobalConfig.Http.Headers,
		}
	}

	client := &Client{}
	/*
	   Req 同时支持 HTTP/1.1，HTTP/2 和 HTTP/3，如果服务端支持，默认情况下首选 HTTP/2，其次 HTTP/1.1，这是由 TLS 握手协商的。
	   如果启用了 HTTP3 (EnableHTTP3)，当探测到服务端支持 HTTP3，会使用 HTTP3 协议进行请求。
	*/
	c := req.C().
		SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36").
		// SetCommonContentType("application/x-www-form-urlencoded; charset=utf-8").
		SetTimeout(time.Duration(o.Timeout) * time.Second)

	// https://github.com/imroc/req/issues/272
	if conf.GlobalConfig.Http.ForceHTTP1 {
		c.EnableForceHTTP1()
	} else {
		c.ImpersonateChrome() // 模拟Chrome浏览器, 不能和 EnableForceXXXX() 同时使用
	}

	c.SetMaxConnsPerHost(o.MaxConnsPerHost)
	c.SetMaxIdleConns(o.MaxConnsPerHost)

	// Add proxy
	if o.Proxy != "" {
		if !strings.HasPrefix(o.Proxy, "http") {
			o.Proxy = "http://" + o.Proxy
		}
		logging.Logger.Infoln("use proxy:", o.Proxy)
		proxyURL, _ := url.Parse(o.Proxy)
		if isSupportedProtocol(proxyURL.Scheme) {
			c.SetProxy(http.ProxyURL(proxyURL))
		} else {
			logging.Logger.Warnln("Unsupported proxy protocol: %s", proxyURL.Scheme)
		}
	}

	if !o.VerifySSL {
		c.EnableInsecureSkipVerify()
	}

	if o.RetryTimes > 0 {
		c.SetCommonRetryCount(o.RetryTimes).
			SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second)
	}

	if o.QPS == 0 {
		o.QPS = conf.GlobalConfig.Http.MaxQps
	}
	if o.QPS == 0 {
		o.QPS = 100
	}
	// Initiate rate limit instance
	client.RateLimiter = ratelimit.New(o.QPS)

	client.Client = c
	client.Options = o
	return client
}

func (c *Client) Basic(target string, method string, body string, header map[string]string, username, password string, moduleName string) (*Response, error) {
	c.Client.SetCommonBasicAuth(username, password)
	return c.Request(target, method, body, header, moduleName)
}

// recordRequestScanMsg 记录请求扫描消息，发送到通道供前端监控
func recordRequestScanMsg(httpMsg RequestScanMsg) {
	// 非阻塞发送，防止通道满时阻塞请求处理
	RequestScanMsgChannel <- httpMsg
	logging.Logger.Debugf("RequestScanMsg recorded: %s [%s] from module: %s", httpMsg.Target, httpMsg.Method, httpMsg.ModuleName)
}

func (c *Client) Request(target string, method string, body string, header map[string]string, moduleName string) (*Response, error) {
	method = strings.ToUpper(method)

	// 生成唯一的请求ID
	requestID := HistoryItemIDGenerator.Add(1)
	startTime := time.Now()

	// https://req.cool/docs/tutorial/debugging/
	var requestDumpBuf, responseDumpBuf bytes.Buffer

	// Enable dump with fully customized settings at client level.
	opt := &req.DumpOptions{
		RequestOutput:  &requestDumpBuf,
		ResponseOutput: &responseDumpBuf,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
		Async:          false,
	}

	// 重定向
	if c.Options.AllowRedirect == 0 {
		c.Client.SetRedirectPolicy(req.NoRedirectPolicy())
	} else {
		c.Client.SetRedirectPolicy(
			// Only allow up to 5 redirects
			req.MaxRedirectPolicy(c.Options.AllowRedirect),
			// Only allow redirect to same domain.
			// e.g. redirect "www.imroc.cc" to "imroc.cc" is allowed, but "google.com" is not
			req.SameDomainRedirectPolicy(),
		)
	}
	// 防止出现一些错误，这次重定向后，修改回去
	c.Options.AllowRedirect = 0

	request := c.Client.R().SetDumpOptions(opt).EnableDump().EnableTrace() // 启用 trace，获取响应的时间

	if c.Options.Headers != nil {
		if strings.Contains(c.Options.Headers["Accept-Encoding"], "gzip, deflate") {
			delete(c.Options.Headers, "Accept-Encoding")
		}
		request.SetHeaders(c.Options.Headers)
	}
	if header != nil {
		// https://github.com/imroc/req/issues/178#issuecomment-1282086128
		if strings.Contains(header["Accept-Encoding"], "gzip, deflate") {
			delete(header, "Accept-Encoding")
		}
		request.SetHeaders(header)
	}

	c.RateLimiter.Take()
	var resp *req.Response
	var err error
	if method == "GET" {
		resp, err = request.Get(target)
	} else if method == "HEAD" {
		resp, err = request.Head(target)
	} else if method == "OPTIONS" {
		resp, err = request.Options(target)
	} else if method == "POST" {
		resp, err = request.
			SetBody(body).
			Post(target)
	} else if method == "PUT" {
		resp, err = request.
			SetBody(body).
			Put(target)
	} else {
		err = errors.New(fmt.Sprintf("Unsupported method: %s", method))
	}

	// 记录请求失败的情况
	if err != nil {
		// 创建失败请求的HttpMsg记录
		failedHttpMsg := RequestScanMsg{
			Id:          requestID,
			ModuleName:  moduleName,
			Target:      target,
			Path:        extractPathFromTarget(target),
			Method:      method,
			Status:      0, // 0 表示请求失败
			Length:      0,
			Title:       "",
			IP:          extractIPFromTarget(target),
			ContentType: "",
			Timestamp:   time.Now().Format(time.RFC3339Nano),
		}

		// 记录失败的请求信息到HttpBody
		httpBody := &HttpBody{
			Id:          requestID,
			RequestRaw:  requestDumpBuf.String(),
			ResponseRaw: fmt.Sprintf("Request failed: %v", err),
		}
		HTTPBodyMap.Store(requestID, httpBody)

		// 发送失败消息到通道
		recordRequestScanMsg(failedHttpMsg)

		logging.Logger.Errorf("Request failed for module %s: %v", moduleName, err)
		return nil, err
	}

	var (
		location string
		respBody string
	)

	if respLocation, err := resp.Location(); err == nil {
		location = respLocation.String()
	}

	var filterHeader bool
	contentType := resp.GetHeader("Content-Type")
	// 过滤掉这种 strings.Contains(contentType, "application/octet-stream") 这种不能过滤 ，有的返回就是这个，比如 /.git/config 相关的
	if strings.Contains(contentType, "image/") || strings.Contains(contentType, "video/") || strings.Contains(contentType, "audio/") {
		filterHeader = true
	}

	if !filterHeader {
		if respBodyByte, err := io.ReadAll(resp.Body); err == nil {
			respBody = string(respBodyByte)
		}
		defer resp.Body.Close()
		// 检测所有的返回包，可能有某个插件导致报错，存在报错信息
		if !strings.Contains(contentType, "application/zip") {
			sensitive.PageErrorMessageCheck(target, requestDumpBuf.String(), respBody)
		}
	}

	if resp.StatusCode == 200 {
		// 检查一下是否为 js 控制的跳转
		if checkJSRedirect(respBody) {
			resp.StatusCode = 302
		}
	}

	// 提取页面标题
	title := ""
	if strings.Contains(contentType, "text/html") {
		title = extractTitleFromHTML(respBody)
	}

	// 创建HttpMsg记录，用于前端监控
	httpMsg := RequestScanMsg{
		Id:          requestID,
		ModuleName:  moduleName,
		Target:      target,
		Path:        extractPathFromTarget(target),
		Method:      method,
		Status:      resp.StatusCode,
		Length:      int(resp.ContentLength),
		Title:       title,
		IP:          extractIPFromTarget(target),
		ContentType: contentType,
		Timestamp:   time.Now().Format(time.RFC3339Nano),
	}

	// 创建HttpBody记录，存储完整的请求响应数据
	httpBody := &HttpBody{
		Id:          requestID,
		RequestRaw:  requestDumpBuf.String(),
		ResponseRaw: responseDumpBuf.String(),
	}

	// 存储到HTTPBodyMap以供后续查询
	HTTPBodyMap.Store(requestID, httpBody)

	// 记录请求扫描消息
	recordRequestScanMsg(httpMsg)

	// 记录请求处理时间
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	logging.Logger.Debugf("Request completed for module %s: %s [%s] - Status: %d, Duration: %v",
		moduleName, target, method, resp.StatusCode, duration)

	return &Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Body:             respBody,
		RequestDump:      requestDumpBuf.String(),
		ResponseDump:     responseDumpBuf.String(),
		RespHeader:       resp.Header,
		ContentLength:    int(resp.ContentLength),
		RequestUrl:       resp.Request.URL.String(),
		Location:         location,
		ServerDurationMs: float64(request.TraceInfo().FirstResponseTime.Milliseconds()),
	}, nil
}

func (c *Client) Upload(target string, params map[string]string, name, fileName string) (*Response, error) {
	// 生成唯一的请求ID
	requestID := HistoryItemIDGenerator.Add(1)
	moduleName := "upload" // Upload方法的默认模块名

	// https://req.cool/docs/tutorial/debugging/
	var requestDumpBuf, responseDumpBuf bytes.Buffer
	// Enable dump with fully customized settings at client level.
	opt := &req.DumpOptions{
		RequestOutput:  &requestDumpBuf,
		ResponseOutput: &responseDumpBuf,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
		Async:          false,
	}

	request := c.Client.R().SetDumpOptions(opt).EnableDump().
		SetHeaders(c.Options.Headers).
		EnableTrace() // 启用 trace，获取响应的时间

	var resp *req.Response
	var err error

	request.
		SetFileBytes(name, fileName, []byte("test")). // 文件名，文件内容
		SetFormData(params)                           // 写入body中额外参数

	if c.Options.Headers != nil {
		if c.Options.Headers["Accept-Encoding"] == "gzip, deflate" {
			delete(c.Options.Headers, "Accept-Encoding")
		}
		request.SetHeaders(c.Options.Headers)
	}

	resp, err = request.Post(target)

	// 记录上传请求失败的情况
	if err != nil {
		failedHttpMsg := RequestScanMsg{
			Id:          requestID,
			ModuleName:  moduleName,
			Target:      target,
			Path:        extractPathFromTarget(target),
			Method:      "POST",
			Status:      0,
			Length:      0,
			Title:       "",
			IP:          extractIPFromTarget(target),
			ContentType: "",
			Timestamp:   time.Now().Format(time.RFC3339Nano),
		}

		httpBody := &HttpBody{
			Id:          requestID,
			RequestRaw:  requestDumpBuf.String(),
			ResponseRaw: fmt.Sprintf("Upload failed: %v", err),
		}
		HTTPBodyMap.Store(requestID, httpBody)
		recordRequestScanMsg(failedHttpMsg)

		return nil, err
	}

	var (
		location string
		respBody string
	)

	if respLocation, err := resp.Location(); err == nil {
		location = respLocation.String()
	}

	if respBodyByte, err := io.ReadAll(resp.Body); err == nil {
		respBody = string(respBodyByte)
	}
	defer resp.Body.Close()

	c.RateLimiter.Take()

	// 提取页面标题
	contentType := resp.GetHeader("Content-Type")
	title := ""
	if strings.Contains(contentType, "text/html") {
		title = extractTitleFromHTML(respBody)
	}

	// 记录上传请求的HttpMsg
	httpMsg := RequestScanMsg{
		Id:          requestID,
		ModuleName:  moduleName,
		Target:      target,
		Path:        extractPathFromTarget(target),
		Method:      "POST",
		Status:      resp.StatusCode,
		Length:      int(resp.ContentLength),
		Title:       title,
		IP:          extractIPFromTarget(target),
		ContentType: contentType,
		Timestamp:   time.Now().Format(time.RFC3339Nano),
	}

	// 存储HttpBody
	httpBody := &HttpBody{
		Id:          requestID,
		RequestRaw:  requestDumpBuf.String(),
		ResponseRaw: responseDumpBuf.String(),
	}
	HTTPBodyMap.Store(requestID, httpBody)
	recordRequestScanMsg(httpMsg)

	return &Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Body:             respBody,
		RequestDump:      requestDumpBuf.String(),
		ResponseDump:     responseDumpBuf.String(),
		RespHeader:       resp.Header,
		ContentLength:    int(resp.ContentLength),
		RequestUrl:       resp.Request.URL.String(),
		Location:         location,
		ServerDurationMs: float64(request.TraceInfo().FirstResponseTime.Milliseconds()),
	}, nil
}

// Request10 发送 http/1.0
func Request10(host, raw string) (*Response, error) {
	defer func() {
		if err := recover(); err != nil {
			logging.Logger.Errorln("Request10 err:", err)
			debugStack := make([]byte, 1024)
			runtime.Stack(debugStack, false)
			logging.Logger.Errorf("Request10 Stack Trace:%v", string(debugStack))
		}
	}()

	// 生成唯一的请求ID
	requestID := HistoryItemIDGenerator.Add(1)
	moduleName := "http10" // Request10的默认模块名

	conn, err := net.Dial("tcp", host)
	if err != nil {
		logging.Logger.Errorln("Error connecting:", err)

		// 记录连接失败的情况
		failedHttpMsg := RequestScanMsg{
			Id:          requestID,
			ModuleName:  moduleName,
			Target:      host,
			Path:        "/",
			Method:      "UNKNOWN",
			Status:      0,
			Length:      0,
			Title:       "",
			IP:          extractIPFromTarget(host),
			ContentType: "",
			Timestamp:   time.Now().Format(time.RFC3339Nano),
		}

		httpBody := &HttpBody{
			Id:          requestID,
			RequestRaw:  raw,
			ResponseRaw: fmt.Sprintf("Connection failed: %v", err),
		}
		HTTPBodyMap.Store(requestID, httpBody)
		recordRequestScanMsg(failedHttpMsg)

		return nil, err
	}
	defer conn.Close()

	// 发送请求
	_, err = fmt.Fprint(conn, raw)
	if err != nil {
		logging.Logger.Errorln("Error sending request:", err)

		// 记录发送失败的情况
		failedHttpMsg := RequestScanMsg{
			Id:          requestID,
			ModuleName:  moduleName,
			Target:      host,
			Path:        "/",
			Method:      "UNKNOWN",
			Status:      0,
			Length:      0,
			Title:       "",
			IP:          extractIPFromTarget(host),
			ContentType: "",
			Timestamp:   time.Now().Format(time.RFC3339Nano),
		}

		httpBody := &HttpBody{
			Id:          requestID,
			RequestRaw:  raw,
			ResponseRaw: fmt.Sprintf("Send request failed: %v", err),
		}
		HTTPBodyMap.Store(requestID, httpBody)
		recordRequestScanMsg(failedHttpMsg)

		return nil, err
	}

	// 读取响应
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		logging.Logger.Errorln("Error reading response:", err)

		// 记录读取响应失败的情况
		failedHttpMsg := RequestScanMsg{
			Id:          requestID,
			ModuleName:  moduleName,
			Target:      host,
			Path:        "/",
			Method:      "UNKNOWN",
			Status:      0,
			Length:      0,
			Title:       "",
			IP:          extractIPFromTarget(host),
			ContentType: "",
			Timestamp:   time.Now().Format(time.RFC3339Nano),
		}

		httpBody := &HttpBody{
			Id:          requestID,
			RequestRaw:  raw,
			ResponseRaw: fmt.Sprintf("Read response failed: %v", err),
		}
		HTTPBodyMap.Store(requestID, httpBody)
		recordRequestScanMsg(failedHttpMsg)

		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	responseDump, _ := httputil.DumpResponse(resp, true)

	// 提取方法从原始请求中
	method := "UNKNOWN"
	if lines := strings.Split(raw, "\n"); len(lines) > 0 {
		if parts := strings.Split(lines[0], " "); len(parts) > 0 {
			method = strings.TrimSpace(parts[0])
		}
	}

	// 提取页面标题
	contentType := resp.Header.Get("Content-Type")
	title := ""
	if strings.Contains(contentType, "text/html") {
		if respBodyBytes, err := io.ReadAll(resp.Body); err == nil {
			title = extractTitleFromHTML(string(respBodyBytes))
		}
	}

	// 记录HTTP/1.0请求的HttpMsg
	httpMsg := RequestScanMsg{
		Id:          requestID,
		ModuleName:  moduleName,
		Target:      host,
		Path:        "/",
		Method:      method,
		Status:      resp.StatusCode,
		Length:      int(resp.ContentLength),
		Title:       title,
		IP:          extractIPFromTarget(host),
		ContentType: contentType,
		Timestamp:   time.Now().Format(time.RFC3339Nano),
	}

	// 存储HttpBody
	httpBody := &HttpBody{
		Id:          requestID,
		RequestRaw:  raw,
		ResponseRaw: string(responseDump),
	}
	HTTPBodyMap.Store(requestID, httpBody)
	recordRequestScanMsg(httpMsg)

	return &Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		RequestDump:   raw,
		ResponseDump:  string(responseDump),
		RespHeader:    resp.Header,
		ContentLength: int(resp.ContentLength),
		RequestUrl:    resp.Request.URL.String(),
	}, nil
}

func Request(target string, method string, body string, header map[string]string, moduleName string) (*Response, error) {
	return NewClient(nil).Request(target, method, body, header, moduleName)
}

func Get(target string, moduleName string) (*Response, error) {
	return NewClient(nil).Request(target, "GET", "", nil, moduleName)
}

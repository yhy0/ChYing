package mitmproxy

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"io"
	"net/http/httputil"

	"github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2025/7/10
  @desc: 被动扫描插件 - 基于 proxify 的插件系统实现
**/

// InitPassiveScanPlugin 初始化被动扫描插件
func InitPassiveScanPlugin() {
	logging.Logger.Infof("初始化被动扫描插件...")

	// 注册响应处理器（使用 ReadOnly 模式，确保不修改原始响应）
	RegisterResponseProcessorWithMode(passiveScanResponseProcessor, ReadOnly)

}

// passiveScanResponseProcessor 是响应处理器函数，实现原来 passiveAddon.go 中的 Response 方法功能
func passiveScanResponseProcessor(resp *http.Response) bool {
	if passiveTask == nil {
		logging.Logger.Errorln("被动扫描任务对象未初始化，跳过处理")
		return false
	}

	// 获取请求对象
	req := resp.Request
	if req == nil {
		logging.Logger.Debugln("响应没有关联请求，跳过处理")
		return false
	}

	// 跳过 CONNECT 请求
	if req.Method == "CONNECT" {
		return false
	}

	// 为异步处理深度克隆响应对象
	// 使用 cloneResponseWithBody(originalResponse, nil) 来确保body也被正确克隆和恢复
	respClone := cloneResponseWithBody(resp, nil)
	if respClone == nil { // 如果克隆失败，记录错误并跳过
		logging.Logger.Errorln("被动扫描：克隆响应对象失败，跳过处理")
		return false
	}
	// 确保克隆出来的请求也是有效的
	clonedReq := respClone.Request
	if clonedReq == nil {
		logging.Logger.Errorln("被动扫描：克隆响应中的请求对象为nil，跳过处理")
		return false
	}

	go func(asyncResp *http.Response) {
		// 在 goroutine 内部使用克隆的 asyncResp 和 asyncReq
		asyncReq := asyncResp.Request

		// 检查主机是否在过滤列表中 (使用克隆的请求URL)
		if Filter(asyncReq.URL.Host) {
			logging.Logger.Debugln("被动扫描(async): 过滤了", asyncReq.URL.Host)
			return
		}

		// 过滤一些干扰项 - 复用原来的逻辑 (使用克隆的请求URL)
		if len(conf.GlobalConfig.Mitmproxy.Exclude) > 0 || !(len(conf.GlobalConfig.Mitmproxy.Exclude) == 1 && conf.GlobalConfig.Mitmproxy.Exclude[0] == "") {
			if !util.RegexpStr(conf.GlobalConfig.Mitmproxy.Exclude, asyncReq.URL.Host) {
				judgeAndDistribute(asyncReq, asyncResp) // 传递克隆的对象
			}
		} else {
			judgeAndDistribute(asyncReq, asyncResp) // 传递克隆的对象
		}
	}(respClone) // 将克隆的响应传递给 goroutine

	// 返回 false 表示没有同步修改响应 (实际工作在goroutine中)
	return false
}

// judgeAndDistribute 判断是否应该处理请求，并分发任务
func judgeAndDistribute(req *http.Request, resp *http.Response) {
	if len(conf.GlobalConfig.Mitmproxy.Include) > 0 && !(len(conf.GlobalConfig.Mitmproxy.Include) == 1 && conf.GlobalConfig.Mitmproxy.Include[0] == "") {
		if util.RegexpStr(conf.GlobalConfig.Mitmproxy.Include, req.URL.Host) {
			distributePassiveScanTask(req, resp)
		}
	} else {
		distributePassiveScanTask(req, resp)
	}
}

// distributePassiveScanTask 分发被动扫描任务 - 实现原来 task.go 中的 distribution 功能
func distributePassiveScanTask(req *http.Request, resp *http.Response) {
	parseUrl, err := url.Parse(req.URL.String())
	if err != nil {
		logging.Logger.Errorln("解析URL错误:", err)
		return
	}

	var host string
	// 有的会带80、443端口号，导致 example.com 和 example.com:80、example.com:443被认为是不同的网站
	port := strings.Split(parseUrl.Host, ":")
	if len(port) > 1 && (port[1] == "443" || port[1] == "80") {
		host = strings.Split(parseUrl.Host, ":")[0]
	} else {
		host = parseUrl.Host
	}

	// 读取解码后的响应体进行分析
	decodedRespBodyBytes, err := GetDecodedResponseBody(resp)
	if err != nil {
		logging.Logger.Errorln("被动扫描：读取或解码响应体失败:", err)
		// 如果无法获取解码后的响应体，可以根据策略选择跳过或使用空响应
		decodedRespBodyBytes = []byte{}
	}

	// 将 http.Header 转换为 map[string]string
	headerMap := make(map[string]string)
	for key, values := range req.Header {
		// 根据HTTP头名称选择分隔符
		separator := ","
		if key == "Set-Cookie" {
			separator = ";"
		}

		// 将多个值连接成一个字符串
		headerMap[key] = strings.Join(values, separator)
	}

	// 创建 CrawlResult 结构
	in := &input.CrawlResult{
		Target:      req.URL.Host,
		Url:         req.URL.String(),
		Host:        host,
		ParseUrl:    parseUrl,
		UniqueId:    util.GenerateUniqueID(req), // 为请求生成唯一ID
		Method:      req.Method,
		RequestBody: getRequestBody(req),
		Headers:     headerMap,
		Resp: &httpx.Response{
			Status:     strconv.Itoa(resp.StatusCode),
			StatusCode: resp.StatusCode,
			Body:       string(decodedRespBodyBytes),
			RespHeader: resp.Header,
		},
		RawRequest:  dumpRequest(req),
		RawResponse: dumpDecodedResponseForPassiveScan(resp, decodedRespBodyBytes),
	}

	// 处理 Content-Type
	if in.Headers == nil {
		in.Headers = make(map[string]string)
	} else {
		if value, ok := in.Headers["Content-Type"]; ok {
			in.ContentType = value
		} else if value, ok = in.Headers["content-type"]; ok {
			in.ContentType = value
		}
	}

	logging.Logger.Debugln("Distribution", in.Url)

	// 将任务添加到任务池处理
	passiveTask.WG.Add(1)
	err = passiveTask.Pool.Submit(passiveTask.Distribution(in))
	if err != nil {
		passiveTask.WG.Done()
		logging.Logger.Errorf("add distribution err:%v, crawlResult:%v", err, in)
	}
}

// readResponseBody 读取HTTP响应体，但不关闭原始Body
func readResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return []byte{}, nil
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 重置响应体，使其可以被下游代码再次读取
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}

// getRequestBody 获取请求体
func getRequestBody(req *http.Request) string {
	if req == nil || req.Body == nil {
		return ""
	}

	// 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logging.Logger.Errorf("读取请求体失败: %v", err)
		return ""
	}

	// 重置请求体，使其可以被下游代码再次读取
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return string(body)
}

// dumpRequest 将 http.Request 转储为原始字符串
func dumpRequest(req *http.Request) string {
	if req == nil {
		return ""
	}

	// 使用 httputil.DumpRequest 获取完整请求
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logging.Logger.Errorf("转储请求失败: %v", err)
		return ""
	}

	return string(dump)
}

// dumpResponse 将 http.Response 转储为原始字符串
func dumpResponse(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	// 使用 httputil.DumpResponse 获取完整响应
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		logging.Logger.Errorf("转储响应失败: %v", err)
		return ""
	}

	return string(dump)
}

// dumpDecodedResponseForPassiveScan 为被动扫描生成包含解码后body的响应dump
func dumpDecodedResponseForPassiveScan(originalResp *http.Response, decodedBody []byte) string {
	if originalResp == nil {
		return ""
	}

	// 创建一个临时响应副本，用于dump，确保使用解码后的body
	respForDump := *originalResp
	respForDump.Body = io.NopCloser(bytes.NewBuffer(decodedBody))
	respForDump.ContentLength = int64(len(decodedBody))
	// 复制头部，并移除可能引起误解的Content-Encoding
	headerCopy := make(http.Header)
	for k, v := range originalResp.Header {
		headerCopy[k] = append([]string(nil), v...)
	}
	headerCopy.Del("Content-Encoding")
	headerCopy.Set("Content-Length", strconv.FormatInt(respForDump.ContentLength, 10))
	respForDump.Header = headerCopy

	dump, err := httputil.DumpResponse(&respForDump, true)
	if err != nil {
		logging.Logger.Errorf("被动扫描：转储解码后的响应失败: %v", err)
		return ""
	}
	return string(dump)
}

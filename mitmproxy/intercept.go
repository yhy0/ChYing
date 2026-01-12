package mitmproxy

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/projectdiscovery/martian/v3"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2025/1/13
   @desc 拦截相关逻辑处理
**/

// InterceptData 拦截数据结构
type InterceptData struct {
	ID           string
	ModifiedBody string
	Action       string // "forward", "drop"
}

// 拦截状态管理
var (
	globalInterceptRequest  atomic.Bool
	globalInterceptResponse atomic.Bool

	// 监控指标
	interceptRequestCount  atomic.Int64 // 拦截的请求总数
	interceptResponseCount atomic.Int64 // 拦截的响应总数
	forwardedCount         atomic.Int64 // 放行的总数
	droppedCount           atomic.Int64 // 丢弃的总数

	// 请求放行状态跟踪
	requestForwardedMap sync.Map // map[string]string - 跟踪哪些请求已经被用户在请求阶段放行，存储请求的原始数据

	// 等待中的拦截请求管理 - 用于按ID分发用户操作
	pendingInterceptsMap sync.Map // map[string]chan InterceptData - 每个拦截ID对应一个channel
)

// InitInterceptSystem 初始化拦截系统
func InitInterceptSystem() {
	fmt.Println("Intercept system initialized.")
}

// SetInterceptMode 控制拦截的开关
func SetInterceptMode(isRequest bool, enable bool) {
	if isRequest {
		globalInterceptRequest.Store(enable)
	} else {
		globalInterceptResponse.Store(enable)
	}
	logging.Logger.Infof("Intercept mode changed: RequestEnabled=%v, ResponseEnabled=%v", globalInterceptRequest.Load(), globalInterceptResponse.Load())
}

// ForwardInterceptedData 从前端接收操作指令（转发/丢弃）
func ForwardInterceptedData(id string, modifiedBody string, action string) {
	logging.Logger.Infof("ForwardInterceptedData called: ID=%s, Action=%s, BodyLen=%d", id, action, len(modifiedBody))

	// 记录修改内容的前几行用于调试
	if action == "forward" && len(modifiedBody) > 0 {
		lines := strings.Split(modifiedBody, "\n")
		if len(lines) > 0 {
			logging.Logger.Infof("ForwardInterceptedData: First line of modified data: %s", lines[0])
		}
		if len(lines) > 5 {
			logging.Logger.Infof("ForwardInterceptedData: Modified data has %d lines", len(lines))
		}
	}

	interceptData := InterceptData{
		ID:           id,
		ModifiedBody: modifiedBody,
		Action:       action,
	}

	// 查找对应ID的等待channel
	if pendingChan, exists := pendingInterceptsMap.Load(id); exists {
		if ch, ok := pendingChan.(chan InterceptData); ok {
			select {
			case ch <- interceptData:
				logging.Logger.Infof("ForwardInterceptedData: Successfully sent to pending channel for ID: %s", id)
				return
			case <-time.After(2 * time.Second):
				logging.Logger.Warnf("ForwardInterceptedData: Timeout sending to pending channel for ID: %s", id)
			}
		}
	} else {
		logging.Logger.Warnf("ForwardInterceptedData: No pending intercept found for ID: %s (may have already timed out or been processed)", id)
	}
}

// HandleRequestIntercept 处理请求拦截逻辑
func HandleRequestIntercept(req *http.Request, ctx *martian.Context, flowID string) error {
	if !globalInterceptRequest.Load() {
		return nil // 未启用请求拦截
	}

	// 检查是否被过滤器过滤
	if req != nil && req.URL != nil {
		if Filter(req.URL.Host) {
			logging.Logger.Debugf("HandleRequestIntercept: Request %s filtered out, skipping intercept", flowID)
			return nil // 被过滤的请求不进行拦截
		}
		if req.Method == "CONNECT" || req.Method == "OPTIONS" {
			logging.Logger.Debugf("HandleRequestIntercept: Request %s is %s method, skipping intercept", flowID, req.Method)
			return nil // CONNECT 和 OPTIONS 方法不进行拦截
		}
	}

	// 统计拦截的请求数量
	interceptRequestCount.Add(1)
	logging.Logger.Infof("[Intercept] Intercepting Request ID: %s, URL: %s", flowID, req.URL.String())

	// 创建该请求专用的等待channel
	pendingChan := make(chan InterceptData, 1)
	pendingInterceptsMap.Store(flowID, pendingChan)
	defer func() {
		pendingInterceptsMap.Delete(flowID)
		close(pendingChan)
	}()

	// 发送拦截事件到前端
	if EventDataChan != nil {
		reqDumpBytes, dumpErr := httputil.DumpRequestOut(req, true)
		if dumpErr != nil {
			logging.Logger.Infof("HandleRequestIntercept: Error dumping request for intercept: %v", dumpErr)
		} else {
			event := &EventData{
				Name: "InterceptRequest",
				Data: map[string]interface{}{
					"id":   flowID,
					"data": string(reqDumpBytes),
					"type": "request",
				},
			}
			EventDataChan <- event
		}
	}

	// 等待用户操作，带超时机制
	var modifiedData InterceptData
	select {
	case data, ok := <-pendingChan:
		if !ok {
			logging.Logger.Warnf("HandleRequestIntercept: Pending channel closed for request %s, forwarding original", flowID)
			return nil
		}
		modifiedData = data
	case <-time.After(5 * time.Minute): // 5分钟超时，自动放行
		logging.Logger.Warnf("HandleRequestIntercept: Timeout waiting for user action on request %s, auto-forwarding", flowID)
		return nil
	}

	// 检查拦截是否仍然启用（用户可能在等待期间关闭了拦截）
	if !globalInterceptRequest.Load() {
		logging.Logger.Infof("HandleRequestIntercept: Intercept disabled during wait for %s, forwarding original", flowID)
		return nil
	}

	switch modifiedData.Action {
	case "forward":
		forwardedCount.Add(1) // 统计放行数量

		// 标记该请求已被用户在请求阶段放行，并存储修改后的请求数据
		var requestDataForStorage string
		if modifiedData.ModifiedBody != "" {
			requestDataForStorage = modifiedData.ModifiedBody
		} else {
			// 如果没有修改，使用原始请求数据
			if reqDumpBytes, dumpErr := httputil.DumpRequestOut(req, true); dumpErr == nil {
				requestDataForStorage = string(reqDumpBytes)
			}
		}
		requestForwardedMap.Store(flowID, requestDataForStorage)

		if modifiedData.ModifiedBody != "" {
			logging.Logger.Infof("HandleRequestIntercept: Attempting to apply modified request for %s", flowID)
			parsedNewReq, parseErr := parseModifiedRequest(modifiedData.ModifiedBody, req.URL)
			if parseErr != nil {
				logging.Logger.Infof("HandleRequestIntercept: Failed to parse modified request %s: %v. Forwarding original.", flowID, parseErr)
			} else {
				req.Method = parsedNewReq.Method
				req.URL = parsedNewReq.URL
				req.Header = parsedNewReq.Header
				req.Body = parsedNewReq.Body
				req.ContentLength = parsedNewReq.ContentLength
				req.Host = parsedNewReq.Host
				logging.Logger.Infof("HandleRequestIntercept: Applied MODIFIED Request for ID: %s", flowID)

				// 同步更新历史记录缓存中的请求信息
				if err := updateHistoryCacheWithModifiedRequest(flowID, parsedNewReq); err != nil {
					logging.Logger.Warnf("HandleRequestIntercept: Failed to update history cache for %s: %v", flowID, err)
				}
			}
		}
		logging.Logger.Infof("HandleRequestIntercept: Request %s forwarded by user, continuing to response", flowID)
		return nil
	case "drop":
		droppedCount.Add(1) // 统计丢弃数量

		// 清理已放行请求的标记（虽然是丢弃，但也需要清理）
		requestForwardedMap.Delete(flowID)

		logging.Logger.Infof("HandleRequestIntercept: Dropping Request ID: %s", flowID)
		return fmt.Errorf("request_dropped_by_user_%s", flowID)
	default:
		logging.Logger.Infof("HandleRequestIntercept: Unknown action for request %s: %s. Forwarding original.", flowID, modifiedData.Action)
		return nil
	}
}

// HandleResponseIntercept 处理响应拦截逻辑
func HandleResponseIntercept(resp *http.Response, ctx *martian.Context, flowID string) (bool, error) {
	if !globalInterceptResponse.Load() {
		// 清理已放行请求的标记
		requestForwardedMap.Delete(flowID)
		return false, nil // 未启用响应拦截，返回false表示未修改
	}

	// 检查是否被过滤器过滤
	req := resp.Request
	if req != nil && req.URL != nil {
		if Filter(req.URL.Host) {
			logging.Logger.Debugf("HandleResponseIntercept: Response %s filtered out, skipping intercept", flowID)
			// 清理已放行请求的标记
			requestForwardedMap.Delete(flowID)
			return false, nil // 被过滤的响应不进行拦截
		}
		if req.Method == "CONNECT" || req.Method == "OPTIONS" {
			logging.Logger.Debugf("HandleResponseIntercept: Response %s is %s method, skipping intercept", flowID, req.Method)
			// 清理已放行请求的标记
			requestForwardedMap.Delete(flowID)
			return false, nil // CONNECT 和 OPTIONS 方法的响应不进行拦截
		}
	}

	// 统计拦截的响应数量
	interceptResponseCount.Add(1)
	logging.Logger.Infof("[Intercept] Intercepting Response ID: %s, URL: %s", flowID, req.URL.String())

	// 创建该响应专用的等待channel
	pendingChan := make(chan InterceptData, 1)
	pendingInterceptsMap.Store(flowID, pendingChan)
	defer func() {
		pendingInterceptsMap.Delete(flowID)
		close(pendingChan)
		// 清理已放行请求的标记
		requestForwardedMap.Delete(flowID)
	}()

	// 获取原始请求数据 - 优先使用已放行的请求数据，回退到临时缓存
	var requestRawString string

	// 首先尝试从已放行的请求数据获取（如果请求被用户手动处理过）
	if forwardedReqData, wasForwarded := requestForwardedMap.Load(flowID); wasForwarded {
		if reqStr, okCast := forwardedReqData.(string); okCast {
			requestRawString = reqStr
			logging.Logger.Infof("HandleResponseIntercept: Using forwarded request data for %s", flowID)
		}
	}

	// 如果没有找到已放行的请求数据，从临时缓存获取
	if requestRawString == "" {
		if rawReqData, rawReqLoaded := TempRequestRawCache.Load(flowID); rawReqLoaded {
			if reqStr, okCast := rawReqData.(string); okCast {
				requestRawString = reqStr
				logging.Logger.Infof("HandleResponseIntercept: Using temp cached request data for %s", flowID)
			}
		}
	}

	// 如果仍然没有找到请求数据，生成一个基本的请求信息
	if requestRawString == "" {
		logging.Logger.Warnf("HandleResponseIntercept: No request data found for %s, generating basic request info", flowID)
		// 从响应对象中提取请求信息作为回退
		if resp.Request != nil {
			if reqDumpBytes, dumpErr := httputil.DumpRequestOut(resp.Request, true); dumpErr == nil {
				requestRawString = string(reqDumpBytes)
				logging.Logger.Infof("HandleResponseIntercept: Generated request data from response.Request for %s", flowID)
			}
		}
	}

	// 生成响应dump
	respDumpBytes, respDumpErr := httputil.DumpResponse(resp, true)
	var responseRawString string
	if respDumpErr != nil {
		logging.Logger.Infof("HandleResponseIntercept: Error dumping response for intercept: %v", respDumpErr)
		responseRawString = fmt.Sprintf("Error dumping response: %v", respDumpErr)
	} else {
		responseRawString = string(respDumpBytes)
	}

	// 检查是否有足够的请求数据用于拦截显示
	if requestRawString == "" {
		logging.Logger.Errorf("HandleResponseIntercept: No request data available for response %s, skipping intercept", flowID)
		return false, nil
	}

	// 发送拦截事件到前端（包含请求和响应）
	if EventDataChan != nil {
		// 确保响应数据是可读的（解压缩的）
		displayResponseString := responseRawString

		// 检查是否被压缩，如果是则尝试解压缩用于显示
		if resp.Header.Get("Content-Encoding") != "" {
			if processedInfo, processErr := ProcessResponseBody(resp); processErr == nil && processedInfo != nil {
				// 创建一个用于显示的响应dump（解压缩的）
				displayResponseString = dumpResponseFromProcessedInfo(resp, processedInfo)
				logging.Logger.Infof("HandleResponseIntercept: Created decompressed response dump for intercept display, length: %d", len(displayResponseString))
			}
		}

		event := &EventData{
			Name: "InterceptResponse",
			Data: map[string]interface{}{
				"id":       flowID,
				"request":  requestRawString,
				"response": displayResponseString, // 使用解压缩的响应数据
				"type":     "response",
			},
		}
		EventDataChan <- event
		logging.Logger.Infof("HandleResponseIntercept: Sent InterceptResponse event for ID: %s", flowID)
	}

	// 等待用户操作，带超时机制
	var modifiedData InterceptData
	select {
	case data, ok := <-pendingChan:
		if !ok {
			logging.Logger.Warnf("HandleResponseIntercept: Pending channel closed for response %s, forwarding original", flowID)
			return false, nil
		}
		modifiedData = data
	case <-time.After(5 * time.Minute): // 5分钟超时，自动放行
		logging.Logger.Warnf("HandleResponseIntercept: Timeout waiting for user action on response %s, auto-forwarding", flowID)
		return false, nil
	}

	// 检查拦截是否仍然启用（用户可能在等待期间关闭了拦截）
	if !globalInterceptResponse.Load() {
		logging.Logger.Infof("HandleResponseIntercept: Intercept disabled during wait for %s, forwarding original", flowID)
		return false, nil
	}

	switch modifiedData.Action {
	case "forward":
		forwardedCount.Add(1) // 统计放行数量

		if modifiedData.ModifiedBody != "" {
			logging.Logger.Infof("HandleResponseIntercept: Attempting to apply modified response for %s", flowID)
			parsedNewResp, parseErr := parseModifiedResponse(modifiedData.ModifiedBody, req)
			if parseErr != nil {
				logging.Logger.Infof("HandleResponseIntercept: Failed to parse modified response %s: %v. Forwarding original.", flowID, parseErr)
				return false, nil
			} else {
				resp.StatusCode = parsedNewResp.StatusCode
				resp.Status = parsedNewResp.Status
				resp.Header = parsedNewResp.Header
				resp.Body = parsedNewResp.Body
				resp.ContentLength = parsedNewResp.ContentLength
				logging.Logger.Infof("HandleResponseIntercept: Applied MODIFIED Response for ID: %s", flowID)
				return true, nil // 返回true表示响应被修改
			}
		}
		logging.Logger.Infof("HandleResponseIntercept: Response %s forwarded, transaction complete", flowID)
		return false, nil
	case "drop":
		droppedCount.Add(1) // 统计丢弃数量

		logging.Logger.Infof("HandleResponseIntercept: Dropping Response ID: %s", flowID)
		return false, fmt.Errorf("response_dropped_by_user_%s", flowID)
	default:
		logging.Logger.Infof("HandleResponseIntercept: Unknown action for response %s: %s. Forwarding original.", flowID, modifiedData.Action)
		return false, nil
	}
}

// parseModifiedRequest 将原始HTTP请求字符串解析为 *http.Request
func parseModifiedRequest(rawReq string, originalURL *url.URL) (*http.Request, error) {
	b := bufio.NewReader(strings.NewReader(rawReq))
	req, err := http.ReadRequest(b)
	if err != nil {
		return nil, fmt.Errorf("http.ReadRequest failed: %w", err)
	}
	if originalURL != nil {
		req.URL.Scheme = originalURL.Scheme
		req.URL.Host = originalURL.Host
	}

	if req.URL.Host == "" {
		req.URL.Host = req.Header.Get("Host")
	}
	req.Host = req.URL.Host
	req.RequestURI = ""

	if req.Body == nil {
		req.Body = http.NoBody
	}

	return req, nil
}

// parseModifiedResponse 将原始HTTP响应字符串解析为 *http.Response
func parseModifiedResponse(rawResp string, originalReq *http.Request) (*http.Response, error) {
	b := bufio.NewReader(strings.NewReader(rawResp))
	resp, err := http.ReadResponse(b, originalReq)
	if err != nil {
		return nil, fmt.Errorf("http.ReadResponse failed: %w", err)
	}
	if resp.Body == nil {
		resp.Body = http.NoBody
	}
	return resp, nil
}

// IsRequestInterceptEnabled 检查请求拦截是否启用
func IsRequestInterceptEnabled() bool {
	return globalInterceptRequest.Load()
}

// IsResponseInterceptEnabled 检查响应拦截是否启用
func IsResponseInterceptEnabled() bool {
	return globalInterceptResponse.Load()
}

// GetInterceptStats 获取拦截统计信息
func GetInterceptStats() map[string]int64 {
	return map[string]int64{
		"interceptedRequests":  interceptRequestCount.Load(),
		"interceptedResponses": interceptResponseCount.Load(),
		"forwarded":            forwardedCount.Load(),
		"dropped":              droppedCount.Load(),
	}
}

// ResetInterceptStats 重置拦截统计信息
func ResetInterceptStats() {
	interceptRequestCount.Store(0)
	interceptResponseCount.Store(0)
	forwardedCount.Store(0)
	droppedCount.Store(0)
}

// updateHistoryCacheWithModifiedRequest 更新历史记录缓存中的请求信息
func updateHistoryCacheWithModifiedRequest(flowID string, modifiedReq *http.Request) error {
	// 从临时缓存中加载历史记录条目
	cachedEntry, exists := TempHistoryCache.Load(flowID)
	if !exists {
		return fmt.Errorf("no history cache entry found for flowID: %s", flowID)
	}

	historyEntry, ok := cachedEntry.(*HTTPHistory)
	if !ok {
		return fmt.Errorf("invalid history cache entry type for flowID: %s", flowID)
	}

	// 更新历史记录条目的信息
	historyEntry.Method = modifiedReq.Method
	historyEntry.FullUrl = modifiedReq.URL.String()
	historyEntry.Host = modifiedReq.URL.Host
	historyEntry.Path = modifiedReq.URL.RequestURI()

	// 将更新后的条目存回缓存
	TempHistoryCache.Store(flowID, historyEntry)

	logging.Logger.Infof("updateHistoryCacheWithModifiedRequest: Updated history cache for %s with new URL: %s", flowID, historyEntry.FullUrl)
	return nil
}

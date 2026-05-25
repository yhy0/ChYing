package claudecode

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	claude "github.com/yhy0/claude-agent-sdk-go"
	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/10
   @desc MCP 工具定义 - 使用 claude-agent-sdk-go 的 SDK MCP Server
**/

// ==================== 工具输入类型定义 ====================

// GetHTTPHistoryInput 获取 HTTP 历史的输入参数
type GetHTTPHistoryInput struct {
	ProjectID    string `json:"project_id"`
	Host         string `json:"host,omitempty"`
	Method       string `json:"method,omitempty"`
	PathContains string `json:"path_contains,omitempty"`
	StatusCode   int    `json:"status_code,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
}

// GetTrafficDetailInput 获取流量详情的输入参数
type GetTrafficDetailInput struct {
	TrafficID int `json:"traffic_id"`
}

// GetVulnerabilitiesInput 获取漏洞列表的输入参数
type GetVulnerabilitiesInput struct {
	ProjectID string `json:"project_id"`
	Severity  string `json:"severity,omitempty"`
	VulnType  string `json:"vuln_type,omitempty"`
	Host      string `json:"host,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// SendHTTPRequestInput 发送 HTTP 请求的输入参数
type SendHTTPRequestInput struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// AnalyzeRequestInput 分析请求的输入参数
type AnalyzeRequestInput struct {
	TrafficID int `json:"traffic_id"`
}

// SearchTrafficInput 搜索流量的输入参数
type SearchTrafficInput struct {
	ProjectID string `json:"project_id"`
	Keyword   string `json:"keyword"`
	SearchIn  string `json:"search_in,omitempty"` // url, request_body, response_body, all
	Limit     int    `json:"limit,omitempty"`
}

// GetSitemapInput 获取网站地图的输入参数
type GetSitemapInput struct {
	ProjectID string `json:"project_id"`
	Host      string `json:"host,omitempty"`
}

// GetStatisticsInput 获取统计信息的输入参数
type GetStatisticsInput struct {
	ProjectID string `json:"project_id"`
}

// GetNewTrafficSinceInput 获取指定时间后的新增流量
type GetNewTrafficSinceInput struct {
	Since     string `json:"since"`                // RFC3339 格式的时间字符串
	SessionID string `json:"session_id,omitempty"` // 可选的 session 隔离
}

// ==================== 创建 MCP Server ====================

// CreateChYingMCPServer 创建 ChYing 安全工具 MCP 服务器
func CreateChYingMCPServer() *claude.MCPServer {
	server := claude.NewMCPServer(
		"chying-security-tools",
		claude.WithVersion("1.0.0"),
		claude.WithMCPTools(
			createGetHTTPHistoryTool(),
			createGetTrafficDetailTool(),
			createGetVulnerabilitiesTool(),
			createSendHTTPRequestTool(),
			createAnalyzeRequestTool(),
			createSearchTrafficTool(),
			createGetSitemapTool(),
			createGetStatisticsTool(),
			createGetNewTrafficSinceTool(),
		),
	)
	return server
}

// ==================== 工具定义 ====================

// createGetHTTPHistoryTool 创建获取 HTTP 历史工具
func createGetHTTPHistoryTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_http_history",
		"获取 HTTP 流量历史记录，可按项目、主机、方法等过滤。返回代理捕获的 HTTP 请求列表。",
		func(input GetHTTPHistoryInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_http_history called: projectID=%s", input.ProjectID)

			if input.ProjectID == "" {
				return claude.ErrorResult("project_id is required"), nil
			}

			limit := input.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 500 {
				limit = 500
			}

			histories, err := db.GetAllHistory(input.ProjectID, "", limit, input.Offset)
			if err != nil {
				logging.Logger.Errorf("Failed to get HTTP history: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get HTTP history: %v", err)), nil
			}

			result := formatTrafficList(histories)
			jsonBytes, _ := json.Marshal(result)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// createGetTrafficDetailTool 创建获取流量详情工具
func createGetTrafficDetailTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_traffic_detail",
		"获取单个 HTTP 流量的详细信息，包括完整的请求头、请求体、响应头和响应体",
		func(input GetTrafficDetailInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_traffic_detail called: trafficID=%d", input.TrafficID)

			if input.TrafficID <= 0 {
				return claude.ErrorResult("traffic_id is required and must be positive"), nil
			}

			httpBody, err := db.GetHttpData(input.TrafficID)
			if err != nil {
				logging.Logger.Errorf("Failed to get traffic detail: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get traffic detail: %v", err)), nil
			}

			if httpBody == nil {
				return claude.ErrorResult("Traffic not found"), nil
			}

			result := fmt.Sprintf("=== Request ===\n%s\n\n=== Response ===\n%s",
				httpBody.RequestRaw, httpBody.ResponseRaw)
			return claude.TextResult(result), nil
		},
	)
}

// createGetVulnerabilitiesTool 创建获取漏洞列表工具
func createGetVulnerabilitiesTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_vulnerabilities",
		"获取已发现的漏洞列表，可按项目、严重程度、类型等过滤",
		func(input GetVulnerabilitiesInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_vulnerabilities called: projectID=%s, severity=%s",
				input.ProjectID, input.Severity)

			if input.ProjectID == "" {
				return claude.ErrorResult("project_id is required"), nil
			}

			limit := input.Limit
			if limit <= 0 {
				limit = 100
			}

			vulns, err := db.GetAllVulnerabilities(input.ProjectID, "", limit, 0)
			if err != nil {
				logging.Logger.Errorf("Failed to get vulnerabilities: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get vulnerabilities: %v", err)), nil
			}

			result := formatVulnerabilities(vulns)
			jsonBytes, _ := json.Marshal(result)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// createSendHTTPRequestTool 创建发送 HTTP 请求工具（流量经过 ChYing 捕获管道）
func createSendHTTPRequestTool() *claude.Tool {
	return claude.NewTypedTool(
		"send_http_request",
		"发送自定义 HTTP 请求并返回响应。请求会被 ChYing 捕获记录，用于测试和验证漏洞。",
		func(input SendHTTPRequestInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("send_http_request called: method=%s, url=%s", input.Method, input.URL)

			if input.Method == "" {
				return claude.ErrorResult("method is required"), nil
			}
			if input.URL == "" {
				return claude.ErrorResult("url is required"), nil
			}

			resp, err := sendHTTPRequest(input.Method, input.URL, input.Headers, input.Body)
			if err != nil {
				logging.Logger.Errorf("Failed to send HTTP request: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to send HTTP request: %v", err)), nil
			}

			jsonBytes, _ := json.Marshal(resp)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// createAnalyzeRequestTool 创建分析请求工具
func createAnalyzeRequestTool() *claude.Tool {
	return claude.NewTypedTool(
		"analyze_request",
		"分析 HTTP 请求，识别潜在的安全问题：参数注入点、认证信息、敏感数据泄露等",
		func(input AnalyzeRequestInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("analyze_request called: trafficID=%d", input.TrafficID)

			if input.TrafficID <= 0 {
				return claude.ErrorResult("traffic_id is required and must be positive"), nil
			}

			analysis, err := analyzeRequest(int64(input.TrafficID))
			if err != nil {
				logging.Logger.Errorf("Failed to analyze request: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to analyze request: %v", err)), nil
			}

			jsonBytes, _ := json.Marshal(analysis)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// createSearchTrafficTool 创建搜索流量工具
func createSearchTrafficTool() *claude.Tool {
	return claude.NewTypedTool(
		"search_traffic",
		"搜索包含特定关键词的 HTTP 流量。支持搜索 URL、请求体、响应体或全部。",
		func(input SearchTrafficInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("search_traffic called: projectID=%s, keyword=%s",
				input.ProjectID, input.Keyword)

			if input.ProjectID == "" {
				return claude.ErrorResult("project_id is required"), nil
			}
			if input.Keyword == "" {
				return claude.ErrorResult("keyword is required"), nil
			}

			searchIn := input.SearchIn
			if searchIn == "" {
				searchIn = "all"
			}

			limit := input.Limit
			if limit <= 0 {
				limit = 50
			}

			results, err := searchTraffic(input.ProjectID, input.Keyword, searchIn, limit)
			if err != nil {
				logging.Logger.Errorf("Failed to search traffic: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to search traffic: %v", err)), nil
			}

			jsonBytes, _ := json.Marshal(results)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// createGetSitemapTool 创建获取网站地图工具
func createGetSitemapTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_sitemap",
		"获取发现的网站地图：所有经过代理的主机及其路径树",
		func(input GetSitemapInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_sitemap called: projectID=%s, host=%s",
				input.ProjectID, input.Host)

			if input.ProjectID == "" {
				return claude.ErrorResult("project_id is required"), nil
			}

			sitemap, err := getSitemap(input.ProjectID, input.Host)
			if err != nil {
				logging.Logger.Errorf("Failed to get sitemap: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get sitemap: %v", err)), nil
			}

			jsonBytes, _ := json.Marshal(sitemap)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// createGetStatisticsTool 创建获取统计信息工具
func createGetStatisticsTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_statistics",
		"获取项目的统计信息：流量数、漏洞数、主机数，以及按方法/状态码/漏洞等级的分布",
		func(input GetStatisticsInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_statistics called: projectID=%s", input.ProjectID)

			if input.ProjectID == "" {
				return claude.ErrorResult("project_id is required"), nil
			}

			stats, err := getStatistics(input.ProjectID)
			if err != nil {
				logging.Logger.Errorf("Failed to get statistics: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get statistics: %v", err)), nil
			}

			jsonBytes, _ := json.Marshal(stats)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// createGetNewTrafficSinceTool 创建增量流量查询工具
func createGetNewTrafficSinceTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_new_traffic_since",
		"获取指定时间之后新增的流量和漏洞。用于定时轮询增量数据。since 参数为 RFC3339 格式（如 2026-05-25T10:00:00Z）。",
		func(input GetNewTrafficSinceInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_new_traffic_since called: since=%s", input.Since)

			if input.Since == "" {
				return claude.ErrorResult("since is required (RFC3339 format, e.g. 2026-05-25T10:00:00Z)"), nil
			}

			since, err := time.Parse(time.RFC3339, input.Since)
			if err != nil {
				return claude.ErrorResult(fmt.Sprintf("Invalid time format: %v. Use RFC3339 (e.g. 2026-05-25T10:00:00Z)", err)), nil
			}

			// 获取新增流量
			newTraffic, err := db.GetNewHistorySince(since, input.SessionID)
			if err != nil {
				return claude.ErrorResult(fmt.Sprintf("Failed to get new traffic: %v", err)), nil
			}

			// 获取新增漏洞
			newVulns, err := db.GetNewVulnerabilitiesSince(since, input.SessionID)
			if err != nil {
				return claude.ErrorResult(fmt.Sprintf("Failed to get new vulnerabilities: %v", err)), nil
			}

			result := map[string]interface{}{
				"since":             input.Since,
				"new_traffic_count": len(newTraffic),
				"new_vuln_count":    len(newVulns),
				"traffic":           formatHistoryItems(newTraffic),
				"vulnerabilities":   formatVulnItems(newVulns),
				"checked_at":        time.Now().Format(time.RFC3339),
			}

			jsonBytes, _ := json.Marshal(result)
			return claude.TextResult(string(jsonBytes)), nil
		},
	)
}

// ==================== 辅助函数 ====================

// formatTrafficList 格式化流量列表
func formatTrafficList(histories []*db.HTTPHistory) map[string]interface{} {
	if histories == nil {
		return map[string]interface{}{
			"count": 0,
			"items": []interface{}{},
		}
	}

	items := formatHistoryItems(histories)
	return map[string]interface{}{
		"count": len(items),
		"items": items,
	}
}

// formatHistoryItems 格式化历史记录条目
func formatHistoryItems(histories []*db.HTTPHistory) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(histories))
	for _, h := range histories {
		items = append(items, map[string]interface{}{
			"id":           h.Hid,
			"host":         h.Host,
			"method":       h.Method,
			"url":          h.FullUrl,
			"path":         h.Path,
			"status":       h.Status,
			"length":       h.Length,
			"content_type": h.ContentType,
			"mime_type":    h.MIMEType,
			"title":        h.Title,
			"ip":           h.IP,
			"created_at":   h.CreatedAt.Format(time.RFC3339),
		})
	}
	return items
}

// formatVulnerabilities 格式化漏洞列表
func formatVulnerabilities(vulns []*db.Vulnerability) map[string]interface{} {
	if vulns == nil {
		return map[string]interface{}{
			"count": 0,
			"items": []interface{}{},
		}
	}

	items := formatVulnItems(vulns)
	return map[string]interface{}{
		"count": len(items),
		"items": items,
	}
}

// formatVulnItems 格式化漏洞条目
func formatVulnItems(vulns []*db.Vulnerability) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(vulns))
	for _, v := range vulns {
		items = append(items, map[string]interface{}{
			"id":          v.ID,
			"vuln_id":     v.VulnID,
			"type":        v.VulnType,
			"target":      v.Target,
			"host":        v.Host,
			"method":      v.Method,
			"path":        v.Path,
			"plugin":      v.Plugin,
			"level":       v.Level,
			"param":       v.Param,
			"payload":     v.Payload,
			"description": v.Description,
			"created_at":  v.CreatedAt.Format(time.RFC3339),
		})
	}
	return items
}

// sendHTTPRequest 发送 HTTP 请求并将流量记录到 ChYing 管道
func sendHTTPRequest(method, targetURL string, headers map[string]string, body string) (map[string]interface{}, error) {
	// 使用 ChYing 内置的 httpx 客户端（遵循代理配置）
	resp, err := httpx.Request(targetURL, method, body, headers, "agent-mcp")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// 将请求/响应注入 ChYing 的流量捕获管道
	captureAgentTraffic(method, targetURL, resp)

	return map[string]interface{}{
		"status_code":    resp.StatusCode,
		"body":           resp.Body,
		"headers":        resp.RespHeader,
		"request_raw":    resp.RequestDump,
		"response_raw":   resp.ResponseDump,
		"content_length": resp.ContentLength,
	}, nil
}

// captureAgentTraffic 将 Agent 发出的请求记录到 ChYing 流量历史中
func captureAgentTraffic(method, targetURL string, resp *httpx.Response) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		logging.Logger.Errorf("captureAgentTraffic: failed to parse URL: %v", err)
		return
	}

	// 生成唯一 ID
	hid := mitmproxy.HistoryItemIDGenerator.Add(1)

	host := parsedURL.Host
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	// 存储到 HTTPBodyMap（前端查看详情时使用）
	httpBody := &mitmproxy.HTTPBody{
		Id:          hid,
		RequestRaw:  resp.RequestDump,
		ResponseRaw: resp.ResponseDump,
	}
	mitmproxy.HTTPBodyMap.Store(hid, httpBody)

	// 发送事件到前端（实时显示）
	historyEvent := &mitmproxy.HTTPHistory{
		Id:       hid,
		Host:     host,
		Method:   method,
		FullUrl:  targetURL,
		Path:     path,
		Status:   fmt.Sprintf("%d", resp.StatusCode),
		Length:   fmt.Sprintf("%d", resp.ContentLength),
		MIMEType: extractMIMEType(resp),
		Note:     "[Agent]", // 标记为 Agent 发出的请求
	}

	if mitmproxy.EventDataChan != nil {
		mitmproxy.SetHTTPUrlMapValue(targetURL, hid)
		mitmproxy.EventDataChan <- &mitmproxy.EventData{
			Name: "HttpHistory",
			Data: historyEvent,
		}
	}
}

// extractMIMEType 从响应中提取 MIME 类型
func extractMIMEType(resp *httpx.Response) string {
	if resp.RespHeader == nil {
		return ""
	}
	ct := resp.RespHeader.Get("Content-Type")
	if ct == "" {
		return ""
	}
	// 只取分号前的部分
	if idx := strings.Index(ct, ";"); idx > 0 {
		ct = ct[:idx]
	}
	return strings.TrimSpace(ct)
}

// analyzeRequest 分析请求，提取安全相关信息
func analyzeRequest(trafficID int64) (map[string]interface{}, error) {
	httpBody, err := db.GetHttpData(int(trafficID))
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic: %w", err)
	}
	if httpBody == nil {
		return nil, fmt.Errorf("traffic %d not found", trafficID)
	}

	analysis := map[string]interface{}{
		"traffic_id": trafficID,
	}

	// 分析请求
	reqRaw := httpBody.RequestRaw
	respRaw := httpBody.ResponseRaw

	// 提取参数（URL 参数 + Body 参数）
	params := extractParameters(reqRaw)
	analysis["parameters"] = params

	// 检测潜在的注入点
	injectionPoints := identifyInjectionPoints(params)
	analysis["injection_points"] = injectionPoints

	// 检测认证信息
	authInfo := detectAuthInfo(reqRaw)
	analysis["auth_info"] = authInfo

	// 检测敏感信息泄露
	leaks := detectInfoLeaks(respRaw)
	analysis["info_leaks"] = leaks

	// 检测安全头部缺失
	missingHeaders := checkSecurityHeaders(respRaw)
	analysis["missing_security_headers"] = missingHeaders

	return analysis, nil
}

// extractParameters 从原始请求中提取参数
func extractParameters(reqRaw string) []map[string]string {
	params := make([]map[string]string, 0)

	lines := strings.Split(reqRaw, "\n")
	if len(lines) == 0 {
		return params
	}

	// 解析请求行中的 URL 参数
	requestLine := strings.TrimSpace(lines[0])
	parts := strings.Fields(requestLine)
	if len(parts) >= 2 {
		path := parts[1]
		if qIdx := strings.Index(path, "?"); qIdx >= 0 {
			queryStr := path[qIdx+1:]
			values, _ := url.ParseQuery(queryStr)
			for k, vs := range values {
				for _, v := range vs {
					params = append(params, map[string]string{
						"name":     k,
						"value":    v,
						"location": "query",
					})
				}
			}
		}
	}

	// 解析 Body 参数（简单处理 form-urlencoded）
	bodyStart := false
	var bodyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			bodyStart = true
			continue
		}
		if bodyStart {
			bodyLines = append(bodyLines, line)
		}
	}
	if len(bodyLines) > 0 {
		body := strings.Join(bodyLines, "\n")
		// 尝试作为 form-urlencoded 解析
		values, err := url.ParseQuery(strings.TrimSpace(body))
		if err == nil && len(values) > 0 {
			for k, vs := range values {
				for _, v := range vs {
					params = append(params, map[string]string{
						"name":     k,
						"value":    v,
						"location": "body",
					})
				}
			}
		} else {
			// 尝试作为 JSON 解析
			var jsonData map[string]interface{}
			if json.Unmarshal([]byte(strings.TrimSpace(body)), &jsonData) == nil {
				for k, v := range jsonData {
					params = append(params, map[string]string{
						"name":     k,
						"value":    fmt.Sprintf("%v", v),
						"location": "json_body",
					})
				}
			}
		}
	}

	return params
}

// identifyInjectionPoints 识别潜在注入点
func identifyInjectionPoints(params []map[string]string) []map[string]string {
	points := make([]map[string]string, 0)

	sqlKeywords := []string{"id", "uid", "user_id", "order", "sort", "page", "limit", "offset", "where", "query", "search", "filter"}
	xssKeywords := []string{"name", "title", "content", "message", "comment", "desc", "text", "value", "label", "q", "keyword"}
	ssrfKeywords := []string{"url", "link", "href", "src", "redirect", "next", "return", "callback", "target", "uri", "path", "file", "image"}
	idorKeywords := []string{"id", "uid", "user_id", "account", "order_id", "doc_id", "file_id", "project_id"}

	for _, p := range params {
		name := strings.ToLower(p["name"])

		for _, kw := range sqlKeywords {
			if strings.Contains(name, kw) {
				points = append(points, map[string]string{
					"param": p["name"],
					"type":  "SQL Injection",
					"reason": fmt.Sprintf("parameter name '%s' matches SQL injection pattern", p["name"]),
				})
				break
			}
		}
		for _, kw := range xssKeywords {
			if strings.Contains(name, kw) {
				points = append(points, map[string]string{
					"param": p["name"],
					"type":  "XSS",
					"reason": fmt.Sprintf("parameter name '%s' likely renders in HTML", p["name"]),
				})
				break
			}
		}
		for _, kw := range ssrfKeywords {
			if strings.Contains(name, kw) {
				points = append(points, map[string]string{
					"param": p["name"],
					"type":  "SSRF",
					"reason": fmt.Sprintf("parameter name '%s' may accept URL input", p["name"]),
				})
				break
			}
		}
		for _, kw := range idorKeywords {
			if name == kw || strings.HasSuffix(name, "_"+kw) {
				points = append(points, map[string]string{
					"param": p["name"],
					"type":  "IDOR",
					"reason": fmt.Sprintf("parameter '%s' is a direct object reference", p["name"]),
				})
				break
			}
		}
	}

	return points
}

// detectAuthInfo 检测请求中的认证信息
func detectAuthInfo(reqRaw string) map[string]interface{} {
	info := map[string]interface{}{
		"has_auth": false,
	}

	lower := strings.ToLower(reqRaw)

	if strings.Contains(lower, "authorization:") {
		info["has_auth"] = true
		// 提取 Authorization 头的值
		for _, line := range strings.Split(reqRaw, "\n") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "authorization:") {
				val := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
				if strings.HasPrefix(strings.ToLower(val), "bearer ") {
					info["type"] = "Bearer Token (JWT?)"
					token := strings.TrimPrefix(val, "Bearer ")
					token = strings.TrimPrefix(token, "bearer ")
					// JWT 有 3 段
					if len(strings.Split(token, ".")) == 3 {
						info["type"] = "JWT"
					}
				} else if strings.HasPrefix(strings.ToLower(val), "basic ") {
					info["type"] = "Basic Auth"
				} else {
					info["type"] = "Custom"
				}
				break
			}
		}
	}

	if strings.Contains(lower, "cookie:") {
		info["has_cookies"] = true
	}

	if strings.Contains(lower, "x-api-key:") || strings.Contains(lower, "api-key:") || strings.Contains(lower, "apikey:") {
		info["has_api_key"] = true
	}

	return info
}

// detectInfoLeaks 检测响应中的敏感信息泄露
func detectInfoLeaks(respRaw string) []string {
	leaks := make([]string, 0)

	lower := strings.ToLower(respRaw)

	// 检测常见泄露模式
	if strings.Contains(lower, "stack trace") || strings.Contains(lower, "stacktrace") || strings.Contains(lower, "traceback") {
		leaks = append(leaks, "Stack trace exposed in response")
	}
	if strings.Contains(lower, "sql syntax") || strings.Contains(lower, "mysql") || strings.Contains(lower, "postgresql") {
		leaks = append(leaks, "Database error message exposed")
	}
	if strings.Contains(lower, "/usr/") || strings.Contains(lower, "c:\\") || strings.Contains(lower, "/home/") || strings.Contains(lower, "/var/www") {
		leaks = append(leaks, "Internal file path exposed")
	}
	if strings.Contains(lower, "x-powered-by:") {
		leaks = append(leaks, "X-Powered-By header reveals technology stack")
	}
	if strings.Contains(lower, "server:") && (strings.Contains(lower, "apache/") || strings.Contains(lower, "nginx/") || strings.Contains(lower, "iis/")) {
		leaks = append(leaks, "Server version exposed in headers")
	}
	if strings.Contains(lower, "swagger") || strings.Contains(lower, "openapi") {
		leaks = append(leaks, "API documentation/Swagger exposed")
	}

	return leaks
}

// checkSecurityHeaders 检查安全头部
func checkSecurityHeaders(respRaw string) []string {
	missing := make([]string, 0)

	lower := strings.ToLower(respRaw)

	if !strings.Contains(lower, "x-content-type-options") {
		missing = append(missing, "X-Content-Type-Options")
	}
	if !strings.Contains(lower, "x-frame-options") {
		missing = append(missing, "X-Frame-Options")
	}
	if !strings.Contains(lower, "strict-transport-security") {
		missing = append(missing, "Strict-Transport-Security")
	}
	if !strings.Contains(lower, "content-security-policy") {
		missing = append(missing, "Content-Security-Policy")
	}
	if !strings.Contains(lower, "x-xss-protection") {
		// X-XSS-Protection is deprecated but still checked
		missing = append(missing, "X-XSS-Protection (deprecated but notable)")
	}

	return missing
}

// searchTraffic 搜索流量
func searchTraffic(projectID, keyword, searchIn string, limit int) (map[string]interface{}, error) {
	histories, err := db.SearchHistory(projectID, keyword, searchIn, limit)
	if err != nil {
		return nil, err
	}

	items := formatHistoryItems(histories)
	return map[string]interface{}{
		"keyword":   keyword,
		"search_in": searchIn,
		"count":     len(items),
		"items":     items,
	}, nil
}

// getSitemap 获取网站地图
func getSitemap(projectID, host string) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if host != "" {
		// 获取指定主机的路径
		paths, err := db.GetDistinctPaths(projectID, host)
		if err != nil {
			return nil, err
		}
		result["host"] = host
		result["paths"] = paths
		result["path_count"] = len(paths)
	} else {
		// 获取所有主机及其路径
		hosts := db.GetHosts()
		hostMap := make(map[string]interface{})
		for _, h := range hosts {
			paths, err := db.GetDistinctPaths(projectID, h)
			if err != nil {
				continue
			}
			hostMap[h] = map[string]interface{}{
				"paths":      paths,
				"path_count": len(paths),
			}
		}
		result["hosts"] = hostMap
		result["host_count"] = len(hosts)
	}

	return result, nil
}

// getStatistics 获取统计信息
func getStatistics(projectID string) (map[string]interface{}, error) {
	// 流量统计
	trafficStats, err := db.GetTrafficStatistics(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic statistics: %w", err)
	}

	// 漏洞统计
	vulnStats, err := db.GetVulnerabilityStatistics(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerability statistics: %w", err)
	}

	return map[string]interface{}{
		"traffic":         trafficStats,
		"vulnerabilities": vulnStats,
	}, nil
}

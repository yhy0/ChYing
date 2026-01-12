package claudecode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/10
   @desc MCP Server - 使用 mcp-go 库为 Claude Code 提供安全工具
**/

// MCPServer MCP 服务器封装
type MCPServer struct {
	server    *server.MCPServer
	sseServer *server.SSEServer
	cmd       *exec.Cmd // stdio 模式下的子进程
	stdin     io.WriteCloser
	stdout    io.ReadCloser

	mu       sync.RWMutex
	running  bool
	mode     string // "sse" 或 "stdio"
	port     int
	execPath string // MCP 可执行文件路径

	// 配置
	config *BuiltinMCPConfig
}

// NewMCPServer 创建 MCP 服务器
func NewMCPServer() *MCPServer {
	return NewMCPServerWithConfig(nil)
}

// NewMCPServerWithConfig 创建带配置的 MCP 服务器
func NewMCPServerWithConfig(config *BuiltinMCPConfig) *MCPServer {
	s := &MCPServer{
		mode:   "sse", // 默认使用 SSE 模式
		config: config,
	}
	if config != nil && config.Mode != "" {
		s.mode = config.Mode
	}
	if config != nil && config.Port > 0 {
		s.port = config.Port
	}
	s.initServer()
	return s
}

// initServer 初始化 MCP 服务器
func (s *MCPServer) initServer() {
	// 创建 MCP 服务器
	s.server = server.NewMCPServer(
		"ChYing Security Tools",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// 注册安全工具
	s.registerTools()
}

// isToolEnabled 检查工具是否应该被启用
func (s *MCPServer) isToolEnabled(toolName string) bool {
	if s.config == nil {
		return true // 无配置时默认启用所有工具
	}

	// 如果在禁用列表中，则禁用
	for _, disabled := range s.config.DisabledTools {
		if disabled == toolName {
			return false
		}
	}

	// 如果启用列表为空，则启用所有未禁用的工具
	if len(s.config.EnabledTools) == 0 {
		return true
	}

	// 如果启用列表不为空，只启用列表中的工具
	for _, enabled := range s.config.EnabledTools {
		if enabled == toolName {
			return true
		}
	}

	return false
}

// addToolIfEnabled 如果工具启用则添加
func (s *MCPServer) addToolIfEnabled(tool mcp.Tool, handler server.ToolHandlerFunc) {
	if s.isToolEnabled(tool.Name) {
		s.server.AddTool(tool, handler)
		logging.Logger.Debugf("MCP tool registered: %s", tool.Name)
	} else {
		logging.Logger.Debugf("MCP tool disabled: %s", tool.Name)
	}
}

// registerTools 注册所有安全工具
func (s *MCPServer) registerTools() {
	// 1. 获取 HTTP 历史记录
	s.addToolIfEnabled(
		mcp.NewTool("get_http_history",
			mcp.WithDescription("获取 HTTP 流量历史记录，可按项目、主机、方法等过滤。返回代理捕获的 HTTP 请求列表。"),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("项目 ID"),
			),
			mcp.WithString("host",
				mcp.Description("主机名过滤，只返回匹配的主机"),
			),
			mcp.WithString("method",
				mcp.Description("HTTP 方法过滤 (GET, POST, PUT, DELETE 等)"),
			),
			mcp.WithString("path_contains",
				mcp.Description("URL 路径包含的关键字"),
			),
			mcp.WithNumber("status_code",
				mcp.Description("HTTP 状态码过滤"),
			),
			mcp.WithNumber("limit",
				mcp.Description("返回数量限制，默认 50，最大 500"),
			),
			mcp.WithNumber("offset",
				mcp.Description("分页偏移量"),
			),
		),
		s.handleGetHTTPHistory,
	)

	// 2. 获取流量详情
	s.addToolIfEnabled(
		mcp.NewTool("get_traffic_detail",
			mcp.WithDescription("获取单个 HTTP 流量的详细信息，包括完整的请求头、请求体、响应头和响应体"),
			mcp.WithNumber("traffic_id",
				mcp.Required(),
				mcp.Description("流量记录 ID"),
			),
		),
		s.handleGetTrafficDetail,
	)

	// 3. 获取漏洞列表
	s.addToolIfEnabled(
		mcp.NewTool("get_vulnerabilities",
			mcp.WithDescription("获取已发现的漏洞列表，可按项目、严重程度、类型等过滤"),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("项目 ID"),
			),
			mcp.WithString("severity",
				mcp.Description("严重程度过滤: critical, high, medium, low, info"),
			),
			mcp.WithString("vuln_type",
				mcp.Description("漏洞类型过滤，如 XSS, SQLi, SSRF 等"),
			),
			mcp.WithString("host",
				mcp.Description("主机名过滤"),
			),
			mcp.WithNumber("limit",
				mcp.Description("返回数量限制"),
			),
		),
		s.handleGetVulnerabilities,
	)

	// 4. 发送 HTTP 请求
	s.addToolIfEnabled(
		mcp.NewTool("send_http_request",
			mcp.WithDescription("发送自定义 HTTP 请求并返回响应。用于测试和验证漏洞。请谨慎使用。"),
			mcp.WithString("method",
				mcp.Required(),
				mcp.Description("HTTP 方法: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS"),
			),
			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("完整的请求 URL"),
			),
			mcp.WithObject("headers",
				mcp.Description("请求头，键值对格式"),
			),
			mcp.WithString("body",
				mcp.Description("请求体内容"),
			),
			mcp.WithString("content_type",
				mcp.Description("Content-Type，默认 application/x-www-form-urlencoded"),
			),
			mcp.WithNumber("timeout",
				mcp.Description("请求超时时间（秒），默认 30"),
			),
			mcp.WithBoolean("follow_redirects",
				mcp.Description("是否跟随重定向，默认 true"),
			),
		),
		s.handleSendHTTPRequest,
	)

	// 5. 分析 HTTP 请求
	s.addToolIfEnabled(
		mcp.NewTool("analyze_request",
			mcp.WithDescription("分析 HTTP 请求，识别潜在的安全问题和攻击向量"),
			mcp.WithString("request",
				mcp.Required(),
				mcp.Description("原始 HTTP 请求文本"),
			),
		),
		s.handleAnalyzeRequest,
	)

	// 6. 搜索流量
	s.addToolIfEnabled(
		mcp.NewTool("search_traffic",
			mcp.WithDescription("在 HTTP 流量中搜索关键字，支持在请求和响应中搜索"),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("项目 ID"),
			),
			mcp.WithString("keyword",
				mcp.Required(),
				mcp.Description("搜索关键字"),
			),
			mcp.WithBoolean("search_request",
				mcp.Description("是否搜索请求内容，默认 true"),
			),
			mcp.WithBoolean("search_response",
				mcp.Description("是否搜索响应内容，默认 true"),
			),
			mcp.WithNumber("limit",
				mcp.Description("返回数量限制，默认 20"),
			),
		),
		s.handleSearchTraffic,
	)

	// 7. 获取站点地图
	s.addToolIfEnabled(
		mcp.NewTool("get_sitemap",
			mcp.WithDescription("获取目标站点的 URL 结构地图"),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("项目 ID"),
			),
			mcp.WithString("host",
				mcp.Description("主机名过滤"),
			),
		),
		s.handleGetSitemap,
	)

	// 8. 获取统计信息
	s.addToolIfEnabled(
		mcp.NewTool("get_statistics",
			mcp.WithDescription("获取项目的统计信息，包括流量数量、漏洞分布等"),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("项目 ID"),
			),
		),
		s.handleGetStatistics,
	)
}

// ==================== 工具处理函数 ====================

// handleGetHTTPHistory 获取 HTTP 历史
func (s *MCPServer) handleGetHTTPHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, _ := args["project_id"].(string)
	host, _ := args["host"].(string)
	method, _ := args["method"].(string)
	pathContains, _ := args["path_contains"].(string)
	statusCode := int(getFloat64(args, "status_code", 0))
	limit := int(getFloat64(args, "limit", 50))
	offset := int(getFloat64(args, "offset", 0))

	if limit > 500 {
		limit = 500
	}
	if limit <= 0 {
		limit = 50
	}

	// 从数据库获取历史记录
	histories, err := db.GetAllHistory(projectID, "", limit, offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取历史记录失败: %v", err)), nil
	}

	// 过滤
	var filtered []*db.HTTPHistory
	for _, h := range histories {
		if host != "" && !strings.Contains(h.Host, host) {
			continue
		}
		if method != "" && h.Method != method {
			continue
		}
		if pathContains != "" && !strings.Contains(h.Path, pathContains) {
			continue
		}
		if statusCode > 0 && h.Status != fmt.Sprintf("%d", statusCode) {
			continue
		}
		filtered = append(filtered, h)
	}

	// 构建简化的结果
	type SimplifiedHistory struct {
		ID       int64  `json:"id"`
		Method   string `json:"method"`
		Host     string `json:"host"`
		Path     string `json:"path"`
		Status   string `json:"status"`
		Length   string `json:"length"`
		MIMEType string `json:"mime_type"`
		Time     string `json:"time"`
	}

	var results []SimplifiedHistory
	for _, h := range filtered {
		results = append(results, SimplifiedHistory{
			ID:       h.ID,
			Method:   h.Method,
			Host:     h.Host,
			Path:     h.Path,
			Status:   h.Status,
			Length:   h.Length,
			MIMEType: h.MIMEType,
			Time:     h.Time,
		})
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"total":   len(results),
		"records": results,
	}, "", "  ")

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetTrafficDetail 获取流量详情
func (s *MCPServer) handleGetTrafficDetail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	trafficID := int64(getFloat64(args, "traffic_id", 0))

	if trafficID <= 0 {
		return mcp.NewToolResultError("traffic_id 必须大于 0"), nil
	}

	history, err := db.GetHistoryByID(trafficID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取流量详情失败: %v", err)), nil
	}

	if history == nil {
		return mcp.NewToolResultError("未找到指定的流量记录"), nil
	}

	data, _ := json.MarshalIndent(history, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetVulnerabilities 获取漏洞列表
func (s *MCPServer) handleGetVulnerabilities(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, _ := args["project_id"].(string)
	severity, _ := args["severity"].(string)
	vulnType, _ := args["vuln_type"].(string)
	host, _ := args["host"].(string)
	limit := int(getFloat64(args, "limit", 100))

	// 从数据库获取漏洞
	vulns, err := db.GetAllVulnerabilities(projectID, "", limit, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取漏洞列表失败: %v", err)), nil
	}

	// 过滤
	var filtered []*db.Vulnerability
	for _, v := range vulns {
		if severity != "" && !strings.EqualFold(v.Level, severity) {
			continue
		}
		if vulnType != "" && !strings.Contains(strings.ToLower(v.VulnType), strings.ToLower(vulnType)) {
			continue
		}
		if host != "" && !strings.Contains(v.Host, host) {
			continue
		}
		filtered = append(filtered, v)
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"total":           len(filtered),
		"vulnerabilities": filtered,
	}, "", "  ")

	return mcp.NewToolResultText(string(data)), nil
}

// handleSendHTTPRequest 发送 HTTP 请求
func (s *MCPServer) handleSendHTTPRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	method, _ := args["method"].(string)
	url, _ := args["url"].(string)
	headers, _ := args["headers"].(map[string]interface{})
	body, _ := args["body"].(string)
	contentType, _ := args["content_type"].(string)
	timeout := int(getFloat64(args, "timeout", 30))
	followRedirects := getBool(args, "follow_redirects", true)

	if method == "" || url == "" {
		return mcp.NewToolResultError("method 和 url 是必需的"), nil
	}

	// 验证 URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return mcp.NewToolResultError("URL 必须以 http:// 或 https:// 开头"), nil
	}

	// 创建请求
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("创建请求失败: %v", err)), nil
	}

	// 设置请求头
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for k, v := range headers {
		if strVal, ok := v.(string); ok {
			req.Header.Set(k, strVal)
		}
	}

	// 创建客户端
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("发送请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	// 读取响应（限制大小）
	maxSize := int64(1024 * 1024) // 1MB
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取响应失败: %v", err)), nil
	}

	// 构建结果
	result := map[string]interface{}{
		"status_code":   resp.StatusCode,
		"status":        resp.Status,
		"headers":       resp.Header,
		"content_length": resp.ContentLength,
		"body":          string(bodyBytes),
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleAnalyzeRequest 分析 HTTP 请求
func (s *MCPServer) handleAnalyzeRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	rawRequest, _ := args["request"].(string)

	if rawRequest == "" {
		return mcp.NewToolResultError("request 参数是必需的"), nil
	}

	analysis := analyzeHTTPRequestSecurity(rawRequest)
	data, _ := json.MarshalIndent(analysis, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleSearchTraffic 搜索流量
func (s *MCPServer) handleSearchTraffic(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, _ := args["project_id"].(string)
	keyword, _ := args["keyword"].(string)
	searchRequest := getBool(args, "search_request", true)
	searchResponse := getBool(args, "search_response", true)
	limit := int(getFloat64(args, "limit", 20))

	if keyword == "" {
		return mcp.NewToolResultError("keyword 参数是必需的"), nil
	}

	// 获取所有历史记录
	histories, err := db.GetAllHistory(projectID, "", 1000, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索失败: %v", err)), nil
	}

	// 搜索
	var matches []map[string]interface{}
	keywordLower := strings.ToLower(keyword)

	for _, h := range histories {
		matched := false
		matchLocations := []string{}

		// 获取请求和响应内容
		req, res := db.GetTraffic(int(h.Hid))

		if searchRequest && req != nil {
			if strings.Contains(strings.ToLower(req.RequestRaw), keywordLower) {
				matched = true
				matchLocations = append(matchLocations, "request")
			}
		}
		if searchResponse && res != nil {
			if strings.Contains(strings.ToLower(res.ResponseRaw), keywordLower) {
				matched = true
				matchLocations = append(matchLocations, "response")
			}
		}

		if matched {
			matches = append(matches, map[string]interface{}{
				"id":              h.ID,
				"method":          h.Method,
				"host":            h.Host,
				"path":            h.Path,
				"status":          h.Status,
				"match_locations": matchLocations,
			})

			if len(matches) >= limit {
				break
			}
		}
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"keyword": keyword,
		"total":   len(matches),
		"matches": matches,
	}, "", "  ")

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetSitemap 获取站点地图
func (s *MCPServer) handleGetSitemap(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	projectID, _ := args["project_id"].(string)
	hostFilter, _ := args["host"].(string)

	// 获取所有历史记录
	histories, err := db.GetAllHistory(projectID, "", 5000, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取站点地图失败: %v", err)), nil
	}

	// 构建站点地图
	sitemap := make(map[string]map[string][]string) // host -> path -> methods

	for _, h := range histories {
		if hostFilter != "" && h.Host != hostFilter {
			continue
		}

		if sitemap[h.Host] == nil {
			sitemap[h.Host] = make(map[string][]string)
		}

		// 检查方法是否已存在
		methods := sitemap[h.Host][h.Path]
		found := false
		for _, m := range methods {
			if m == h.Method {
				found = true
				break
			}
		}
		if !found {
			sitemap[h.Host][h.Path] = append(methods, h.Method)
		}
	}

	// 转换为更友好的格式
	var result []map[string]interface{}
	for host, paths := range sitemap {
		hostEntry := map[string]interface{}{
			"host":        host,
			"total_paths": len(paths),
			"paths":       []map[string]interface{}{},
		}

		for path, methods := range paths {
			hostEntry["paths"] = append(hostEntry["paths"].([]map[string]interface{}), map[string]interface{}{
				"path":    path,
				"methods": methods,
			})
		}

		result = append(result, hostEntry)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetStatistics 获取统计信息
func (s *MCPServer) handleGetStatistics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectID, _ := args["project_id"].(string)

	// 获取流量统计
	histories, _ := db.GetAllHistory(projectID, "", 10000, 0)

	// 获取漏洞统计
	vulns, _ := db.GetAllVulnerabilities(projectID, "", 10000, 0)

	// 统计
	hostCount := make(map[string]int)
	methodCount := make(map[string]int)
	statusCount := make(map[string]int)

	for _, h := range histories {
		hostCount[h.Host]++
		methodCount[h.Method]++
		statusCount[h.Status]++
	}

	vulnBySeverity := make(map[string]int)
	vulnByType := make(map[string]int)

	for _, v := range vulns {
		vulnBySeverity[v.Level]++
		vulnByType[v.VulnType]++
	}

	stats := map[string]interface{}{
		"traffic": map[string]interface{}{
			"total":         len(histories),
			"by_host":       hostCount,
			"by_method":     methodCount,
			"by_status":     statusCount,
		},
		"vulnerabilities": map[string]interface{}{
			"total":       len(vulns),
			"by_severity": vulnBySeverity,
			"by_type":     vulnByType,
		},
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ==================== 服务器控制 ====================

// Start 启动 MCP 服务器 (SSE 模式)
func (s *MCPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// 使用 SSE 模式
	s.sseServer = server.NewSSEServer(s.server)

	// 启动服务器
	go func() {
		// 使用配置的端口或默认端口
		startPort := 18080
		if s.port > 0 {
			startPort = s.port
		}

		// 动态选择端口
		for i := 0; i < 100; i++ {
			port := startPort + i
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			logging.Logger.Infof("Trying to start MCP SSE server on %s", addr)
			err := s.sseServer.Start(addr)
			if err == nil {
				s.mu.Lock()
				s.port = port
				s.mu.Unlock()
				return
			}
			logging.Logger.Warnf("Port %d in use, trying next...", port)
		}
		logging.Logger.Error("Failed to start MCP SSE server: no available port")
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)
	s.running = true
	s.mode = "sse"

	logging.Logger.Info("MCP Server started in SSE mode")
	return nil
}

// StartStdio 启动 stdio 模式（用于 Claude Code CLI 直接调用）
// 这个函数会阻塞，应该在独立的可执行文件中调用
func (s *MCPServer) StartStdio() error {
	return server.ServeStdio(s.server)
}

// Stop 停止 MCP 服务器
func (s *MCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.sseServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.sseServer.Shutdown(ctx)
	}

	s.running = false
	logging.Logger.Info("MCP Server stopped")
	return nil
}

// IsRunning 检查服务器是否运行中
func (s *MCPServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetPort 获取服务器端口
func (s *MCPServer) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

// GetURL 获取服务器 URL
func (s *MCPServer) GetURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.mode == "sse" && s.port > 0 {
		return fmt.Sprintf("http://127.0.0.1:%d/sse", s.port)
	}
	return ""
}

// GetMode 获取运行模式
func (s *MCPServer) GetMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

// ==================== 辅助函数 ====================

// getFloat64 安全获取 float64 值
func getFloat64(args map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := args[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return defaultVal
}

// getBool 安全获取 bool 值
func getBool(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key]; ok {
		if boolVal, ok := v.(bool); ok {
			return boolVal
		}
	}
	return defaultVal
}

// analyzeHTTPRequestSecurity 分析 HTTP 请求的安全问题
func analyzeHTTPRequestSecurity(request string) map[string]interface{} {
	findings := []map[string]interface{}{}
	requestLower := strings.ToLower(request)

	// 检查敏感参数
	sensitiveParams := []struct {
		pattern     string
		description string
		severity    string
	}{
		{"password=", "发现明文密码参数", "high"},
		{"passwd=", "发现明文密码参数", "high"},
		{"pwd=", "发现可能的密码参数", "medium"},
		{"token=", "发现令牌参数", "medium"},
		{"api_key=", "发现 API 密钥参数", "high"},
		{"apikey=", "发现 API 密钥参数", "high"},
		{"secret=", "发现密钥参数", "high"},
		{"auth=", "发现认证参数", "medium"},
		{"session=", "发现会话参数", "medium"},
		{"credit_card", "发现信用卡相关参数", "critical"},
		{"ssn=", "发现社会安全号参数", "critical"},
	}

	for _, sp := range sensitiveParams {
		if strings.Contains(requestLower, sp.pattern) {
			findings = append(findings, map[string]interface{}{
				"type":        "sensitive_data",
				"description": sp.description,
				"severity":    sp.severity,
				"pattern":     sp.pattern,
			})
		}
	}

	// 检查注入风险
	injectionPatterns := []struct {
		pattern     string
		description string
		vulnType    string
	}{
		{"select ", "可能存在 SQL 注入", "sqli"},
		{"union ", "可能存在 SQL 注入 (UNION)", "sqli"},
		{"insert ", "可能存在 SQL 注入 (INSERT)", "sqli"},
		{"update ", "可能存在 SQL 注入 (UPDATE)", "sqli"},
		{"delete ", "可能存在 SQL 注入 (DELETE)", "sqli"},
		{"<script", "可能存在 XSS", "xss"},
		{"javascript:", "可能存在 XSS", "xss"},
		{"onerror=", "可能存在 XSS (事件处理)", "xss"},
		{"onload=", "可能存在 XSS (事件处理)", "xss"},
		{"../", "可能存在路径遍历", "path_traversal"},
		{"..\\", "可能存在路径遍历", "path_traversal"},
		{"; ", "可能存在命令注入", "command_injection"},
		{"| ", "可能存在命令注入", "command_injection"},
		{"` ", "可能存在命令注入", "command_injection"},
		{"$((", "可能存在命令注入", "command_injection"},
		{"{{", "可能存在模板注入", "ssti"},
		{"${", "可能存在表达式注入", "expression_injection"},
	}

	for _, ip := range injectionPatterns {
		if strings.Contains(requestLower, ip.pattern) {
			findings = append(findings, map[string]interface{}{
				"type":        ip.vulnType,
				"description": ip.description,
				"severity":    "high",
				"pattern":     ip.pattern,
			})
		}
	}

	// 检查不安全的头部
	unsafeHeaders := []struct {
		header      string
		description string
	}{
		{"x-forwarded-for:", "发现 X-Forwarded-For 头，可能用于 IP 欺骗"},
		{"x-real-ip:", "发现 X-Real-IP 头，可能用于 IP 欺骗"},
		{"x-custom-ip-authorization:", "发现自定义 IP 授权头"},
	}

	for _, uh := range unsafeHeaders {
		if strings.Contains(requestLower, uh.header) {
			findings = append(findings, map[string]interface{}{
				"type":        "header_manipulation",
				"description": uh.description,
				"severity":    "medium",
				"header":      uh.header,
			})
		}
	}

	// 计算风险等级
	riskLevel := "low"
	criticalCount := 0
	highCount := 0

	for _, f := range findings {
		switch f["severity"] {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}

	if criticalCount > 0 {
		riskLevel = "critical"
	} else if highCount > 0 {
		riskLevel = "high"
	} else if len(findings) > 0 {
		riskLevel = "medium"
	}

	return map[string]interface{}{
		"findings":      findings,
		"total_issues":  len(findings),
		"risk_level":    riskLevel,
		"request_size":  len(request),
		"analyzed_at":   time.Now().Format(time.RFC3339),
	}
}

// ==================== Stdio 模式入口 ====================

// RunStdioServer 运行 stdio 模式的 MCP 服务器
// 这个函数用于创建独立的 MCP 可执行文件
func RunStdioServer() {
	s := NewMCPServer()
	if err := s.StartStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP Server error: %v\n", err)
		os.Exit(1)
	}
}

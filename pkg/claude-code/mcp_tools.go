package claudecode

import (
	"fmt"

	claude "github.com/yhy0/claude-agent-sdk-go"
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

			// 从数据库获取流量历史
			histories, err := db.GetAllHistory(input.ProjectID, "", limit, input.Offset)
			if err != nil {
				logging.Logger.Errorf("Failed to get HTTP history: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get HTTP history: %v", err)), nil
			}

			// 格式化结果
			result := formatTrafficList(histories)
			return claude.TextResult(fmt.Sprintf("%v", result)), nil
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

			// 从数据库获取流量详情
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

			// 从数据库获取漏洞列表
			vulns, err := db.GetAllVulnerabilities(input.ProjectID, "", limit, 0)
			if err != nil {
				logging.Logger.Errorf("Failed to get vulnerabilities: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get vulnerabilities: %v", err)), nil
			}

			result := formatVulnerabilities(vulns)
			return claude.TextResult(fmt.Sprintf("%v", result)), nil
		},
	)
}

// createSendHTTPRequestTool 创建发送 HTTP 请求工具
func createSendHTTPRequestTool() *claude.Tool {
	return claude.NewTypedTool(
		"send_http_request",
		"发送自定义 HTTP 请求并返回响应。用于测试和验证漏洞。请谨慎使用。",
		func(input SendHTTPRequestInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("send_http_request called: method=%s, url=%s", input.Method, input.URL)

			if input.Method == "" {
				return claude.ErrorResult("method is required"), nil
			}
			if input.URL == "" {
				return claude.ErrorResult("url is required"), nil
			}

			// 发送 HTTP 请求
			resp, err := sendHTTPRequest(input.Method, input.URL, input.Headers, input.Body)
			if err != nil {
				logging.Logger.Errorf("Failed to send HTTP request: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to send HTTP request: %v", err)), nil
			}

			return claude.TextResult(fmt.Sprintf("%v", resp)), nil
		},
	)
}

// createAnalyzeRequestTool 创建分析请求工具
func createAnalyzeRequestTool() *claude.Tool {
	return claude.NewTypedTool(
		"analyze_request",
		"分析 HTTP 请求，识别潜在的安全问题和攻击向量",
		func(input AnalyzeRequestInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("analyze_request called: trafficID=%d", input.TrafficID)

			if input.TrafficID <= 0 {
				return claude.ErrorResult("traffic_id is required and must be positive"), nil
			}

			// 分析请求
			analysis, err := analyzeRequest(int64(input.TrafficID))
			if err != nil {
				logging.Logger.Errorf("Failed to analyze request: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to analyze request: %v", err)), nil
			}

			return claude.TextResult(fmt.Sprintf("%v", analysis)), nil
		},
	)
}

// createSearchTrafficTool 创建搜索流量工具
func createSearchTrafficTool() *claude.Tool {
	return claude.NewTypedTool(
		"search_traffic",
		"搜索包含特定关键词的 HTTP 流量",
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

			// 搜索流量
			results, err := searchTraffic(input.ProjectID, input.Keyword, searchIn, limit)
			if err != nil {
				logging.Logger.Errorf("Failed to search traffic: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to search traffic: %v", err)), nil
			}

			return claude.TextResult(fmt.Sprintf("%v", results)), nil
		},
	)
}

// createGetSitemapTool 创建获取网站地图工具
func createGetSitemapTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_sitemap",
		"获取发现的网站地图（所有 URL 路径）",
		func(input GetSitemapInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_sitemap called: projectID=%s, host=%s",
				input.ProjectID, input.Host)

			if input.ProjectID == "" {
				return claude.ErrorResult("project_id is required"), nil
			}

			// 获取网站地图
			sitemap, err := getSitemap(input.ProjectID, input.Host)
			if err != nil {
				logging.Logger.Errorf("Failed to get sitemap: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get sitemap: %v", err)), nil
			}

			return claude.TextResult(fmt.Sprintf("%v", sitemap)), nil
		},
	)
}

// createGetStatisticsTool 创建获取统计信息工具
func createGetStatisticsTool() *claude.Tool {
	return claude.NewTypedTool(
		"get_statistics",
		"获取项目的统计信息（流量数、漏洞数、主机数等）",
		func(input GetStatisticsInput) (*claude.ToolResult, error) {
			logging.Logger.Infof("get_statistics called: projectID=%s", input.ProjectID)

			if input.ProjectID == "" {
				return claude.ErrorResult("project_id is required"), nil
			}

			// 获取统计信息
			stats, err := getStatistics(input.ProjectID)
			if err != nil {
				logging.Logger.Errorf("Failed to get statistics: %v", err)
				return claude.ErrorResult(fmt.Sprintf("Failed to get statistics: %v", err)), nil
			}

			return claude.TextResult(fmt.Sprintf("%v", stats)), nil
		},
	)
}

// ==================== 辅助函数 ====================

// formatTrafficList 格式化流量列表
func formatTrafficList(histories interface{}) interface{} {
	if histories == nil {
		return map[string]interface{}{
			"count": 0,
			"items": []interface{}{},
		}
	}
	// TODO: 根据实际的 histories 类型进行格式化
	return map[string]interface{}{
		"count": 0,
		"items": histories,
	}
}

// formatVulnerabilities 格式化漏洞列表
func formatVulnerabilities(vulns interface{}) interface{} {
	if vulns == nil {
		return map[string]interface{}{
			"count": 0,
			"items": []interface{}{},
		}
	}
	return map[string]interface{}{
		"count": 0,
		"items": vulns,
	}
}

// sendHTTPRequest 发送 HTTP 请求
func sendHTTPRequest(method, url string, headers map[string]string, body string) (interface{}, error) {
	// TODO: 实现实际的 HTTP 请求发送
	return map[string]interface{}{
		"status_code": 200,
		"headers":     map[string]string{},
		"body":        "",
	}, nil
}

// analyzeRequest 分析请求
func analyzeRequest(trafficID int64) (interface{}, error) {
	// TODO: 实现实际的请求分析
	return map[string]interface{}{
		"traffic_id": trafficID,
		"issues":     []interface{}{},
	}, nil
}

// searchTraffic 搜索流量
func searchTraffic(projectID, keyword, searchIn string, limit int) (interface{}, error) {
	// TODO: 实现实际的流量搜索
	return map[string]interface{}{
		"count": 0,
		"items": []interface{}{},
	}, nil
}

// getSitemap 获取网站地图
func getSitemap(projectID, host string) (interface{}, error) {
	// TODO: 实现实际的网站地图获取
	return map[string]interface{}{
		"urls": []interface{}{},
	}, nil
}

// getStatistics 获取统计信息
func getStatistics(projectID string) (interface{}, error) {
	// TODO: 实现实际的统计信息获取
	return map[string]interface{}{
		"traffic_count":       0,
		"vulnerability_count": 0,
		"host_count":          0,
	}, nil
}

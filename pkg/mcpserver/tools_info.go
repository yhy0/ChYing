package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/db"
)

// --- get_hosts ---

func getHostsTool() mcp.Tool {
	return mcp.NewTool("get_hosts",
		mcp.WithDescription("Get all unique hostnames from the HTTP traffic history."),
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
	)
}

func handleGetHosts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := req.GetString("session_id", "")
	hosts := db.GetHostsBySession(sessionID)
	return jsonResult(hosts), nil
}

// --- get_statistics ---

func getStatisticsTool() mcp.Tool {
	return mcp.NewTool("get_statistics",
		mcp.WithDescription("Get project statistics including traffic count, host count, and vulnerability breakdown by level and type."),
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
	)
}

func handleGetStatistics(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := req.GetString("session_id", "")
	hosts := db.GetHostsBySession(sessionID)

	// 获取流量总数（使用 COUNT 查询避免加载全部记录到内存）
	var trafficCount int64
	if db.GlobalDB != nil {
		tq := db.GlobalDB.Model(&db.HTTPHistory{}).Where("project_id = ?", db.CurrentProjectName)
		if sessionID != "" {
			tq = tq.Where("session_id = ?", sessionID)
		}
		tq.Count(&trafficCount)
	}

	// 获取漏洞统计
	vulnStats, _ := db.GetVulnerabilityStatistics(db.CurrentProjectName)

	stats := map[string]interface{}{
		"project_name":    db.CurrentProjectName,
		"traffic_count":   trafficCount,
		"host_count":      len(hosts),
		"hosts":           hosts,
		"vulnerabilities": vulnStats,
	}

	return jsonResult(stats), nil
}

// --- get_current_project ---

func getCurrentProjectTool() mcp.Tool {
	return mcp.NewTool("get_current_project",
		mcp.WithDescription("Get current project information including project name, proxy configuration, and MCP server status."),
	)
}

func handleGetCurrentProject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectInfo := map[string]interface{}{
		"project_name":  db.CurrentProjectName,
		"proxy_host":    conf.AppConf.Proxy.Host,
		"proxy_port":    conf.AppConf.Proxy.Port,
		"proxy_enabled": conf.AppConf.Proxy.Enabled,
		"mcp_port":      conf.AppConf.MCPPort,
		"version":       conf.Version,
	}

	return jsonResult(projectInfo), nil
}

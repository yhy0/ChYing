package mcpserver

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/db"
)

var startTime = time.Now()

// --- get_scan_status ---

func getScanStatusTool() mcp.Tool {
	return mcp.NewTool("get_scan_status",
		mcp.WithDescription("Get current passive scan status including proxy state, request counts, and active sessions."),
		mcp.WithString("session_id",
			mcp.Description("Optional: filter status by session ID"),
		),
	)
}

func handleGetScanStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := req.GetString("session_id", "")

	var trafficCount, vulnCount int64
	if db.GlobalDB != nil {
		tq := db.GlobalDB.Model(&db.HTTPHistory{})
		vq := db.GlobalDB.Model(&db.Vulnerability{})
		if sessionID != "" {
			tq = tq.Where("session_id = ?", sessionID)
			vq = vq.Where("session_id = ?", sessionID)
		}
		tq.Count(&trafficCount)
		vq.Count(&vulnCount)
	}

	sessions := ListSessions()

	status := map[string]interface{}{
		"proxy_running":         true,
		"proxy_host":            conf.AppConf.Proxy.Host,
		"proxy_port":            conf.AppConf.Proxy.Port,
		"total_requests":        trafficCount,
		"total_vulnerabilities": vulnCount,
		"active_sessions":       len(sessions),
		"uptime":                time.Since(startTime).Truncate(time.Second).String(),
	}
	return jsonResult(status), nil
}

// --- get_new_findings_since ---

func getNewFindingsSinceTool() mcp.Tool {
	return mcp.NewTool("get_new_findings_since",
		mcp.WithDescription("Get new vulnerabilities and traffic discovered after a given timestamp. Use for incremental polling."),
		mcp.WithString("since",
			mcp.Required(),
			mcp.Description("ISO 8601 timestamp (e.g. 2026-03-31T14:00:00Z). Returns findings after this time."),
		),
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by session ID"),
		),
		mcp.WithString("type",
			mcp.Description("What to return: 'vulnerabilities', 'traffic', or 'all' (default: 'all')"),
		),
	)
}

func handleGetNewFindingsSince(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sinceStr, err := req.RequireString("since")
	if err != nil {
		return errorResult("since is required (ISO 8601 format)"), nil
	}

	since, parseErr := time.Parse(time.RFC3339, sinceStr)
	if parseErr != nil {
		since, parseErr = time.Parse("2006-01-02T15:04:05Z07:00", sinceStr)
		if parseErr != nil {
			return errorResult("invalid timestamp format, use ISO 8601 (e.g. 2026-03-31T14:00:00Z)"), nil
		}
	}

	sessionID := req.GetString("session_id", "")
	findingType := req.GetString("type", "all")
	queryTime := time.Now()

	result := map[string]interface{}{
		"query_time": queryTime.Format(time.RFC3339),
	}

	if findingType == "all" || findingType == "vulnerabilities" {
		vulns, _ := db.GetNewVulnerabilitiesSince(since, sessionID)
		result["vulnerabilities"] = vulns
		result["vuln_count"] = len(vulns)
	}

	if findingType == "all" || findingType == "traffic" {
		histories, _ := db.GetNewHistorySince(since, sessionID)
		result["traffic_count"] = len(histories)
		hostSet := make(map[string]bool)
		for _, h := range histories {
			hostSet[h.Host] = true
		}
		var newHosts []string
		for h := range hostSet {
			newHosts = append(newHosts, h)
		}
		result["new_hosts"] = newHosts
	}

	return jsonResult(result), nil
}

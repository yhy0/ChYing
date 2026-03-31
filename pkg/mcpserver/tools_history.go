package mcpserver

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/pkg/db"
)

// historyItem 流量历史简要信息
type historyItem struct {
	ID        int64  `json:"id"`
	Hid       int64  `json:"hid"`
	Host      string `json:"host"`
	Method    string `json:"method"`
	FullUrl   string `json:"full_url"`
	Path      string `json:"path"`
	Status    string `json:"status"`
	Length    string `json:"length"`
	MIMEType  string `json:"mime_type"`
	Extension string `json:"extension"`
	Title     string `json:"title"`
	IP        string `json:"ip"`
}

func convertHistory(h *db.HTTPHistory) historyItem {
	return historyItem{
		ID:        h.ID,
		Hid:       h.Hid,
		Host:      h.Host,
		Method:    h.Method,
		FullUrl:   h.FullUrl,
		Path:      h.Path,
		Status:    h.Status,
		Length:    h.Length,
		MIMEType:  h.MIMEType,
		Extension: h.Extension,
		Title:     h.Title,
		IP:        h.IP,
	}
}

// --- get_http_history ---

func getHttpHistoryTool() mcp.Tool {
	return mcp.NewTool("get_http_history",
		mcp.WithDescription("Get HTTP traffic history with pagination. Returns a list of HTTP requests captured by the proxy."),
		mcp.WithString("source",
			mcp.Description("Filter by source: 'local' (proxy captured), 'remote' (remote node), or 'all' (default)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of records to return (default 50, max 500)"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of records to skip (default 0)"),
		),
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
	)
}

func handleGetHttpHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := req.GetString("source", "all")
	limit := req.GetInt("limit", 50)
	offset := req.GetInt("offset", 0)
	sessionID := req.GetString("session_id", "")

	if limit > 500 {
		limit = 500
	}
	if limit <= 0 {
		limit = 50
	}

	if source == "all" {
		source = ""
	}

	histories, err := db.GetAllHistory(db.CurrentProjectName, source, limit, offset, sessionID)
	if err != nil {
		return errorResult("failed to get history: %v", err), nil
	}

	items := make([]historyItem, 0, len(histories))
	for _, h := range histories {
		items = append(items, convertHistory(h))
	}

	return jsonResult(items), nil
}

// --- get_traffic_by_host ---

var defaultExcludeExtensions = []string{
	"js", "css", "png", "jpg", "gif", "svg", "ico",
	"woff", "woff2", "ttf", "eot", "mp4", "mp3",
}

func getTrafficByHostTool() mcp.Tool {
	return mcp.NewTool("get_traffic_by_host",
		mcp.WithDescription("Get HTTP traffic filtered by host. By default excludes static resources (js, css, images, fonts, media)."),
		mcp.WithString("host",
			mcp.Required(),
			mcp.Description("The hostname to filter traffic for (e.g., 'example.com')"),
		),
		mcp.WithString("exclude_extensions",
			mcp.Description("Comma-separated list of file extensions to exclude (e.g., 'js,css,png'). Set to 'none' to include all resources. Default excludes common static resources."),
		),
	)
}

func handleGetTrafficByHost(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	host, err := req.RequireString("host")
	if err != nil {
		return errorResult("host is required"), nil
	}

	histories := db.GetHistory([]string{host})

	// 构建排除扩展名集合
	excludeExts := make(map[string]bool)
	excludeStr := req.GetString("exclude_extensions", "")
	if excludeStr == "" {
		// 未指定时使用默认排除列表
		for _, ext := range defaultExcludeExtensions {
			excludeExts[ext] = true
		}
	} else if excludeStr != "none" {
		// 用户自定义排除列表
		for _, part := range strings.Split(excludeStr, ",") {
			ext := strings.TrimSpace(part)
			if ext != "" {
				excludeExts[ext] = true
			}
		}
	}
	// excludeStr == "none" 时不排除任何扩展名

	items := make([]historyItem, 0)
	for _, h := range histories {
		if len(excludeExts) > 0 && excludeExts[h.Extension] {
			continue
		}
		items = append(items, convertHistory(h))
	}

	return jsonResult(items), nil
}

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/pkg/db"
)

// --- get_vulnerabilities ---

func getVulnerabilitiesTool() mcp.Tool {
	return mcp.NewTool("get_vulnerabilities",
		mcp.WithDescription("Get discovered vulnerabilities list with pagination. Returns vulnerability details including type, severity, target, and description."),
		mcp.WithString("source",
			mcp.Description("Filter by source: 'local', 'remote', or 'all' (default)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of records to return (default 100, max 500)"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of records to skip (default 0)"),
		),
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
	)
}

func handleGetVulnerabilities(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := req.GetString("source", "all")
	limit := req.GetInt("limit", 100)
	offset := req.GetInt("offset", 0)
	sessionID := req.GetString("session_id", "")

	if limit > 500 {
		limit = 500
	}
	if limit <= 0 {
		limit = 100
	}

	if source == "all" {
		source = ""
	}

	vulns, err := db.GetAllVulnerabilities(db.CurrentProjectName, source, limit, offset, sessionID)
	if err != nil {
		return errorResult("failed to get vulnerabilities: %v", err), nil
	}

	type vulnItem struct {
		ID          int64  `json:"id"`
		VulnID      string `json:"vuln_id"`
		VulnType    string `json:"vuln_type"`
		Target      string `json:"target"`
		Host        string `json:"host"`
		Method      string `json:"method"`
		Path        string `json:"path"`
		Plugin      string `json:"plugin"`
		Level       string `json:"level"`
		Param       string `json:"param"`
		Payload     string `json:"payload"`
		Description string `json:"description"`
		CurlCommand string `json:"curl_command"`
		Request     string `json:"request"`
		Response    string `json:"response"`
		CreatedAt   string `json:"created_at"`
	}

	items := make([]vulnItem, 0, len(vulns))
	for _, v := range vulns {
		items = append(items, vulnItem{
			ID:          v.ID,
			VulnID:      v.VulnID,
			VulnType:    v.VulnType,
			Target:      v.Target,
			Host:        v.Host,
			Method:      v.Method,
			Path:        v.Path,
			Plugin:      v.Plugin,
			Level:       v.Level,
			Param:       v.Param,
			Payload:     v.Payload,
			Description: v.Description,
			CurlCommand: v.CurlCommand,
			Request:     v.Request,
			Response:    v.Response,
			CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return jsonResult(items), nil
}

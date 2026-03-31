package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/pkg/db"
)

// --- register_session ---

func registerSessionTool() mcp.Tool {
	return mcp.NewTool("register_session",
		mcp.WithDescription("Register a new scan session for traffic isolation. Returns a session_id that should be sent as X-ChYing-Session header in proxied requests."),
		mcp.WithString("targets",
			mcp.Required(),
			mcp.Description("JSON array of target hostnames, e.g. [\"target.com\", \"*.target.com\"]"),
		),
		mcp.WithString("description",
			mcp.Description("Optional description for this session"),
		),
	)
}

func handleRegisterSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetsStr, err := req.RequireString("targets")
	if err != nil {
		return errorResult("targets is required"), nil
	}

	var targets []string
	if jsonErr := json.Unmarshal([]byte(targetsStr), &targets); jsonErr != nil {
		targets = []string{targetsStr}
	}

	if len(targets) == 0 {
		return errorResult("at least one target is required"), nil
	}

	description := req.GetString("description", "")
	session := RegisterSession(targets, description)
	return jsonResult(session), nil
}

// --- configure_session ---

func configureSessionTool() mcp.Tool {
	return mcp.NewTool("configure_session",
		mcp.WithDescription("Modify an existing session's target list."),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID to configure"),
		),
		mcp.WithString("add_targets",
			mcp.Description("JSON array of hostnames to add"),
		),
		mcp.WithString("remove_targets",
			mcp.Description("JSON array of hostnames to remove"),
		),
	)
}

func handleConfigureSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := req.RequireString("session_id")
	if err != nil {
		return errorResult("session_id is required"), nil
	}

	var addTargets, removeTargets []string

	if addStr := req.GetString("add_targets", ""); addStr != "" {
		json.Unmarshal([]byte(addStr), &addTargets)
	}
	if removeStr := req.GetString("remove_targets", ""); removeStr != "" {
		json.Unmarshal([]byte(removeStr), &removeTargets)
	}

	session, ok := ConfigureSession(sessionID, addTargets, removeTargets)
	if !ok {
		return errorResult("session not found: %s", sessionID), nil
	}
	return jsonResult(session), nil
}

// --- close_session ---

func closeSessionTool() mcp.Tool {
	return mcp.NewTool("close_session",
		mcp.WithDescription("Close a scan session and return a summary of findings."),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID to close"),
		),
	)
}

func handleCloseSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := req.RequireString("session_id")
	if err != nil {
		return errorResult("session_id is required"), nil
	}

	if !CloseSession(sessionID) {
		return errorResult("session not found: %s", sessionID), nil
	}

	var trafficCount, vulnCount int64
	if db.GlobalDB != nil {
		db.GlobalDB.Model(&db.HTTPHistory{}).Where("session_id = ?", sessionID).Count(&trafficCount)
		db.GlobalDB.Model(&db.Vulnerability{}).Where("session_id = ?", sessionID).Count(&vulnCount)
	}

	summary := map[string]interface{}{
		"closed":         true,
		"session_id":     sessionID,
		"total_requests": trafficCount,
		"total_vulns":    vulnCount,
	}
	return jsonResult(summary), nil
}

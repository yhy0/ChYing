package mcpserver

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/db"
)

var repeaterTitleRe = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)

type historyReference struct {
	ID        int64  `json:"id"`
	Hid       int64  `json:"hid"`
	ProjectID string `json:"project_id"`
	SessionID string `json:"session_id,omitempty"`
}

func sessionAttribution(sessionID string) (string, string, error) {
	projectID := db.CurrentProjectName
	if projectID == "" {
		projectID = "default"
	}
	if sessionID == "" {
		return projectID, "", nil
	}
	session, ok := GetSession(sessionID)
	if !ok || !session.Active {
		return "", "", fmt.Errorf("session not found or inactive: %s", sessionID)
	}
	if session.ProjectID != projectID {
		return "", "", fmt.Errorf("session belongs to project %s, current project is %s", session.ProjectID, projectID)
	}
	return projectID, sessionID, nil
}

// --- send_request ---

func sendRequestTool() mcp.Tool {
	return mcp.NewTool("send_request",
		mcp.WithDescription(`Send a raw HTTP request (Repeater). Useful for testing and verifying vulnerabilities.

Example raw_request:
GET /api/users HTTP/1.1
Host: example.com
Cookie: session=abc123

Example with POST body:
POST /api/login HTTP/1.1
Host: example.com
Content-Type: application/json

{"username":"admin","password":"test"}`),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target URL including scheme (e.g., 'https://example.com')"),
		),
		mcp.WithString("raw_request",
			mcp.Required(),
			mcp.Description("Raw HTTP request text (headers and optional body)"),
		),
		mcp.WithString("session_id",
			mcp.Description("Optional scan session ID used to attribute this request to an agent task"),
		),
	)
}

func handleSendRequest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, err := req.RequireString("target")
	if err != nil {
		return errorResult("target is required"), nil
	}

	rawRequest, err := req.RequireString("raw_request")
	if err != nil {
		return errorResult("raw_request is required"), nil
	}

	projectID, sessionID, attributionErr := sessionAttribution(req.GetString("session_id", ""))
	if attributionErr != nil {
		return errorResult("%v", attributionErr), nil
	}

	resp, reqErr := httpx.Raw(rawRequest, target)
	if reqErr != nil {
		return errorResult("request failed: %v", reqErr), nil
	}

	// 写入 ChYing history，使 Repeater 请求在"代理拦截"界面可见。
	// send_request 不经过 9080 代理，默认不入 history；这里补登记，让所有流量统一在 ChYing 查看。
	reference := saveRepeaterHistory(target, rawRequest, resp, projectID, sessionID)
	return jsonResult(map[string]interface{}{
		"project_id": projectID, "session_id": sessionID,
		"history_id": reference.ID, "hid": reference.Hid,
		"status": resp.Status, "status_code": resp.StatusCode,
		"content_length": resp.ContentLength, "time_ms": resp.ServerDurationMs,
		"request": resp.RequestDump, "response": resp.ResponseDump,
	}), nil
}

// saveRepeaterHistory 把 send_request 的请求/响应写入 db history + traffic，标记 source=repeater。
func saveRepeaterHistory(target string, rawRequest string, resp *httpx.Response, projectID, sessionID string) historyReference {
	if resp == nil {
		return historyReference{ProjectID: projectID, SessionID: sessionID}
	}
	u, err := url.Parse(target)
	if err != nil {
		return historyReference{ProjectID: projectID, SessionID: sessionID}
	}
	host := u.Host
	path := u.Path
	method := "GET"
	if firstLine := strings.SplitN(rawRequest, "\n", 2); len(firstLine) > 0 {
		if parts := strings.Fields(strings.TrimSpace(firstLine[0])); len(parts) > 0 {
			method = strings.ToUpper(parts[0])
			if len(parts) > 1 && strings.HasPrefix(parts[1], "/") {
				path = parts[1]
			}
		}
	}
	if path == "" {
		path = "/"
	}
	title := ""
	if m := repeaterTitleRe.FindStringSubmatch(resp.Body); m != nil {
		title = strings.TrimSpace(m[1])
	}
	hid := time.Now().UnixNano()
	ct := resp.RespHeader.Get("Content-Type")
	now := time.Now().Format("2006-01-02 15:04:05")

	history := &db.HTTPHistory{
		Hid:         hid,
		Host:        host,
		Method:      method,
		FullUrl:     target,
		Path:        path,
		Status:      strconv.Itoa(resp.StatusCode),
		Length:      strconv.Itoa(resp.ContentLength),
		ContentType: ct,
		MIMEType:    ct,
		Title:       title,
		IP:          u.Hostname(),
		Source:      "repeater",
		SourceID:    "mcp",
		NodeName:    "MCP",
		ProjectID:   projectID,
		SessionID:   sessionID,
		Time:        now,
	}
	db.AddHistory(history)

	db.AddRequest(&db.Request{
		RequestId:  uint(hid),
		Url:        target,
		Host:       host,
		Path:       path,
		RequestRaw: resp.RequestDump,
	}, &db.Response{
		RequestId:   uint(hid),
		Url:         target,
		Host:        host,
		Path:        path,
		ResponseRaw: resp.ResponseDump,
		ContentType: ct,
	})
	return historyReference{ID: history.ID, Hid: hid, ProjectID: projectID, SessionID: sessionID}
}

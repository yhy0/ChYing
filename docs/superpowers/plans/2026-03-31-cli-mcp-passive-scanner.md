# ChYing CLI + MCP 被动扫描服务 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a standalone CLI entry point that runs ChYing's proxy + passive scanner + MCP server without the Wails GUI, with session-based isolation for multi-Agent concurrency.

**Architecture:** New `cmd/chying-cli/` entry point reuses existing packages (`mitmproxy/`, `pkg/db/`, `pkg/mcpserver/`). Session isolation via `X-ChYing-Session` header extracted in proxy callbacks. DB models gain a `session_id` field; MCP tools gain an optional `session_id` filter parameter.

**Tech Stack:** Go 1.25, cobra (CLI), mcp-go (MCP SSE), GORM/SQLite, proxify/martian (proxy)

---

## File Structure

| Operation | File | Responsibility |
|-----------|------|----------------|
| Create | `cmd/chying-cli/main.go` | Cobra root command + version |
| Create | `cmd/chying-cli/serve.go` | `serve` subcommand: init all services, block until signal |
| Create | `cmd/chying-cli/output.go` | CLI terminal output formatting (vuln, traffic, status) |
| Create | `pkg/mcpserver/session.go` | Session CRUD (register, get, list, configure, close) in memory |
| Create | `pkg/mcpserver/tools_session.go` | MCP tools: register_session, configure_session, close_session |
| Create | `pkg/mcpserver/tools_realtime.go` | MCP tools: get_scan_status, get_new_findings_since |
| Create | `Dockerfile.cli` | Multi-stage Docker build for CLI binary |
| Create | `docker-compose.cli.yml` | Docker Compose with chying + agent example |
| Modify | `pkg/db/history.go` | Add `SessionID` field to `HTTPHistory` model |
| Modify | `pkg/db/vulnerability.go` | Add `SessionID` field to `Vulnerability` model |
| Modify | `mitmproxy/type.go` | Add `SessionID` field to `mitmproxy.HTTPHistory` DTO |
| Modify | `mitmproxy/proxify.go` | Extract+strip `X-ChYing-Session` header in `onRequestCallback`; pass session_id through to event |
| Modify | `main.go` | Pass `SessionID` from mitmproxy event to `db.HTTPHistory` in `EventNotification()` |
| Modify | `app_utils.go` | Pass `SessionID` from vuln channel to `db.Vulnerability` in `startEventLoop()` |
| Modify | `pkg/mcpserver/server.go` | Register new tools; `StartHTTPServer` accepts bind address |
| Modify | `pkg/mcpserver/tools_history.go` | Add optional `session_id` param to history query tools |
| Modify | `pkg/mcpserver/tools_vuln.go` | Add optional `session_id` param to vuln query tool |
| Modify | `pkg/mcpserver/tools_traffic.go` | Add optional `session_id` param to traffic query tools |
| Modify | `pkg/mcpserver/tools_info.go` | Add optional `session_id` param to statistics/hosts tools |

---

### Task 1: Add cobra dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add cobra dependency**

```bash
cd /Users/yhy/Documents/Github/ChYing && go get github.com/spf13/cobra@latest
```

- [ ] **Step 2: Verify it resolved**

```bash
cd /Users/yhy/Documents/Github/ChYing && grep cobra go.mod
```

Expected: a line like `github.com/spf13/cobra v1.x.x`

- [ ] **Step 3: Tidy**

```bash
cd /Users/yhy/Documents/Github/ChYing && go mod tidy
```

- [ ] **Step 4: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add go.mod go.sum && git commit -m "deps: add spf13/cobra for CLI entry point"
```

---

### Task 2: Add `SessionID` field to database models

**Files:**
- Modify: `pkg/db/history.go`
- Modify: `pkg/db/vulnerability.go`

- [ ] **Step 1: Add `SessionID` to `HTTPHistory` model**

In `pkg/db/history.go`, add a new field to the `HTTPHistory` struct, after the `ProjectID` field:

```go
	// Session 标识字段
	SessionID string `gorm:"index;default:''" json:"session_id"` // 扫描会话ID，用于多Agent隔离
```

- [ ] **Step 2: Add `SessionID` to `Vulnerability` model**

In `pkg/db/vulnerability.go`, add a new field to the `Vulnerability` struct, after the `ProjectID` field:

```go
	// Session 标识字段
	SessionID string `gorm:"index;default:''" json:"session_id"` // 扫描会话ID，用于多Agent隔离
```

- [ ] **Step 3: Update `GetAllHistory` to support session filtering**

In `pkg/db/history.go`, add session_id filtering to `GetAllHistory`. After the existing `source` filter block, add:

```go
	// 添加 session_id 过滤
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
```

The full updated function signature becomes:

```go
func GetAllHistory(projectID string, source string, limit, offset int, sessionID ...string) ([]*HTTPHistory, error) {
```

And at the start of the function, extract the optional param:

```go
	var sid string
	if len(sessionID) > 0 {
		sid = sessionID[0]
	}
```

Then use `sid` in the filter. This variadic approach ensures all existing callers (which pass 4 args) continue to work without changes.

- [ ] **Step 4: Update `GetAllVulnerabilities` to support session filtering**

In `pkg/db/vulnerability.go`, same pattern. Updated signature:

```go
func GetAllVulnerabilities(projectID string, source string, limit, offset int, sessionID ...string) ([]*Vulnerability, error) {
```

Add at the start:

```go
	var sid string
	if len(sessionID) > 0 {
		sid = sessionID[0]
	}
```

Add after the existing `source` filter:

```go
	// 添加 session_id 过滤
	if sid != "" {
		query = query.Where("session_id = ?", sid)
	}
```

- [ ] **Step 5: Add `GetNewVulnerabilitiesSince` query method**

Add to `pkg/db/vulnerability.go`:

```go
// GetNewVulnerabilitiesSince 获取指定时间之后的新漏洞
func GetNewVulnerabilitiesSince(since time.Time, sessionID string) ([]*Vulnerability, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var data []*Vulnerability
	query := GlobalDB.Model(&Vulnerability{}).Where("created_at > ?", since)
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	query.Order("created_at ASC").Find(&data)
	return data, nil
}
```

- [ ] **Step 6: Add `GetNewHistorySince` query method**

Add to `pkg/db/history.go`:

```go
// GetNewHistorySince 获取指定时间之后的新历史记录
func GetNewHistorySince(since time.Time, sessionID string) ([]*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var data []*HTTPHistory
	query := GlobalDB.Model(&HTTPHistory{}).Where("created_at > ?", since)
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	query.Order("created_at ASC").Find(&data)
	return data, nil
}
```

- [ ] **Step 7: Add `GetHostsBySession` query method**

Add to `pkg/db/history.go`:

```go
// GetHostsBySession 获取指定 session 的所有域名
func GetHostsBySession(sessionID string) []string {
	var hosts []string
	if GlobalDB == nil {
		return hosts
	}
	query := GlobalDB.Model(&HTTPHistory{}).Select("DISTINCT host")
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	var histories []*HTTPHistory
	query.Find(&histories)
	for _, h := range histories {
		hosts = append(hosts, h.Host)
	}
	return hosts
}
```

- [ ] **Step 8: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./pkg/db/
```

Expected: no errors

- [ ] **Step 9: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add pkg/db/history.go pkg/db/vulnerability.go && git commit -m "feat(db): add session_id field and session-aware query methods"
```

---

### Task 3: Add `SessionID` to mitmproxy DTO and proxy callbacks

**Files:**
- Modify: `mitmproxy/type.go`
- Modify: `mitmproxy/proxify.go`

- [ ] **Step 1: Add `SessionID` field to `mitmproxy.HTTPHistory`**

In `mitmproxy/type.go`, add to the `HTTPHistory` struct after the `Color` field:

```go
	SessionID string `json:"session_id,omitempty"`
```

- [ ] **Step 2: Extract and strip `X-ChYing-Session` in `onRequestCallback`**

In `mitmproxy/proxify.go`, in `onRequestCallback`, after the flowID assignment (line ~415) and before the body read block, add:

```go
	// 提取 session ID 并从请求中剥离（不转发给目标）
	sessionID := req.Header.Get("X-ChYing-Session")
	if sessionID != "" {
		req.Header.Del("X-ChYing-Session")
	}
```

Then when creating `tempEntry` (the `HTTPHistory` struct around line ~469), add the `SessionID` field:

```go
	tempEntry := &HTTPHistory{
		Id:        tempHistoryID,
		FlowID:    flowID,
		Host:      req.URL.Host,
		Method:    req.Method,
		FullUrl:   req.URL.String(),
		Path:      req.URL.RequestURI(),
		Timestamp: time.Now().Format(time.RFC3339Nano),
		SessionID: sessionID, // 新增
	}
```

- [ ] **Step 3: Propagate `SessionID` in `onResponseCallback`**

In `mitmproxy/proxify.go`, in `onResponseCallback`, when building `historyForEvent` (around line ~633), add the `SessionID` field:

```go
	historyForEvent := &HTTPHistory{
		// ... existing fields ...
		SessionID:         tempEntry.SessionID, // 新增：从请求阶段传递
	}
```

Find the block that creates `historyForEvent` and add `SessionID: tempEntry.SessionID,` after the `ResponseTimestamp` line.

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./mitmproxy/
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add mitmproxy/type.go mitmproxy/proxify.go && git commit -m "feat(proxy): extract X-ChYing-Session header and propagate session_id"
```

---

### Task 4: Pass `SessionID` through event pipeline to DB

**Files:**
- Modify: `main.go` (the `EventNotification` function)
- Modify: `app_utils.go` (the `startEventLoop` function)

- [ ] **Step 1: Pass `SessionID` in `EventNotification`**

In `main.go`, in the `EventNotification()` function, when creating `historyData` (around line ~141), add:

```go
					SessionID: _http.SessionID, // 新增：传递 session 标识
```

Add this line after the `NodeName` field assignment in the `historyData` struct literal.

- [ ] **Step 2: Pass `SessionID` in `startEventLoop` for vulnerabilities**

In `app_utils.go`, in the `startEventLoop()` function, when creating `vulnData` (around line ~206), the vulnerability needs to inherit the session from the traffic that triggered it. For now, use the project default. We will address per-vuln session tracking in Task 6 (passive scan plugin).

This step is a placeholder — the actual session propagation for vulns happens through a different channel (see Task 6).

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build .
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add main.go app_utils.go && git commit -m "feat: propagate session_id from proxy events to database"
```

---

### Task 5: Session manager

**Files:**
- Create: `pkg/mcpserver/session.go`

- [ ] **Step 1: Create session manager**

Create `pkg/mcpserver/session.go`:

```go
package mcpserver

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ScanSession 扫描会话
type ScanSession struct {
	SessionID   string    `json:"session_id"`
	Targets     []string  `json:"targets"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `json:"active"`
}

// sessionStore 内存中的 session 存储
var sessionStore sync.Map

// RegisterSession 注册新的扫描会话
func RegisterSession(targets []string, description string) *ScanSession {
	session := &ScanSession{
		SessionID:   uuid.New().String(),
		Targets:     targets,
		Description: description,
		CreatedAt:   time.Now(),
		Active:      true,
	}
	sessionStore.Store(session.SessionID, session)
	return session
}

// GetSession 获取指定会话
func GetSession(sessionID string) (*ScanSession, bool) {
	val, ok := sessionStore.Load(sessionID)
	if !ok {
		return nil, false
	}
	return val.(*ScanSession), true
}

// ListSessions 列出所有活跃会话
func ListSessions() []*ScanSession {
	var sessions []*ScanSession
	sessionStore.Range(func(key, value interface{}) bool {
		s := value.(*ScanSession)
		if s.Active {
			sessions = append(sessions, s)
		}
		return true
	})
	return sessions
}

// ConfigureSession 修改会话的目标列表
func ConfigureSession(sessionID string, addTargets, removeTargets []string) (*ScanSession, bool) {
	val, ok := sessionStore.Load(sessionID)
	if !ok {
		return nil, false
	}
	session := val.(*ScanSession)

	// 移除指定目标
	if len(removeTargets) > 0 {
		removeSet := make(map[string]bool)
		for _, t := range removeTargets {
			removeSet[t] = true
		}
		var filtered []string
		for _, t := range session.Targets {
			if !removeSet[t] {
				filtered = append(filtered, t)
			}
		}
		session.Targets = filtered
	}

	// 添加新目标
	if len(addTargets) > 0 {
		existing := make(map[string]bool)
		for _, t := range session.Targets {
			existing[t] = true
		}
		for _, t := range addTargets {
			if !existing[t] {
				session.Targets = append(session.Targets, t)
			}
		}
	}

	sessionStore.Store(sessionID, session)
	return session, true
}

// CloseSession 关闭会话
func CloseSession(sessionID string) bool {
	val, ok := sessionStore.Load(sessionID)
	if !ok {
		return false
	}
	session := val.(*ScanSession)
	session.Active = false
	sessionStore.Store(sessionID, session)
	return true
}
```

- [ ] **Step 2: Add uuid dependency**

```bash
cd /Users/yhy/Documents/Github/ChYing && go get github.com/google/uuid@latest && go mod tidy
```

Note: Check if `uuid` is already in go.mod (the project uses `github.com/satori/go.uuid`). If so, you can use that instead. Replace `uuid.New().String()` with `uuid.NewV4().String()` and the import with `uuid "github.com/satori/go.uuid"`.

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./pkg/mcpserver/
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add pkg/mcpserver/session.go go.mod go.sum && git commit -m "feat(mcp): add session manager for multi-Agent isolation"
```

---

### Task 6: Session MCP tools

**Files:**
- Create: `pkg/mcpserver/tools_session.go`

- [ ] **Step 1: Create session tools**

Create `pkg/mcpserver/tools_session.go`:

```go
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
		// 如果不是 JSON 数组，当作单个目标处理
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

	// 获取该 session 的统计信息
	var trafficCount int64
	if db.GlobalDB != nil {
		db.GlobalDB.Model(&db.HTTPHistory{}).Where("session_id = ?", sessionID).Count(&trafficCount)
	}
	var vulnCount int64
	if db.GlobalDB != nil {
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./pkg/mcpserver/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add pkg/mcpserver/tools_session.go && git commit -m "feat(mcp): add session management tools (register, configure, close)"
```

---

### Task 7: Realtime MCP tools (scan status + incremental findings)

**Files:**
- Create: `pkg/mcpserver/tools_realtime.go`

- [ ] **Step 1: Create realtime tools**

Create `pkg/mcpserver/tools_realtime.go`:

```go
package mcpserver

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/db"
)

// startTime 记录服务启动时间
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

	var trafficCount int64
	var vulnCount int64

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
		// 尝试其他格式
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

		// 提取新发现的域名
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./pkg/mcpserver/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add pkg/mcpserver/tools_realtime.go && git commit -m "feat(mcp): add scan status and incremental findings tools"
```

---

### Task 8: Add `session_id` filter to existing MCP tools

**Files:**
- Modify: `pkg/mcpserver/tools_history.go`
- Modify: `pkg/mcpserver/tools_vuln.go`
- Modify: `pkg/mcpserver/tools_traffic.go`
- Modify: `pkg/mcpserver/tools_info.go`

- [ ] **Step 1: Update `get_http_history` tool**

In `pkg/mcpserver/tools_history.go`:

Add `session_id` parameter to `getHttpHistoryTool()`:

```go
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
```

In `handleGetHttpHistory`, extract and pass it:

```go
	sessionID := req.GetString("session_id", "")
```

Update the call to `db.GetAllHistory` to pass sessionID:

```go
	history, err := db.GetAllHistory(db.CurrentProjectName, source, limit, offset, sessionID)
```

- [ ] **Step 2: Update `get_vulnerabilities` tool**

In `pkg/mcpserver/tools_vuln.go`:

Add `session_id` parameter to `getVulnerabilitiesTool()`:

```go
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
```

In `handleGetVulnerabilities`, extract and pass it:

```go
	sessionID := req.GetString("session_id", "")
	vulns, err := db.GetAllVulnerabilities(db.CurrentProjectName, source, limit, offset, sessionID)
```

- [ ] **Step 3: Update `get_hosts` tool**

In `pkg/mcpserver/tools_info.go`:

Add `session_id` parameter to `getHostsTool()`:

```go
	return mcp.NewTool("get_hosts",
		mcp.WithDescription("Get all unique hostnames from the HTTP traffic history."),
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
	)
```

Update `handleGetHosts`:

```go
func handleGetHosts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := req.GetString("session_id", "")
	hosts := db.GetHostsBySession(sessionID)
	return jsonResult(hosts), nil
}
```

- [ ] **Step 4: Update `get_statistics` tool**

In `pkg/mcpserver/tools_info.go`:

Add `session_id` parameter to `getStatisticsTool()`:

```go
		mcp.WithString("session_id",
			mcp.Description("Optional: filter by scan session ID"),
		),
```

In `handleGetStatistics`, extract and use it:

```go
	sessionID := req.GetString("session_id", "")
	hosts := db.GetHostsBySession(sessionID)

	var trafficCount int64
	if db.GlobalDB != nil {
		tq := db.GlobalDB.Model(&db.HTTPHistory{}).Where("project_id = ?", db.CurrentProjectName)
		if sessionID != "" {
			tq = tq.Where("session_id = ?", sessionID)
		}
		tq.Count(&trafficCount)
	}
```

- [ ] **Step 5: Update `get_traffic_by_host` and `query_by_dsl`**

In `pkg/mcpserver/tools_traffic.go`:

Add `session_id` parameter to both tools. For `get_traffic_by_host`, the `db.GetHistory` function operates by host and does not currently support session filtering. For now, add the parameter definition but note this is a future enhancement — the host filter already provides reasonable isolation.

For `query_by_dsl`, the DSL system needs separate work to support session filtering. Add the parameter for API consistency, but document it as not yet functional for DSL queries.

- [ ] **Step 6: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./pkg/mcpserver/
```

Expected: no errors

- [ ] **Step 7: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add pkg/mcpserver/tools_history.go pkg/mcpserver/tools_vuln.go pkg/mcpserver/tools_traffic.go pkg/mcpserver/tools_info.go && git commit -m "feat(mcp): add session_id filter to all existing query tools"
```

---

### Task 9: Register new tools in MCP server + configurable bind address

**Files:**
- Modify: `pkg/mcpserver/server.go`

- [ ] **Step 1: Register new tools**

In `pkg/mcpserver/server.go`, add the new tool registrations in `NewChYingMCPServer()`:

```go
	// Session 管理工具
	s.AddTool(registerSessionTool(), handleRegisterSession)
	s.AddTool(configureSessionTool(), handleConfigureSession)
	s.AddTool(closeSessionTool(), handleCloseSession)

	// 实时状态工具
	s.AddTool(getScanStatusTool(), handleGetScanStatus)
	s.AddTool(getNewFindingsSinceTool(), handleGetNewFindingsSince)
```

- [ ] **Step 2: Add bind address parameter to `StartHTTPServer`**

Update the function signature:

```go
func StartHTTPServer(port int, bindAddr ...string) (string, error) {
	s := NewChYingMCPServer()
	httpServer := server.NewStreamableHTTPServer(s)

	host := "127.0.0.1"
	if len(bindAddr) > 0 && bindAddr[0] != "" {
		host = bindAddr[0]
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	logging.Logger.Infof("MCP server listening on %s", addr)
```

This variadic approach keeps all existing callers working (they pass only `port`).

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./pkg/mcpserver/ && go build .
```

Expected: no errors (the GUI main.go calls `StartHTTPServer(mcpPort)` with one arg, which still works)

- [ ] **Step 4: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add pkg/mcpserver/server.go && git commit -m "feat(mcp): register session/realtime tools, support configurable bind address"
```

---

### Task 10: CLI entry point — main + serve command

**Files:**
- Create: `cmd/chying-cli/main.go`
- Create: `cmd/chying-cli/serve.go`

- [ ] **Step 1: Create CLI main.go**

Create `cmd/chying-cli/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yhy0/ChYing/conf"
)

var rootCmd = &cobra.Command{
	Use:   "chying-cli",
	Short: "ChYing CLI - 被动扫描服务",
	Long:  "ChYing 被动扫描服务，提供 HTTP 代理 + 被动扫描 + MCP 接口。",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ChYing CLI v%s\n", conf.Version)
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Create serve command**

Create `cmd/chying-cli/serve.go`:

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/panjf2000/ants/v2"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/spf13/cobra"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/mitmproxy"
	JieConf "github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/mode"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/mcpserver"
	"github.com/yhy0/logging"
)

var (
	proxyPort int
	mcpPort   int
	bindAddr  string
	project   string
	quiet     bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动被动扫描服务",
	Long:  "启动 HTTP 代理 + 被动扫描 + MCP SSE 服务",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntVar(&proxyPort, "proxy-port", 9080, "代理监听端口")
	serveCmd.Flags().IntVar(&mcpPort, "mcp-port", 9090, "MCP SSE 服务端口")
	serveCmd.Flags().StringVar(&bindAddr, "bind", "127.0.0.1", "监听地址 (Docker 场景用 0.0.0.0)")
	serveCmd.Flags().StringVar(&project, "project", "default", "项目名称")
	serveCmd.Flags().BoolVar(&quiet, "quiet", false, "静默模式，不输出流量和漏洞到终端")
}

func runServe(cmd *cobra.Command, args []string) error {
	// 1. 日志初始化
	logging.Logger = logging.New(true, file.ChyingDir, "ChYing-CLI", true)
	logging.Logger.Infoln("Starting ChYing CLI...")

	// 2. 文件系统
	file.New()

	// 3. 协程池
	pool, err := ants.NewPool(conf.Parallelism)
	if err != nil {
		return fmt.Errorf("创建协程池失败: %w", err)
	}
	defer pool.Release()

	// 4. 指纹识别引擎
	JieConf.Wappalyzer, _ = wappalyzer.New()

	// 5. 配置加载
	conf.InitConfig()
	conf.HotConf()
	conf.SyncJieConfig()

	// 处理过滤配置
	for _, suffix := range strings.Split(conf.AppConf.Mitmproxy.FilterSuffix, ", ") {
		conf.Config.FilterSuffix = append(conf.Config.FilterSuffix, suffix)
	}
	for index, v := range conf.AppConf.Mitmproxy.Exclude {
		if v == "" {
			continue
		}
		conf.Config.Exclude = append(conf.Config.Exclude, &conf.Scope{
			Id: index, Enabled: true, Prefix: v, Regexp: true, Type: "exclude",
		})
	}
	for index, v := range conf.AppConf.Mitmproxy.Include {
		if v == "" {
			continue
		}
		conf.Config.Include = append(conf.Config.Include, &conf.Scope{
			Id: index, Enabled: true, Prefix: v, Regexp: true, Type: "include",
		})
	}

	// 关闭所有插件（被动扫描模式由 proxy 插件系统管理）
	for k := range JieConf.Plugin {
		JieConf.Plugin[k] = false
	}

	// 6. 数据库
	db.Init(project, "sqlite")

	// 7. 被动模式
	mode.Passive()

	// 8. 覆盖代理配置为 CLI 参数
	conf.ProxyHost = bindAddr
	conf.ProxyPort = proxyPort
	conf.AppConf.Proxy.Host = bindAddr
	conf.AppConf.Proxy.Port = proxyPort

	// 9. 启动代理
	go func() {
		logging.Logger.Infof("Starting proxy on %s:%d", bindAddr, proxyPort)
		mitmproxy.Proxify()
		logging.Logger.Errorln("Proxy server has stopped.")
	}()
	time.Sleep(2 * time.Second) // 等待代理启动

	printStatus("Proxy listening on %s:%d", bindAddr, proxyPort)

	// 10. 启动 MCP Server
	mcpAddr, mcpErr := mcpserver.StartHTTPServer(mcpPort, bindAddr)
	if mcpErr != nil {
		return fmt.Errorf("MCP server 启动失败: %w", mcpErr)
	}
	printStatus("MCP server on %s/mcp", mcpAddr)

	// 11. 事件循环（消费 EventDataChan 和 OutChannel，入库 + 终端输出）
	go cliEventNotification(pool)
	go cliVulnLoop()

	printStatus("ChYing CLI ready. Press Ctrl+C to stop.")

	// 12. 优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	printStatus("Shutting down...")
	return nil
}

// cliEventNotification 消费代理事件，入库（替代 GUI 模式的 EventNotification）
func cliEventNotification(pool *ants.Pool) {
	for _data := range mitmproxy.EventDataChan {
		if _data.Name == "HttpHistory" {
			_http, ok := _data.Data.(*mitmproxy.HTTPHistory)
			if !ok {
				continue
			}

			historyData := &db.HTTPHistory{
				Hid:         _http.Id,
				Host:        _http.Host,
				Method:      _http.Method,
				FullUrl:     _http.FullUrl,
				Path:        _http.Path,
				Status:      _http.Status,
				Length:      _http.Length,
				ContentType: _http.ContentType,
				MIMEType:    _http.MIMEType,
				Extension:   _http.Extension,
				Title:       _http.Title,
				IP:          _http.IP,
				Color:       _http.Color,
				Note:        _http.Note,
				Source:      "local",
				SourceID:    "localhost",
				NodeName:    "CLI",
				SessionID:   _http.SessionID,
			}
			db.AddHistory(historyData)

			// 存储请求/响应体
			if body, loaded := mitmproxy.HTTPBodyMap.LoadAndDelete(_http.Id); loaded {
				if httpBody, typeOk := body.(*mitmproxy.HTTPBody); typeOk {
					db.AddRequest(&db.Request{
						RequestId:  int(historyData.Hid),
						RequestRaw: httpBody.RequestRaw,
						Url:        historyData.FullUrl,
						Host:       historyData.Host,
					}, &db.Response{
						RequestId:   int(historyData.Hid),
						ResponseRaw: httpBody.ResponseRaw,
						Url:         historyData.FullUrl,
						Host:        historyData.Host,
					})
				}
			}

			if !quiet {
				printTraffic(_http.Method, _http.FullUrl, _http.Status, _http.Length)
			}
		}
	}
}

// cliVulnLoop 消费漏洞通道，入库 + 终端输出
func cliVulnLoop() {
	for vuln := range output.OutChannel {
		vulnData := &db.Vulnerability{
			VulnID:      fmt.Sprintf("%s-%s-%d", vuln.VulnData.VulnType, vuln.VulnData.Target, time.Now().UnixNano()),
			VulnType:    vuln.VulnData.VulnType,
			Target:      vuln.VulnData.Target,
			Host:        vuln.VulnData.Target,
			Method:      vuln.VulnData.Method,
			Plugin:      vuln.Plugin,
			Level:       vuln.Level,
			IP:          vuln.VulnData.Ip,
			Param:       vuln.VulnData.Param,
			Payload:     vuln.VulnData.Payload,
			Description: vuln.VulnData.Description,
			CurlCommand: vuln.VulnData.CURLCommand,
			Request:     vuln.VulnData.Request,
			Response:    vuln.VulnData.Response,
			Source:      "local",
			SourceID:    "localhost",
			NodeName:    "CLI",
			ProjectID:   project,
			// SessionID 将通过 TODO: 从流量关联中继承
		}
		db.AddVulnerability(vulnData)

		if !quiet {
			printVuln(vuln.Level, vuln.VulnData.VulnType, vuln.VulnData.Target, vuln.VulnData.Param)
		}
	}
}
```

- [ ] **Step 3: Check `db.Request` and `db.Response` models exist**

The `cliEventNotification` calls `db.AddRequest`. Verify the function signature:

```bash
cd /Users/yhy/Documents/Github/ChYing && grep -n "func AddRequest" pkg/db/*.go
```

Adapt the call if the signature differs from what's shown above.

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./cmd/chying-cli/
```

Expected: compilation error for `printStatus`, `printTraffic`, `printVuln` (not yet defined — that's Task 11)

- [ ] **Step 5: Commit (partial — output.go comes next)**

Hold this commit until Task 11 is done.

---

### Task 11: CLI terminal output formatting

**Files:**
- Create: `cmd/chying-cli/output.go`

- [ ] **Step 1: Create output.go**

Create `cmd/chying-cli/output.go`:

```go
package main

import (
	"fmt"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

func timestamp() string {
	return time.Now().Format("15:04:05")
}

func printStatus(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s]%s %s%s%s\n", colorGray, timestamp(), colorReset, colorCyan, msg, colorReset)
}

func printTraffic(method, url, status, length string) {
	statusColor := colorGreen
	if len(status) > 0 && status[0] != '2' {
		statusColor = colorYellow
	}
	fmt.Printf("%s[%s]%s %s→%s %s %s %s[%s]%s %s\n",
		colorGray, timestamp(), colorReset,
		colorBlue, colorReset,
		method, url,
		statusColor, status, colorReset,
		length,
	)
}

func printVuln(level, vulnType, target, param string) {
	var icon, levelColor string
	switch level {
	case "Critical":
		icon = "🔴"
		levelColor = colorRed
	case "High":
		icon = "🔴"
		levelColor = colorRed
	case "Medium":
		icon = "🟡"
		levelColor = colorYellow
	case "Low":
		icon = "🟢"
		levelColor = colorGreen
	default:
		icon = "⚪"
		levelColor = colorGray
	}

	paramInfo := ""
	if param != "" {
		paramInfo = fmt.Sprintf(" | Param: %s", param)
	}

	fmt.Printf("%s[%s]%s %s %s[%s]%s %s | %s%s\n",
		colorGray, timestamp(), colorReset,
		icon,
		levelColor, level, colorReset,
		vulnType, target, paramInfo,
	)
}
```

- [ ] **Step 2: Verify full CLI compilation**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build ./cmd/chying-cli/
```

Expected: successful build

- [ ] **Step 3: Commit Task 10 + 11 together**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add cmd/chying-cli/ && git commit -m "feat: add CLI entry point with serve command and terminal output"
```

---

### Task 12: Docker deployment files

**Files:**
- Create: `Dockerfile.cli`
- Create: `docker-compose.cli.yml`

- [ ] **Step 1: Create Dockerfile.cli**

Create `Dockerfile.cli`:

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# 依赖缓存
COPY go.mod go.sum ./
RUN go mod download

# 源码编译
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o chying-cli ./cmd/chying-cli/

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite-libs

COPY --from=builder /app/chying-cli /usr/local/bin/chying-cli

# 数据持久化目录
RUN mkdir -p /root/.config/ChYing
VOLUME /root/.config/ChYing

EXPOSE 9080 9090

ENTRYPOINT ["chying-cli"]
CMD ["serve", "--bind", "0.0.0.0"]
```

- [ ] **Step 2: Create docker-compose.cli.yml**

Create `docker-compose.cli.yml`:

```yaml
services:
  chying:
    build:
      context: .
      dockerfile: Dockerfile.cli
    ports:
      - "9080:9080"
      - "9090:9090"
    volumes:
      - chying-data:/root/.config/ChYing
    command: ["serve", "--proxy-port", "9080", "--mcp-port", "9090", "--bind", "0.0.0.0"]
    restart: unless-stopped

volumes:
  chying-data:
```

- [ ] **Step 3: Commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add Dockerfile.cli docker-compose.cli.yml && git commit -m "feat: add Docker deployment files for CLI mode"
```

---

### Task 13: Integration verification

**Files:** None (verification only)

- [ ] **Step 1: Build CLI binary**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build -o bin/chying-cli ./cmd/chying-cli/
```

Expected: successful build, binary at `bin/chying-cli`

- [ ] **Step 2: Test version command**

```bash
cd /Users/yhy/Documents/Github/ChYing && ./bin/chying-cli version
```

Expected: `ChYing CLI v2.1.3`

- [ ] **Step 3: Test serve command help**

```bash
cd /Users/yhy/Documents/Github/ChYing && ./bin/chying-cli serve --help
```

Expected: shows all flags (--proxy-port, --mcp-port, --bind, --project, --quiet)

- [ ] **Step 4: Test serve startup**

```bash
cd /Users/yhy/Documents/Github/ChYing && timeout 10 ./bin/chying-cli serve --proxy-port 19080 --mcp-port 19090 --project cli-test || true
```

Expected: output showing proxy and MCP server started, then timeout exit

- [ ] **Step 5: Verify GUI mode still compiles**

```bash
cd /Users/yhy/Documents/Github/ChYing && go build .
```

Expected: successful build (no regressions)

- [ ] **Step 6: Clean up test artifacts**

```bash
cd /Users/yhy/Documents/Github/ChYing && rm -f bin/chying-cli && rm -rf ~/.config/ChYing/db/cli-test
```

- [ ] **Step 7: Final commit**

```bash
cd /Users/yhy/Documents/Github/ChYing && git add -A && git commit -m "chore: integration verification complete" --allow-empty
```

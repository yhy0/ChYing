package mcpserver

import (
	"net/http"
	"testing"

	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func resetSessions() {
	sessionStore.Range(func(key, _ interface{}) bool {
		sessionStore.Delete(key)
		return true
	})
}

func TestSessionAttributionTracksCurrentProject(t *testing.T) {
	resetSessions()
	previousProject := db.CurrentProjectName
	t.Cleanup(func() {
		db.CurrentProjectName = previousProject
		resetSessions()
	})

	db.CurrentProjectName = "src-auto"
	session := RegisterSession([]string{"example.com"}, "TaiE case", "src-auto")
	projectID, sessionID, err := sessionAttribution(session.SessionID)
	if err != nil {
		t.Fatalf("sessionAttribution() error = %v", err)
	}
	if projectID != "src-auto" || sessionID != session.SessionID {
		t.Fatalf("unexpected attribution: project=%q session=%q", projectID, sessionID)
	}

	db.CurrentProjectName = "another-project"
	if _, _, err := sessionAttribution(session.SessionID); err == nil {
		t.Fatal("expected stale cross-project session to be rejected")
	}
}

func TestSaveRepeaterHistoryPersistsProjectAndSession(t *testing.T) {
	previousDB := db.GlobalDB
	previousProject := db.CurrentProjectName
	database, err := gorm.Open(sqlite.Open("file:session-attribution?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(&db.HTTPHistory{}, &db.Request{}, &db.Response{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	db.GlobalDB = database
	db.CurrentProjectName = "src-auto"
	t.Cleanup(func() {
		db.GlobalDB = previousDB
		db.CurrentProjectName = previousProject
	})

	response := &httpx.Response{
		Status: "200 OK", StatusCode: 200, Body: "<title>Example</title>",
		RequestDump:  "GET /api HTTP/1.1\r\nHost: example.com\r\n\r\n",
		ResponseDump: "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{}",
		RespHeader:   http.Header{"Content-Type": []string{"application/json"}}, ContentLength: 2,
	}
	reference := saveRepeaterHistory("https://example.com/api", response.RequestDump, response, "src-auto", "case-session")
	if reference.ID == 0 || reference.Hid == 0 {
		t.Fatalf("expected persisted reference, got %#v", reference)
	}

	var history db.HTTPHistory
	if err := database.Where("hid = ?", reference.Hid).First(&history).Error; err != nil {
		t.Fatalf("read history: %v", err)
	}
	if history.ProjectID != "src-auto" || history.SessionID != "case-session" {
		t.Fatalf("history attribution lost: project=%q session=%q", history.ProjectID, history.SessionID)
	}
}

func TestSessionPersistsAcrossMemoryReset(t *testing.T) {
	previousDB := db.GlobalDB
	previousProject := db.CurrentProjectName
	database, err := gorm.Open(sqlite.Open("file:persistent-mcp-session?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(&db.MCPScanSession{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	db.GlobalDB = database
	db.CurrentProjectName = "src-auto"
	resetSessions()
	t.Cleanup(func() {
		db.GlobalDB = previousDB
		db.CurrentProjectName = previousProject
		resetSessions()
	})

	registered := RegisterSession([]string{"api.example.com"}, "TaiE case", "src-auto")
	resetSessions()
	restored, ok := GetSession(registered.SessionID)
	if !ok {
		t.Fatal("expected persisted session to be restored")
	}
	if restored.ProjectID != "src-auto" || len(restored.Targets) != 1 || restored.Targets[0] != "api.example.com" {
		t.Fatalf("unexpected restored session: %#v", restored)
	}
	resetSessions()
	configured, ok := ConfigureSession(registered.SessionID, []string{"edge.example.com"}, nil)
	if !ok || len(configured.Targets) != 2 {
		t.Fatalf("expected configure to restore persisted session, got %#v", configured)
	}
}

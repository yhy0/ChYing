package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yhy0/ChYing/pkg/desktop"
)

func TestParseOpenOptions(t *testing.T) {
	options, err := parseOpenOptions([]string{"--project", "src-auto", "--wait-mcp", "--json", "--timeout", "3s"}, os.Stderr)
	if err != nil {
		t.Fatalf("parseOpenOptions failed: %v", err)
	}
	if options.project != "src-auto" || !options.waitMCP || !options.json || options.timeout != 3*time.Second {
		t.Fatalf("unexpected options: %#v", options)
	}
}

func TestWaitForProject(t *testing.T) {
	runtimePath := filepath.Join(t.TempDir(), "runtime.json")
	requestedAt := time.Now().UTC()
	go func() {
		time.Sleep(30 * time.Millisecond)
		_ = desktop.WriteRuntimeState(runtimePath, desktop.RuntimeState{
			Version: 1, Status: desktop.StatusReady, Project: "src-auto",
			MCPURL: "http://127.0.0.1:9090/mcp", PID: os.Getpid(),
			StartedAt: requestedAt, UpdatedAt: time.Now().UTC(),
		})
	}()

	state, err := waitForProject(runtimePath, "src-auto", requestedAt, time.Second, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("waitForProject failed: %v", err)
	}
	if state.MCPURL != "http://127.0.0.1:9090/mcp" {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestWaitForProjectRejectsDifferentReadyProject(t *testing.T) {
	runtimePath := filepath.Join(t.TempDir(), "runtime.json")
	now := time.Now().UTC()
	if err := desktop.WriteRuntimeState(runtimePath, desktop.RuntimeState{
		Version: 1, Status: desktop.StatusReady, Project: "other",
		MCPURL: "http://127.0.0.1:9090/mcp", PID: os.Getpid(),
		StartedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := waitForProject(runtimePath, "src-auto", now, time.Second, 10*time.Millisecond); err == nil {
		t.Fatal("expected a project conflict")
	}
}

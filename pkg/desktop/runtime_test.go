package desktop

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestProjectFromArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		project string
		wantErr bool
	}{
		{name: "absent", args: []string{"--debug"}},
		{name: "separate", args: []string{"--project", "src-auto"}, project: "src-auto"},
		{name: "inline", args: []string{"--project=src-auto"}, project: "src-auto"},
		{name: "missing", args: []string{"--project"}, wantErr: true},
		{name: "path", args: []string{"--project", "../src-auto"}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project, err := ProjectFromArgs(test.args)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ProjectFromArgs failed: %v", err)
			}
			if project != test.project {
				t.Fatalf("project = %q, want %q", project, test.project)
			}
		})
	}
}

func TestRuntimeStateRoundTripIsPrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "runtime.json")
	now := time.Now().UTC().Truncate(time.Second)
	want := RuntimeState{
		Version: 1, Status: StatusReady, Project: "src-auto",
		MCPURL: "http://127.0.0.1:9090/mcp", PID: os.Getpid(),
		StartedAt: now, UpdatedAt: now,
	}
	if err := WriteRuntimeState(path, want); err != nil {
		t.Fatalf("WriteRuntimeState failed: %v", err)
	}
	got, err := ReadRuntimeState(path)
	if err != nil {
		t.Fatalf("ReadRuntimeState failed: %v", err)
	}
	if got.Status != want.Status || got.Project != want.Project || got.MCPURL != want.MCPURL {
		t.Fatalf("unexpected round trip: %#v", got)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("runtime permissions = %o, want 600", info.Mode().Perm())
		}
	}
}

func TestProcessIsRunning(t *testing.T) {
	if !ProcessIsRunning(os.Getpid()) {
		t.Fatal("current process should be running")
	}
	if ProcessIsRunning(-1) {
		t.Fatal("negative pid should not be running")
	}
}

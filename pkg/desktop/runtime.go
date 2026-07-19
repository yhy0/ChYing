package desktop

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
)

const (
	StatusSelecting = "selecting"
	StatusOpening   = "opening"
	StatusReady     = "ready"
	StatusFailed    = "failed"
	StatusStopped   = "stopped"
)

// RuntimeState is the local desktop readiness contract consumed by chyingctl.
// It intentionally contains no credentials or captured traffic.
type RuntimeState struct {
	Version   int       `json:"version"`
	Status    string    `json:"status"`
	Project   string    `json:"project,omitempty"`
	MCPURL    string    `json:"mcp_url,omitempty"`
	PID       int       `json:"pid"`
	Progress  int       `json:"progress,omitempty"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func RuntimeFilePath() string {
	if override := strings.TrimSpace(os.Getenv("CHYING_RUNTIME_FILE")); override != "" {
		return override
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".config", "ChYing", "runtime.json")
	}
	return filepath.Join(home, ".config", "ChYing", "runtime.json")
}

func NormalizeProjectName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", errors.New("project name is required")
	}
	if name == "." || name == ".." || strings.ContainsAny(name, `/\\\x00`) || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid project name: %q", value)
	}
	return name, nil
}

// ProjectFromArgs extracts --project without rejecting unrelated Wails or OS arguments.
func ProjectFromArgs(args []string) (string, error) {
	var project string
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--project":
			if index+1 >= len(args) {
				return "", errors.New("--project requires a value")
			}
			index++
			project = args[index]
		case strings.HasPrefix(arg, "--project="):
			project = strings.TrimPrefix(arg, "--project=")
		}
	}
	if project == "" {
		return "", nil
	}
	return NormalizeProjectName(project)
}

func ReadRuntimeState(path string) (RuntimeState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RuntimeState{}, err
	}
	var state RuntimeState
	if err := json.Unmarshal(data, &state); err != nil {
		return RuntimeState{}, fmt.Errorf("decode runtime state: %w", err)
	}
	return state, nil
}

func WriteRuntimeState(path string, state RuntimeState) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create runtime directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode runtime state: %w", err)
	}
	temp, err := os.CreateTemp(dir, ".runtime-*.json")
	if err != nil {
		return fmt.Errorf("create runtime state: %w", err)
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}
	if err := temp.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("secure runtime state: %w", err)
	}
	if _, err := temp.Write(append(data, '\n')); err != nil {
		cleanup()
		return fmt.Errorf("write runtime state: %w", err)
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync runtime state: %w", err)
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("close runtime state: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("publish runtime state: %w", err)
	}
	return os.Chmod(path, 0o600)
}

func ProcessIsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	running, err := process.PidExists(int32(pid))
	return err == nil && running
}

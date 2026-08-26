package desktop

import (
	"fmt"
	"os"
	"time"
)

func AssertCanServeProject(project string) error {
	state, err := ReadRuntimeState(RuntimeFilePath())
	if err != nil {
		return nil
	}
	if !ProcessIsRunning(state.PID) || state.PID == os.Getpid() {
		return nil
	}
	if state.Project != "" && state.Project != project {
		return fmt.Errorf("ChYing 已打开项目 %q（pid %d），不会静默切换到 %q", state.Project, state.PID, project)
	}
	if state.Status == StatusReady && state.MCPURL != "" {
		return fmt.Errorf("ChYing 项目 %q 已在运行（pid %d）", state.Project, state.PID)
	}
	return nil
}

func WriteOpeningState(project string) {
	now := time.Now().UTC()
	_ = WriteRuntimeState(RuntimeFilePath(), RuntimeState{
		Version: 1, Status: StatusOpening, Project: project, PID: os.Getpid(),
		Progress: 20, Message: "正在启动 CLI 服务", StartedAt: now, UpdatedAt: now,
	})
}

func WriteReadyState(project, mcpURL, proxyAddr, caCert string, ca CAInstallResult) error {
	now := time.Now().UTC()
	state := RuntimeState{
		Version: 1, Status: StatusReady, Project: project,
		MCPURL: mcpURL, ProxyAddr: proxyAddr, CaCert: caCert,
		CaInstalled: ca.Installed, CaMessage: ca.Message,
		PID: os.Getpid(), Progress: 100, Message: "CLI 已就绪",
		StartedAt: now, UpdatedAt: now,
	}
	if existing, err := ReadRuntimeState(RuntimeFilePath()); err == nil && !existing.StartedAt.IsZero() {
		state.StartedAt = existing.StartedAt
	}
	return WriteRuntimeState(RuntimeFilePath(), state)
}

func WriteFailedState(project, message string) {
	now := time.Now().UTC()
	_ = WriteRuntimeState(RuntimeFilePath(), RuntimeState{
		Version: 1, Status: StatusFailed, Project: project, PID: os.Getpid(),
		Error: message, Message: "CLI 启动失败", StartedAt: now, UpdatedAt: now,
	})
}

func WriteStoppedState() {
	now := time.Now().UTC()
	state := RuntimeState{
		Version: 1, Status: StatusStopped, PID: os.Getpid(),
		Message: "CLI 已退出", UpdatedAt: now, StartedAt: now,
	}
	if existing, err := ReadRuntimeState(RuntimeFilePath()); err == nil {
		state.Project = existing.Project
		if !existing.StartedAt.IsZero() {
			state.StartedAt = existing.StartedAt
		}
	}
	_ = WriteRuntimeState(RuntimeFilePath(), state)
}

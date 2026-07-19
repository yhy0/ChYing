package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/desktop"
	"github.com/yhy0/logging"
)

const desktopWindowName = "承影"

func (a *App) ensureDesktopRuntime() {
	a.desktopRuntimeOnce.Do(func() {
		now := time.Now().UTC()
		a.desktopRequests = make(chan []string, 16)
		a.runtimeState = desktop.RuntimeState{
			Version:   1,
			Status:    desktop.StatusSelecting,
			PID:       os.Getpid(),
			Message:   "等待选择项目",
			StartedAt: now,
			UpdatedAt: now,
		}
	})
}

// startDesktopLaunchController starts the single serialized project-open queue.
// It is called once Wails has started so window and event operations are safe.
func (a *App) startDesktopLaunchController(initialArgs []string) {
	a.ensureDesktopRuntime()
	a.desktopControllerOnce.Do(func() {
		a.publishDesktopState(a.currentDesktopState())
		go func() {
			for args := range a.desktopRequests {
				a.processDesktopLaunchArgs(args)
			}
		}()
	})
	if len(initialArgs) > 0 {
		a.handleDesktopLaunchArgs(initialArgs)
	}
}

// handleDesktopLaunchArgs is safe to call from Wails' second-instance callback,
// including during application construction before the main window is ready.
func (a *App) handleDesktopLaunchArgs(args []string) {
	a.ensureDesktopRuntime()
	request := append([]string(nil), args...)
	a.desktopRequests <- request
}

func (a *App) processDesktopLaunchArgs(args []string) {
	projectName, err := desktop.ProjectFromArgs(args)
	if err != nil {
		_ = a.failDesktopOpen("", fmt.Errorf("启动参数无效: %w", err))
		a.focusMainWindow()
		return
	}
	if projectName == "" {
		a.focusMainWindow()
		return
	}

	result := a.OpenExistingProject(projectName)
	if result.Error != "" {
		logging.Logger.Warnf("Agent 打开项目失败: %s", result.Error)
	}
}

// OpenExistingProject is the single project-opening entry point shared by the
// desktop UI and agent launch requests.
func (a *App) OpenExistingProject(projectName string) Result {
	a.ensureDesktopRuntime()
	a.projectOpenMu.Lock()
	defer a.projectOpenMu.Unlock()

	name, err := desktop.NormalizeProjectName(projectName)
	if err != nil {
		return a.failDesktopOpen(projectName, err)
	}

	state := a.currentDesktopState()
	if state.Status == desktop.StatusReady {
		if state.Project != name {
			return Result{Error: fmt.Sprintf("ChYing 已打开项目 %q；为避免污染数据，不会静默切换到 %q，请先退出桌面应用", state.Project, name)}
		}
		state.UpdatedAt = time.Now().UTC()
		state.Message = "项目已就绪"
		a.setAndPublishDesktopState(state)
		a.focusMainWindow()
		return Result{Data: state}
	}
	if state.Status == desktop.StatusFailed || db.CurrentProjectName != "" {
		current := state.Project
		if current == "" {
			current = db.CurrentProjectName
		}
		return Result{Error: fmt.Sprintf("ChYing 当前项目 %q 的初始化状态不可复用，请退出桌面应用后重试", current)}
	}
	if err := validateExistingProject(name); err != nil {
		return a.failDesktopOpen(name, err)
	}

	a.updateDesktopState(desktop.StatusOpening, name, 5, "正在打开项目", "")

	steps := []struct {
		progress int
		message  string
		run      func() Result
	}{
		{15, "正在准备项目", func() Result { return a.StartInitialization("Open existing project", name) }},
		{25, "正在初始化基础组件", a.StepBasicInitialization},
		{35, "正在加载配置", a.StepConfigurationLoad},
		{50, "正在连接项目数据库", func() Result { return a.StepDatabaseConnection(name) }},
		{65, "正在检查数据库结构", a.StepSchemaValidation},
		{80, "正在启动代理服务", a.StepProxyServerStart},
		{90, "正在加载项目数据", func() Result { return a.StepProjectDataLoad("Open existing project", name) }},
		{100, "正在启动 MCP 服务", a.StepInitializationComplete},
	}

	for _, step := range steps {
		a.updateDesktopState(desktop.StatusOpening, name, step.progress, step.message, "")
		if err := resultError(step.run()); err != nil {
			return a.failDesktopOpen(name, err)
		}
	}

	state = a.currentDesktopState()
	if state.Status != desktop.StatusReady || state.Project != name || state.MCPURL == "" {
		message := state.Error
		if message == "" {
			message = "MCP 服务未能进入就绪状态"
		}
		return a.failDesktopOpen(name, fmt.Errorf("%s", message))
	}

	a.focusMainWindow()
	return Result{Data: state}
}

// GetDesktopRuntimeState exposes the same readiness contract used by chyingctl.
func (a *App) GetDesktopRuntimeState() Result {
	a.ensureDesktopRuntime()
	return Result{Data: a.currentDesktopState()}
}

func resultError(result Result) error {
	if result.Error != "" {
		return fmt.Errorf("%s", result.Error)
	}
	if progress, ok := result.Data.(*InitProgress); ok && !progress.Success {
		if progress.Error != "" {
			return fmt.Errorf("%s", progress.Error)
		}
		return fmt.Errorf("%s", progress.Message)
	}
	if result.Data == nil {
		return fmt.Errorf("初始化步骤没有返回结果")
	}
	return nil
}

func validateExistingProject(projectName string) error {
	return validateExistingProjectAt(filepath.Join(file.ChyingDir, "db"), projectName)
}

func validateExistingProjectAt(root string, projectName string) error {
	projectDir := filepath.Join(root, projectName)
	rel, err := filepath.Rel(root, projectDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("项目路径不合法")
	}

	dirInfo, err := os.Lstat(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("项目 %q 不存在", projectName)
		}
		return fmt.Errorf("检查项目目录失败: %w", err)
	}
	if dirInfo.Mode()&os.ModeSymlink != 0 || !dirInfo.IsDir() {
		return fmt.Errorf("项目目录必须是真实目录，不能是符号链接")
	}

	databasePath := filepath.Join(projectDir, projectName+".db")
	databaseInfo, err := os.Lstat(databasePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("项目数据库不存在: %s", databasePath)
		}
		return fmt.Errorf("检查项目数据库失败: %w", err)
	}
	if databaseInfo.Mode()&os.ModeSymlink != 0 || !databaseInfo.Mode().IsRegular() {
		return fmt.Errorf("项目数据库必须是普通文件，不能是符号链接")
	}
	return nil
}

func (a *App) failDesktopOpen(projectName string, err error) Result {
	message := err.Error()
	if a.currentDesktopState().Status == desktop.StatusReady {
		return Result{Error: message}
	}
	a.publishDesktopFailure(projectName, message)
	return Result{Error: message}
}

func (a *App) publishDesktopFailure(projectName string, message string) {
	a.updateDesktopState(desktop.StatusFailed, projectName, 0, "项目打开失败", message)
}

func (a *App) markDesktopProjectReady(projectName string, mcpURL string) {
	a.ensureDesktopRuntime()
	a.runtimeMu.Lock()
	state := a.runtimeState
	state.Status = desktop.StatusReady
	state.Project = projectName
	state.MCPURL = mcpURL
	state.PID = os.Getpid()
	state.Progress = 100
	state.Message = "项目已就绪"
	state.Error = ""
	state.UpdatedAt = time.Now().UTC()
	a.runtimeState = state
	a.runtimeMu.Unlock()
	a.publishDesktopState(state)
	if wailsApp != nil {
		wailsApp.Event.Emit("DesktopProjectOpened", state)
	}
}

func (a *App) markDesktopProjectFailed(projectName string, message string) {
	a.publishDesktopFailure(projectName, message)
}

func (a *App) markDesktopStopped() {
	a.ensureDesktopRuntime()
	state := a.currentDesktopState()
	state.Status = desktop.StatusStopped
	state.Progress = 0
	state.Message = "ChYing 已退出"
	state.MCPURL = ""
	state.UpdatedAt = time.Now().UTC()
	a.setAndPublishDesktopState(state)
}

func (a *App) updateDesktopState(status string, projectName string, progress int, message string, errorMessage string) {
	a.ensureDesktopRuntime()
	state := a.currentDesktopState()
	state.Version = 1
	state.Status = status
	state.Project = projectName
	state.PID = os.Getpid()
	state.Progress = progress
	state.Message = message
	state.Error = errorMessage
	state.UpdatedAt = time.Now().UTC()
	if status != desktop.StatusReady {
		state.MCPURL = ""
	}
	a.setAndPublishDesktopState(state)
}

func (a *App) setAndPublishDesktopState(state desktop.RuntimeState) {
	a.ensureDesktopRuntime()
	a.runtimeMu.Lock()
	a.runtimeState = state
	a.runtimeMu.Unlock()
	a.publishDesktopState(state)
}

func (a *App) currentDesktopState() desktop.RuntimeState {
	a.ensureDesktopRuntime()
	a.runtimeMu.RLock()
	defer a.runtimeMu.RUnlock()
	return a.runtimeState
}

func (a *App) publishDesktopState(state desktop.RuntimeState) {
	if err := desktop.WriteRuntimeState(desktop.RuntimeFilePath(), state); err != nil {
		logging.Logger.Warnf("写入桌面运行状态失败: %v", err)
	}
	if wailsApp != nil {
		wailsApp.Event.Emit("DesktopProjectOpenProgress", state)
		if state.Status == desktop.StatusFailed {
			wailsApp.Event.Emit("DesktopProjectOpenFailed", state)
		}
	}
}

func (a *App) focusMainWindow() {
	if wailsApp == nil {
		return
	}
	window, ok := wailsApp.Window.GetByName(desktopWindowName)
	if !ok || window == nil {
		return
	}
	window.Show()
	window.Restore()
	window.Focus()
}

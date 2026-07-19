package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/yhy0/ChYing/pkg/desktop"
)

const defaultTimeout = 60 * time.Second

type openOptions struct {
	project  string
	waitMCP  bool
	timeout  time.Duration
	json     bool
	appPath  string
	runtime  string
	pollRate time.Duration
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "open":
		options, err := parseOpenOptions(args[1:], stderr)
		if err != nil {
			printFailure(stdout, stderr, hasJSONFlag(args[1:]), "", err)
			return 2
		}
		state, err := openProject(options)
		if err != nil {
			printFailure(stdout, stderr, options.json, options.project, err)
			return 1
		}
		printState(stdout, state, options.json)
		return 0

	case "status":
		statusFlags := flag.NewFlagSet("status", flag.ContinueOnError)
		statusFlags.SetOutput(stderr)
		jsonOutput := statusFlags.Bool("json", false, "以 JSON 输出")
		runtimePath := statusFlags.String("runtime", desktop.RuntimeFilePath(), "运行状态文件")
		if err := statusFlags.Parse(args[1:]); err != nil {
			return 2
		}
		state, err := desktop.ReadRuntimeState(*runtimePath)
		if err != nil {
			printFailure(stdout, stderr, *jsonOutput, "", fmt.Errorf("读取 ChYing 状态失败: %w", err))
			return 1
		}
		if !desktop.ProcessIsRunning(state.PID) {
			state.Status = desktop.StatusStopped
			state.MCPURL = ""
			state.Error = "运行状态已过期，ChYing 进程不存在"
		}
		printState(stdout, state, *jsonOutput)
		return 0

	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "未知命令: %s\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func parseOpenOptions(args []string, output io.Writer) (openOptions, error) {
	flags := flag.NewFlagSet("open", flag.ContinueOnError)
	flags.SetOutput(output)
	options := openOptions{pollRate: 150 * time.Millisecond}
	flags.StringVar(&options.project, "project", "", "要打开的本地项目名称")
	flags.BoolVar(&options.waitMCP, "wait-mcp", true, "等待 MCP 服务就绪")
	flags.DurationVar(&options.timeout, "timeout", defaultTimeout, "等待超时")
	flags.BoolVar(&options.json, "json", false, "以 JSON 输出")
	flags.StringVar(&options.appPath, "app", "", "ChYing.app 或 ChYing 可执行文件路径")
	flags.StringVar(&options.runtime, "runtime", desktop.RuntimeFilePath(), "运行状态文件")
	if err := flags.Parse(args); err != nil {
		return openOptions{}, err
	}
	if options.project == "" && flags.NArg() == 1 {
		options.project = flags.Arg(0)
	}
	project, err := desktop.NormalizeProjectName(options.project)
	if err != nil {
		return openOptions{}, err
	}
	options.project = project
	if options.timeout <= 0 {
		return openOptions{}, errors.New("--timeout 必须大于 0")
	}
	return options, nil
}

func openProject(options openOptions) (desktop.RuntimeState, error) {
	appPath, err := resolveAppPath(options.appPath)
	if err != nil {
		return desktop.RuntimeState{}, err
	}
	requestedAt := time.Now().UTC()
	if err := launchApp(appPath, options.project); err != nil {
		return desktop.RuntimeState{}, fmt.Errorf("启动 ChYing 失败: %w", err)
	}
	if !options.waitMCP {
		return desktop.RuntimeState{
			Version: 1, Status: desktop.StatusOpening, Project: options.project,
			Message: "打开请求已发送", UpdatedAt: requestedAt,
		}, nil
	}
	return waitForProject(options.runtime, options.project, requestedAt, options.timeout, options.pollRate)
}

func waitForProject(runtimePath string, project string, requestedAt time.Time, timeout time.Duration, pollRate time.Duration) (desktop.RuntimeState, error) {
	deadline := time.Now().Add(timeout)
	var lastState desktop.RuntimeState
	var lastReadErr error

	for time.Now().Before(deadline) {
		state, err := desktop.ReadRuntimeState(runtimePath)
		if err != nil {
			lastReadErr = err
			time.Sleep(pollRate)
			continue
		}
		lastState = state
		if !desktop.ProcessIsRunning(state.PID) {
			time.Sleep(pollRate)
			continue
		}
		if state.Status == desktop.StatusReady && state.Project != "" && state.Project != project {
			return state, fmt.Errorf("ChYing 已打开项目 %q，不会静默切换到 %q", state.Project, project)
		}
		if state.UpdatedAt.Before(requestedAt.Add(-time.Second)) {
			time.Sleep(pollRate)
			continue
		}
		if state.Project == project {
			switch state.Status {
			case desktop.StatusReady:
				if state.MCPURL == "" {
					return state, errors.New("ChYing 已就绪，但没有发布 MCP 地址")
				}
				return state, nil
			case desktop.StatusFailed:
				if state.Error == "" {
					state.Error = "项目打开失败"
				}
				return state, errors.New(state.Error)
			case desktop.StatusStopped:
				return state, errors.New("ChYing 在项目就绪前退出")
			}
		}
		time.Sleep(pollRate)
	}

	if lastReadErr != nil && lastState.PID == 0 {
		return lastState, fmt.Errorf("等待 ChYing 状态超时: %w", lastReadErr)
	}
	if lastState.Status != "" {
		return lastState, fmt.Errorf("等待项目 %q 的 MCP 服务超时（当前状态: %s）", project, lastState.Status)
	}
	return lastState, fmt.Errorf("等待项目 %q 的 MCP 服务超时", project)
}

func resolveAppPath(explicit string) (string, error) {
	if explicit == "" {
		explicit = strings.TrimSpace(os.Getenv("CHYING_APP"))
	}
	if explicit != "" {
		return requirePath(explicit)
	}

	var candidates []string
	if executable, err := os.Executable(); err == nil {
		executable, _ = filepath.Abs(executable)
		binaryDir := filepath.Dir(executable)
		if filepath.Base(binaryDir) == "MacOS" && filepath.Base(executable) == "chyingctl" {
			candidates = append(candidates, filepath.Dir(filepath.Dir(binaryDir)))
		}
		candidates = append(candidates, filepath.Join(binaryDir, "ChYing"))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "bin", "ChYing.dev.app"),
			filepath.Join(cwd, "bin", "ChYing.app"),
		)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "Applications", "ChYing.app"))
	}
	candidates = append(candidates, "/Applications/ChYing.app")
	if executable, err := exec.LookPath("ChYing"); err == nil {
		candidates = append(candidates, executable)
	}

	for _, candidate := range candidates {
		if path, err := requirePath(candidate); err == nil {
			return path, nil
		}
	}
	return "", errors.New("找不到 ChYing.app；请安装应用，或使用 --app / CHYING_APP 指定路径")
}

func requirePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("ChYing 路径不可用 %q: %w", absPath, err)
	}
	if !info.IsDir() && info.Mode()&0o111 == 0 {
		return "", fmt.Errorf("ChYing 文件不可执行: %s", absPath)
	}
	return absPath, nil
}

func launchApp(appPath string, project string) error {
	if runtime.GOOS == "darwin" && strings.HasSuffix(strings.ToLower(appPath), ".app") {
		return exec.Command("open", "-n", appPath, "--args", "--project", project).Run()
	}
	command := exec.Command(appPath, "--project", project)
	command.Stdin = nil
	command.Stdout = nil
	command.Stderr = nil
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}

func printState(output io.Writer, state desktop.RuntimeState, jsonOutput bool) {
	if jsonOutput {
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(state)
		return
	}
	fmt.Fprintf(output, "ChYing: %s\n项目: %s\n", state.Status, state.Project)
	if state.MCPURL != "" {
		fmt.Fprintf(output, "MCP: %s\n", state.MCPURL)
	}
	fmt.Fprintf(output, "PID: %d\n", state.PID)
	if state.Error != "" {
		fmt.Fprintf(output, "错误: %s\n", state.Error)
	}
}

func printFailure(stdout io.Writer, stderr io.Writer, jsonOutput bool, project string, err error) {
	if jsonOutput {
		now := time.Now().UTC()
		_ = json.NewEncoder(stdout).Encode(desktop.RuntimeState{
			Version: 1, Status: desktop.StatusFailed, Project: project,
			Error: err.Error(), StartedAt: now, UpdatedAt: now,
		})
		return
	}
	fmt.Fprintln(stderr, err)
}

func hasJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" || arg == "--json=true" {
			return true
		}
	}
	return false
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "用法:")
	fmt.Fprintln(output, "  chyingctl open --project <名称> --wait-mcp --json")
	fmt.Fprintln(output, "  chyingctl status --json")
}

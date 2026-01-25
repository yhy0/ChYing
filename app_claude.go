package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yhy0/ChYing/conf"
	claudecode "github.com/yhy0/ChYing/pkg/claude-code"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/12/01
   @desc Claude Code CLI 集成 - 通过 CLI 调用 Claude Code
**/

// claudeClient Claude Code 客户端（全局单例）
var claudeClient *claudecode.Client

// a2aClient A2A 客户端（全局单例）
var a2aClient *claudecode.A2AClient

// getClaudeClient 获取或创建 Claude Code 客户端
func getClaudeClient() *claudecode.Client {
	if claudeClient == nil {
		claudeClient = claudecode.NewClient()
	}
	return claudeClient
}

// getA2AClient 获取或创建 A2A 客户端
func getA2AClient() *claudecode.A2AClient {
	if a2aClient == nil {
		appConfig := conf.GetAppConfig()
		a2aClient = claudecode.NewA2AClient(&claudecode.A2AClientConfig{
			AgentURL:  appConfig.AI.A2A.AgentURL,
			Headers:   appConfig.AI.A2A.Headers,
			Timeout:   appConfig.AI.A2A.Timeout,
			EnableSSE: appConfig.AI.A2A.EnableSSE,
		})
	}
	return a2aClient
}

// getAgentMode 获取当前 Agent 模式
func getAgentMode() string {
	appConfig := conf.GetAppConfig()
	mode := appConfig.AI.AgentMode
	if mode == "" {
		mode = "claude-code" // 默认使用 Claude Code CLI
	}
	return mode
}

// ClaudeInitialize 初始化 Claude Code 服务
func (a *App) ClaudeInitialize() Result {
	client := getClaudeClient()

	// 从配置获取设置
	appConfig := conf.GetAppConfig()
	config := &claudecode.Config{
		CLIPath:        appConfig.AI.Claude.CLIPath,
		Model:          appConfig.AI.Claude.Model,
		MaxTurns:       appConfig.AI.Claude.MaxTurns,
		SystemPrompt:   appConfig.AI.Claude.SystemPrompt,
		PermissionMode: appConfig.AI.Claude.PermissionMode,
	}

	// 更新配置
	client.UpdateConfig(config)

	// 初始化客户端
	if err := client.Initialize(config); err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "Claude Code service initialized successfully"}
}

// ClaudeIsInitialized 检查服务是否已初始化
func (a *App) ClaudeIsInitialized() Result {
	client := getClaudeClient()
	return Result{Data: client.IsInitialized()}
}

// ClaudeCreateSession 创建新会话
func (a *App) ClaudeCreateSession(projectID string) Result {
	client := getClaudeClient()
	session := client.CreateSession(projectID, nil)

	return Result{Data: map[string]interface{}{
		"session_id": session.ID,
		"project_id": session.ProjectID,
		"created_at": session.CreatedAt,
	}}
}

// ClaudeCreateSessionWithContext 创建带上下文的新会话
func (a *App) ClaudeCreateSessionWithContext(projectID string, agentContext *claudecode.AgentContext) Result {
	client := getClaudeClient()
	session := client.CreateSession(projectID, agentContext)

	return Result{Data: map[string]interface{}{
		"session_id": session.ID,
		"project_id": session.ProjectID,
		"created_at": session.CreatedAt,
	}}
}

// ClaudeGetSession 获取会话
func (a *App) ClaudeGetSession(sessionID string) Result {
	client := getClaudeClient()
	session := client.GetSession(sessionID)
	if session == nil {
		return Result{Error: "Session not found"}
	}

	// 尝试从 transcript 文件获取历史
	history := session.History
	if session.TranscriptPath != "" {
		transcriptHistory, err := claudecode.ReadTranscript(session.TranscriptPath)
		if err == nil && len(transcriptHistory) > 0 {
			history = transcriptHistory
		}
	}

	return Result{Data: map[string]interface{}{
		"session_id":      session.ID,
		"project_id":      session.ProjectID,
		"created_at":      session.CreatedAt,
		"updated_at":      session.UpdatedAt,
		"context":         session.Context,
		"history":         history,
		"conversation_id": session.ConversationID,
		"transcript_path": session.TranscriptPath,
	}}
}

// ClaudeGetSessionHistory 获取会话历史
func (a *App) ClaudeGetSessionHistory(sessionID string) Result {
	client := getClaudeClient()
	session := client.GetSession(sessionID)
	if session == nil {
		return Result{Error: "Session not found"}
	}

	// 优先从 transcript 文件获取历史
	if session.TranscriptPath != "" {
		transcriptHistory, err := claudecode.ReadTranscript(session.TranscriptPath)
		if err == nil && len(transcriptHistory) > 0 {
			return Result{Data: transcriptHistory}
		}
		logging.Logger.Warnf("Failed to read transcript for session %s: %v, falling back to session history", sessionID, err)
	}

	return Result{Data: session.History}
}

// ClaudeClearSession 清除会话
func (a *App) ClaudeClearSession(sessionID string) Result {
	client := getClaudeClient()
	client.ClearSession(sessionID)
	return Result{Data: "Session cleared"}
}

// ClaudeDeleteSession 删除会话
func (a *App) ClaudeDeleteSession(sessionID string) Result {
	client := getClaudeClient()
	client.DeleteSession(sessionID)
	return Result{Data: "Session deleted"}
}

// ClaudeListSessions 列出所有会话
func (a *App) ClaudeListSessions() Result {
	client := getClaudeClient()
	sessions := client.ListSessions()
	var result []map[string]interface{}
	for _, s := range sessions {
		result = append(result, map[string]interface{}{
			"id":         s.ID,
			"session_id": s.ID,
			"project_id": s.ProjectID,
			"projectId":  s.ProjectID,
			"created_at": s.CreatedAt,
			"createdAt":  s.CreatedAt,
			"updated_at": s.UpdatedAt,
			"updatedAt":  s.UpdatedAt,
			"history":    s.History,
			"context":    s.Context,
		})
	}
	return Result{Data: result}
}

// ClaudeListSessionsByProject 列出项目的所有会话
func (a *App) ClaudeListSessionsByProject(projectID string) Result {
	client := getClaudeClient()
	sessions := client.ListSessionsByProject(projectID)
	var result []map[string]interface{}
	for _, s := range sessions {
		result = append(result, map[string]interface{}{
			"id":         s.ID,
			"session_id": s.ID,
			"project_id": s.ProjectID,
			"projectId":  s.ProjectID,
			"created_at": s.CreatedAt,
			"createdAt":  s.CreatedAt,
			"updated_at": s.UpdatedAt,
			"updatedAt":  s.UpdatedAt,
			"history":    s.History,
			"context":    s.Context,
		})
	}
	return Result{Data: result}
}

// ClaudeListProjects 列出所有有会话的项目
func (a *App) ClaudeListProjects() Result {
	client := getClaudeClient()
	projects := client.ListProjects()
	return Result{Data: projects}
}

// ClaudeSendMessage 发送消息（流式）- 通过 Wails 事件推送
func (a *App) ClaudeSendMessage(sessionID, message string) Result {
	if wailsApp == nil {
		return Result{Error: "Wails app not set"}
	}

	// 根据 Agent 模式选择客户端
	agentMode := getAgentMode()

	if agentMode == "a2a" {
		return a.sendMessageViaA2A(sessionID, message)
	}

	// 默认使用 Claude Code CLI
	return a.sendMessageViaClaude(sessionID, message)
}

// sendMessageViaClaude 通过 Claude Code CLI 发送消息
func (a *App) sendMessageViaClaude(sessionID, message string) Result {
	client := getClaudeClient()

	// 创建事件通道
	eventChan := make(chan claudecode.StreamEvent, 100)

	// 启动 goroutine 处理流式响应
	go func() {
		ctx := context.Background()
		err := client.SendMessage(ctx, sessionID, message, eventChan)
		if err != nil {
			logging.Logger.Errorf("Claude Code streaming error: %v", err)
			// 发送错误事件
			wailsApp.Event.Emit(claudecode.EventClaudeError, claudecode.StreamEvent{
				Type:      "error",
				Error:     err.Error(),
				SessionID: sessionID,
			})
		}
	}()

	// 启动 goroutine 转发事件到前端
	go func() {
		for event := range eventChan {
			event.SessionID = sessionID
			switch event.Type {
			case "text":
				wailsApp.Event.Emit(claudecode.EventClaudeText, event)
			case "tool_use":
				wailsApp.Event.Emit(claudecode.EventClaudeToolUse, event)
			case "tool_result":
				wailsApp.Event.Emit(claudecode.EventClaudeToolResult, event)
			case "error":
				wailsApp.Event.Emit(claudecode.EventClaudeError, event)
			case "done":
				wailsApp.Event.Emit(claudecode.EventClaudeDone, event)
			case "cost":
				wailsApp.Event.Emit(claudecode.EventClaudeCost, event)
			default:
				wailsApp.Event.Emit("claude:stream", event)
			}
		}
	}()

	return Result{Data: "Streaming started"}
}

// sendMessageViaA2A 通过 A2A 协议发送消息
func (a *App) sendMessageViaA2A(sessionID, message string) Result {
	client := getA2AClient()

	// 检查是否已连接
	if !client.IsConnected() {
		// 尝试连接
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := client.Connect(ctx); err != nil {
			return Result{Error: fmt.Sprintf("Failed to connect to A2A agent: %v", err)}
		}
	}

	// 创建事件通道
	eventChan := make(chan claudecode.StreamEvent, 100)

	// 启动 goroutine 处理流式响应
	go func() {
		ctx := context.Background()
		err := client.SendMessage(ctx, sessionID, message, eventChan)
		if err != nil {
			logging.Logger.Errorf("A2A streaming error: %v", err)
			// 发送错误事件
			wailsApp.Event.Emit(claudecode.EventClaudeError, claudecode.StreamEvent{
				Type:      "error",
				Error:     err.Error(),
				SessionID: sessionID,
			})
		}
	}()

	// 启动 goroutine 转发事件到前端
	go func() {
		for event := range eventChan {
			event.SessionID = sessionID
			switch event.Type {
			case "text":
				wailsApp.Event.Emit(claudecode.EventClaudeText, event)
			case "tool_use":
				wailsApp.Event.Emit(claudecode.EventClaudeToolUse, event)
			case "tool_result":
				wailsApp.Event.Emit(claudecode.EventClaudeToolResult, event)
			case "error":
				wailsApp.Event.Emit(claudecode.EventClaudeError, event)
			case "done":
				wailsApp.Event.Emit(claudecode.EventClaudeDone, event)
			case "cost":
				wailsApp.Event.Emit(claudecode.EventClaudeCost, event)
			default:
				wailsApp.Event.Emit("claude:stream", event)
			}
		}
	}()

	return Result{Data: "Streaming started via A2A"}
}

// ClaudeConfirmToolExecution 确认工具执行（Claude Code CLI 自动处理工具，此方法保留兼容性）
func (a *App) ClaudeConfirmToolExecution(sessionID, toolUseID string, confirmed bool) Result {
	if !confirmed {
		return Result{Data: "Tool execution rejected"}
	}
	// Claude Code CLI 模式下，工具执行由 CLI 自动处理
	// 此方法保留用于前端兼容性
	return Result{Data: "Tool execution confirmed (handled by Claude Code CLI)"}
}

// ClaudeUpdateContext 更新会话上下文
func (a *App) ClaudeUpdateContext(sessionID string, agentContext *claudecode.AgentContext) Result {
	client := getClaudeClient()
	client.UpdateContext(sessionID, agentContext)
	return Result{Data: "Context updated"}
}

// ClaudeGetConfig 获取 Claude Code 配置
func (a *App) ClaudeGetConfig() Result {
	// 始终从应用配置文件读取配置，确保配置界面能正确显示已保存的配置
	appConfig := conf.GetAppConfig()

	return Result{Data: map[string]interface{}{
		"agent_mode":      appConfig.AI.AgentMode,
		"cli_path":        appConfig.AI.Claude.CLIPath,
		"model":           appConfig.AI.Claude.Model,
		"max_turns":       appConfig.AI.Claude.MaxTurns,
		"system_prompt":   appConfig.AI.Claude.SystemPrompt,
		"permission_mode": appConfig.AI.Claude.PermissionMode,
		// 注意: API Key、代理、MCP 服务器等配置请在 ~/.claude/settings.json 中设置
		// ChYing 会自动复用 Claude CLI 的用户配置
		"a2a": map[string]interface{}{
			"enabled":    appConfig.AI.A2A.Enabled,
			"agent_url":  appConfig.AI.A2A.AgentURL,
			"headers":    appConfig.AI.A2A.Headers,
			"timeout":    appConfig.AI.A2A.Timeout,
			"enable_sse": appConfig.AI.A2A.EnableSSE,
		},
	}}
}

// A2AConfigInput A2A 配置输入结构
type A2AConfigInput struct {
	Enabled   bool              `json:"enabled"`
	AgentURL  string            `json:"agent_url"`
	Headers   map[string]string `json:"headers"`
	Timeout   int               `json:"timeout"`
	EnableSSE bool              `json:"enable_sse"`
}

// ClaudeUpdateConfig 更新 Claude Code 配置
func (a *App) ClaudeUpdateConfig(cliPath string, model string, maxTurns int, systemPrompt string, permissionMode string) Result {
	client := getClaudeClient()
	config := &claudecode.Config{
		CLIPath:        cliPath,
		Model:          model,
		MaxTurns:       maxTurns,
		SystemPrompt:   systemPrompt,
		PermissionMode: permissionMode,
	}

	// 同时更新应用配置文件
	appConfig := conf.GetAppConfig()
	appConfig.AI.Claude.CLIPath = cliPath
	appConfig.AI.Claude.Model = model
	appConfig.AI.Claude.MaxTurns = maxTurns
	appConfig.AI.Claude.SystemPrompt = systemPrompt
	appConfig.AI.Claude.PermissionMode = permissionMode

	// 保存配置到文件
	if err := conf.SaveConfig(); err != nil {
		logging.Logger.Warnf("保存配置文件失败: %v", err)
	}

	client.UpdateConfig(config)

	return Result{Data: "Config updated"}
}

// ClaudeStopSession 停止会话（取消正在进行的请求）
func (a *App) ClaudeStopSession(sessionID string) Result {
	client := getClaudeClient()
	client.StopSession(sessionID)
	return Result{Data: "Session stopped"}
}

// ClaudeGetMCPServerURL 获取 MCP 服务器 URL
func (a *App) ClaudeGetMCPServerURL() Result {
	client := getClaudeClient()
	url := client.GetMCPServerURL()
	if url == "" {
		return Result{Error: "MCP server not running"}
	}
	return Result{Data: url}
}

// ClaudeTestConnection 测试 LLM 连接
// 发送一个简单的测试消息来验证 Claude CLI 是否可用
func (a *App) ClaudeTestConnection() Result {
	client := getClaudeClient()

	// 从配置获取设置
	appConfig := conf.GetAppConfig()
	config := &claudecode.Config{
		CLIPath:        appConfig.AI.Claude.CLIPath,
		Model:          appConfig.AI.Claude.Model,
		PermissionMode: appConfig.AI.Claude.PermissionMode,
	}

	// 更新配置
	client.UpdateConfig(config)

	// 测试连接
	ctx := context.Background()
	if err := client.TestConnection(ctx); err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "Claude CLI connection test successful"}
}

// ClaudeTestExternalMCPServer 测试外部 MCP 服务器连接
// 支持 SSE 和 STDIO 两种类型的 MCP 服务器
func (a *App) ClaudeTestExternalMCPServer(serverType string, url string, headers map[string]string, command string, args []string) Result {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if serverType == "sse" {
		// 测试 SSE 类型的 MCP 服务器
		if url == "" {
			return Result{Error: "SSE server URL is required"}
		}

		// 创建 HTTP 请求测试连接
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return Result{Error: fmt.Sprintf("Failed to create request: %v", err)}
		}

		// 添加自定义 headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// 发送请求
		client := &http.Client{
			Timeout: 15 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			return Result{Error: fmt.Sprintf("Connection failed: %v", err)}
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			return Result{Error: fmt.Sprintf("Server returned error status %d: %s", resp.StatusCode, string(body))}
		}

		return Result{Data: map[string]interface{}{
			"success":     true,
			"message":     fmt.Sprintf("SSE server connection successful (status: %d)", resp.StatusCode),
			"status_code": resp.StatusCode,
		}}

	} else if serverType == "stdio" {
		// 测试 STDIO 类型的 MCP 服务器
		if command == "" {
			return Result{Error: "STDIO server command is required"}
		}

		// 检查命令是否存在
		_, err := exec.LookPath(command)
		if err != nil {
			return Result{Error: fmt.Sprintf("Command not found: %s", command)}
		}

		// 尝试启动进程并立即关闭
		cmd := exec.CommandContext(ctx, command, args...)

		// 获取 stdin 以便发送初始化消息
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return Result{Error: fmt.Sprintf("Failed to create stdin pipe: %v", err)}
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return Result{Error: fmt.Sprintf("Failed to create stdout pipe: %v", err)}
		}

		// 启动进程
		if err := cmd.Start(); err != nil {
			return Result{Error: fmt.Sprintf("Failed to start process: %v", err)}
		}

		// 发送 MCP 初始化消息
		initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"ChYing","version":"1.0.0"}}}` + "\n"
		_, err = stdin.Write([]byte(initMsg))
		if err != nil {
			cmd.Process.Kill()
			return Result{Error: fmt.Sprintf("Failed to send init message: %v", err)}
		}

		// 读取响应（带超时）
		responseChan := make(chan string, 1)
		errorChan := make(chan error, 1)
		go func() {
			buf := make([]byte, 4096)
			n, err := stdout.Read(buf)
			if err != nil {
				errorChan <- err
				return
			}
			responseChan <- string(buf[:n])
		}()

		select {
		case response := <-responseChan:
			cmd.Process.Kill()
			// 检查是否是有效的 JSON-RPC 响应
			if strings.Contains(response, "jsonrpc") {
				return Result{Data: map[string]interface{}{
					"success": true,
					"message": "STDIO server connection successful",
				}}
			}
			return Result{Data: map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Process started, response: %s", response[:min(len(response), 200)]),
			}}
		case err := <-errorChan:
			cmd.Process.Kill()
			return Result{Error: fmt.Sprintf("Failed to read response: %v", err)}
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			return Result{Data: map[string]interface{}{
				"success": true,
				"message": "Process started successfully (no immediate response, which may be normal)",
			}}
		}

	} else {
		return Result{Error: fmt.Sprintf("Unknown server type: %s", serverType)}
	}
}

// ==================== A2A Agent API ====================

// A2AUpdateConfig 更新 A2A 配置
func (a *App) A2AUpdateConfig(agentMode string, a2aConfig *A2AConfigInput) Result {
	appConfig := conf.GetAppConfig()

	// 更新 Agent 模式
	appConfig.AI.AgentMode = agentMode

	// 更新 A2A 配置
	if a2aConfig != nil {
		appConfig.AI.A2A.Enabled = a2aConfig.Enabled
		appConfig.AI.A2A.AgentURL = a2aConfig.AgentURL
		appConfig.AI.A2A.Headers = a2aConfig.Headers
		appConfig.AI.A2A.Timeout = a2aConfig.Timeout
		appConfig.AI.A2A.EnableSSE = a2aConfig.EnableSSE

		// 更新全局 A2A 客户端配置
		if a2aClient != nil {
			a2aClient.Disconnect()
			a2aClient = nil
		}
	}

	// 保存配置到文件
	if err := conf.SaveConfig(); err != nil {
		logging.Logger.Warnf("保存配置文件失败: %v", err)
		return Result{Error: fmt.Sprintf("Failed to save config: %v", err)}
	}

	return Result{Data: "A2A config updated"}
}

// A2ATestConnection 测试 A2A Agent 连接
func (a *App) A2ATestConnection(agentURL string, headers map[string]string) Result {
	if agentURL == "" {
		return Result{Error: "Agent URL is required"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建临时客户端测试连接
	client := claudecode.NewA2AClient(&claudecode.A2AClientConfig{
		AgentURL: agentURL,
		Headers:  headers,
	})

	if err := client.TestConnection(ctx); err != nil {
		return Result{Error: fmt.Sprintf("Connection failed: %v", err)}
	}

	// 获取 Agent 信息
	if err := client.Connect(ctx); err != nil {
		return Result{Error: fmt.Sprintf("Failed to connect: %v", err)}
	}
	defer client.Disconnect()

	info := client.GetAgentInfo()

	return Result{Data: map[string]interface{}{
		"success": true,
		"message": "A2A Agent connection successful",
		"agent":   info,
	}}
}

// A2AGetAgentInfo 获取 A2A Agent 信息
func (a *App) A2AGetAgentInfo() Result {
	client := getA2AClient()

	if !client.IsConnected() {
		// 尝试连接
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := client.Connect(ctx); err != nil {
			return Result{Error: fmt.Sprintf("Failed to connect to A2A agent: %v", err)}
		}
	}

	info := client.GetAgentInfo()
	if info == nil {
		return Result{Error: "No agent info available"}
	}

	return Result{Data: info}
}

// A2ADisconnect 断开 A2A 连接
func (a *App) A2ADisconnect() Result {
	if a2aClient != nil {
		a2aClient.Disconnect()
	}
	return Result{Data: "A2A disconnected"}
}

// ==================== Claude CLI Settings API ====================

// ClaudeGetCLISettings 获取 Claude CLI 的 settings.json 内容
// 读取 ~/.claude/settings.json 文件
func (a *App) ClaudeGetCLISettings() Result {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return Result{Error: fmt.Sprintf("Failed to get home directory: %v", err)}
		}
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// 检查文件是否存在
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// 文件不存在，返回空配置
		return Result{Data: map[string]any{
			"path":     settingsPath,
			"exists":   false,
			"settings": map[string]any{},
		}}
	}

	// 读取文件内容
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return Result{Error: fmt.Sprintf("Failed to read settings file: %v", err)}
	}

	// 解析 JSON
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return Result{Error: fmt.Sprintf("Failed to parse settings JSON: %v", err)}
	}

	return Result{Data: map[string]any{
		"path":     settingsPath,
		"exists":   true,
		"settings": settings,
		"raw":      string(data),
	}}
}

// ClaudeUpdateCLISettings 更新 Claude CLI 的 settings.json 内容
// 写入 ~/.claude/settings.json 文件
func (a *App) ClaudeUpdateCLISettings(settingsJSON string) Result {
	// 先验证 JSON 格式
	var settings map[string]any
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return Result{Error: fmt.Sprintf("Invalid JSON format: %v", err)}
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return Result{Error: fmt.Sprintf("Failed to get home directory: %v", err)}
		}
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// 确保 .claude 目录存在
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return Result{Error: fmt.Sprintf("Failed to create .claude directory: %v", err)}
	}

	// 格式化 JSON（美化输出）
	formattedJSON, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return Result{Error: fmt.Sprintf("Failed to format JSON: %v", err)}
	}

	// 写入文件
	if err := os.WriteFile(settingsPath, formattedJSON, 0644); err != nil {
		return Result{Error: fmt.Sprintf("Failed to write settings file: %v", err)}
	}

	logging.Logger.Infof("Claude CLI settings updated: %s", settingsPath)

	return Result{Data: map[string]any{
		"path":    settingsPath,
		"message": "Settings saved successfully",
	}}
}

// ClaudeValidateCLISettings 验证 Claude CLI settings JSON 格式
func (a *App) ClaudeValidateCLISettings(settingsJSON string) Result {
	var settings map[string]any
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return Result{Error: fmt.Sprintf("Invalid JSON: %v", err)}
	}

	return Result{Data: map[string]any{
		"valid":   true,
		"message": "JSON format is valid",
	}}
}

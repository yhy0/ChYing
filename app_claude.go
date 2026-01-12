package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
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

// getClaudeClient 获取或创建 Claude Code 客户端
func getClaudeClient() *claudecode.Client {
	if claudeClient == nil {
		claudeClient = claudecode.NewClient(nil)
	}
	return claudeClient
}

// ClaudeInitialize 初始化 Claude Code 服务
func (a *App) ClaudeInitialize() Result {
	client := getClaudeClient()

	// 从配置获取设置
	appConfig := conf.GetAppConfig()
	config := &claudecode.Config{
		CLIPath:            appConfig.AI.Claude.CLIPath,
		WorkDir:            appConfig.AI.Claude.WorkDir,
		Model:              appConfig.AI.Claude.Model,
		MaxTurns:           appConfig.AI.Claude.MaxTurns,
		SystemPrompt:       appConfig.AI.Claude.SystemPrompt,
		AllowedTools:       appConfig.AI.Claude.AllowedTools,
		DisallowedTools:    appConfig.AI.Claude.DisallowedTools,
		PermissionMode:     appConfig.AI.Claude.PermissionMode,
		RequireToolConfirm: appConfig.AI.Claude.RequireToolConfirm,
		APIKey:             appConfig.AI.Claude.APIKey,
		BaseURL:            appConfig.AI.Claude.BaseURL,
		Temperature:        appConfig.AI.Claude.Temperature,
	}

	// 更新配置
	client.UpdateConfig(config)

	// 初始化客户端
	if err := client.Initialize(); err != nil {
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

	return Result{Data: map[string]interface{}{
		"session_id": session.ID,
		"project_id": session.ProjectID,
		"created_at": session.CreatedAt,
		"updated_at": session.UpdatedAt,
		"context":    session.Context,
		"history":    session.History,
	}}
}

// ClaudeGetSessionHistory 获取会话历史
func (a *App) ClaudeGetSessionHistory(sessionID string) Result {
	client := getClaudeClient()
	session := client.GetSession(sessionID)
	if session == nil {
		return Result{Error: "Session not found"}
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
	client := getClaudeClient()
	if wailsApp == nil {
		return Result{Error: "Wails app not set"}
	}

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

	// 转换外部 MCP 服务器配置
	externalServers := make([]map[string]interface{}, len(appConfig.AI.Claude.MCP.ExternalServers))
	for i, s := range appConfig.AI.Claude.MCP.ExternalServers {
		externalServers[i] = map[string]interface{}{
			"id":          s.ID,
			"name":        s.Name,
			"type":        s.Type,
			"enabled":     s.Enabled,
			"description": s.Description,
			"url":         s.URL,
			"headers":     s.Headers,
			"command":     s.Command,
			"args":        s.Args,
			"env":         s.Env,
		}
	}

	return Result{Data: map[string]interface{}{
		"cli_path":             appConfig.AI.Claude.CLIPath,
		"work_dir":             appConfig.AI.Claude.WorkDir,
		"model":                appConfig.AI.Claude.Model,
		"max_turns":            appConfig.AI.Claude.MaxTurns,
		"system_prompt":        appConfig.AI.Claude.SystemPrompt,
		"allowed_tools":        appConfig.AI.Claude.AllowedTools,
		"disallowed_tools":     appConfig.AI.Claude.DisallowedTools,
		"permission_mode":      appConfig.AI.Claude.PermissionMode,
		"require_tool_confirm": appConfig.AI.Claude.RequireToolConfirm,
		"api_key":              appConfig.AI.Claude.APIKey,
		"base_url":             appConfig.AI.Claude.BaseURL,
		"temperature":          appConfig.AI.Claude.Temperature,
		"mcp": map[string]interface{}{
			"enabled":          appConfig.AI.Claude.MCP.Enabled,
			"mode":             appConfig.AI.Claude.MCP.Mode,
			"port":             appConfig.AI.Claude.MCP.Port,
			"enabled_tools":    appConfig.AI.Claude.MCP.EnabledTools,
			"disabled_tools":   appConfig.AI.Claude.MCP.DisabledTools,
			"external_servers": externalServers,
		},
	}}
}

// MCPConfigInput MCP 配置输入结构
type MCPConfigInput struct {
	Enabled         bool                     `json:"enabled"`
	Mode            string                   `json:"mode"`
	Port            int                      `json:"port"`
	EnabledTools    []string                 `json:"enabled_tools"`
	DisabledTools   []string                 `json:"disabled_tools"`
	ExternalServers []ExternalMCPServerInput `json:"external_servers"`
}

// ExternalMCPServerInput 外部 MCP 服务器输入结构
type ExternalMCPServerInput struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Enabled     bool              `json:"enabled"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         []string          `json:"env"`
}

// ClaudeUpdateConfig 更新 Claude Code 配置
func (a *App) ClaudeUpdateConfig(cliPath string, workDir string, model string, maxTurns int, systemPrompt string, permissionMode string, requireToolConfirm bool, apiKey string, baseURL string, temperature float64, mcpConfig *MCPConfigInput) Result {
	client := getClaudeClient()
	config := &claudecode.Config{
		CLIPath:            cliPath,
		WorkDir:            workDir,
		Model:              model,
		MaxTurns:           maxTurns,
		SystemPrompt:       systemPrompt,
		PermissionMode:     permissionMode,
		RequireToolConfirm: requireToolConfirm,
		APIKey:             apiKey,
		BaseURL:            baseURL,
		Temperature:        temperature,
	}

	// 同时更新应用配置文件
	appConfig := conf.GetAppConfig()
	appConfig.AI.Claude.CLIPath = cliPath
	appConfig.AI.Claude.WorkDir = workDir
	appConfig.AI.Claude.Model = model
	appConfig.AI.Claude.MaxTurns = maxTurns
	appConfig.AI.Claude.SystemPrompt = systemPrompt
	appConfig.AI.Claude.PermissionMode = permissionMode
	appConfig.AI.Claude.RequireToolConfirm = requireToolConfirm
	appConfig.AI.Claude.APIKey = apiKey
	appConfig.AI.Claude.BaseURL = baseURL
	appConfig.AI.Claude.Temperature = temperature

	// 更新 MCP 配置
	if mcpConfig != nil {
		appConfig.AI.Claude.MCP.Enabled = mcpConfig.Enabled
		appConfig.AI.Claude.MCP.Mode = mcpConfig.Mode
		appConfig.AI.Claude.MCP.Port = mcpConfig.Port
		appConfig.AI.Claude.MCP.EnabledTools = mcpConfig.EnabledTools
		appConfig.AI.Claude.MCP.DisabledTools = mcpConfig.DisabledTools

		// 更新外部 MCP 服务器配置
		if mcpConfig.ExternalServers != nil {
			externalServers := make([]conf.ExternalMCPServer, len(mcpConfig.ExternalServers))
			for i, s := range mcpConfig.ExternalServers {
				externalServers[i] = conf.ExternalMCPServer{
					ID:          s.ID,
					Name:        s.Name,
					Type:        s.Type,
					Enabled:     s.Enabled,
					Description: s.Description,
					URL:         s.URL,
					Headers:     s.Headers,
					Command:     s.Command,
					Args:        s.Args,
					Env:         s.Env,
				}
			}
			appConfig.AI.Claude.MCP.ExternalServers = externalServers
		}

		// 同步到 claudecode.Config
		config.BuiltinMCP = claudecode.BuiltinMCPConfig{
			Enabled:       mcpConfig.Enabled,
			Mode:          mcpConfig.Mode,
			Port:          mcpConfig.Port,
			EnabledTools:  mcpConfig.EnabledTools,
			DisabledTools: mcpConfig.DisabledTools,
		}
	}

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
// 发送一个简单的测试消息来验证 API 配置是否正确
func (a *App) ClaudeTestConnection() Result {
	client := getClaudeClient()

	// 从配置获取设置
	appConfig := conf.GetAppConfig()
	config := &claudecode.Config{
		CLIPath:     appConfig.AI.Claude.CLIPath,
		WorkDir:     appConfig.AI.Claude.WorkDir,
		Model:       appConfig.AI.Claude.Model,
		APIKey:      appConfig.AI.Claude.APIKey,
		BaseURL:     appConfig.AI.Claude.BaseURL,
		Temperature: appConfig.AI.Claude.Temperature,
	}

	// 更新配置
	client.UpdateConfig(config)

	// 测试连接
	ctx := context.Background()
	if err := client.TestConnection(ctx); err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "LLM connection test successful"}
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

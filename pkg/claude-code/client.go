package claudecode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/10
   @desc Claude Code CLI 客户端封装
**/

// Client Claude Code CLI 客户端
type Client struct {
	config         *Config
	sessionManager *SessionManager
	mcpServer      *MCPServer
	mu             sync.RWMutex
	initialized    bool

	// 当前运行的 CLI 进程
	activeProcesses map[string]*CLIProcess
	processMu       sync.RWMutex
}

// CLIProcess CLI 进程
type CLIProcess struct {
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	cancel    context.CancelFunc
	sessionID string
}

// 全局客户端实例
var (
	globalClient *Client
	clientOnce   sync.Once
)

// GetClient 获取全局客户端实例
func GetClient() *Client {
	clientOnce.Do(func() {
		globalClient = &Client{
			config:          &Config{},
			sessionManager:  NewSessionManager(),
			activeProcesses: make(map[string]*CLIProcess),
		}
	})
	return globalClient
}

// NewClient 创建新的客户端实例
func NewClient(config *Config) *Client {
	client := &Client{
		sessionManager:  NewSessionManager(),
		activeProcesses: make(map[string]*CLIProcess),
	}
	if config != nil {
		client.config = config
	} else {
		client.config = &Config{}
	}
	return client
}

// UpdateConfig 更新配置
func (c *Client) UpdateConfig(config *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if config != nil {
		// 应用默认值，防止空值覆盖
		if config.CLIPath == "" {
			config.CLIPath = "claude"
		}
		if config.Model == "" {
			config.Model = "claude-sonnet-4"
		}
		if config.MaxTurns == 0 {
			config.MaxTurns = 50
		}
		if config.PermissionMode == "" {
			config.PermissionMode = "default"
		}
		if config.Timeout == 0 {
			config.Timeout = 300
		}
		c.config = config
	}
}

// Initialize 初始化客户端
func (c *Client) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	config := c.config
	if config == nil {
		config = &Config{}
		c.config = config
	}

	// 设置默认值
	if config.CLIPath == "" {
		config.CLIPath = "claude"
	}
	if config.Model == "" {
		config.Model = "claude-sonnet-4"
	}
	if config.MaxTurns == 0 {
		config.MaxTurns = 50
	}
	if config.PermissionMode == "" {
		config.PermissionMode = "default"
	}
	if config.Timeout == 0 {
		config.Timeout = 300 // 默认 5 分钟超时
	}

	// 检查 CLI 是否可用
	if err := c.checkCLI(); err != nil {
		return fmt.Errorf("Claude CLI not available: %v", err)
	}

	// 启动 MCP Server（如果启用）
	if config.BuiltinMCP.Enabled {
		c.mcpServer = NewMCPServerWithConfig(&config.BuiltinMCP)
		if err := c.mcpServer.Start(); err != nil {
			logging.Logger.Warnf("Failed to start MCP server: %v", err)
			// MCP Server 启动失败不阻止初始化
		}
	} else {
		logging.Logger.Info("MCP Server disabled by configuration")
	}

	// 从数据库加载历史会话
	c.sessionManager.LoadAllSessionsFromDB()

	c.initialized = true
	logging.Logger.Info("Claude Code client initialized successfully")
	return nil
}

// checkCLI 检查 CLI 是否可用
func (c *Client) checkCLI() error {
	cmd := exec.Command(c.config.CLIPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	logging.Logger.Infof("Claude CLI version: %s", string(output))
	return nil
}

// IsInitialized 检查是否已初始化
func (c *Client) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// CreateSession 创建会话
func (c *Client) CreateSession(projectID string, ctx *AgentContext) *Session {
	if !c.IsInitialized() {
		// 如果未初始化，尝试初始化
		if err := c.Initialize(); err != nil {
			logging.Logger.Warnf("Failed to initialize client: %v", err)
		}
	}
	return c.sessionManager.CreateSession(projectID, ctx)
}

// GetSession 获取会话
func (c *Client) GetSession(sessionID string) *Session {
	return c.sessionManager.GetSession(sessionID)
}

// DeleteSession 删除会话
func (c *Client) DeleteSession(sessionID string) {
	// 先停止可能运行的进程
	c.StopProcess(sessionID)
	c.sessionManager.DeleteSession(sessionID)
}

// ClearSession 清除会话
func (c *Client) ClearSession(sessionID string) {
	c.StopProcess(sessionID)
	c.sessionManager.ClearSession(sessionID)
}

// ListSessions 列出所有会话
func (c *Client) ListSessions() []*Session {
	return c.sessionManager.ListSessions()
}

// ListSessionsByProject 列出项目的所有会话
func (c *Client) ListSessionsByProject(projectID string) []*Session {
	return c.sessionManager.ListSessionsByProject(projectID)
}

// ListProjects 列出所有有会话的项目
func (c *Client) ListProjects() []string {
	return c.sessionManager.ListProjects()
}

// GetSessionHistory 获取会话历史
func (c *Client) GetSessionHistory(sessionID string) ([]ChatMessage, error) {
	session := c.sessionManager.GetSession(sessionID)
	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session.History, nil
}

// UpdateContext 更新会话上下文
func (c *Client) UpdateContext(sessionID string, ctx *AgentContext) error {
	session := c.sessionManager.GetSession(sessionID)
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	c.sessionManager.UpdateContext(sessionID, ctx)
	return nil
}

// SendMessage 发送消息（流式）
func (c *Client) SendMessage(ctx context.Context, sessionID, message string, eventChan chan<- StreamEvent) error {
	logging.Logger.Infof("SendMessage called: sessionID=%s, message=%s", sessionID, message)

	if !c.IsInitialized() {
		return fmt.Errorf("client not initialized")
	}

	session := c.sessionManager.GetSession(sessionID)
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// 添加用户消息到历史
	c.sessionManager.AddUserMessage(sessionID, message)

	// 构建 CLI 参数
	args := c.buildCLIArgs(session, message)
	logging.Logger.Infof("CLI args: %v", args)

	// 创建带超时的上下文
	timeout := time.Duration(c.config.Timeout) * time.Second
	procCtx, cancel := context.WithTimeout(ctx, timeout)

	// 创建 CLI 进程
	cmd := exec.CommandContext(procCtx, c.config.CLIPath, args...)
	logging.Logger.Infof("Starting CLI: %s", c.config.CLIPath)
	logging.Logger.Debugf("CLI args: %v", args)

	// 设置工作目录
	if c.config.WorkDir != "" {
		cmd.Dir = c.config.WorkDir
	}

	// 设置环境变量 - 映射模型名称
	cmd.Env = c.buildEnv()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// 保存进程引用
	process := &CLIProcess{
		cmd:       cmd,
		stdout:    stdout,
		stderr:    stderr,
		cancel:    cancel,
		sessionID: sessionID,
	}
	c.setProcess(sessionID, process)

	// 启动进程
	if err := cmd.Start(); err != nil {
		cancel()
		c.removeProcess(sessionID)
		return fmt.Errorf("failed to start CLI: %v", err)
	}
	logging.Logger.Infof("CLI process started, PID: %d", cmd.Process.Pid)

	// 启动 goroutine 处理输出
	go c.handleOutput(procCtx, sessionID, stdout, stderr, eventChan)

	// 启动 goroutine 等待进程结束
	go func() {
		cmd.Wait()
		c.removeProcess(sessionID)
	}()

	return nil
}

// buildCLIArgs 构建 CLI 参数
func (c *Client) buildCLIArgs(session *Session, message string) []string {
	// Claude CLI 不直接支持第三方模型名，需要通过环境变量映射
	// 使用 haiku 别名，通过 ANTHROPIC_DEFAULT_HAIKU_MODEL 指定实际模型
	modelToUse := "haiku"

	args := []string{
		"--output-format", "stream-json",
		"--verbose",
		"--model", modelToUse,
		"--max-turns", fmt.Sprintf("%d", c.config.MaxTurns),
	}

	// 如果有 Claude Code 会话 ID，使用 --resume 恢复会话
	// 注意：恢复会话时不需要重新设置 system-prompt
	if session.ConversationID != "" {
		args = append(args, "--resume", session.ConversationID)
	} else {
		// 新会话：添加系统提示词
		if c.config.SystemPrompt != "" {
			args = append(args, "--system-prompt", c.config.SystemPrompt)
		} else {
			// 使用默认的安全分析提示词
			args = append(args, "--system-prompt", c.buildDefaultSystemPrompt(session.Context))
		}
	}

	// 添加允许的工具
	if len(c.config.AllowedTools) > 0 {
		for _, tool := range c.config.AllowedTools {
			args = append(args, "--allowed-tools", tool)
		}
	}

	// 添加禁用的工具 - 默认禁用 Bash 等危险工具
	disallowedTools := c.config.DisallowedTools
	if len(disallowedTools) == 0 {
		// 默认禁用的内置工具
		disallowedTools = []string{
			"Bash",         // 禁用 Bash 命令执行
			"Write",        // 禁用文件写入
			"Edit",         // 禁用文件编辑
			"NotebookEdit", // 禁用 Notebook 编辑
		}
	}
	for _, tool := range disallowedTools {
		args = append(args, "--disallowed-tools", tool)
	}

	// 添加 MCP 服务器
	if c.mcpServer != nil && c.mcpServer.IsRunning() {
		args = append(args, "--mcp-server", fmt.Sprintf("chying:%s", c.mcpServer.GetURL()))
	}

	// 添加用户消息作为最后一个参数（使用 --print 模式，消息直接作为参数）
	// 注意：使用 "--" 分隔符确保消息内容不会被解析为命令行选项
	args = append(args, "--print", "--", c.buildPromptWithContext(message, session.Context))

	return args
}

// buildEnv 构建环境变量
func (c *Client) buildEnv() []string {
	// 继承当前环境变量
	env := os.Environ()

	// 获取配置的模型名称
	model := c.config.Model
	if model == "" {
		model = "claude-sonnet-4"
	}

	// Claude CLI 不直接支持第三方模型名，需要通过环境变量映射
	// 将所有别名都映射到用户配置的模型
	env = append(env,
		fmt.Sprintf("ANTHROPIC_DEFAULT_OPUS_MODEL=%s", model),
		fmt.Sprintf("ANTHROPIC_DEFAULT_SONNET_MODEL=%s", model),
		fmt.Sprintf("ANTHROPIC_DEFAULT_HAIKU_MODEL=%s", model),
		fmt.Sprintf("CLAUDE_CODE_SUBAGENT_MODEL=%s", model),
	)

	return env
}

// buildDefaultSystemPrompt 构建默认系统提示词
func (c *Client) buildDefaultSystemPrompt(ctx *AgentContext) string {
	prompt := `You are an AI security analyst assistant integrated into ChYing, a comprehensive web application security testing platform similar to Burp Suite.

## Your Role
You help security researchers and penetration testers analyze targets, identify vulnerabilities, and suggest testing strategies.

## Your Capabilities
You have access to ChYing's security tools through MCP:
- get_http_history: Retrieve HTTP traffic history from the proxy
- get_http_detail: Get detailed request/response for a specific traffic record
- get_vulnerabilities: List discovered vulnerabilities
- get_fingerprints: Get technology fingerprints (frameworks, servers, languages)
- get_site_map: Get the site map of discovered URLs
- get_collection_info: Get collected information (domains, IPs, emails, APIs, etc.)
- send_http_request: Send a custom HTTP request (use with caution)
- run_scan: Execute a security scan (use with caution)
- analyze_target: Get a comprehensive analysis of a target host

## Guidelines
1. Always analyze available data before making conclusions
2. Explain your reasoning clearly and provide actionable recommendations
3. For potentially destructive operations, explain what you're about to do and why
4. Respond in the same language as the user's message
5. Be concise but thorough in your analysis
6. Prioritize findings by severity and exploitability
7. Suggest specific payloads or techniques when appropriate

## Security Testing Methodology
When analyzing a target, consider:
- Input validation vulnerabilities (XSS, SQL injection, command injection)
- Authentication and session management issues
- Access control flaws
- Information disclosure
- Business logic vulnerabilities
- Server misconfigurations
- Known CVEs based on fingerprints`

	if ctx != nil {
		prompt += fmt.Sprintf(`

## Current Context
Project ID: %s
Project Name: %s`, ctx.ProjectID, ctx.ProjectName)

		if len(ctx.SelectedTrafficIDs) > 0 {
			prompt += fmt.Sprintf(`
Selected Traffic Records: %d items`, len(ctx.SelectedTrafficIDs))
		}

		if len(ctx.SelectedVulnIDs) > 0 {
			prompt += fmt.Sprintf(`
Selected Vulnerabilities: %d items`, len(ctx.SelectedVulnIDs))
		}

		if ctx.CustomData != "" {
			prompt += fmt.Sprintf(`

## Additional Context from User
%s`, ctx.CustomData)
		}
	}

	return prompt
}

// buildPromptWithContext 构建带上下文的提示词
func (c *Client) buildPromptWithContext(message string, ctx *AgentContext) string {
	// 如果有选中的流量或漏洞，添加到消息中
	if ctx != nil && (len(ctx.SelectedTrafficIDs) > 0 || len(ctx.SelectedVulnIDs) > 0) {
		prefix := ""
		if len(ctx.SelectedTrafficIDs) > 0 {
			prefix += fmt.Sprintf("[Selected %d traffic records] ", len(ctx.SelectedTrafficIDs))
		}
		if len(ctx.SelectedVulnIDs) > 0 {
			prefix += fmt.Sprintf("[Selected %d vulnerabilities] ", len(ctx.SelectedVulnIDs))
		}
		return prefix + message
	}
	return message
}

// handleOutput 处理 CLI 输出
func (c *Client) handleOutput(ctx context.Context, sessionID string, stdout, stderr io.ReadCloser, eventChan chan<- StreamEvent) {
	logging.Logger.Info("handleOutput started")
	defer func() {
		logging.Logger.Info("handleOutput finished, closing eventChan")
		close(eventChan)
	}()

	var textContent string
	var toolUses []ToolUse

	scanner := bufio.NewScanner(stdout)
	// 增加缓冲区大小以处理大输出
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineCount := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			logging.Logger.Info("Context done, exiting handleOutput")
			return
		default:
		}

		line := scanner.Text()
		lineCount++
		logging.Logger.Debugf("CLI output line %d: %s", lineCount, line)
		if line == "" {
			continue
		}

		var msg CLIOutputMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			logging.Logger.Warnf("Failed to parse CLI output: %v, line: %s", err, line)
			continue
		}
		logging.Logger.Debugf("Parsed message type: %s", msg.Type)

		event := c.processOutputMessage(sessionID, &msg, &textContent, &toolUses)
		if event != nil {
			logging.Logger.Debugf("Sending event: type=%s", event.Type)
			eventChan <- *event
		}
	}
	logging.Logger.Infof("Scanner finished, read %d lines", lineCount)

	if err := scanner.Err(); err != nil {
		logging.Logger.Errorf("Scanner error: %v", err)
		eventChan <- StreamEvent{
			Type:      "error",
			Error:     err.Error(),
			SessionID: sessionID,
		}
	}

	// 读取 stderr 并处理错误信息
	stderrData, _ := io.ReadAll(stderr)
	if len(stderrData) > 0 {
		stderrStr := string(stderrData)
		logging.Logger.Debugf("CLI stderr: %s", stderrStr)

		// 提取真正的错误信息，过滤掉调试日志
		errorMsg := extractErrorFromStderr(stderrStr)
		if errorMsg != "" {
			eventChan <- StreamEvent{
				Type:      "error",
				Error:     errorMsg,
				SessionID: sessionID,
			}
		}
	}

	// 保存助手消息到历史
	if textContent != "" || len(toolUses) > 0 {
		c.sessionManager.AddAssistantMessage(sessionID, textContent, toolUses)
	}

	eventChan <- StreamEvent{
		Type:      "done",
		SessionID: sessionID,
	}
}

// processOutputMessage 处理输出消息
func (c *Client) processOutputMessage(sessionID string, msg *CLIOutputMessage, textContent *string, toolUses *[]ToolUse) *StreamEvent {
	switch msg.Type {
	case "system":
		// 系统消息，可能包含会话 ID（从 init 子类型中获取）
		if msg.SessionID != "" && msg.Subtype == "init" {
			session := c.sessionManager.GetSession(sessionID)
			if session != nil && session.ConversationID == "" {
				session.ConversationID = msg.SessionID
				logging.Logger.Infof("Saved Claude Code session ID: %s", msg.SessionID)
				// 保存到数据库
				c.sessionManager.UpdateSession(session)
			}
		}
		return nil

	case "assistant":
		// 助手消息
		if msg.Message != nil {
			for _, block := range msg.Message.Content {
				switch block.Type {
				case "text":
					*textContent += block.Text
					return &StreamEvent{
						Type:      "text",
						Content:   block.Text,
						SessionID: sessionID,
					}
				case "tool_use":
					toolUse := ToolUse{
						ID:     block.ID,
						Name:   block.Name,
						Input:  block.Input,
						Status: "pending",
					}
					*toolUses = append(*toolUses, toolUse)
					return &StreamEvent{
						Type:      "tool_use",
						ToolUse:   &toolUse,
						SessionID: sessionID,
					}
				}
			}
		}
		return nil

	case "content_block_delta":
		// 流式文本增量
		if len(msg.Content) > 0 {
			var delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(msg.Content, &delta); err == nil && delta.Text != "" {
				*textContent += delta.Text
				return &StreamEvent{
					Type:      "text",
					Content:   delta.Text,
					SessionID: sessionID,
				}
			}
		}
		return nil

	case "tool_use":
		// 工具调用
		toolUse := ToolUse{
			ID:     msg.ToolUseID,
			Name:   msg.ToolName,
			Input:  msg.ToolInput,
			Status: "running",
		}
		*toolUses = append(*toolUses, toolUse)
		return &StreamEvent{
			Type:      "tool_use",
			ToolUse:   &toolUse,
			SessionID: sessionID,
		}

	case "tool_result":
		// 工具结果
		for i := range *toolUses {
			if (*toolUses)[i].ID == msg.ToolUseID {
				if msg.ToolError != "" {
					(*toolUses)[i].Status = "error"
					(*toolUses)[i].Error = msg.ToolError
				} else {
					(*toolUses)[i].Status = "completed"
					(*toolUses)[i].Result = msg.ToolResult
				}
				return &StreamEvent{
					Type:      "tool_result",
					ToolUse:   &(*toolUses)[i],
					SessionID: sessionID,
				}
			}
		}
		return nil

	case "cost":
		// 费用统计
		return &StreamEvent{
			Type:         "cost",
			SessionID:    sessionID,
			CostUSD:      msg.TotalCostUSD,
			InputTokens:  msg.TotalInputTokens,
			OutputTokens: msg.TotalOutputTokens,
		}

	case "result":
		// 结果消息（包含最终费用统计）
		// 同时检查是否有 session_id 需要保存
		if msg.SessionID != "" {
			session := c.sessionManager.GetSession(sessionID)
			if session != nil && session.ConversationID == "" {
				session.ConversationID = msg.SessionID
				logging.Logger.Infof("Saved Claude Code session ID from result: %s", msg.SessionID)
				c.sessionManager.UpdateSession(session)
			}
		}
		return &StreamEvent{
			Type:         "cost",
			SessionID:    sessionID,
			CostUSD:      msg.TotalCostUSD,
			InputTokens:  msg.TotalInputTokens,
			OutputTokens: msg.TotalOutputTokens,
		}

	case "error":
		// 错误
		var errorMsg string
		if len(msg.Content) > 0 {
			json.Unmarshal(msg.Content, &errorMsg)
		}
		return &StreamEvent{
			Type:      "error",
			Error:     errorMsg,
			SessionID: sessionID,
		}
	}

	return nil
}

// SendPermissionResponse 发送权限响应
// 注意：在 A 模式（--print）下，此方法不可用，因为没有 stdin 通道
// 保留此方法是为了 API 兼容性，但会返回错误
func (c *Client) SendPermissionResponse(sessionID, toolUseID string, allowed bool) error {
	return fmt.Errorf("permission response not supported in --print mode")
}

// StopProcess 停止进程
func (c *Client) StopProcess(sessionID string) {
	process := c.getProcess(sessionID)
	if process != nil {
		process.cancel()
		c.removeProcess(sessionID)
	}
}

// StopSession 停止会话（别名）
func (c *Client) StopSession(sessionID string) {
	c.StopProcess(sessionID)
}

// GetMCPServerURL 获取 MCP 服务器 URL
func (c *Client) GetMCPServerURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.mcpServer != nil && c.mcpServer.IsRunning() {
		return c.mcpServer.GetURL()
	}
	return ""
}

// setProcess 设置进程
func (c *Client) setProcess(sessionID string, process *CLIProcess) {
	c.processMu.Lock()
	defer c.processMu.Unlock()
	c.activeProcesses[sessionID] = process
}

// getProcess 获取进程
func (c *Client) getProcess(sessionID string) *CLIProcess {
	c.processMu.RLock()
	defer c.processMu.RUnlock()
	return c.activeProcesses[sessionID]
}

// removeProcess 移除进程
func (c *Client) removeProcess(sessionID string) {
	c.processMu.Lock()
	defer c.processMu.Unlock()
	delete(c.activeProcesses, sessionID)
}

// TestConnection 测试 LLM 连接
// 发送一个简单的测试消息来验证 API 配置是否正确
func (c *Client) TestConnection(ctx context.Context) error {
	if c.config == nil {
		return fmt.Errorf("client not configured")
	}

	// 设置默认值
	cliPath := c.config.CLIPath
	if cliPath == "" {
		cliPath = "claude"
	}

	// 构建测试参数 - 使用 --print 模式进行简单测试
	args := []string{
		"--output-format", "stream-json",
		"--print", // 非交互模式，直接输出结果
		"--model", "haiku",
		"--max-turns", "1",
		"Say 'OK' if you can hear me.", // 简单的测试消息
	}

	// 创建带超时的上下文（30秒超时）
	testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(testCtx, cliPath, args...)

	// 设置工作目录
	if c.config.WorkDir != "" {
		cmd.Dir = c.config.WorkDir
	}

	// 设置环境变量
	cmd.Env = c.buildEnv()

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 检查是否是超时
		if testCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("connection test timed out after 30 seconds")
		}
		return fmt.Errorf("LLM connection test failed: %v, output: %s", err, string(output))
	}

	// 检查输出是否包含有效响应
	outputStr := string(output)
	if outputStr == "" {
		return fmt.Errorf("LLM returned empty response")
	}

	logging.Logger.Infof("LLM connection test successful, response: %s", outputStr[:min(len(outputStr), 200)])
	return nil
}

// Shutdown 关闭客户端
func (c *Client) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 停止所有进程
	c.processMu.Lock()
	for sessionID, process := range c.activeProcesses {
		process.cancel()
		delete(c.activeProcesses, sessionID)
	}
	c.processMu.Unlock()

	// 停止 MCP Server
	if c.mcpServer != nil {
		c.mcpServer.Stop()
	}

	c.initialized = false
}

// extractErrorFromStderr 从 stderr 中提取真正的错误信息
// 过滤掉调试日志，只保留关键错误信息
func extractErrorFromStderr(stderr string) string {
	if stderr == "" {
		return ""
	}

	lines := strings.Split(stderr, "\n")
	var errorLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 跳过调试日志行（通常以时间戳或 DEBUG/INFO 开头）
		if strings.Contains(line, "[DEBUG]") ||
			strings.Contains(line, "[INFO]") ||
			strings.Contains(line, "DEBUG ") ||
			strings.Contains(line, "INFO ") {
			continue
		}

		// 提取 "Error" 开头的行或包含关键错误信息的行
		if strings.HasPrefix(line, "Error") ||
			strings.HasPrefix(line, "error:") ||
			strings.HasPrefix(line, "Error:") ||
			strings.Contains(line, "TypeError:") ||
			strings.Contains(line, "SyntaxError:") ||
			strings.Contains(line, "ReferenceError:") ||
			strings.Contains(line, "failed") ||
			strings.Contains(line, "Failed") {
			errorLines = append(errorLines, line)
		}
	}

	if len(errorLines) == 0 {
		// 如果没有找到明确的错误行，但 stderr 不为空
		// 检查是否整个内容都是错误信息（短内容）
		if len(stderr) < 500 && !strings.Contains(stderr, "[DEBUG]") {
			return strings.TrimSpace(stderr)
		}
		return ""
	}

	// 返回第一个错误信息（通常是最关键的）
	// 如果错误信息太长，截断它
	result := strings.Join(errorLines, "; ")
	if len(result) > 500 {
		result = result[:500] + "..."
	}
	return result
}

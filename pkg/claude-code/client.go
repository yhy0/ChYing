package claudecode

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	claude "github.com/yhy0/claude-agent-sdk-go"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/10
   @desc Claude Code SDK 客户端（使用 claude-agent-sdk-go）
**/

// Client Claude Code SDK 客户端
type Client struct {
	config         *Config
	sessionManager *SessionManager
	mu             sync.RWMutex
	initialized    bool
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
			config:         &Config{},
			sessionManager: NewSessionManager(),
		}
	})
	return globalClient
}

// NewClient 创建新的客户端实例
func NewClient() *Client {
	return &Client{
		sessionManager: NewSessionManager(),
		config:         &Config{},
	}
}

// UpdateConfig 更新配置
func (c *Client) UpdateConfig(config *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if config != nil {
		// 应用默认值
		if config.Model == "" {
			config.Model = "claude-3-5-sonnet-20241022"
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
func (c *Client) Initialize(config *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if config != nil {
		c.config = config
	}

	cfg := c.config
	if cfg == nil {
		cfg = &Config{}
		c.config = cfg
	}

	// 设置默认值
	if cfg.Model == "" {
		cfg.Model = "claude-3-5-sonnet-20241022"
	}
	if cfg.MaxTurns == 0 {
		cfg.MaxTurns = 50
	}
	if cfg.PermissionMode == "" {
		cfg.PermissionMode = "default"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 300
	}

	// 从数据库加载历史会话
	c.sessionManager.LoadAllSessionsFromDB()

	c.initialized = true
	logging.Logger.Info("Claude Code SDK client initialized successfully")
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
		if err := c.Initialize(nil); err != nil {
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
	c.sessionManager.DeleteSession(sessionID)
}

// ClearSession 清除会话
func (c *Client) ClearSession(sessionID string) {
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

	// 构建带上下文的提示词
	enhancedPrompt := c.buildPromptWithContext(message, session.Context)

	// 获取 Claude 会话 ID（用于复用会话）
	claudeSessionID := session.ConversationID

	// 启动 goroutine 处理消息
	go c.handleMessage(ctx, sessionID, claudeSessionID, enhancedPrompt, eventChan)

	return nil
}

// handleMessage 处理消息 - 使用 claude-agent-sdk-go 的 ClaudeSDKClient（流式模式）
// 注意：SDK MCP Server 需要使用 ClaudeSDKClient，因为工具发现通过双向控制协议进行
// claudeSessionID: Claude CLI 的会话 ID，用于复用会话上下文
func (c *Client) handleMessage(ctx context.Context, sessionID, claudeSessionID, message string, eventChan chan<- StreamEvent) {
	defer close(eventChan)

	c.mu.RLock()
	config := c.config
	c.mu.RUnlock()

	// 构建 SDK 选项
	opts := []claude.Option{
		claude.WithModel(config.Model),
		claude.WithMaxTurns(config.MaxTurns),
	}

	// 如果有之前的 Claude 会话 ID，使用 WithResume 恢复会话
	if claudeSessionID != "" {
		opts = append(opts, claude.WithResume(claudeSessionID))
		logging.Logger.Infof("Resuming Claude session: %s", claudeSessionID)
	}

	// 设置 CLI 路径（只有当用户指定了完整路径时才设置）
	// 如果是空字符串或 "claude"，让 SDK 自动从 PATH 查找
	if config.CLIPath != "" && config.CLIPath != "claude" {
		opts = append(opts, claude.WithCLIPath(config.CLIPath))
		logging.Logger.Infof("Using custom Claude CLI path: %s", config.CLIPath)
	}

	// 设置工作目录为当前项目目录
	// 项目目录格式: ~/.config/ChYing/db/{projectName}/
	// 使用 db.CurrentProjectName 获取当前项目名称（在数据库初始化时设置）
	if db.CurrentProjectName != "" {
		projectDir := filepath.Join(file.ChyingDir, "db", db.CurrentProjectName)
		opts = append(opts, claude.WithWorkingDir(projectDir))
		logging.Logger.Infof("Claude CLI working directory set to: %s", projectDir)
	}

	// 设置 setting sources: 加载用户全局设置 + 项目设置
	// - user: ~/.claude/settings.json (API Key、代理、全局权限等)
	// - project: {projectDir}/.claude/settings.json 和 CLAUDE.md (项目特定配置)
	opts = append(opts, claude.WithSettingSources(
		claude.SettingSourceUser,
		claude.SettingSourceProject,
	))

	// 添加权限模式
	if config.PermissionMode != "" {
		permMode := c.parsePermissionMode(config.PermissionMode)
		opts = append(opts, claude.WithPermissionMode(permMode))
	}

	// 添加系统提示词
	if config.SystemPrompt != "" {
		opts = append(opts, claude.WithSystemPrompt(config.SystemPrompt))
	}

	// 添加内置的 ChYing 安全工具 MCP Server
	chyingMCPServer := CreateChYingMCPServer()
	mcpServers := map[string]claude.MCPServerConfig{
		"chying": chyingMCPServer.ToConfig(),
	}
	opts = append(opts, claude.WithMCPServers(mcpServers))

	// 允许 ChYing MCP 工具（格式：mcp__<server-name>__<tool-name>）
	opts = append(opts, claude.WithAllowedTools(
		"mcp__chying__get_http_history",
		"mcp__chying__get_traffic_detail",
		"mcp__chying__get_vulnerabilities",
		"mcp__chying__send_http_request",
		"mcp__chying__analyze_request",
		"mcp__chying__search_traffic",
		"mcp__chying__get_sitemap",
		"mcp__chying__get_statistics",
	))

	// 创建 ClaudeSDKClient（流式模式，支持 SDK MCP Server 的工具发现）
	sdkClient := claude.NewClaudeSDKClient(opts...)

	// 连接到 Claude CLI
	logging.Logger.Infof("Connecting to Claude CLI for message: %s", message[:min(len(message), 100)])
	if err := sdkClient.Connect(ctx); err != nil {
		logging.Logger.Errorf("Failed to connect to Claude CLI: %v", err)
		eventChan <- StreamEvent{
			Type:      "error",
			Error:     fmt.Sprintf("Failed to connect to Claude CLI: %v", err),
			SessionID: sessionID,
		}
		return
	}
	defer sdkClient.Disconnect()

	// 发送查询
	if err := sdkClient.Query(ctx, message); err != nil {
		logging.Logger.Errorf("Failed to send query: %v", err)
		eventChan <- StreamEvent{
			Type:      "error",
			Error:     fmt.Sprintf("Failed to send query: %v", err),
			SessionID: sessionID,
		}
		return
	}

	logging.Logger.Info("Claude Query started, waiting for messages...")
	// 计算工作目录（用于 transcript path）
	workingDir := ""
	if db.CurrentProjectName != "" {
		workingDir = filepath.Join(file.ChyingDir, "db", db.CurrentProjectName)
	}

	// 处理消息流
	for msg := range sdkClient.ReceiveMessages() {
		select {
		case <-ctx.Done():
			logging.Logger.Warnf("Context cancelled: %v", ctx.Err())
			eventChan <- StreamEvent{
				Type:      "error",
				Error:     ctx.Err().Error(),
				SessionID: sessionID,
			}
			return
		default:
			// 转换 SDK 消息为流事件，并处理会话 ID 保存
			logging.Logger.Infof("Received message type: %T", msg)
			c.convertSDKMessageToEvent(msg, sessionID, workingDir, eventChan)
		}
	}

	// 消息通道关闭，处理完成
	logging.Logger.Info("Message processing completed")
}

// convertSDKMessageToEvent 将 SDK 消息转换为流事件
func (c *Client) convertSDKMessageToEvent(msg claude.Message, sessionID, workingDir string, eventChan chan<- StreamEvent) {
	switch m := msg.(type) {
	case *claude.UserMessage:
		for _, block := range m.Content {
			if textBlock, ok := block.(*claude.TextBlock); ok {
				eventChan <- StreamEvent{
					Type:      "text",
					Content:   textBlock.Text,
					SessionID: sessionID,
				}
			}
		}

	case *claude.AssistantMessage:
		for _, block := range m.Content {
			switch b := block.(type) {
			case *claude.TextBlock:
				eventChan <- StreamEvent{
					Type:      "text",
					Content:   b.Text,
					SessionID: sessionID,
				}
			case *claude.ToolUseBlock:
				eventChan <- StreamEvent{
					Type: "tool_use",
					ToolUse: &ToolUse{
						ID:     b.ID,
						Name:   b.Name,
						Input:  c.marshalInput(b.Input),
						Status: "pending",
					},
					SessionID: sessionID,
				}
			case *claude.ToolResultBlock:
				eventChan <- StreamEvent{
					Type: "tool_result",
					ToolUse: &ToolUse{
						ID:     b.ToolUseID,
						Status: "completed",
						Result: fmt.Sprintf("%v", b.Content),
					},
					SessionID: sessionID,
				}
			case *claude.ThinkingBlock:
				// 可选：处理思考块
				logging.Logger.Debugf("Thinking: %s", b.Thinking)
			}
		}

	case *claude.ResultMessage:
		// 保存 Claude 会话 ID 以便后续复用
		if m.SessionID != "" {
			c.sessionManager.SetConversationID(sessionID, m.SessionID)
			logging.Logger.Infof("Saved Claude session ID: %s for session: %s", m.SessionID, sessionID)

			// 计算并保存 transcript path
			if workingDir != "" {
				transcriptPath := ComputeTranscriptPath(workingDir, m.SessionID)
				if transcriptPath != "" {
					c.sessionManager.SetTranscriptPath(sessionID, transcriptPath)
					logging.Logger.Infof("Saved transcript path: %s for session: %s", transcriptPath, sessionID)
				}
			}
		}

		// 发送费用信息
		inputTokens := 0
		outputTokens := 0
		if m.Usage != nil {
			if it, ok := m.Usage["input_tokens"].(float64); ok {
				inputTokens = int(it)
			}
			if ot, ok := m.Usage["output_tokens"].(float64); ok {
				outputTokens = int(ot)
			}
		}
		costUSD := 0.0
		if m.TotalCostUSD != nil {
			costUSD = *m.TotalCostUSD
		}
		eventChan <- StreamEvent{
			Type:         "cost",
			CostUSD:      costUSD,
			SessionID:    sessionID,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
		}
		// ResultMessage 表示查询完成，发送 done 事件
		eventChan <- StreamEvent{
			Type:      "done",
			SessionID: sessionID,
		}

	case *claude.StreamEvent:
		// 处理流事件
		logging.Logger.Debugf("Stream event: %v", m.Event)
	}
}

// marshalInput 将输入转换为 JSON
func (c *Client) marshalInput(input map[string]any) []byte {
	data, err := json.Marshal(input)
	if err != nil {
		logging.Logger.Warnf("Failed to marshal input: %v", err)
		return []byte("{}")
	}
	return data
}

// parsePermissionMode 解析权限模式
func (c *Client) parsePermissionMode(mode string) claude.PermissionMode {
	switch strings.ToLower(mode) {
	case "acceptedits":
		return claude.PermissionModeAcceptEdits
	case "plan":
		return claude.PermissionModePlan
	case "bypasspermissions":
		return claude.PermissionModeBypassPermissions
	default:
		return claude.PermissionModeDefault
	}
}

// buildPromptWithContext 构建带上下文的提示词
func (c *Client) buildPromptWithContext(message string, ctx *AgentContext) string {
	if ctx == nil || (len(ctx.SelectedTrafficIDs) == 0 && len(ctx.SelectedVulnIDs) == 0) {
		return message
	}

	var contextParts []string

	// 处理选中的流量
	if len(ctx.SelectedTrafficIDs) > 0 {
		trafficContext := c.buildTrafficContext(ctx.SelectedTrafficIDs)
		if trafficContext != "" {
			contextParts = append(contextParts, trafficContext)
		}
	}

	// 处理选中的漏洞
	if len(ctx.SelectedVulnIDs) > 0 {
		contextParts = append(contextParts, fmt.Sprintf("[Selected %d vulnerabilities: %s]",
			len(ctx.SelectedVulnIDs), strings.Join(ctx.SelectedVulnIDs, ", ")))
	}

	if len(contextParts) == 0 {
		return message
	}

	return strings.Join(contextParts, "\n") + "\n\n" + message
}

// buildTrafficContext 构建流量上下文信息
func (c *Client) buildTrafficContext(trafficIDs []string) string {
	if len(trafficIDs) == 0 {
		return ""
	}

	// 转换 ID 为 int64
	var hids []int64
	for _, idStr := range trafficIDs {
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
			hids = append(hids, id)
		}
	}

	if len(hids) == 0 {
		return fmt.Sprintf("[Selected %d traffic records (IDs: %s)]",
			len(trafficIDs), strings.Join(trafficIDs, ", "))
	}

	// 从数据库获取流量摘要
	histories, err := db.GetHistoriesByHids(hids)
	if err != nil {
		logging.Logger.Warnf("Failed to get traffic histories: %v", err)
		return fmt.Sprintf("[Selected %d traffic records (IDs: %s)]",
			len(trafficIDs), strings.Join(trafficIDs, ", "))
	}

	if len(histories) == 0 {
		return fmt.Sprintf("[Selected %d traffic records (IDs: %s)]",
			len(trafficIDs), strings.Join(trafficIDs, ", "))
	}

	// 构建摘要信息
	var summaries []string
	for _, h := range histories {
		summary := fmt.Sprintf("  - ID:%d | %s %s%s | Status:%s",
			h.Hid, h.Method, h.Host, h.Path, h.Status)
		summaries = append(summaries, summary)
	}

	return fmt.Sprintf(`[Selected %d HTTP traffic records]
Traffic Summary (use get_traffic_detail tool with traffic_id to get full request/response):
%s`,
		len(histories), strings.Join(summaries, "\n"))
}

// StopSession 停止会话
func (c *Client) StopSession(sessionID string) error {
	// 占位符实现
	return nil
}

// TestConnection 测试 LLM 连接
func (c *Client) TestConnection(ctx context.Context) error {
	if c.config == nil {
		return fmt.Errorf("client not configured")
	}

	// 占位符实现
	logging.Logger.Info("LLM connection test (placeholder)")
	return nil
}

// Shutdown 关闭客户端
func (c *Client) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.initialized = false
}

// GetMCPServerURL 获取 MCP 服务器 URL（占位符实现）
func (c *Client) GetMCPServerURL() string {
	// 占位符实现 - 返回空字符串表示 MCP 服务器未运行
	return ""
}

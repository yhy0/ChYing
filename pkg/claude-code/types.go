package claudecode

import (
	"encoding/json"
	"time"
)

/**
   @author yhy
   @since 2026/01/10
   @desc Claude Code 类型定义（SDK 迁移版本）
**/

// ==================== 配置类型 ====================

// Config Claude Code 配置
type Config struct {
	// Claude CLI 配置
	CLIPath        string `json:"cli_path" yaml:"cli_path"`
	Model          string `json:"model" yaml:"model"`
	MaxTurns       int    `json:"max_turns" yaml:"max_turns"`
	SystemPrompt   string `json:"system_prompt" yaml:"system_prompt"`
	PermissionMode string `json:"permission_mode" yaml:"permission_mode"`

	// ChYing 内部使用
	Timeout int `json:"timeout" yaml:"timeout"`

	// 注意: API Key、代理、MCP 服务器等配置由 Claude CLI 从 ~/.claude/settings.json 读取
	// ChYing 会自动复用用户的 Claude CLI 配置
}

// ==================== 会话类型 ====================

// Session 会话
type Session struct {
	ID             string        `json:"id"`
	ProjectID      string        `json:"project_id"`
	Context        *AgentContext `json:"context"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	History        []ChatMessage `json:"history"`
	ConversationID string        `json:"conversation_id,omitempty"` // Claude Code 会话 ID
	TranscriptPath string        `json:"transcript_path,omitempty"` // Claude CLI transcript 文件路径
	TotalCost      float64       `json:"total_cost,omitempty"`      // 累计费用
}

// AgentContext 代理上下文
type AgentContext struct {
	ProjectID            string   `json:"project_id"`
	ProjectName          string   `json:"project_name"`
	SelectedTrafficIDs   []string `json:"selected_traffic_ids,omitempty"`
	SelectedVulnIDs      []string `json:"selected_vuln_ids,omitempty"`
	SelectedFingerprints []string `json:"selected_fingerprints,omitempty"`
	CustomData           string   `json:"custom_data,omitempty"`
	AutoCollect          bool     `json:"auto_collect"`
}

// ChatMessage 对话消息
type ChatMessage struct {
	ID           string    `json:"id"`
	Role         string    `json:"role"` // "user", "assistant", "system"
	Content      string    `json:"content"`
	Timestamp    time.Time `json:"timestamp"`
	ToolUses     []ToolUse `json:"tool_uses,omitempty"`
	CostUSD      float64   `json:"cost_usd,omitempty"`
	InputTokens  int       `json:"input_tokens,omitempty"`
	OutputTokens int       `json:"output_tokens,omitempty"`
}

// ToolUse 工具调用
type ToolUse struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	Status string          `json:"status"` // "pending", "running", "completed", "error"
	Result string          `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// ==================== 事件类型 ====================

// StreamEvent 流式事件（发送到前端）
type StreamEvent struct {
	Type         string   `json:"type"` // "text", "tool_use", "tool_result", "error", "done", "cost"
	Content      string   `json:"content,omitempty"`
	ToolUse      *ToolUse `json:"tool_use,omitempty"`
	Error        string   `json:"error,omitempty"`
	SessionID    string   `json:"session_id,omitempty"`
	CostUSD      float64  `json:"cost_usd,omitempty"`
	InputTokens  int      `json:"input_tokens,omitempty"`
	OutputTokens int      `json:"output_tokens,omitempty"`
}

// ==================== 请求响应类型 ====================

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	SessionID string        `json:"session_id"`
	Message   string        `json:"message"`
	Context   *AgentContext `json:"context,omitempty"`
}

// NewSessionRequest 创建会话请求
type NewSessionRequest struct {
	ProjectID string        `json:"project_id"`
	Context   *AgentContext `json:"context,omitempty"`
}

// ToolConfirmRequest 工具确认请求
type ToolConfirmRequest struct {
	SessionID string `json:"session_id"`
	ToolUseID string `json:"tool_use_id"`
	Confirmed bool   `json:"confirmed"`
}

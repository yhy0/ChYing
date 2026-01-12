package claudecode

import (
	"encoding/json"
	"time"
)

/**
   @author yhy
   @since 2026/01/10
   @desc Claude Code CLI 类型定义
**/

// ==================== 配置类型 ====================

// Config Claude Code 配置
type Config struct {
	// CLI 路径，默认使用 PATH 中的 claude
	CLIPath string `json:"cli_path" yaml:"cli_path"`
	// 工作目录
	WorkDir string `json:"work_dir" yaml:"work_dir"`
	// 模型
	Model string `json:"model" yaml:"model"`
	// 最大回合数
	MaxTurns int `json:"max_turns" yaml:"max_turns"`
	// 系统提示词
	SystemPrompt string `json:"system_prompt" yaml:"system_prompt"`
	// 允许的工具列表
	AllowedTools []string `json:"allowed_tools" yaml:"allowed_tools"`
	// 禁用的工具列表
	DisallowedTools []string `json:"disallowed_tools" yaml:"disallowed_tools"`
	// MCP 服务器配置（外部 MCP 服务器）
	MCPServers []MCPServerConfig `json:"mcp_servers" yaml:"mcp_servers"`
	// 内置 MCP 服务器配置
	BuiltinMCP BuiltinMCPConfig `json:"builtin_mcp" yaml:"builtin_mcp"`
	// 权限模式
	PermissionMode string `json:"permission_mode" yaml:"permission_mode"`
	// 是否需要工具确认
	RequireToolConfirm bool `json:"require_tool_confirm" yaml:"require_tool_confirm"`
	// CLI 超时时间（秒），默认 300 秒（5分钟）
	Timeout int `json:"timeout" yaml:"timeout"`
	// 环境变量配置
	APIKey      string  `json:"api_key" yaml:"api_key"`           // ANTHROPIC_API_KEY
	BaseURL     string  `json:"base_url" yaml:"base_url"`         // ANTHROPIC_BASE_URL
	Temperature float64 `json:"temperature" yaml:"temperature"`   // AI_TEMPERATURE
}

// BuiltinMCPConfig 内置 MCP 服务器配置
type BuiltinMCPConfig struct {
	// 是否启用内置 MCP 服务器
	Enabled bool `json:"enabled" yaml:"enabled"`
	// 运行模式: "sse" 或 "stdio"
	Mode string `json:"mode" yaml:"mode"`
	// SSE 模式端口（0 表示自动选择）
	Port int `json:"port" yaml:"port"`
	// 启用的工具列表（空表示全部启用）
	EnabledTools []string `json:"enabled_tools" yaml:"enabled_tools"`
	// 禁用的工具列表
	DisabledTools []string `json:"disabled_tools" yaml:"disabled_tools"`
}

// MCPServerConfig MCP 服务器配置
type MCPServerConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ==================== 会话类型 ====================

// Session 会话
type Session struct {
	ID           string         `json:"id"`
	ProjectID    string         `json:"project_id"`
	Context      *AgentContext  `json:"context"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	History      []ChatMessage  `json:"history"`
	ConversationID string       `json:"conversation_id,omitempty"` // Claude Code 会话 ID
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
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	ToolUses  []ToolUse `json:"tool_uses,omitempty"`
	// Claude Code 特有字段
	CostUSD      float64 `json:"cost_usd,omitempty"`
	InputTokens  int     `json:"input_tokens,omitempty"`
	OutputTokens int     `json:"output_tokens,omitempty"`
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

// ==================== CLI 输入输出类型 ====================

// CLIInputMessage CLI 输入消息 (stream-json 格式)
// Claude CLI stream-json 格式要求: {"type":"user","message":{"role":"user","content":"..."}}
type CLIInputMessage struct {
	Type             string                  `json:"type"`                        // "user" 或 "control"
	Message          *CLIInputMessageContent `json:"message,omitempty"`           // 消息内容
	PermissionResult *PermissionResult       `json:"permission_result,omitempty"` // 权限响应
}

// CLIInputMessageContent CLI 输入消息内容
type CLIInputMessageContent struct {
	Role    string `json:"role"`    // "user"
	Content string `json:"content"` // 消息内容
}

// PermissionResult 权限响应
type PermissionResult struct {
	ToolUseID string `json:"tool_use_id"`
	Allowed   bool   `json:"allowed"`
}

// CLIOutputMessage CLI 输出消息 (stream-json 格式)
type CLIOutputMessage struct {
	Type string `json:"type"`
	// 根据 type 不同，以下字段可能存在
	Subtype          string          `json:"subtype,omitempty"`
	SessionID        string          `json:"session_id,omitempty"`
	Message          *MessageContent `json:"message,omitempty"`
	Content          json.RawMessage `json:"content,omitempty"`
	ToolUseID        string          `json:"tool_use_id,omitempty"`
	ToolName         string          `json:"tool_name,omitempty"`
	ToolInput        json.RawMessage `json:"tool_input,omitempty"`
	ToolResult       string          `json:"tool_result,omitempty"`
	ToolError        string          `json:"tool_error,omitempty"`
	CostUSD          float64         `json:"cost_usd,omitempty"`
	InputTokens      int             `json:"input_tokens,omitempty"`
	OutputTokens     int             `json:"output_tokens,omitempty"`
	TotalCostUSD     float64         `json:"total_cost_usd,omitempty"`
	TotalInputTokens int             `json:"total_input_tokens,omitempty"`
	TotalOutputTokens int            `json:"total_output_tokens,omitempty"`
}

// MessageContent 消息内容
type MessageContent struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type  string          `json:"type"` // "text", "tool_use", "tool_result"
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// ==================== 事件类型 ====================

// StreamEvent 流式事件（发送到前端）
type StreamEvent struct {
	Type      string   `json:"type"` // "text", "tool_use", "tool_result", "error", "done", "cost"
	Content   string   `json:"content,omitempty"`
	ToolUse   *ToolUse `json:"tool_use,omitempty"`
	Error     string   `json:"error,omitempty"`
	SessionID string   `json:"session_id,omitempty"`
	// 费用统计
	CostUSD      float64 `json:"cost_usd,omitempty"`
	InputTokens  int     `json:"input_tokens,omitempty"`
	OutputTokens int     `json:"output_tokens,omitempty"`
}

// ==================== MCP 工具类型 ====================

// MCPTool MCP 工具定义
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolResult MCP 工具执行结果
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent MCP 内容
type MCPContent struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
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

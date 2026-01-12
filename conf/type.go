package conf

import (
	JieConf "github.com/yhy0/ChYing/pkg/Jie/conf"
)

/**
   @author yhy
   @since 2024/9/2
   @desc 配置结构体定义
   @note AppConfig 中的 Jie 相关配置直接引用 Jie 的类型定义，避免重复
**/

// Configure 代理配置 (保持向后兼容)
type Configure struct {
	ProxyPort    int      `json:"port"`
	Exclude      []*Scope `json:"exclude"`      // Exclude 排除显示的域名
	Include      []*Scope `json:"include"`      // 只允许某些域名
	FilterSuffix []string `json:"filterSuffix"` // 过滤后缀
}

// Scope 作用域配置 (保持向后兼容)
type Scope struct {
	Id      int    `json:"id"`
	Enabled bool   `json:"enabled"`
	Prefix  string `json:"prefix"`
	Regexp  bool   `json:"regexp"`
	Type    string `json:"type"`
}

// AIConfig AI 配置结构体
type AIConfig struct {
	Claude struct {
		// Claude Code CLI 配置
		CLIPath         string   `json:"cli_path" yaml:"cli_path" mapstructure:"cli_path"`                            // CLI 路径，默认使用 PATH 中的 claude
		WorkDir         string   `json:"work_dir" yaml:"work_dir" mapstructure:"work_dir"`                            // 工作目录
		Model           string   `json:"model" yaml:"model" mapstructure:"model"`                                     // 模型
		MaxTurns        int      `json:"max_turns" yaml:"max_turns" mapstructure:"max_turns"`                         // 最大回合数
		SystemPrompt    string   `json:"system_prompt" yaml:"system_prompt" mapstructure:"system_prompt"`             // 系统提示词
		AllowedTools    []string `json:"allowed_tools" yaml:"allowed_tools" mapstructure:"allowed_tools"`             // 允许的工具列表
		DisallowedTools []string `json:"disallowed_tools" yaml:"disallowed_tools" mapstructure:"disallowed_tools"`    // 禁用的工具列表
		PermissionMode  string   `json:"permission_mode" yaml:"permission_mode" mapstructure:"permission_mode"`       // 权限模式
		RequireToolConfirm bool  `json:"require_tool_confirm" yaml:"require_tool_confirm" mapstructure:"require_tool_confirm"` // 是否需要工具确认
		// 环境变量配置
		APIKey      string  `json:"api_key" yaml:"api_key" mapstructure:"api_key"`              // ANTHROPIC_API_KEY
		BaseURL     string  `json:"base_url" yaml:"base_url" mapstructure:"base_url"`           // ANTHROPIC_BASE_URL
		Temperature float64 `json:"temperature" yaml:"temperature" mapstructure:"temperature"` // AI_TEMPERATURE
		// MCP 服务器配置
		MCP MCPConfig `json:"mcp" yaml:"mcp" mapstructure:"mcp"`
	} `json:"claude" yaml:"claude" mapstructure:"claude"`
}

// MCPConfig MCP 服务器配置
type MCPConfig struct {
	Enabled       bool     `json:"enabled" yaml:"enabled" mapstructure:"enabled"`                         // 是否启用内置 MCP 服务器
	Mode          string   `json:"mode" yaml:"mode" mapstructure:"mode"`                                   // 运行模式: "sse" 或 "stdio"
	Port          int      `json:"port" yaml:"port" mapstructure:"port"`                                   // SSE 模式端口（0 表示自动选择）
	EnabledTools  []string `json:"enabled_tools" yaml:"enabled_tools" mapstructure:"enabled_tools"`       // 启用的工具列表（空表示全部启用）
	DisabledTools []string `json:"disabled_tools" yaml:"disabled_tools" mapstructure:"disabled_tools"`    // 禁用的工具列表
	// 外部 MCP 服务器配置
	ExternalServers []ExternalMCPServer `json:"external_servers" yaml:"external_servers" mapstructure:"external_servers"` // 外部 MCP 服务器列表
}

// ExternalMCPServer 外部 MCP 服务器配置
type ExternalMCPServer struct {
	ID          string            `json:"id" yaml:"id" mapstructure:"id"`                           // 唯一标识符
	Name        string            `json:"name" yaml:"name" mapstructure:"name"`                     // 显示名称
	Type        string            `json:"type" yaml:"type" mapstructure:"type"`                     // 类型: "sse" 或 "stdio"
	Enabled     bool              `json:"enabled" yaml:"enabled" mapstructure:"enabled"`            // 是否启用
	Description string            `json:"description" yaml:"description" mapstructure:"description"` // 描述
	// SSE 模式配置
	URL     string            `json:"url" yaml:"url" mapstructure:"url"`               // SSE 服务器 URL
	Headers map[string]string `json:"headers" yaml:"headers" mapstructure:"headers"`   // 自定义请求头（如认证）
	// STDIO 模式配置
	Command string   `json:"command" yaml:"command" mapstructure:"command"` // 命令路径
	Args    []string `json:"args" yaml:"args" mapstructure:"args"`          // 命令参数
	Env     []string `json:"env" yaml:"env" mapstructure:"env"`             // 环境变量 (KEY=VALUE 格式)
}

// AppConfig 应用完整配置结构体
// Jie 扫描相关配置直接引用 Jie 的类型定义，保持类型一致性
type AppConfig struct {
	Version string `json:"version" yaml:"version" mapstructure:"version"`

	// === ChYing 独有配置 ===

	// 代理配置
	Proxy struct {
		Host    string `json:"host" yaml:"host" mapstructure:"host"`
		Port    int    `json:"port" yaml:"port" mapstructure:"port"`
		Enabled bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	} `json:"proxy" yaml:"proxy" mapstructure:"proxy"`

	// AI 配置
	AI AIConfig `json:"ai" yaml:"ai" mapstructure:"ai"`

	// 扫描配置
	Scan struct {
		EnablePortScan bool `json:"enable_port_scan" yaml:"enable_port_scan" mapstructure:"enable_port_scan"`
		EnableDirScan  bool `json:"enable_dir_scan" yaml:"enable_dir_scan" mapstructure:"enable_dir_scan"`
		EnableVulnScan bool `json:"enable_vuln_scan" yaml:"enable_vuln_scan" mapstructure:"enable_vuln_scan"`
		Threads        int  `json:"threads" yaml:"threads" mapstructure:"threads"`
		Timeout        int  `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
		Parallel       int  `json:"parallel" yaml:"parallel" mapstructure:"parallel"` // 对一个网站同时扫描的最大 url 个数
	} `json:"scan" yaml:"scan" mapstructure:"scan"`

	// 日志配置
	Logging struct {
		Level string `json:"level" yaml:"level" mapstructure:"level"`
		File  string `json:"file" yaml:"file" mapstructure:"file"`
	} `json:"logging" yaml:"logging" mapstructure:"logging"`

	// === Jie 扫描配置 - 直接引用 Jie 的类型 ===

	// HTTP 发包配置
	Http JieConf.Http `json:"http" yaml:"http" mapstructure:"http"`

	// 插件配置
	Plugins JieConf.Plugins `json:"plugins" yaml:"plugins" mapstructure:"plugins"`

	// 反连平台配置
	Reverse JieConf.Reverse `json:"reverse" yaml:"reverse" mapstructure:"reverse"`

	// 爬虫配置
	BasicCrawler JieConf.BasicCrawler `json:"basicCrawler" yaml:"basicCrawler" mapstructure:"basiccrawler"`

	// mitmproxy配置
	Mitmproxy JieConf.Mitmproxy `json:"mitmproxy" yaml:"mitmproxy" mapstructure:"mitmproxy"`

	// 信息收集配置
	Collection JieConf.Collection `json:"collection" yaml:"collection" mapstructure:"collection"`
}

func init() {
	Config = &Configure{
		ProxyPort:    ProxyPort,
		Exclude:      []*Scope{},
		Include:      []*Scope{},
		FilterSuffix: []string{},
	}
}

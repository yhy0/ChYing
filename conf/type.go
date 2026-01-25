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

// ProxyListener 代理监听器配置
type ProxyListener struct {
	ID      string `json:"id" yaml:"id" mapstructure:"id"`
	Host    string `json:"host" yaml:"host" mapstructure:"host"`
	Port    int    `json:"port" yaml:"port" mapstructure:"port"`
	Enabled bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Running bool   `json:"running" yaml:"-" mapstructure:"-"` // 运行时状态，不持久化
}

// AIConfig AI 配置结构体
type AIConfig struct {
	// Agent 模式: "claude-code" 或 "a2a"
	AgentMode string `json:"agent_mode" yaml:"agent_mode" mapstructure:"agent_mode"`

	Claude struct {
		// Claude Code CLI 配置
		CLIPath        string `json:"cli_path" yaml:"cli_path" mapstructure:"cli_path"`                   // CLI 路径，默认使用 PATH 中的 claude
		Model          string `json:"model" yaml:"model" mapstructure:"model"`                            // 模型
		MaxTurns       int    `json:"max_turns" yaml:"max_turns" mapstructure:"max_turns"`                // 最大回合数
		SystemPrompt   string `json:"system_prompt" yaml:"system_prompt" mapstructure:"system_prompt"`    // 系统提示词
		PermissionMode string `json:"permission_mode" yaml:"permission_mode" mapstructure:"permission_mode"` // 权限模式
		// 注意: API Key、代理、MCP 服务器等配置请在 ~/.claude/settings.json 中设置
		// ChYing 会自动复用 Claude CLI 的用户配置
	} `json:"claude" yaml:"claude" mapstructure:"claude"`

	// A2A Agent 配置
	A2A A2AConfig `json:"a2a" yaml:"a2a" mapstructure:"a2a"`
}

// A2AConfig A2A Agent 配置
type A2AConfig struct {
	Enabled   bool              `json:"enabled" yaml:"enabled" mapstructure:"enabled"`       // 是否启用
	AgentURL  string            `json:"agent_url" yaml:"agent_url" mapstructure:"agent_url"` // Agent URL
	Headers   map[string]string `json:"headers" yaml:"headers" mapstructure:"headers"`       // 自定义请求头
	Timeout   int               `json:"timeout" yaml:"timeout" mapstructure:"timeout"`       // 超时时间（秒）
	EnableSSE bool              `json:"enable_sse" yaml:"enable_sse" mapstructure:"enable_sse"` // 是否启用 SSE 流式响应
}

// AppConfig 应用完整配置结构体
// Jie 扫描相关配置直接引用 Jie 的类型定义，保持类型一致性
type AppConfig struct {
	Version string `json:"version" yaml:"version" mapstructure:"version"`

	// === ChYing 独有配置 ===

	// 代理配置
	Proxy struct {
		Host      string          `json:"host" yaml:"host" mapstructure:"host"`
		Port      int             `json:"port" yaml:"port" mapstructure:"port"`
		Enabled   bool            `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
		Listeners []ProxyListener `json:"listeners" yaml:"listeners" mapstructure:"listeners"`
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

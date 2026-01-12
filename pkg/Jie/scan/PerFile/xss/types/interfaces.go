package types

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/config"
)

// Detector 是XSS检测器的统一接口
type Detector interface {
	// Detect 针对一个具体的注入点执行XSS检测
	Detect(ctx *DetectionContext, point InjectionPoint) []DetectionResult
	// Configure 使用提供的配置来设置检测器
	Configure(cfg *config.Config) error
	// Name 返回检测器的名称
	Name() string
}

// Parser 是HTML/JavaScript解析器的统一接口
type Parser interface {
	// Parse 从响应中解析出可供分析的结构
	Parse(resp *httpx.Response) (ParsedResponse, error)
}

// ContextAnalyzer 是上下文分析器的接口
type ContextAnalyzer interface {
	// Analyze 分析解析后的响应，识别出所有潜在的注入点
	Analyze(parsed ParsedResponse, originalValue string) []InjectionPoint
}

// PayloadManager 是Payload管理器的接口
type PayloadManager interface {
	// Generate 为给定的注入点生成测试Payloads
	Generate(point InjectionPoint) <-chan string
	// Configure configures the payload manager
	Configure(cfg *config.Config)
}

// Verifier 是漏洞验证器的接口
type Verifier interface {
	// Verify 验证一个响应是否表明存在XSS漏洞
	Verify(ctx *DetectionContext, result *DetectionResult, resp *httpx.Response)
}

// AttackSurfaceDiscoverer 是一个用于从响应中发现新攻击参数（例如表单输入）的接口。
type AttackSurfaceDiscoverer interface {
	// Discover 解析响应并返回找到的任何新参数。
	Discover(resp *httpx.Response) []httpx.Param
}

// Engine 是检测引擎的接口，负责编排整个扫描流程
type Engine interface {
	// Run 启动扫描过程
	Run(initialCtx *input.CrawlResult, client *httpx.Client) error
	// Configure 配置引擎
	Configure(cfg *config.Config) error
}

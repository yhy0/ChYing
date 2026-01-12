package types

import (
	"encoding/json"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"golang.org/x/net/html"
)

// ContextType 表示回显发生的上下文环境
type ContextType string

const (
	// HTML 上下文类型
	HTMLTagContext       ContextType = "html_tag"
	HTMLCommentContext   ContextType = "html_comment"
	HTMLAttributeContext ContextType = "html_attribute"

	// JavaScript 上下文类型
	JSContext              ContextType = "js_generic"
	JSStringLiteralContext ContextType = "js_string_literal"
	JSCommentContext       ContextType = "js_comment"

	// 其他上下文类型
	URLContext     ContextType = "url"
	UnknownContext ContextType = "unknown"
)

// Parameter 代表一个需要测试的参数
type Parameter struct {
	Name     string // 参数名
	Position string // 参数位置 (e.g., "query", "body", "header", "cookie")
}

// DetectionContext 封装了单次检测所需的所有上下文信息
type DetectionContext struct {
	// 原始响应，其中包含了原始请求的信息 (RequestUrl, RequestDump)
	BaseResponse *httpx.Response
	// 用于发出请求的HTTP客户端
	Client *httpx.Client
	// 可变的参数，用于重放攻击
	Variations *httpx.Variations
	// 原始的请求方法
	Method string
}

// ParsedResponse 封装了解析后的响应内容
type ParsedResponse struct {
	// 传统AST解析树的根节点
	HTMLAstRoot *html.Node
	// 基于goquery的DOM对象，用于高级解析
	GoqueryDoc *goquery.Document
	// 响应中提取的脚本块
	Scripts []string
	// 原始响应体
	Body string
	// 性能优化：缓存解析结果
	CachedTagSet  map[string]struct{}            // 缓存的标签集合
	CachedAttrSet map[string]map[string]struct{} // 缓存的属性集合，按标签分组
}

// QuoteType 定义了JS字符串或HTML属性中使用的引号类型
type QuoteType string

const (
	SingleQuote   QuoteType = "single"
	DoubleQuote   QuoteType = "double"
	BacktickQuote QuoteType = "backtick"
	NoQuote       QuoteType = "none"
)

// InjectionPoint 描述了一个输入在响应中的具体回显位置和环境
type InjectionPoint struct {
	// 探测阶段解析出的响应，包含了原始的AST信息
	BaseParsedResponse ParsedResponse
	// 回显所在的参数
	Parameter Parameter
	// 上下文类型
	Type ContextType
	// 引号类型 (仅对 JS 字符串和 HTML 属性相关)
	QuoteType QuoteType
	// 上下文相关的详细信息, e.g., 标签名, 属性名
	ContextDetails map[string]string
	// 原始回显值
	ReflectedValue string
	// 置信度
	Confidence float64
}

// DetectionResult 封装了单次漏洞检测的结果
type DetectionResult struct {
	// 是否确认发现漏洞
	Found bool
	// 漏洞详情，用于最终报告
	VulnInfo *output.VulMessage
	// 触发漏洞的Payload
	TriggeringPayload string
	// 发现漏洞的注入点
	InjectionPoint InjectionPoint
	// 风险等级
	RiskLevel string
	// 置信度
	Confidence float64
	// 使用的检测方法
	DetectionMethod string
}

// Vulnerability 定义了一个已发现的漏洞的详细信息。
// 这是整个扫描引擎向外部报告结果的标准格式。
type Vulnerability struct {
	Type    string                 `json:"type"`    // 漏洞类型, e.g., "Reflected XSS", "DOM-based XSS"
	URL     string                 `json:"url"`     // 存在漏洞的URL
	Payload string                 `json:"payload"` // 触发漏洞的Payload (对于DOM XSS可能不适用)
	Details map[string]interface{} `json:"details"` // 其他详细信息, e.g., {"source": "location.hash", "sink": "innerHTML"}
}

// ToVulMessage 将内部的Vulnerability结构转换为外部报告所需的VulMessage格式。
func (v *Vulnerability) ToVulMessage() *output.VulMessage {
	detailsBytes, _ := json.Marshal(v.Details)

	vulnData := output.VulnData{
		Target:      v.URL,
		Payload:     v.Payload,
		Description: fmt.Sprintf("发现 %s 漏洞. 详情: %s", v.Type, string(detailsBytes)),
		VulnType:    "xss", // 统一设置为xss
	}

	return &output.VulMessage{
		DataType: "vul",
		VulnData: vulnData,
		Plugin:   "xss",
		Level:    output.High, // 默认将XSS漏洞标记为高危
	}
}

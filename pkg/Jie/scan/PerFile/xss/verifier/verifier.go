package verifier

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/parser"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/logging"
)

// A list of HTML tags that can execute code, used for verification.
var executableTags = map[string]struct{}{
	"script": {}, "svg": {}, "img": {}, "video": {}, "audio": {},
	"iframe": {}, "object": {}, "embed": {}, "details": {}, "isindex": {},
	"body": {}, "form": {},
}

// defaultVerifier is types.Verifier's default implementation
type defaultVerifier struct {
	parser types.Parser
}

// NewVerifier creates a new verifier instance
func NewVerifier() types.Verifier {
	return &defaultVerifier{
		parser: parser.NewParser(), // Verifier internal holds its own parser instance
	}
}

// getTagSetCached 带缓存的标签集合获取函数
func getTagSetCached(doc *goquery.Document, parsed *types.ParsedResponse) map[string]struct{} {
	// 如果已有缓存，直接返回
	if parsed != nil && parsed.CachedTagSet != nil {
		return parsed.CachedTagSet
	}

	// 否则计算并缓存
	tags := make(map[string]struct{})
	if doc != nil {
		doc.Find("*").Each(func(i int, s *goquery.Selection) {
			tags[goquery.NodeName(s)] = struct{}{}
		})
	}

	// 缓存结果
	if parsed != nil {
		parsed.CachedTagSet = tags
	}

	return tags
}

// getAttributeSetCached 带缓存的属性集合获取函数
func getAttributeSetCached(doc *goquery.Document, selector string, parsed *types.ParsedResponse) map[string]struct{} {
	// 如果已有缓存，直接返回
	if parsed != nil && parsed.CachedAttrSet != nil {
		if cachedAttrs, exists := parsed.CachedAttrSet[selector]; exists {
			return cachedAttrs
		}
	}

	// 否则计算并缓存
	attrs := make(map[string]struct{})
	if doc != nil {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			for _, attr := range s.Get(0).Attr {
				attrs[attr.Key] = struct{}{}
			}
		})
	}

	// 缓存结果
	if parsed != nil {
		if parsed.CachedAttrSet == nil {
			parsed.CachedAttrSet = make(map[string]map[string]struct{})
		}
		parsed.CachedAttrSet[selector] = attrs
	}

	return attrs
}

// getTagSet collects all unique tag names from a goquery document.
func getTagSet(doc *goquery.Document) map[string]struct{} {
	return getTagSetCached(doc, nil)
}

// getAttributeSet collects all unique attribute keys for a given selector.
func getAttributeSet(doc *goquery.Document, selector string) map[string]struct{} {
	attrs := make(map[string]struct{})
	if doc != nil {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			for _, attr := range s.Get(0).Attr {
				attrs[attr.Key] = struct{}{}
			}
		})
	}
	return attrs
}

// isIdentifierInScript uses a JS lexer to precisely check if a keyword exists in executable code.
func isIdentifierInScript(scriptContent, keyword string) bool {
	l := js.NewLexer(parse.NewInputString(scriptContent))
	for {
		tt, text := l.Next()
		if tt == js.ErrorToken {
			if l.Err() != io.EOF {
				logging.Logger.Debugf("JS lexer error during verification: %v", l.Err())
			}
			break
		}
		// We are looking for keywords that are part of the code
		if strings.Contains(string(text), keyword) {
			// Crucial check: ensure it's not part of a string, comment, or regex
			if tt != js.StringToken && tt != js.CommentToken && tt != js.RegExpToken {
				return true
			}
		}
	}
	return false
}

// buildDOMSignature 为一个DOM文档构建一个签名，用于后续的比对。
// 签名是一个map，键是标签名，值是该标签所有属性的集合。
func buildDOMSignature(parsed *types.ParsedResponse) map[string]map[string]struct{} {
	signature := make(map[string]map[string]struct{})
	if parsed == nil || parsed.GoqueryDoc == nil {
		return signature
	}

	parsed.GoqueryDoc.Find("*").Each(func(i int, s *goquery.Selection) {
		tagName := goquery.NodeName(s)
		if _, exists := signature[tagName]; !exists {
			signature[tagName] = make(map[string]struct{})
		}
		for _, attr := range s.Get(0).Attr {
			signature[tagName][strings.ToLower(attr.Key)] = struct{}{}
		}
	})
	return signature
}

// Verify 验证一个响应是否表明存在XSS漏洞。
// 此函数实现了基于DOM结构差异比对的验证逻辑。
func (v *defaultVerifier) Verify(ctx *types.DetectionContext, result *types.DetectionResult, attackResp *httpx.Response) {
	result.Found = false

	// 1. 解析攻击后的响应，获取新的DOM结构
	newParsed, err := v.parser.Parse(attackResp)
	if err != nil {
		logging.Logger.Warnf("AST验证阶段解析攻击响应失败: %v", err)
		return
	}

	// 2. 获取探测阶段的原始DOM结构
	originalParsed := &result.InjectionPoint.BaseParsedResponse
	if originalParsed.GoqueryDoc == nil {
		logging.Logger.Warn("无法执行AST验证，原始DOM为空")
		return
	}

	// 3. 执行DOM差异比对
	v.verifyDOMChange(result, originalParsed, &newParsed)

	// 4. (备用) 如果DOM比对未发现漏洞，且是JS上下文，则尝试验证JS突破
	if !result.Found {
		contextType := result.InjectionPoint.Type
		if contextType == types.JSContext || contextType == types.JSStringLiteralContext || contextType == types.JSCommentContext {
			v.verifyJSBreakout(result, newParsed.GoqueryDoc)
		}
	}
}

// verifyDOMChange 通过比对攻击前后的DOM签名，来验证注入是否成功。
func (v *defaultVerifier) verifyDOMChange(result *types.DetectionResult, originalParsed, newParsed *types.ParsedResponse) {
	originalSignature := buildDOMSignature(originalParsed)
	newSignature := buildDOMSignature(newParsed)

	// 策略1: 检查是否注入了新的、可执行的HTML标签
	for tag := range newSignature {
		if _, exists := originalSignature[tag]; !exists {
			if _, isExec := executableTags[tag]; isExec {
				// 为提高置信度，确保新标签确实存在于DOM中
				if newParsed.GoqueryDoc.Find(tag).Length() > 0 {
					result.Found = true
					result.Confidence = 0.98
					result.RiskLevel = "High"
					logging.Logger.Infof("AST验证成功: 在 %s 上下文中检测到新的可执行标签 <%s>。", result.InjectionPoint.Type, tag)
					return
				}
			}
		}
	}

	// 策略2: 检查是否为已有标签增添了新的事件处理器属性
	for tag, newAttrs := range newSignature {
		if originalAttrs, exists := originalSignature[tag]; exists {
			for attr := range newAttrs {
				if _, attrExists := originalAttrs[attr]; !attrExists {
					// 发现了一个新的属性，检查它是否是事件处理器
					if strings.HasPrefix(attr, "on") {
						result.Found = true
						result.Confidence = 0.98
						result.RiskLevel = "High"
						logging.Logger.Infof("AST验证成功: 在 <%s> 标签上检测到新的事件处理器 '%s'。", tag, attr)
						return
					}
				}
			}
		}
	}
}

// verifyJSBreakout (原verifyJSInjection) 检查payload是否成功突破了JS上下文。
func (v *defaultVerifier) verifyJSBreakout(result *types.DetectionResult, newDoc *goquery.Document) {
	keywordsToVerify := v.extractKeywords(result.TriggeringPayload)
	scriptContent := newDoc.Find("script").Text()

	for _, keyword := range keywordsToVerify {
		if isIdentifierInScript(scriptContent, keyword) {
			result.Found = true
			result.Confidence = 0.88 // 置信度稍低，因为仅基于关键词匹配
			result.RiskLevel = "High"
			logging.Logger.Infof("AST验证成功: 在脚本上下文中检测到JS突破，关键词: '%s'", keyword)
			return
		}
	}
}

// extractKeywords 从payload中提取关键词的辅助函数
func (v *defaultVerifier) extractKeywords(payload string) []string {
	var keywords []string

	// 常见的危险函数
	dangerousFunctions := []string{"alert", "eval", "setTimeout", "setInterval", "Function", "confirm", "prompt", "console"}

	payloadLower := strings.ToLower(payload)
	for _, fn := range dangerousFunctions {
		if strings.Contains(payloadLower, fn) {
			keywords = append(keywords, fn)
		}
	}

	// 如果没有找到已知函数，尝试提取我们的统一测试标识
	if len(keywords) == 0 {
		if strings.Contains(payload, "ChYingXSS") {
			keywords = append(keywords, "ChYingXSS")
		}
		// 备用：检查常见的XSS测试标识
		testPatterns := []string{"xss", "test", "1", "2", "3"}
		for _, pattern := range testPatterns {
			if strings.Contains(payloadLower, pattern) {
				keywords = append(keywords, pattern)
				break // 只取第一个匹配的模式
			}
		}
	}

	// 如果还是没有关键词，使用默认的alert
	if len(keywords) == 0 {
		keywords = append(keywords, "alert")
	}

	return keywords
}

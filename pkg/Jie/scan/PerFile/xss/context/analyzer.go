package context

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/logging"
)

// defaultContextAnalyzer 是 types.ContextAnalyzer 接口的默认实现
type defaultContextAnalyzer struct{}

// NewContextAnalyzer 创建一个新的上下文分析器实例
func NewContextAnalyzer() types.ContextAnalyzer {
	return &defaultContextAnalyzer{}
}

// analyzeJsContent 对JavaScript代码进行词法分析，以确定探针所在的微观上下文
func (a *defaultContextAnalyzer) analyzeJsContent(scriptContent string, probe string) []types.InjectionPoint {
	var points []types.InjectionPoint
	l := js.NewLexer(parse.NewInputString(scriptContent))
	for {
		tt, text := l.Next()
		tokenText := string(text)

		if tt == js.ErrorToken {
			if l.Err() != io.EOF {
				logging.Logger.Debugf("JS词法分析错误: %v", l.Err())
			}
			break
		}

		if strings.Contains(tokenText, probe) {
			details := make(map[string]string)
			point := types.InjectionPoint{
				ContextDetails: details,
			}

			switch tt {
			case js.StringToken:
				point.Type = types.JSStringLiteralContext
				details["sub_type"] = "string_literal"
				if len(tokenText) > 1 {
					if tokenText[0] == '\'' {
						details["quote_type"] = "single"
						point.QuoteType = types.SingleQuote
					} else if tokenText[0] == '"' {
						details["quote_type"] = "double"
						point.QuoteType = types.DoubleQuote
					} else if tokenText[0] == '`' {
						details["quote_type"] = "backtick"
						point.QuoteType = types.BacktickQuote
					} else {
						details["quote_type"] = "none"
						point.QuoteType = types.NoQuote
					}
				} else {
					point.QuoteType = types.NoQuote
				}
			case js.CommentToken:
				point.Type = types.JSCommentContext
				details["sub_type"] = "comment"
				point.QuoteType = types.NoQuote
			case js.IdentifierToken:
				point.Type = types.JSContext
				details["sub_type"] = "identifier"
				point.QuoteType = types.NoQuote
			case js.TemplateToken:
				point.Type = types.JSStringLiteralContext
				details["sub_type"] = "template_literal"
				point.QuoteType = types.BacktickQuote
			default:
				point.Type = types.JSContext
				details["sub_type"] = "unknown"
				point.QuoteType = types.NoQuote
			}
			points = append(points, point)
		}
	}
	return points
}

// isDangerousAttribute 判断属性是否为可能直接执行代码的危险属性
func isDangerousAttribute(attrName string) bool {
	dangerousAttrs := []string{
		"href", "src", "action", "formaction", "data", "xlink:href", "manifest",
		"style", "background", "lowsrc", "dynsrc", "poster",
	}

	attrLower := strings.ToLower(attrName)
	for _, dangerous := range dangerousAttrs {
		if attrLower == dangerous {
			return true
		}
	}

	// 事件处理器属性
	return strings.HasPrefix(attrLower, "on")
}

// detectAttributeQuoteType 检测属性值中的引号类型
func detectAttributeQuoteType(attrValue, probe string) types.QuoteType {
	// 在属性值中查找探针位置，然后检查周围的引号
	probeIndex := strings.Index(attrValue, probe)
	if probeIndex == -1 {
		return types.NoQuote
	}

	// 检查探针前后是否有引号
	if probeIndex > 0 {
		before := attrValue[probeIndex-1]
		if before == '\'' {
			return types.SingleQuote
		} else if before == '"' {
			return types.DoubleQuote
		} else if before == '`' {
			return types.BacktickQuote
		}
	}

	// 检查探针后面是否有引号
	probeEnd := probeIndex + len(probe)
	if probeEnd < len(attrValue) {
		after := attrValue[probeEnd]
		if after == '\'' {
			return types.SingleQuote
		} else if after == '"' {
			return types.DoubleQuote
		} else if after == '`' {
			return types.BacktickQuote
		}
	}

	return types.NoQuote
}

// Analyze 对解析后的响应进行分析，识别所有潜在的注入点。
// 新的实现优化了遍历逻辑，并增强了对HTML和JS混合上下文的分析能力。
func (a *defaultContextAnalyzer) Analyze(resp types.ParsedResponse, probe string) []types.InjectionPoint {
	var points []types.InjectionPoint
	if resp.GoqueryDoc == nil {
		return points
	}

	// 统一处理所有节点
	resp.GoqueryDoc.Find("*").Each(func(i int, s *goquery.Selection) {
		nodeName := goquery.NodeName(s)

		// 优先处理 <script> 标签，因为其内容是纯JS上下文
		if nodeName == "script" {
			scriptContent, _ := s.Html()
			if strings.Contains(scriptContent, probe) {
				jsPoints := a.analyzeJsContent(scriptContent, probe)
				for k := range jsPoints {
					jsPoints[k].ContextDetails["parent_tag"] = "script"
				}
				points = append(points, jsPoints...)
			}
			// 处理完script标签后，不再对其子节点进行重复分析
			return
		}

		// 检查属性中的回显
		for _, attr := range s.Get(0).Attr {
			details := make(map[string]string)
			details["tag_name"] = nodeName

			if strings.Contains(attr.Val, probe) {
				details["reflected_in"] = "value"
				details["attribute_key"] = attr.Key
				// 增强：判断是否为危险属性
				if isDangerousAttribute(attr.Key) {
					details["is_dangerous_attr"] = "true"
				}
				points = append(points, types.InjectionPoint{
					Type:           types.HTMLAttributeContext,
					ContextDetails: details,
					// 增强：精确判断属性值的引号类型
					QuoteType: detectAttributeQuoteType(attr.Val, probe),
				})
			} else if strings.Contains(attr.Key, probe) {
				details["reflected_in"] = "key"
				details["attribute_key"] = attr.Key
				points = append(points, types.InjectionPoint{
					Type:           types.HTMLAttributeContext,
					ContextDetails: details,
					QuoteType:      types.NoQuote, // 属性名中无引号
				})
			}
		}

		// 检查文本节点和注释节点中的回显
		s.Contents().Each(func(j int, child *goquery.Selection) {
			// 如果子节点是 <script>，它已经在上面被处理过，跳过
			if goquery.NodeName(child) == "script" {
				return
			}

			nodeText, _ := child.Html()
			if !strings.Contains(nodeText, probe) {
				return
			}

			// 使用 goquery.NodeName(child) 来判断节点类型
			switch goquery.NodeName(child) {
			case "#text":
				details := make(map[string]string)
				details["parent_tag"] = nodeName
				points = append(points, types.InjectionPoint{
					Type:           types.HTMLTagContext,
					ContextDetails: details,
				})
			case "#comment":
				points = append(points, types.InjectionPoint{
					Type:           types.HTMLCommentContext,
					ContextDetails: make(map[string]string),
				})
			}
		})
	})

	// 去重逻辑可以放在这里，如果需要的话
	return points
}

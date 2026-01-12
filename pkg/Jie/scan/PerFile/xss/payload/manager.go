package payload

import (
	"fmt"
	"strings"

	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/config"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/logging"
)

// defaultPayloadManager is the default implementation of the PayloadManager interface.
type defaultPayloadManager struct {
	config *config.Config
}

// NewPayloadManager creates a new default payload manager.
func NewPayloadManager(cfg *config.Config) types.PayloadManager {
	return &defaultPayloadManager{
		config: cfg,
	}
}

// Configure updates the payload manager's configuration.
func (pm *defaultPayloadManager) Configure(cfg *config.Config) {
	pm.config = cfg
}

// Generate generates a list of XSS payloads based on the injection point and configuration.
func (pm *defaultPayloadManager) Generate(point types.InjectionPoint) <-chan string {
	payloadChan := make(chan string, 100) // Add a buffer to prevent blocking

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.Logger.Errorf("Payload生成过程中发生panic: %v", r)
			}
			close(payloadChan)
		}()

		payloadSet := make(map[string]struct{})

		// Use a consistent, harmless, and unique script for payloads
		// This makes verification easier and more reliable.
		harmlessScript := "alert('ChYingXSS')"

		// 1. Generate highly context-specific base payloads
		basePayloads := pm.generateContextAwarePayloads(&point, harmlessScript)
		for _, p := range basePayloads {
			payloadSet[p] = struct{}{}
		}

		// 2. Generate encoded variants from the base payloads if enabled
		if pm.config != nil && pm.config.Payload.EnableMixedEncoding {
			encodedPayloads := pm.generateEncodedVariants(basePayloads)
			for _, p := range encodedPayloads {
				payloadSet[p] = struct{}{}
			}
		}

		// 3. Add generic WAF bypass and modern JS payloads for broader coverage
		genericPayloads := pm.generateGenericPayloads(harmlessScript)
		for _, p := range genericPayloads {
			payloadSet[p] = struct{}{}
		}

		// 4. Send all unique payloads through the channel (with limit)
		count := 0
		maxPayloads := pm.getContextSpecificLimit(point.Type)

		for payload := range payloadSet {
			if count >= maxPayloads {
				logging.Logger.Debugf("已达到上下文特定的payload数量限制 (%d)，停止生成", maxPayloads)
				break
			}
			payloadChan <- payload
			count++
		}
	}()

	return payloadChan
}

// getContextSpecificLimit 根据上下文智能调整payload数量限制
func (pm *defaultPayloadManager) getContextSpecificLimit(contextType types.ContextType) int {
	baseLimit := 50 // 默认限制
	if pm.config != nil && pm.config.Payload.MaxPayloadsPerContext > 0 {
		baseLimit = pm.config.Payload.MaxPayloadsPerContext
	}

	// 根据上下文类型调整限制
	switch contextType {
	case types.JSStringLiteralContext:
		return min(baseLimit, 20) // JS字符串上下文相对简单，需要的payload较少
	case types.HTMLAttributeContext:
		return min(baseLimit, baseLimit) // 属性上下文需要更多变种，使用完整限制
	case types.HTMLTagContext:
		return min(baseLimit, 30) // HTML标签上下文中等复杂度
	case types.JSContext:
		return min(baseLimit, 25) // JS代码上下文
	case types.HTMLCommentContext:
		return min(baseLimit, 15) // 注释上下文最简单
	default:
		return baseLimit
	}
}

// generateContextAwarePayloads creates initial payloads based on the precise injection context.
func (pm *defaultPayloadManager) generateContextAwarePayloads(point *types.InjectionPoint, script string) []string {
	var payloads []string

	switch point.Type {
	case types.JSStringLiteralContext:
		var breakoutKey string
		switch point.QuoteType {
		case types.SingleQuote:
			breakoutKey = "single_quote"
		case types.DoubleQuote:
			breakoutKey = "double_quote"
		case types.BacktickQuote:
			breakoutKey = "backtick"
		default:
			breakoutKey = "double_quote" // Default fallback
		}
		template := JsBreakoutPayloads[breakoutKey]
		payloads = append(payloads, strings.Replace(template, "{{script}}", script, -1))
		// Also try HTML-encoded breakouts
		if val, ok := JsBreakoutPayloads[breakoutKey+"_html"]; ok {
			payloads = append(payloads, strings.Replace(val, "{{script}}", script, -1))
		}

	case types.JSCommentContext:
		payloads = append(payloads, fmt.Sprintf("*/%s/*", script))
		payloads = append(payloads, fmt.Sprintf("\n%s//", script)) // For single line comments

	case types.HTMLAttributeContext:
		// Check if it's a "dangerous" attribute like href, src, etc.
		attrName := strings.ToLower(point.ContextDetails["attribute_key"])
		isURLAttr := false
		urlLikeAttrs := []string{"href", "src", "action", "data", "formaction", "xlink:href", "manifest"}
		for _, u := range urlLikeAttrs {
			if u == attrName {
				isURLAttr = true
				break
			}
		}

		if isURLAttr {
			for _, protoPayload := range ProtocolPayloads {
				payloads = append(payloads, strings.Replace(protoPayload, "{{script}}", script, -1))
			}
		} else {
			// It's a regular attribute, try to break out and inject an event handler
			payloads = append(payloads, fmt.Sprintf(`" onmouseover=%s `, script))
			payloads = append(payloads, fmt.Sprintf(`' onmouseover=%s `, script))
			payloads = append(payloads, fmt.Sprintf(` onmouseover=%s `, script)) // No quote breakout
		}

	case types.HTMLTagContext, types.HTMLCommentContext:
		var breakout string
		if point.Type == types.HTMLCommentContext {
			breakout = HTMLBreakoutPayloads["comment_breakout_1"]
		} else {
			breakout = HTMLBreakoutPayloads["tag_breakout"]
		}
		// Generate standard tag injections
		payloads = append(payloads, fmt.Sprintf(`%s<script>%s</script>`, breakout, script))
		payloads = append(payloads, fmt.Sprintf(`%s<svg onload=%s>`, breakout, script))
		payloads = append(payloads, fmt.Sprintf(`%s<img src=x onerror=%s>`, breakout, script))
		payloads = append(payloads, fmt.Sprintf(`%s<details open ontoggle=%s>`, breakout, script))

	default:
		// Generic fallback for unknown contexts
		payloads = append(payloads, fmt.Sprintf(`<script>%s</script>`, script))
		payloads = append(payloads, fmt.Sprintf(`<svg onload=%s>`, script))
	}

	return payloads
}

// min 返回两个整数中的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateEncodedVariants applies various encoding schemes to a list of payloads.
func (pm *defaultPayloadManager) generateEncodedVariants(basePayloads []string) []string {
	var variants []string
	if pm.config == nil {
		return variants
	}

	for _, p := range basePayloads {
		if pm.config.Payload.EnableURLEncoding {
			variants = append(variants, urlEncode(p, false))
			variants = append(variants, urlEncode(p, true))
		}
		if pm.config.Payload.EnableHTMLEncoding {
			variants = append(variants, htmlEncode(p))
		}
		if pm.config.Payload.EnableUnicodeEncoding {
			variants = append(variants, unicodeEncode(p))
		}
		if pm.config.Payload.EnableHexEncoding {
			variants = append(variants, hexEncode(p))
		}
		// The main Generate function now controls mixed encoding.
		variants = append(variants, MixedEncode(p)...)
	}
	return variants
}

// generateGenericPayloads creates payloads from the predefined lists for broad coverage.
func (pm *defaultPayloadManager) generateGenericPayloads(script string) []string {
	var payloads []string
	if pm.config == nil {
		return payloads
	}

	// Add modern JS payloads, replacing the placeholder
	for _, p := range ModernJSPayloads {
		payloads = append(payloads, strings.Replace(p, "{{script}}", script, -1))
	}

	// Add standalone WAF bypass payloads
	if pm.config.Payload.EnableWAFBypass {
		payloads = append(payloads, WafBypassPayloads...)
	}

	// Generate payloads from a list of all event handlers
	if pm.config.Payload.EnableEventHandlerInjection {
		// 根据配置动态调整事件处理器数量限制
		limit := 20
		if pm.config.Payload.MaxPayloadsPerContext > 0 {
			limit = min(limit, pm.config.Payload.MaxPayloadsPerContext/3) // 事件处理器占总数的1/3
		}
		for i, event := range EventHandlerPayloads {
			if i >= limit {
				break
			}
			payloads = append(payloads, fmt.Sprintf(`<div %s=%s></div>`, event, script))
		}
	}

	return payloads
}

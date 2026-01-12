package detectors

import (
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
)

// --- JS String Literal Detector ---

// JSStringLiteralDetector 专门用于处理JS字符串字面量中的注入。
type JSStringLiteralDetector struct {
	BaseDetector
}

// NewJSStringLiteralDetector 创建一个新的JS字符串字面量检测器。
func NewJSStringLiteralDetector() types.Detector {
	return &JSStringLiteralDetector{
		BaseDetector: BaseDetector{DetectorName: "JSStringLiteralDetector"},
	}
}

// Detect 仅在上下文为JS字符串时执行检测。
func (d *JSStringLiteralDetector) Detect(ctx *types.DetectionContext, point types.InjectionPoint) []types.DetectionResult {
	if point.Type != types.JSStringLiteralContext {
		return nil
	}
	vulnType := "Reflected XSS (JS String Literal Context)"
	vulnDesc := "在参数 '%s' 的JS字符串字面量上下文中发现反射型XSS。Payload: %s"
	return d.BaseDetector.RunDetectionCycle(ctx, point, vulnType, vulnDesc)
}

// --- JS Comment Detector ---

// JSCommentDetector 专门用于处理JS注释中的注入。
type JSCommentDetector struct {
	BaseDetector
}

// NewJSCommentDetector 创建一个新的JS注释检测器。
func NewJSCommentDetector() types.Detector {
	return &JSCommentDetector{
		BaseDetector: BaseDetector{DetectorName: "JSCommentDetector"},
	}
}

// Detect 仅在上下文为JS注释时执行检测。
func (d *JSCommentDetector) Detect(ctx *types.DetectionContext, point types.InjectionPoint) []types.DetectionResult {
	if point.Type != types.JSCommentContext {
		return nil
	}
	vulnType := "Reflected XSS (JS Comment Context)"
	vulnDesc := "在参数 '%s' 的JS注释上下文中发现反射型XSS。Payload: %s"
	return d.BaseDetector.RunDetectionCycle(ctx, point, vulnType, vulnDesc)
}

// --- JS Context Detector ---

// JSContextDetector 专门用于处理通用JS代码执行上下文中的注入。
type JSContextDetector struct {
	BaseDetector
}

// NewJSContextDetector 创建一个新的通用JS代码执行上下文检测器。
func NewJSContextDetector() types.Detector {
	return &JSContextDetector{
		BaseDetector: BaseDetector{DetectorName: "JSContextDetector"},
	}
}

// Detect 仅在上下文为通用JS代码时执行检测。
func (d *JSContextDetector) Detect(ctx *types.DetectionContext, point types.InjectionPoint) []types.DetectionResult {
	if point.Type != types.JSContext {
		return nil
	}
	vulnType := "Reflected XSS (JS Execution Context)"
	vulnDesc := "在参数 '%s' 的JS代码执行上下文中发现反射型XSS。Payload: %s"
	return d.BaseDetector.RunDetectionCycle(ctx, point, vulnType, vulnDesc)
}

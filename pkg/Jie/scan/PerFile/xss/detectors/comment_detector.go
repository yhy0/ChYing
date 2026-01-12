package detectors

import (
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
)

// CommentContextDetector 专门用于检测HTML注释上下文中的XSS。
type CommentContextDetector struct {
	BaseDetector
}

// NewCommentContextDetector 创建一个注释上下文检测器实例。
func NewCommentContextDetector() types.Detector {
	return &CommentContextDetector{
		BaseDetector: BaseDetector{
			DetectorName: "CommentContextDetector",
		},
	}
}

// Detect 针对一个具体的注入点执行XSS检测。
func (d *CommentContextDetector) Detect(ctx *types.DetectionContext, point types.InjectionPoint) []types.DetectionResult {
	if point.Type != types.HTMLCommentContext {
		return nil // 此检测器只处理HTML注释上下文。
	}

	vulnType := "Reflected XSS (HTML Comment Context)"
	vulnDesc := "在参数 '%s' 的HTML注释上下文中发现反射型XSS (AST验证通过)。Payload: %s"
	return d.BaseDetector.RunDetectionCycle(ctx, point, vulnType, vulnDesc)
}

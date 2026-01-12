package detectors

import (
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
)

// AttributeContextDetector 专门用于检测HTML属性上下文中的XSS
type AttributeContextDetector struct {
	BaseDetector
}

// NewAttributeContextDetector 创建一个属性上下文检测器实例
func NewAttributeContextDetector() types.Detector {
	return &AttributeContextDetector{
		BaseDetector: BaseDetector{
			DetectorName: "AttributeContextDetector",
		},
	}
}

// Detect 针对一个具体的注入点执行XSS检测
func (d *AttributeContextDetector) Detect(ctx *types.DetectionContext, point types.InjectionPoint) []types.DetectionResult {
	if point.Type != types.HTMLAttributeContext {
		return nil // 此检测器只处理HTML属性上下文
	}

	vulnType := "Reflected XSS (HTML Attribute Context)"
	vulnDesc := "在参数 '%s' 的HTML属性上下文中发现反射型XSS (AST验证通过)。Payload: %s"
	return d.BaseDetector.RunDetectionCycle(ctx, point, vulnType, vulnDesc)
}

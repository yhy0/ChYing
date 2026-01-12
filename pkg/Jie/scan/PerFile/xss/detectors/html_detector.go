package detectors

import (
	"fmt"
	"strconv"
	"time"

	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/config"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/payload"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/verifier"
	"github.com/yhy0/logging"
)

// BaseDetector 提供了所有检测器共享的通用功能。
// 它封装了配置、payload管理、验证和请求发送的样板逻辑。
type BaseDetector struct {
	Config         *config.Config
	PayloadManager types.PayloadManager
	Verifier       types.Verifier
	DetectorName   string
}

// Configure 为基础检测器设置配置。
func (b *BaseDetector) Configure(cfg *config.Config) error {
	b.Config = cfg
	// 每个检测器都拥有自己的Payload管理器和验证器实例
	b.PayloadManager = payload.NewPayloadManager(cfg)
	b.Verifier = verifier.NewVerifier()
	return nil
}

// Name 返回检测器的名称。
func (b *BaseDetector) Name() string {
	return b.DetectorName
}

// RunDetectionCycle 运行核心的检测循环，包括payload生成、请求发送和结果验证。
func (b *BaseDetector) RunDetectionCycle(ctx *types.DetectionContext, point types.InjectionPoint, vulnType, vulnDescFormat string) []types.DetectionResult {
	var results []types.DetectionResult
	paramIndexStr, ok := point.ContextDetails["param_index"]
	if !ok {
		logging.Logger.Warnf("在 %s 中, 无法为 %s 的注入点找到参数索引", b.DetectorName, ctx.BaseResponse.RequestUrl)
		return nil
	}
	paramIndex, err := strconv.Atoi(paramIndexStr)
	if err != nil {
		logging.Logger.Warnf("在 %s 中发现无效的参数索引 '%s'", b.DetectorName, paramIndexStr)
		return nil
	}

	for p := range b.PayloadManager.Generate(point) {
		modifiedRequestData := ctx.Variations.SetPayloadByIndex(paramIndex, ctx.BaseResponse.RequestUrl, p, ctx.Method)
		resp, reqErr := ctx.Client.Request(ctx.BaseResponse.RequestUrl, ctx.Method, modifiedRequestData, nil, "XSS")

		if reqErr != nil {
			logging.Logger.Debugf("在 %s 中请求失败: %v", b.DetectorName, reqErr)
			continue // 网络错误是常见情况，静默处理
		}

		result := types.DetectionResult{
			Found:             false,
			InjectionPoint:    point,
			TriggeringPayload: p,
			DetectionMethod:   b.Name(),
		}

		b.Verifier.Verify(ctx, &result, resp)

		if result.Found {
			logging.Logger.Infoln(fmt.Sprintf("%s XSS验证成功!", b.DetectorName))
			result.VulnInfo = b.createVulnMessage(ctx, &point, p, resp, vulnType, vulnDescFormat)
			results = append(results, result)
			// 为提高效率，一旦发现漏洞就停止对当前注入点的测试
			break
		}
	}
	return results
}

// createVulnMessage 创建标准格式的漏洞信息。
func (b *BaseDetector) createVulnMessage(ctx *types.DetectionContext, point *types.InjectionPoint, payload string, resp *httpx.Response, vulnType, vulnDescFormat string) *output.VulMessage {
	return &output.VulMessage{
		DataType: "vul",
		Plugin:   "XSS-Refactored",
		Level:    "High", // Verifier可以覆盖风险等级
		VulnData: output.VulnData{
			CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
			VulnType:    vulnType,
			Target:      ctx.BaseResponse.RequestUrl,
			Payload:     payload,
			Request:     resp.RequestDump,
			Response:    resp.ResponseDump,
			Description: fmt.Sprintf(vulnDescFormat, point.Parameter.Name, payload),
		},
	}
}

// HTMLContextDetector 专门用于检测HTML上下文中的XSS。
type HTMLContextDetector struct {
	BaseDetector
}

// NewHTMLContextDetector 创建一个HTML上下文检测器实例。
func NewHTMLContextDetector() types.Detector {
	return &HTMLContextDetector{
		BaseDetector: BaseDetector{
			DetectorName: "HTMLContextDetector",
		},
	}
}

// Detect 针对一个具体的注入点执行XSS检测。
func (d *HTMLContextDetector) Detect(ctx *types.DetectionContext, point types.InjectionPoint) []types.DetectionResult {
	if point.Type != types.HTMLTagContext {
		return nil // 此检测器只处理HTML标签上下文。
	}

	vulnType := "Reflected XSS (HTML Tag Context)"
	vulnDesc := "在参数 '%s' 的HTML标签上下文中发现反射型XSS (AST验证通过)。Payload: %s"
	return d.RunDetectionCycle(ctx, point, vulnType, vulnDesc)
}

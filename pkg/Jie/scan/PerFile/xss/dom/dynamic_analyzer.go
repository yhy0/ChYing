package dom

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/browser"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/config"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/logging"
	"github.com/ysmood/gson"
)

// DynamicAnalyzer 是第三层筛选器，使用无头浏览器进行动态污点分析。
type DynamicAnalyzer struct {
	config  *config.Config
	browser *rod.Browser
}

// NewDynamicAnalyzer 创建一个新的动态分析器。
func NewDynamicAnalyzer(cfg *config.Config) (*DynamicAnalyzer, error) {
	br, err := browser.GetManager().GetBrowser()
	if err != nil {
		logging.Logger.Errorf("无法获取浏览器实例: %v", err)
		return nil, err
	}
	return &DynamicAnalyzer{
		config:  cfg,
		browser: br,
	}, nil
}

// TaintAnalysisResult 封装了动态污点分析的结果
type TaintAnalysisResult struct {
	Found     bool
	Source    string
	Sink      string
	URL       string
	TaintPath []string // 记录污点传播路径 (可选)
}

// Analyze 执行动态分析。
func (d *DynamicAnalyzer) Analyze(url string) (*types.Vulnerability, error) {
	logging.Logger.Infof("【DOM XSS Tier 3】启动对 %s 的动态分析...", url)

	// 创建一个新的浏览器页面
	page, err := d.browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, fmt.Errorf("创建浏览器页面失败: %w", err)
	}

	// 确保资源正确释放
	defer func() {
		if err := page.Close(); err != nil {
			logging.Logger.Warnf("关闭浏览器页面时出错: %v", err)
		}
	}()

	// 设置一个总体的超时上下文
	timeout := 30 * time.Second // 默认30秒
	if d.config != nil && d.config.Engine.DynamicAnalysisTimeoutSec > 0 {
		timeout = time.Duration(d.config.Engine.DynamicAnalysisTimeoutSec) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	page = page.Context(ctx)

	// 用于接收JS中报告的结果
	resultChan := make(chan TaintAnalysisResult, 1)

	// 步骤1: 暴露一个Go函数给页面的JS环境
	stopExpo, err := page.Expose("reportTaint", func(json gson.JSON) (interface{}, error) {
		var result TaintAnalysisResult
		if err := json.Unmarshal(&result); err != nil {
			logging.Logger.Errorf("无法解析来自JS的污点报告: %v", err)
			return nil, err
		}

		logging.Logger.Warnf("【DOM XSS Found!】在 %s 检测到污点流动: Source='%s', Sink='%s'", url, result.Source, result.Sink)
		select {
		case resultChan <- result:
		default:
			logging.Logger.Warn("结果通道已满或关闭，丢弃污点报告。")
		}
		return nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("暴露JS函数失败: %w", err)
	}
	defer func() {
		if err := stopExpo(); err != nil {
			logging.Logger.Warnf("停止JS函数暴露时出错: %v", err)
		}
	}()

	// 步骤2: 设置请求拦截
	router := page.HijackRequests()
	defer func() {
		if err := router.Stop(); err != nil {
			logging.Logger.Warnf("停止请求拦截器时出错: %v", err)
		}
	}()

	router.MustAdd("*", func(hijackCtx *rod.Hijack) {
		defer func() {
			if r := recover(); r != nil {
				logging.Logger.Errorf("请求拦截处理中发生panic: %v", r)
				// 继续请求以避免阻塞
				hijackCtx.ContinueRequest(&proto.FetchContinueRequest{})
			}
		}()

		// 只拦截JS文件
		if hijackCtx.Request.Type() == proto.NetworkResourceTypeScript {
			logging.Logger.Debugf("拦截到JS文件: %s", hijackCtx.Request.URL())

			hijackCtx.MustLoadResponse()

			// 【关键步骤】对JS代码进行插桩
			instrumentedJS, err := InstrumentJS(hijackCtx.Response.Body())
			if err != nil {
				logging.Logger.Warnf("JS插桩失败: %v, URL: %s", err, hijackCtx.Request.URL())
				// 插桩失败，返回原始代码
				hijackCtx.Response.SetBody(hijackCtx.Response.Body())
			} else {
				// 插桩成功，返回修改后的代码
				hijackCtx.Response.SetBody(instrumentedJS)
			}
		}
		// 其他类型的请求直接放行
		hijackCtx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	go router.Run()

	// 步骤3: 导航到目标页面
	if err := page.Navigate(url); err != nil {
		return nil, fmt.Errorf("导航到目标页面失败: %w", err)
	}

	// 步骤4: 等待结果或超时
	select {
	case result := <-resultChan:
		// 发现了漏洞，构造并返回漏洞信息
		vuln := &types.Vulnerability{
			Type:    "DOM-based XSS",
			URL:     url,
			Payload: "N/A (Taint Flow)",
			Details: map[string]interface{}{
				"source": result.Source,
				"sink":   result.Sink,
			},
		}
		return vuln, nil
	case <-ctx.Done():
		// 超时或取消
		if ctx.Err() == context.DeadlineExceeded {
			logging.Logger.Debugf("动态分析超时: %s", url)
		} else {
			logging.Logger.Debugf("动态分析被取消: %s", url)
		}
		return nil, nil
	}
}

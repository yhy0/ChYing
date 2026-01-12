package engine

import (
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/config"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/context"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/detectors"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/discoverer"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/dom"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/parser"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/payload"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/verifier"
	"github.com/yhy0/logging"
)

// CoreEngine 是XSS扫描的核心引擎，负责协调各个模块完成扫描任务。
type CoreEngine struct {
	config          *config.Config
	parser          types.Parser
	contextAnalyzer types.ContextAnalyzer
	payloadManager  types.PayloadManager
	verifier        types.Verifier
	discoverer      types.AttackSurfaceDiscoverer
	// 将检测器列表改为Map，用于实现智能调度。
	// 键是上下文类型（如HTML、JS字符串等），值是对应的检测器实例。
	detectors       map[types.ContextType]types.Detector
	domPrefilter    *dom.Prefilter    // DOM XSS 静态预筛选器 (第一层)
	domASTValidator *dom.ASTValidator // DOM XSS AST语法验证器 (第二层)
}

// NewEngine 创建一个新的核心扫描引擎实例。
func NewEngine(cfg *config.Config) (types.Engine, error) {
	if cfg == nil {
		// 如果外部没有提供配置，则默认使用智能模式。
		logging.Logger.Info("未提供配置，使用默认的智能模式。")
		cfg = config.NewConfig(config.Intelligent)
	}

	engine := &CoreEngine{
		config:          cfg,
		parser:          parser.NewParser(),
		contextAnalyzer: context.NewContextAnalyzer(),
		payloadManager:  payload.NewPayloadManager(cfg),
		verifier:        verifier.NewVerifier(),
		discoverer:      discoverer.NewFormDiscoverer(),
		// 初始化用于智能调度的检测器Map
		detectors:       make(map[types.ContextType]types.Detector),
		domPrefilter:    dom.NewPrefilter(),    // 初始化DOM预筛选器
		domASTValidator: dom.NewASTValidator(), // 初始化DOM AST验证器
	}

	// 注册并配置所有可用的检测器
	if err := engine.registerAndConfigureDetectors(); err != nil {
		return nil, fmt.Errorf("初始化检测器失败: %w", err)
	}

	return engine, nil
}

// registerAndConfigureDetectors 注册所有检测器，并将它们与各自能处理的上下文类型关联起来。
func (e *CoreEngine) registerAndConfigureDetectors() error {
	// HTML
	htmlDetector := detectors.NewHTMLContextDetector()
	if err := htmlDetector.Configure(e.config); err != nil {
		return fmt.Errorf("配置HTML检测器失败: %w", err)
	}
	e.detectors[types.HTMLTagContext] = htmlDetector

	// Comment
	commentDetector := detectors.NewCommentContextDetector()
	if err := commentDetector.Configure(e.config); err != nil {
		return fmt.Errorf("配置Comment检测器失败: %w", err)
	}
	e.detectors[types.HTMLCommentContext] = commentDetector

	// Attribute
	attributeDetector := detectors.NewAttributeContextDetector()
	if err := attributeDetector.Configure(e.config); err != nil {
		return fmt.Errorf("配置Attribute检测器失败: %w", err)
	}
	e.detectors[types.HTMLAttributeContext] = attributeDetector

	// JS String Literal
	jsStringDetector := detectors.NewJSStringLiteralDetector()
	if err := jsStringDetector.Configure(e.config); err != nil {
		return fmt.Errorf("配置JSStringLiteralDetector失败: %w", err)
	}
	e.detectors[types.JSStringLiteralContext] = jsStringDetector

	// JS Comment
	jsCommentDetector := detectors.NewJSCommentDetector()
	if err := jsCommentDetector.Configure(e.config); err != nil {
		return fmt.Errorf("配置JSCommentDetector失败: %w", err)
	}
	e.detectors[types.JSCommentContext] = jsCommentDetector

	// JS Context
	jsContextDetector := detectors.NewJSContextDetector()
	if err := jsContextDetector.Configure(e.config); err != nil {
		return fmt.Errorf("配置JSContextDetector失败: %w", err)
	}
	e.detectors[types.JSContext] = jsContextDetector

	return nil
}

// Configure applies a new configuration to the engine
func (e *CoreEngine) Configure(cfg *config.Config) error {
	if cfg != nil {
		e.config = cfg

		// Propagate the new config to all sub-modules
		e.payloadManager.Configure(cfg)
		for contextType, d := range e.detectors {
			if err := d.Configure(cfg); err != nil {
				return fmt.Errorf("重新配置检测器 %s 失败: %w", contextType, err)
			}
		}
	}
	return nil
}

// Run starts the scanning process
func (e *CoreEngine) Run(initialCtx *input.CrawlResult, client *httpx.Client) error {
	logging.Logger.Infof("开始对 %s 进行XSS扫描 (重构版引擎)", initialCtx.Url)

	// =================================================================
	// 阶段一: 反射型XSS检测
	// =================================================================
	injectionPoints, variations, err := e.findInjectionPoints(initialCtx, client)
	if err != nil {
		logging.Logger.Errorf("在回显检测阶段发生错误: %v", err)
		return err // 关键阶段错误，直接返回
	}

	var allResults []types.DetectionResult
	if len(injectionPoints) > 0 {
		logging.Logger.Infof("在 %s 发现 %d 个潜在注入点，开始漏洞检测...", initialCtx.Url, len(injectionPoints))
		baseCtx := &types.DetectionContext{
			BaseResponse: initialCtx.Resp,
			Client:       client,
			Variations:   variations,
			Method:       initialCtx.Method,
		}

		for _, point := range injectionPoints {
			// 【智能调度】根据上下文类型，从Map中直接选择正确的检测器
			detector, found := e.detectors[point.Type]
			if !found {
				logging.Logger.Warnf("未找到针对上下文类型 '%s' 的特定检测器，跳过此注入点。", point.Type)
				continue
			}

			logging.Logger.Debugf("为上下文类型 '%s' 选择检测器: %T", point.Type, detector)
			results := detector.Detect(baseCtx, point)
			allResults = append(allResults, results...)
		}
	} else {
		logging.Logger.Infof("在 %s 未发现回显点。", initialCtx.Url)
	}

	// =================================================================
	// 阶段二: DOM型XSS分层检测
	// =================================================================
	// 无论是否存在反射点，都对原始响应进行一次静态分析

	// 第一层：快速字符串预筛选
	prefilterResult := e.domPrefilter.Analyze(initialCtx.Resp.Body)
	if prefilterResult.IsSuspicious {
		logging.Logger.Debugf("【DOM XSS Tier 1 Pass】URL: %s. 预筛选器发现 Sources: %v, Sinks: %v", initialCtx.Url, prefilterResult.FoundSources, prefilterResult.FoundSinks)

		// 第二层：AST语法验证
		astValidationResult := e.domASTValidator.Validate(initialCtx.Resp.Body, &prefilterResult)
		if astValidationResult.IsStillSuspicious {
			logging.Logger.Infof("【DOM XSS Tier 2 Pass】URL: %s. AST验证通过。Sources: %v, Sinks: %v. 高度可疑，启动动态分析...",
				initialCtx.Url, astValidationResult.ValidatedSources, astValidationResult.ValidatedSinks)

			// 第三层：启动基于 'rod' 的动态分析器
			dynamicDomDetector, err := dom.NewDynamicAnalyzer(e.config)
			if err != nil {
				logging.Logger.Errorf("创建动态DOM分析器失败: %v", err)
			} else {
				domVuln, err := dynamicDomDetector.Analyze(initialCtx.Url)
				if err != nil {
					logging.Logger.Errorf("动态DOM分析期间发生错误: %v", err)
				}
				if domVuln != nil {
					// 将Vulnerability包装在DetectionResult中
					detectionResult := types.DetectionResult{
						Found:    true,
						VulnInfo: domVuln.ToVulMessage(),
					}
					allResults = append(allResults, detectionResult)
				}
			}
		} else {
			logging.Logger.Debugf("【DOM XSS Tier 2 Fail】URL: %s. AST验证未通过，目标被确认为低风险。", initialCtx.Url)
		}
	} else {
		logging.Logger.Debugf("【DOM XSS Tier 1 Fail】URL: %s. 未发现可疑的Source/Sink组合。", initialCtx.Url)
	}

	// =================================================================
	// 阶段三: 结果处理
	// =================================================================
	e.processResults(allResults)

	logging.Logger.Infof("XSS扫描完成 (重构版引擎): %s", initialCtx.Url)
	return nil
}

// findInjectionPoints 执行回显检测，发现所有潜在的注入点
func (e *CoreEngine) findInjectionPoints(in *input.CrawlResult, client *httpx.Client) ([]types.InjectionPoint, *httpx.Variations, error) {
	var allPoints []types.InjectionPoint

	// 从请求中解析出可变参数
	requestHeaders, requestBody := parseRequestDump(in.Resp.RequestDump)
	contentType := requestHeaders["Content-Type"]
	variations, err := httpx.ParseUri(in.Url, []byte(requestBody), in.Method, contentType, requestHeaders)
	if err != nil {
		// 对于没有参数的GET请求等情况，这可能是正常的。
		// 即使解析失败，我们仍然继续，因为响应中可能包含我们可以攻击的表单。
		logging.Logger.Debugf("解析URI参数失败 (可能无参数): %v。将创建一个空的Variations对象继续。", err)
		variations = &httpx.Variations{
			Params: []httpx.Param{},
		}
	}

	// 从响应体中发现新的攻击参数（例如，表单）
	discoveredParams := e.discoverer.Discover(in.Resp)

	// 将发现的参数合并到现有参数列表中，并避免重复
	existingParamNames := make(map[string]struct{})
	for _, p := range variations.Params {
		existingParamNames[p.Name] = struct{}{}
	}

	for _, newParam := range discoveredParams {
		if _, exists := existingParamNames[newParam.Name]; !exists {
			variations.Params = append(variations.Params, newParam)
			existingParamNames[newParam.Name] = struct{}{}
			logging.Logger.Debugf("添加了从响应中发现的新参数进行测试: Name=%s", newParam.Name)
		}
	}

	// 如果最终没有任何参数可供测试，则直接返回
	if len(variations.Params) == 0 {
		logging.Logger.Debugf("在请求和响应中均未找到可测试的参数。")
		return nil, nil, nil
	}

	// 并发控制：使用信号量限制并发数
	maxConcurrent := 3
	if e.config != nil && e.config.Engine.MaxConcurrentRequests > 0 {
		maxConcurrent = e.config.Engine.MaxConcurrentRequests
	}
	semaphore := make(chan struct{}, maxConcurrent)

	// 用于收集结果的通道和同步
	type paramResult struct {
		points []types.InjectionPoint
		param  httpx.Param
		index  int
		err    error
	}
	resultChan := make(chan paramResult, len(variations.Params))

	// 并发测试每个参数
	for i, param := range variations.Params {
		go func(paramIndex int, p httpx.Param) {
			defer func() {
				if r := recover(); r != nil {
					logging.Logger.Errorf("参数检测过程中发生panic (参数: %s): %v", p.Name, r)
					resultChan <- paramResult{nil, p, paramIndex, fmt.Errorf("panic occurred: %v", r)}
				}
			}()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 测试单个参数
			points, err := e.testSingleParam(paramIndex, p, in, client, variations, requestHeaders)
			resultChan <- paramResult{points, p, paramIndex, err}
		}(i, param)
	}

	// 收集所有结果
	for i := 0; i < len(variations.Params); i++ {
		result := <-resultChan
		if result.err != nil {
			logging.Logger.Warnf("参数检测失败 (参数: %s): %v", result.param.Name, result.err)
			continue
		}
		allPoints = append(allPoints, result.points...)
	}

	return allPoints, variations, nil
}

// testSingleParam 测试单个参数的回显情况
func (e *CoreEngine) testSingleParam(paramIndex int, param httpx.Param, in *input.CrawlResult, client *httpx.Client, variations *httpx.Variations, requestHeaders map[string]string) ([]types.InjectionPoint, error) {
	// 增强的探针生成
	probe := e.generateProbe(param.Name)

	// 设置payload，发送请求
	payload := probe
	modifiedRequestData := variations.SetPayloadByIndex(paramIndex, in.Url, payload, in.Method)

	// 使用增强的重试机制发送请求
	resp, reqErr := e.requestWithRetry(client, in.Url, in.Method, modifiedRequestData, requestHeaders)

	if reqErr != nil {
		return nil, fmt.Errorf("发送回显探测请求失败 (参数: %s): %v", param.Name, reqErr)
	}

	// 使用增强的探针检测
	if e.detectProbeInResponse(resp.Body, probe) {
		logging.Logger.Infof("参数 '%s' 的值在响应中找到回显: %s", param.Name, probe)

		// 解析响应
		parsedResp, parseErr := e.parser.Parse(resp)
		if parseErr != nil {
			return nil, fmt.Errorf("解析回显响应失败: %v", parseErr)
		}

		// 分析上下文
		points := e.contextAnalyzer.Analyze(parsedResp, probe)
		for j := range points {
			points[j].BaseParsedResponse = parsedResp // 存储探测时的AST快照
			// 补充在 analyzer 中无法获取的参数信息
			points[j].Parameter = types.Parameter{
				Name:     param.Name,
				Position: getParamPosition(in.Method),
			}
			// 修复：只设置一次参数索引，避免重复
			if points[j].ContextDetails == nil {
				points[j].ContextDetails = make(map[string]string)
			}
			points[j].ContextDetails["param_index"] = strconv.Itoa(paramIndex)
		}
		return points, nil
	}

	return nil, nil
}

// generateProbe 生成一个多样化且不易被WAF检测的探针。
func (e *CoreEngine) generateProbe(paramName string) string {
	prefix := util.RandomLetters(6)
	suffix := util.RandomNumbers(2)
	// 引入一些变化，例如使用不同的分隔符
	separator := []string{"-", "_", "."}
	sep := separator[util.RandomInt(0, len(separator))]

	// 组合成最终的探针，例如: "aBcVxyz-paramName99"
	return fmt.Sprintf("%s%s%s%s", prefix, sep, paramName, suffix)
}

// detectProbeInResponse 增强的探针检测，支持常见编码
func (e *CoreEngine) detectProbeInResponse(body, probe string) bool {
	// 1. 检测原始探针
	if strings.Contains(body, probe) {
		return true
	}

	// 2. 检测HTML实体编码的探针
	htmlEncoded := html.EscapeString(probe)
	if htmlEncoded != probe && strings.Contains(body, htmlEncoded) {
		logging.Logger.Debugf("检测到HTML编码的探针: %s", htmlEncoded)
		return true
	}

	// 3. 检测URL编码的探针
	urlEncoded := url.QueryEscape(probe)
	if urlEncoded != probe && strings.Contains(body, urlEncoded) {
		logging.Logger.Debugf("检测到URL编码的探针: %s", urlEncoded)
		return true
	}

	// 4. 检测大小写变换的探针
	lowerProbe := strings.ToLower(probe)
	if lowerProbe != probe && strings.Contains(strings.ToLower(body), lowerProbe) {
		logging.Logger.Debugf("检测到小写转换的探针: %s", lowerProbe)
		return true
	}

	upperProbe := strings.ToUpper(probe)
	if upperProbe != probe && strings.Contains(strings.ToUpper(body), upperProbe) {
		logging.Logger.Debugf("检测到大写转换的探针: %s", upperProbe)
		return true
	}

	// 5. 检测部分探针（防止截断）
	if len(probe) > 6 {
		partialProbe := probe[:len(probe)-2] // 去掉最后2个字符
		if strings.Contains(body, partialProbe) {
			logging.Logger.Debugf("检测到部分探针（可能被截断）: %s", partialProbe)
			return true
		}
	}

	return false
}

// getParamPosition 根据HTTP方法判断参数位置
func getParamPosition(method string) string {
	if method == "GET" {
		return "query"
	}
	return "body"
}

// processResults 处理所有检测结果，并发送报告
func (e *CoreEngine) processResults(results []types.DetectionResult) {
	for _, result := range results {
		if result.Found && result.VulnInfo != nil {
			// 将漏洞信息发送到全局的输出channel
			output.OutChannel <- *result.VulnInfo
			logging.Logger.Infoln(fmt.Sprintf("已报告XSS漏洞: %s, Payload: %s", result.VulnInfo.VulnData.Target, result.VulnInfo.VulnData.Payload))
		}
	}
}

// parseRequestDump 解析原始HTTP请求字符串，返回头和body
func parseRequestDump(dump string) (map[string]string, string) {
	headers := make(map[string]string)
	// 标准化换行符
	dump = strings.ReplaceAll(dump, "\r\n", "\n")
	parts := strings.SplitN(dump, "\n\n", 2)
	if len(parts) < 1 {
		return headers, ""
	}

	headerLines := strings.Split(parts[0], "\n")
	for i, line := range headerLines {
		if i == 0 { // 跳过请求行 (e.g., "GET / HTTP/1.1")
			continue
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			headers[headerParts[0]] = headerParts[1]
		}
	}

	if len(parts) == 2 {
		return headers, parts[1]
	}

	return headers, ""
}

// requestWithRetry 执行带重试机制的HTTP请求
func (e *CoreEngine) requestWithRetry(client *httpx.Client, url, method string, data string, headers map[string]string) (*httpx.Response, error) {
	maxRetries := 3               // 默认重试3次
	retryDelay := 1 * time.Second // 默认延迟1秒

	// 如果有配置，使用配置的值
	if e.config != nil {
		if e.config.Engine.MaxRetries > 0 {
			maxRetries = e.config.Engine.MaxRetries
		}
		if e.config.Engine.RetryDelaySeconds > 0 {
			retryDelay = time.Duration(e.config.Engine.RetryDelaySeconds) * time.Second
		}
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		var resp *httpx.Response
		var err error

		if method == "GET" {
			resp, err = client.Request(data, method, "", headers, "XSS")
		} else {
			resp, err = client.Request(url, method, data, headers, "XSS")
		}

		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 如果不是最后一次尝试，等待后重试
		if attempt < maxRetries-1 {
			logging.Logger.Debugf("请求失败，%v后将进行第%d次重试: %v", retryDelay, attempt+2, err)
			time.Sleep(retryDelay)
			// 指数退避：每次重试延迟翻倍
			retryDelay *= 2
		}
	}

	return nil, fmt.Errorf("请求失败，已重试%d次: %w", maxRetries, lastErr)
}

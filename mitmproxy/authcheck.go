package mitmproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2025/6/24
   @desc 越权检测功能实现
   该文件实现了HTTP请求中的权限检测功能，可以自动替换或移除请求头来检测越权漏洞

   @update 2025/7/12: 改进为动态规则处理，规则修改后立即生效于新的请求，无需重启服务
**/

// ReplaceRule 表示一条替换规则
type ReplaceRule struct {
	ID            int    `json:"id"`
	Enabled       bool   `json:"enabled"`
	Type          string `json:"type"` // replace 或 remove
	HeaderName    string `json:"headerName"`
	OriginalValue string `json:"originalValue"`
	NewValue      string `json:"newValue"`
	Description   string `json:"description"`
}

// FilterCondition 表示过滤条件
type FilterCondition struct {
	OnlyInScope      bool     `json:"onlyInScope"`
	IncludeUrls      []string `json:"includeUrls"`
	ExcludeUrls      []string `json:"excludeUrls"`
	IncludeFileTypes []string `json:"includeFileTypes"`
	ExcludeFileTypes []string `json:"excludeFileTypes"`
}

// AuthorizationRules 表示所有越权检测规则
type AuthorizationRules struct {
	Rules           []ReplaceRule   `json:"rules"`
	FilterCondition FilterCondition `json:"filterCondition"`
}

// AuthTestResult 表示一次越权测试的结果
type AuthTestResult struct {
	ID               int64     `json:"id"`
	OriginalRequest  string    `json:"originalRequest"`
	OriginalResponse string    `json:"originalResponse"` // 原始请求的响应
	ModifiedRequest  string    `json:"modifiedRequest"`
	ModifiedResponse string    `json:"modifiedResponse"` // 修改后请求的响应
	RuleID           int       `json:"ruleID"`
	RuleDescription  string    `json:"ruleDescription"`
	StatusCode       int       `json:"statusCode"`     // 修改后请求的状态码
	OriginalStatus   int       `json:"originalStatus"` // 原始请求的状态码
	URL              string    `json:"url"`
	Method           string    `json:"method"` // HTTP方法
	Timestamp        time.Time `json:"timestamp"`
}

var (
	// 用于存储和管理越权检测规则
	authorizationRules AuthorizationRules

	// 规则文件路径
	authRulesFilePath string

	// 是否启用越权检测
	AuthorizationCheckEnabled atomic.Bool

	// 用于存储越权检测结果
	authTestResults     []AuthTestResult
	authTestResultsLock sync.RWMutex
	authTestResultID    atomic.Int64

	// 用于将请求转发到越权检测界面的通道
	authTestResultChan chan AuthTestResult

	// 全局httpx客户端实例，用于复用连接池
	authHTTPClient *httpx.Client
	authClientOnce sync.Once
)

// getAuthHTTPClient 获取全局的httpx客户端实例（懒加载）
func getAuthHTTPClient() *httpx.Client {
	authClientOnce.Do(func() {
		options := &httpx.Options{
			Timeout:         10, // 10秒超时
			VerifySSL:       false,
			AllowRedirect:   0,  // 不允许重定向
			RetryTimes:      0,  // 不重试
			QPS:             50, // 设置QPS限制，避免过于频繁的请求
			MaxConnsPerHost: 10, // 每个host最大连接数
		}
		authHTTPClient = httpx.NewClient(options)
		logging.Logger.Debugf("Created global auth HTTP client for authorization checking")
	})
	return authHTTPClient
}

// RegisterAuthorizationProcessor 注册越权检测处理器到插件系统
func RegisterAuthorizationProcessor() {
	logging.Logger.Infof("正在注册越权检测处理器...")
	// 注册响应处理器，在响应阶段进行越权检测
	// 使用 ReadOnly 模式确保不会修改原始响应
	RegisterResponseProcessorWithMode(authorizationResponseProcessor, ReadOnly)

	// 初始化结果通道
	authTestResultChan = make(chan AuthTestResult, 100)

	// 启动处理线程
	go processAuthTestResults()
}

// 处理越权测试结果的后台线程
func processAuthTestResults() {
	for result := range authTestResultChan {
		authTestResultsLock.Lock()
		authTestResults = append(authTestResults, result)
		authTestResultsLock.Unlock()
		logging.Logger.Infof("越权测试结果: URL=%s, 规则=%s, 状态码=%d",
			result.URL, result.RuleDescription, result.StatusCode)
	}
}

// authorizationResponseProcessor 响应处理器，在响应阶段进行越权检测
func authorizationResponseProcessor(resp *http.Response) bool {
	if !AuthorizationCheckEnabled.Load() {
		return false
	}

	// 检查过滤条件
	if resp.Request == nil || shouldSkipCheck(resp.Request) {
		return false
	}

	// 异步进行越权测试，使用已有的响应数据
	go testRequestWithAuthRulesUsingExistingResponse(resp.Request, resp)

	// 返回false表示没有修改原始响应
	return false
}

// convertHTTPRequestToParams 将 http.Request 转换为 httpx.Request 需要的参数
func convertHTTPRequestToParams(req *http.Request) (target string, method string, body string, header map[string]string, err error) {
	if req == nil {
		return "", "", "", nil, fmt.Errorf("request is nil")
	}

	// 获取完整的URL
	target = req.URL.String()

	// 获取方法
	method = req.Method

	// 获取请求体
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return "", "", "", nil, fmt.Errorf("failed to read request body: %v", err)
		}
		body = string(bodyBytes)
		// 重新设置Body，以便后续可以继续使用
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 转换请求头
	header = make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			header[key] = values[0] // 只取第一个值
		}
	}

	return target, method, body, header, nil
}

// testRequestWithAuthRulesUsingExistingResponse 使用已有响应数据进行越权测试
func testRequestWithAuthRulesUsingExistingResponse(originalReq *http.Request, originalResp *http.Response) {
	// 获取原始请求的dump
	originalReqDump, err := httputil.DumpRequestOut(originalReq, true)
	if err != nil {
		logging.Logger.Errorf("Failed to dump original request: %v", err)
		return
	}

	// 获取原始响应的dump（使用解码后的响应body）
	originalRespDumpString, err := GetResponseDumpWithDecodedBody(originalResp)
	if err != nil {
		logging.Logger.Errorf("Failed to dump original response: %v", err)
		// 如果获取解码后的响应失败，尝试使用原始响应
		if originalResp != nil {
			originalRespDumpBytes, dumpErr := httputil.DumpResponse(originalResp, true)
			if dumpErr == nil {
				originalRespDumpString = string(originalRespDumpBytes)
			}
		}
	}

	// 提取原始响应的状态码
	originalStatus := 200 // 默认值
	if originalResp != nil {
		originalStatus = originalResp.StatusCode
	}

	// 对每个启用的规则进行测试
	enabledRulesCount := 0
	for _, rule := range authorizationRules.Rules {
		if !rule.Enabled {
			continue
		}
		enabledRulesCount++

		// 深度复制请求进行修改测试
		testReq := authCloneRequest(originalReq)
		if testReq == nil {
			logging.Logger.Errorf("Failed to clone request for rule %d, URL: %s", rule.ID, originalReq.URL.String())
			continue
		}

		// 应用规则修改请求头
		headerModified := applyRuleToRequest(testReq, rule)
		if !headerModified {
			logging.Logger.Debugf("Rule %d skipped - header %s not found or not modified in request to %s", rule.ID, rule.HeaderName, originalReq.URL.String())
			continue
		}

		// 获取修改后请求的dump
		modifiedReqDump, err := httputil.DumpRequestOut(testReq, true)
		if err != nil {
			logging.Logger.Errorf("Failed to dump modified request: %v", err)
			continue
		}

		// 使用全局httpx客户端实例（复用连接池）
		client := getAuthHTTPClient()

		// 转换修改后的请求参数，只发送修改后的请求
		modifiedTarget, modifiedMethod, modifiedBody, modifiedHeader, err := convertHTTPRequestToParams(testReq)
		if err != nil {
			logging.Logger.Errorf("Failed to convert modified request parameters: %v", err)
			continue
		}
		// 只执行修改后的请求获取响应
		modifiedResp, err := client.Request(modifiedTarget, modifiedMethod, modifiedBody, modifiedHeader, "Authcheck-Modified")
		if err != nil {
			logging.Logger.Errorf("Failed to execute modified request: %v", err)
			continue
		}

		var modifiedRespDumpString string
		if modifiedResp != nil {
			modifiedRespDumpString = modifiedResp.ResponseDump
		}

		// 创建测试结果
		result := AuthTestResult{
			ID:               authTestResultID.Add(1),
			OriginalRequest:  string(originalReqDump),
			OriginalResponse: originalRespDumpString, // 使用代理系统已有的响应数据
			ModifiedRequest:  string(modifiedReqDump),
			ModifiedResponse: modifiedRespDumpString,
			RuleID:           rule.ID,
			RuleDescription:  rule.Description,
			StatusCode:       modifiedResp.StatusCode,
			OriginalStatus:   originalStatus, // 使用已有响应的状态码
			URL:              originalReq.URL.String(),
			Method:           originalReq.Method,
			Timestamp:        time.Now(),
		}

		authTestResultChan <- result
	}
}

// InitAuthorizationChecker 初始化越权检测模块
func InitAuthorizationChecker() error {
	// 确定规则文件路径 - 使用统一的配置目录
	proxyDataDir := filepath.Join(file.ChyingDir, "proxify_data")
	authRulesFilePath = filepath.Join(proxyDataDir, "auth_checker_rules.json")
	// 确保目录存在
	if err := os.MkdirAll(proxyDataDir, 0755); err != nil {
		logging.Logger.Errorf("could not create config directory for authorization checker: %v", err)
		return fmt.Errorf("could not create config directory for authorization checker: %v", err)
	}

	// 加载规则
	err := LoadAuthorizationRules()
	if err != nil {
		logging.Logger.Errorf("LoadAuthorizationRules error: %v", err)
		return err
	}

	// 注册到插件系统
	RegisterAuthorizationProcessor()

	return nil
}

// LoadAuthorizationRules 从文件加载越权检测规则
func LoadAuthorizationRules() error {
	// 检查文件是否存在
	if _, err := os.Stat(authRulesFilePath); os.IsNotExist(err) {
		// 文件不存在，初始化为默认值
		authorizationRules = AuthorizationRules{
			Rules: []ReplaceRule{
				{
					ID:            1,
					Enabled:       true,
					Type:          "replace",
					HeaderName:    "Authorization",
					OriginalValue: "Bearer .*",
					NewValue:      "Bearer invalid-token",
					Description:   "替换JWT令牌以测试认证",
				},
				{
					ID:            2,
					Enabled:       true,
					Type:          "remove",
					HeaderName:    "Authorization",
					OriginalValue: "",
					NewValue:      "",
					Description:   "移除Authorization头以测试未认证访问",
				},
				{
					ID:            3,
					Enabled:       true,
					Type:          "replace",
					HeaderName:    "Cookie",
					OriginalValue: "session=.*?;",
					NewValue:      "session=guest-session;",
					Description:   "替换会话cookie为访客会话",
				},
			},
			FilterCondition: FilterCondition{
				OnlyInScope:      false,                                        // 改为 false，允许所有请求
				IncludeUrls:      []string{},                                   // 清空包含列表，或保留但不作为强制要求
				ExcludeUrls:      []string{"/public/", "/static/", "/assets/"}, // 保留常见静态资源排除
				IncludeFileTypes: []string{},
				ExcludeFileTypes: []string{"jpg", "png", "gif", "css", "js", "ico", "svg", "woff", "woff2", "ttf"}, // 扩展静态文件排除
			},
		}
		// 保存默认配置
		return SaveAuthorizationRulesToFile()
	}

	logging.Logger.Infof("authRulesFilePath: %s", authRulesFilePath)
	// 读取文件内容
	data, err := os.ReadFile(authRulesFilePath)
	if err != nil {
		return fmt.Errorf("could not read authorization rules file: %v", err)
	}

	// 解析JSON数据
	if err := json.Unmarshal(data, &authorizationRules); err != nil {
		// 解析失败，初始化为默认值
		authorizationRules = AuthorizationRules{
			Rules: []ReplaceRule{},
			FilterCondition: FilterCondition{
				OnlyInScope: false,
			},
		}
		return fmt.Errorf("could not parse authorization rules file: %v", err)
	}

	return nil
}

// SaveAuthorizationRulesToFile 将规则保存到文件
func SaveAuthorizationRulesToFile() error {
	// 转换为JSON
	data, err := json.MarshalIndent(authorizationRules, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal authorization rules: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(authRulesFilePath, data, 0644); err != nil {
		return fmt.Errorf("could not write authorization rules file: %v", err)
	}

	return nil
}

// GetAuthorizationRules 获取所有越权检测规则
func GetAuthorizationRules() (AuthorizationRules, error) {
	// 返回当前规则副本
	return authorizationRules, nil
}

// SaveAuthorizationRules 保存越权检测规则
func SaveAuthorizationRules(rules AuthorizationRules) error {
	// 更新规则
	authorizationRules = rules

	// 保存到文件
	if err := SaveAuthorizationRulesToFile(); err != nil {
		return err
	}

	// 记录明确的日志，表明规则已更新并将立即生效
	logging.Logger.Infof("越权检测规则已更新，包含 %d 条规则，立即生效于新的请求",
		len(rules.Rules))

	return nil
}

// StartAuthorizationCheck 开始越权检测
func StartAuthorizationCheck() error {
	if AuthorizationCheckEnabled.Load() {
		logging.Logger.Infof("越权检测已经在运行中")
		return nil // 已经在运行
	}

	// 启用越权检测
	AuthorizationCheckEnabled.Store(true)
	logging.Logger.Infof("越权检测已启动，将对符合条件的请求执行检测")
	return nil
}

// StopAuthorizationCheck 停止越权检测
func StopAuthorizationCheck() error {
	if !AuthorizationCheckEnabled.Load() {
		logging.Logger.Infof("越权检测已经停止")
		return nil // 已经停止
	}

	// 禁用越权检测
	AuthorizationCheckEnabled.Store(false)
	logging.Logger.Infof("越权检测已停止，不再对请求执行检测")
	return nil
}

// shouldSkipCheck 检查是否应该跳过此请求的检测
func shouldSkipCheck(req *http.Request) bool {
	url := req.URL.String()

	// 检查是否仅对作用域内请求应用规则
	if authorizationRules.FilterCondition.OnlyInScope {
		// 只有当 OnlyInScope 为 true 且 IncludeUrls 不为空时才进行严格过滤
		if len(authorizationRules.FilterCondition.IncludeUrls) > 0 {
			inScope := false
			// 检查是否匹配包含URL列表
			for _, includeUrl := range authorizationRules.FilterCondition.IncludeUrls {
				if strings.Contains(url, includeUrl) {
					inScope = true
					break
				}
			}

			if !inScope {
				logging.Logger.Debugf("Auth check skipped (not in scope): %s", url)
				return true
			}
		}
	}

	// 检查是否匹配排除URL列表
	for _, excludeUrl := range authorizationRules.FilterCondition.ExcludeUrls {
		if strings.Contains(url, excludeUrl) {
			logging.Logger.Debugf("Auth check skipped (excluded URL): %s", url)
			return true
		}
	}

	// 检查文件类型
	ext := filepath.Ext(req.URL.Path)
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")

		// 检查是否在排除文件类型列表中
		for _, excludeType := range authorizationRules.FilterCondition.ExcludeFileTypes {
			if strings.EqualFold(ext, excludeType) {
				logging.Logger.Debugf("Auth check skipped (excluded file type %s): %s", ext, url)
				return true
			}
		}

		// 如果包含文件类型列表不为空，检查是否在其中
		if len(authorizationRules.FilterCondition.IncludeFileTypes) > 0 {
			included := false
			for _, includeType := range authorizationRules.FilterCondition.IncludeFileTypes {
				if strings.EqualFold(ext, includeType) {
					included = true
					break
				}
			}

			if !included {
				logging.Logger.Debugf("Auth check skipped (file type not included %s): %s", ext, url)
				return true
			}
		}
	}

	logging.Logger.Debugf("Auth check processing: %s", url)
	return false
}

// GetAuthorizationTestResults 获取越权检测结果
func GetAuthorizationTestResults() []AuthTestResult {
	authTestResultsLock.RLock()
	defer authTestResultsLock.RUnlock()

	// 返回结果副本
	results := make([]AuthTestResult, len(authTestResults))
	copy(results, authTestResults)
	return results
}

// ClearAuthorizationTestResults 清除越权检测结果
func ClearAuthorizationTestResults() {
	authTestResultsLock.Lock()
	defer authTestResultsLock.Unlock()

	authTestResults = nil
}

// cloneRequest 创建请求的深度副本
func authCloneRequest(originalReq *http.Request) *http.Request {
	if originalReq == nil {
		return nil
	}

	// 创建新的URL对象
	newURL := *originalReq.URL

	// 读取body
	var bodyBytes []byte
	if originalReq.Body != nil && originalReq.Body != http.NoBody {
		bodyBytes, _ = io.ReadAll(originalReq.Body)
		originalReq.Body.Close()
		// 恢复原始请求的body
		originalReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 创建新请求
	newReq := &http.Request{
		Method:        originalReq.Method,
		URL:           &newURL,
		Proto:         originalReq.Proto,
		ProtoMajor:    originalReq.ProtoMajor,
		ProtoMinor:    originalReq.ProtoMinor,
		Header:        make(http.Header),
		Host:          originalReq.Host,
		ContentLength: originalReq.ContentLength,
		Close:         originalReq.Close,
	}

	// 复制header
	for key, values := range originalReq.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	// 设置body
	if bodyBytes != nil {
		newReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	return newReq
}

// applyRuleToRequest 应用规则到请求，返回是否修改了请求头
func applyRuleToRequest(req *http.Request, rule ReplaceRule) bool {
	headerValue := req.Header.Get(rule.HeaderName)
	if headerValue == "" {
		return false // 请求中不存在该头部
	}

	// 为正则匹配构造完整的 header 字符串 (HeaderName: HeaderValue)
	fullHeaderValue := rule.HeaderName + ": " + headerValue

	if rule.Type == "remove" {
		// 移除头部
		req.Header.Del(rule.HeaderName)
		return true
	} else if rule.Type == "replace" {
		// 如果指定了原始值模式，检查是否匹配
		if rule.OriginalValue != "" {
			re, err := regexp.Compile(rule.OriginalValue)
			if err != nil {
				logging.Logger.Infof("Invalid regex pattern: %v", err)
				return false
			}

			// 使用完整的 header 字符串进行匹配
			if !re.MatchString(fullHeaderValue) {
				return false // 不匹配模式
			}

			// 如果匹配，对完整字符串进行替换，然后提取新的值部分
			replacedFullHeader := re.ReplaceAllString(fullHeaderValue, rule.NewValue)
			// 从替换后的完整字符串中提取新的 header 值 (去掉 "HeaderName: " 前缀)
			if strings.HasPrefix(replacedFullHeader, rule.HeaderName+": ") {
				newValue := strings.TrimPrefix(replacedFullHeader, rule.HeaderName+": ")
				req.Header.Set(rule.HeaderName, newValue)
				return true
			} else {
				// 如果替换后不包含 header 前缀，直接使用替换结果
				req.Header.Set(rule.HeaderName, replacedFullHeader)
				return true
			}
		} else {
			// 没有指定原始值模式，直接替换
			req.Header.Set(rule.HeaderName, rule.NewValue)
			return true
		}
	}

	return false
}

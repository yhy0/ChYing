package mitmproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2025/6/15
   @desc 匹配替换规则管理功能，用于动态修改请求和响应
   该文件实现了基于正则表达式的匹配替换功能，支持对请求头、请求体、响应头、响应体等进行修改
**/

// 规则类型常量
const (
	RuleTypeRequestHeader     = "request_header"
	RuleTypeRequestBody       = "request_body"
	RuleTypeResponseHeader    = "response_header"
	RuleTypeResponseBody      = "response_body"
	RuleTypeRequestFirstLine  = "request_first_line"
	RuleTypeRequestParamName  = "request_param_name"
	RuleTypeRequestParamValue = "request_param_value"
)

// MatchReplaceRule 表示单个匹配替换规则
type MatchReplaceRule struct {
	ID      int    `json:"id"`
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Match   string `json:"match"`
	Replace string `json:"replace"`
	Comment string `json:"comment"`
}

// MatchReplaceRules 表示所有匹配替换规则
type MatchReplaceRules struct {
	Rules []MatchReplaceRule `json:"rules"`
}

var (
	// 用于存储和管理匹配替换规则
	matchReplaceRules MatchReplaceRules

	// 规则文件路径
	rulesFilePath string
)

// InitMatchReplaceRules 初始化匹配替换规则系统
func InitMatchReplaceRules() error {
	// 确定规则文件路径
	proxyDataDir := filepath.Join(file.ChyingDir, "proxify_data")
	rulesFilePath = filepath.Join(proxyDataDir, "match_replace_rules.json")

	// 确保目录存在
	if err := os.MkdirAll(proxyDataDir, 0755); err != nil {
		return fmt.Errorf("could not create config directory for match replace rules: %v", err)
	}

	// 注册请求和响应处理器（使用 Modifying 模式，允许直接修改原始请求/响应）
	RegisterRequestProcessorWithMode(MatchReplaceRequestProcessor, Modifying)
	RegisterResponseProcessorWithMode(MatchReplaceResponseProcessor, Modifying)
	logging.Logger.Infoln("匹配替换规则处理器已注册（Modifying 模式）")

	// 加载规则
	return LoadMatchReplaceRules()
}

// LoadMatchReplaceRules 从文件加载匹配替换规则
func LoadMatchReplaceRules() error {
	// 检查文件是否存在
	if _, err := os.Stat(rulesFilePath); os.IsNotExist(err) {
		// 文件不存在，初始化为默认值，并添加预设规则示例
		matchReplaceRules = MatchReplaceRules{
			Rules: createDefaultRules(),
		}
		// 保存默认配置
		return SaveMatchReplaceRulesToFile()
	}

	// 读取文件内容
	data, err := os.ReadFile(rulesFilePath)
	if err != nil {
		return fmt.Errorf("could not read match replace rules file: %v", err)
	}

	// 解析JSON数据
	if err := json.Unmarshal(data, &matchReplaceRules); err != nil {
		// 解析失败，初始化为默认值，并添加预设规则示例
		matchReplaceRules = MatchReplaceRules{
			Rules: createDefaultRules(),
		}
		return fmt.Errorf("could not parse match replace rules file: %v", err)
	}

	// 如果规则列表为空，添加预设规则示例
	if len(matchReplaceRules.Rules) == 0 {
		matchReplaceRules.Rules = createDefaultRules()
		// 保存到文件
		if err := SaveMatchReplaceRulesToFile(); err != nil {
			fmt.Printf("Warning: failed to save default rules: %v\n", err)
		}
	}

	logging.Logger.Infof("已加载 %d 条匹配替换规则", len(matchReplaceRules.Rules))
	return nil
}

// SaveMatchReplaceRulesToFile 将规则保存到文件
func SaveMatchReplaceRulesToFile() error {
	// 转换为JSON
	data, err := json.MarshalIndent(matchReplaceRules, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal match replace rules: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(rulesFilePath, data, 0644); err != nil {
		return fmt.Errorf("could not write match replace rules file: %v", err)
	}

	return nil
}

// GetMatchReplaceRules 获取所有匹配替换规则
func GetMatchReplaceRules() (MatchReplaceRules, error) {
	// 返回当前规则副本
	return matchReplaceRules, nil
}

// SaveMatchReplaceRule 保存或更新单个规则
func SaveMatchReplaceRule(rule MatchReplaceRule) error {
	// 检查规则类型是否有效
	if !isValidRuleType(rule.Type) {
		return fmt.Errorf("invalid rule type: %s", rule.Type)
	}

	// 检查规则的ID是更新还是新增
	if rule.ID > 0 {
		// 查找并更新现有规则
		found := false
		for i, existingRule := range matchReplaceRules.Rules {
			if existingRule.ID == rule.ID {
				matchReplaceRules.Rules[i] = rule
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("rule with ID %d not found", rule.ID)
		}
	} else {
		// 新规则，分配ID
		maxID := 0
		for _, existingRule := range matchReplaceRules.Rules {
			if existingRule.ID > maxID {
				maxID = existingRule.ID
			}
		}
		rule.ID = maxID + 1

		// 添加到规则列表
		matchReplaceRules.Rules = append(matchReplaceRules.Rules, rule)
	}

	// 保存到文件
	if err := SaveMatchReplaceRulesToFile(); err != nil {
		return err
	}

	logging.Logger.Infof("已保存规则: ID=%d, 名称=%s", rule.ID, rule.Name)
	return nil
}

// DeleteMatchReplaceRule 删除单个规则
func DeleteMatchReplaceRule(ruleID int) error {
	// 查找并删除规则
	found := false
	var newRules []MatchReplaceRule
	for _, rule := range matchReplaceRules.Rules {
		if rule.ID != ruleID {
			newRules = append(newRules, rule)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("rule with ID %d not found", ruleID)
	}

	matchReplaceRules.Rules = newRules

	// 保存到文件
	if err := SaveMatchReplaceRulesToFile(); err != nil {
		return err
	}

	logging.Logger.Infof("已删除规则: ID=%d", ruleID)
	return nil
}

// ApplyMatchReplaceRulesWithConfig 使用提供的配置应用规则
func ApplyMatchReplaceRulesWithConfig(config MatchReplaceRules) error {
	// 更新规则列表
	matchReplaceRules.Rules = config.Rules

	// 保存到文件
	logging.Logger.Infof("更新匹配替换规则配置，共 %d 条规则", len(config.Rules))
	return SaveMatchReplaceRulesToFile()
}

// MatchReplaceRequestProcessor 处理请求的匹配替换规则
func MatchReplaceRequestProcessor(req *http.Request) bool {
	modified := false

	// 筛选出已启用的请求规则
	var enabledRules []MatchReplaceRule
	for _, rule := range matchReplaceRules.Rules {
		if rule.Enabled && isRequestRuleType(rule.Type) {
			enabledRules = append(enabledRules, rule)
		}
	}

	if len(enabledRules) == 0 {
		return false // 没有启用的规则
	}

	// 处理请求首行 (method + path + protocol)
	for _, rule := range enabledRules {
		if rule.Type == RuleTypeRequestFirstLine {
			firstLine := fmt.Sprintf("%s %s %s", req.Method, req.URL.RequestURI(), req.Proto)
			re, err := regexp.Compile(rule.Match)
			if err != nil {
				logging.Logger.Errorf("规则 %d 正则表达式错误: %v", rule.ID, err)
				continue
			}

			newFirstLine := re.ReplaceAllString(firstLine, rule.Replace)
			if newFirstLine != firstLine {
				// 解析新的首行
				parts := strings.SplitN(newFirstLine, " ", 3)
				if len(parts) >= 2 {
					// 修改方法
					if parts[0] != req.Method {
						req.Method = parts[0]
						modified = true
					}

					// 修改URL路径
					if parts[1] != req.URL.RequestURI() {
						newURL, err := url.Parse(parts[1])
						if err == nil {
							req.URL.Path = newURL.Path
							req.URL.RawQuery = newURL.RawQuery
							modified = true
						}
					}
				}

				logging.Logger.Infof("已应用请求首行替换规则 %d: %s", rule.ID, rule.Name)
			}
		}
	}

	// 处理请求参数名和值
	for _, rule := range enabledRules {
		if (rule.Type == RuleTypeRequestParamName || rule.Type == RuleTypeRequestParamValue) && len(req.URL.Query()) > 0 {
			// 创建一个新的query参数集合
			newQuery := url.Values{}
			re, err := regexp.Compile(rule.Match)
			if err != nil {
				logging.Logger.Errorf("规则 %d 正则表达式错误: %v", rule.ID, err)
				continue
			}

			paramModified := false

			// 处理每个参数
			for key, values := range req.URL.Query() {
				newKey := key

				// 处理参数名
				if rule.Type == RuleTypeRequestParamName {
					newKey = re.ReplaceAllString(key, rule.Replace)
					if newKey != key {
						paramModified = true
					}
				}

				// 处理参数值
				newValues := make([]string, len(values))
				for i, value := range values {
					newValues[i] = value
					if rule.Type == RuleTypeRequestParamValue {
						newValues[i] = re.ReplaceAllString(value, rule.Replace)
						if newValues[i] != value {
							paramModified = true
						}
					}
				}

				// 添加到新的参数集合
				for _, v := range newValues {
					newQuery.Add(newKey, v)
				}
			}

			// 如果有修改，更新URL
			if paramModified {
				req.URL.RawQuery = newQuery.Encode()
				modified = true
				logging.Logger.Infof("已应用请求参数替换规则 %d: %s", rule.ID, rule.Name)
			}
		}
	}

	// 应用请求头规则
	for _, rule := range enabledRules {
		if rule.Type == RuleTypeRequestHeader {
			headerModified := applyHeaderReplaceRule(req.Header, rule)
			if headerModified {
				modified = true
				logging.Logger.Infof("已应用请求头替换规则 %d: %s, %+v", rule.ID, rule.Name, req.Header)
			}
		}
	}

	// 应用请求体规则
	for _, rule := range enabledRules {
		if rule.Type == RuleTypeRequestBody && req.Body != nil {
			// 读取请求体
			body, err := readRequestBody(req)
			if err != nil {
				logging.Logger.Errorf("读取请求体失败: %v", err)
				continue
			}

			// 应用替换
			newBody, bodyModified := applyBodyReplaceRule(body, rule)
			if bodyModified {
				resetRequestBody(req, newBody)
				modified = true
				logging.Logger.Infof("已应用请求体替换规则 %d: %s", rule.ID, rule.Name)
			}
		}
	}

	return modified
}

// MatchReplaceResponseProcessor 处理响应的匹配替换规则
func MatchReplaceResponseProcessor(resp *http.Response) bool {
	modified := false

	// 筛选出已启用的响应规则
	var enabledRules []MatchReplaceRule
	for _, rule := range matchReplaceRules.Rules {
		if rule.Enabled && isResponseRuleType(rule.Type) {
			enabledRules = append(enabledRules, rule)
		}
	}

	if len(enabledRules) == 0 {
		return false // 没有启用的规则
	}

	// 应用响应头规则
	for _, rule := range enabledRules {
		if rule.Type == RuleTypeResponseHeader {
			headerModified := applyHeaderReplaceRule(resp.Header, rule)
			if headerModified {
				modified = true
				logging.Logger.Infof("已应用响应头替换规则 %d: %s", rule.ID, rule.Name)
			}
		}
	}

	// 应用响应体规则
	for _, rule := range enabledRules {
		if rule.Type == RuleTypeResponseBody && resp.Body != nil {
			// 使用 httputil.go 中的 ProcessResponseBody 处理响应体
			processed, err := ProcessResponseBody(resp)
			if err != nil {
				logging.Logger.Errorf("处理响应体失败: %v", err)
				continue
			}

			// 只处理文本内容
			if processed.IsText {
				// 将处理后的内容转换为字符串并应用替换规则
				bodyStr := string(processed.Content)
				logging.Logger.Infof("正在处理文本内容，应用替换规则: %s", rule.Name)

				// 应用替换规则
				newBody, bodyModified := applyBodyReplaceRule(bodyStr, rule)
				if bodyModified {
					// 标记内容已修改
					processed.Content = []byte(newBody)
					SetContentModified(processed)

					// 使用 httputil.go 中的 UpdateResponseWithProcessedBody 更新响应
					if err := UpdateResponseWithProcessedBody(resp, processed, true); err != nil {
						logging.Logger.Errorf("更新响应体失败: %v", err)
					} else {
						modified = true
						logging.Logger.Infof("已应用响应体替换规则 %d: %s，内容长度: %d -> %d 字节",
							rule.ID, rule.Name, len(bodyStr), len(newBody))
					}
				}
			} else {
				// 非文本内容，恢复原始内容
				if err := UpdateResponseWithProcessedBody(resp, processed, true); err != nil {
					logging.Logger.Errorf("恢复原始内容失败: %v", err)
				}
				logging.Logger.Infof("跳过非文本内容的响应体替换，内容类型: %s", processed.ContentType)
			}
		}
	}

	return modified
}

// 判断规则类型是否有效
func isValidRuleType(ruleType string) bool {
	validTypes := []string{
		RuleTypeRequestHeader,
		RuleTypeRequestBody,
		RuleTypeResponseHeader,
		RuleTypeResponseBody,
		RuleTypeRequestFirstLine,
		RuleTypeRequestParamName,
		RuleTypeRequestParamValue,
	}

	for _, validType := range validTypes {
		if ruleType == validType {
			return true
		}
	}

	return false
}

// 判断规则是否适用于请求
func isRequestRuleType(ruleType string) bool {
	return ruleType == RuleTypeRequestHeader ||
		ruleType == RuleTypeRequestBody ||
		ruleType == RuleTypeRequestFirstLine ||
		ruleType == RuleTypeRequestParamName ||
		ruleType == RuleTypeRequestParamValue
}

// 判断规则是否适用于响应
func isResponseRuleType(ruleType string) bool {
	return ruleType == RuleTypeResponseHeader ||
		ruleType == RuleTypeResponseBody
}

// 应用头部替换规则
func applyHeaderReplaceRule(header http.Header, rule MatchReplaceRule) bool {
	modified := false

	// 使用正则表达式实现规则匹配和替换
	re, err := regexp.Compile(rule.Match)
	if err != nil {
		logging.Logger.Errorf("规则 %d 正则表达式错误: %v", rule.ID, err)
		return false
	}

	// 创建一个临时map来存储修改后的头部
	// 这样做是因为在遍历map时直接修改它可能导致迭代问题
	newHeaders := make(map[string][]string)
	deletedHeaders := make(map[string]bool)

	// 遍历所有头部
	for name, values := range header {
		// 检查是否匹配整个头部，如 "User-Agent: xxx"
		headerFullFormat := fmt.Sprintf("%s: %s", name, strings.Join(values, ", "))
		if re.MatchString(headerFullFormat) {
			// 对整个头部应用替换
			newHeaderLine := re.ReplaceAllString(headerFullFormat, rule.Replace)
			// 解析回头部名称和值
			parts := strings.SplitN(newHeaderLine, ": ", 2)
			if len(parts) == 2 {
				newName, newValue := parts[0], parts[1]

				// 如果头部名称发生变化，标记原头部需要删除
				if newName != name {
					deletedHeaders[name] = true
				}

				// 添加新头部
				newHeaders[newName] = []string{newValue}
				modified = true
				continue
			}
		}

		// 否则逐个检查头部值
		newValues := make([]string, 0, len(values))
		valueModified := false

		for _, value := range values {
			if re.MatchString(value) {
				newValue := re.ReplaceAllString(value, rule.Replace)
				if newValue != value {
					valueModified = true
				}
				newValues = append(newValues, newValue)
			} else {
				newValues = append(newValues, value)
			}
		}

		if valueModified {
			newHeaders[name] = newValues
			modified = true
		}
	}

	// 应用头部修改
	for name := range deletedHeaders {
		header.Del(name)
	}

	for name, values := range newHeaders {
		// 如果是需要删除并重新添加的头部
		if deletedHeaders[name] {
			header.Del(name)
		}

		// 添加新值
		for _, value := range values {
			// 先删除，在新增
			header.Del(name)
			header.Add(name, value)
			logging.Logger.Debug(name, value)
		}
	}
	return modified
}

// 应用正文替换规则
func applyBodyReplaceRule(body string, rule MatchReplaceRule) (string, bool) {
	re, err := regexp.Compile(rule.Match)
	if err != nil {
		logging.Logger.Errorf("规则 %d 正则表达式错误: %v", rule.ID, err)
		return body, false
	}

	newBody := re.ReplaceAllString(body, rule.Replace)
	return newBody, newBody != body
}

// 读取请求体并重置以便后续读取
func readRequestBody(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	// 重置请求体以便后续处理
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return string(body), nil
}

// 重置请求体
func resetRequestBody(req *http.Request, newBody string) {
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBufferString(newBody))
	req.ContentLength = int64(len(newBody))
	// 如果有Content-Length头部，也更新它
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
}

// createDefaultRules 创建默认的示例规则集
func createDefaultRules() []MatchReplaceRule {
	return []MatchReplaceRule{
		{
			ID:      1,
			Enabled: false,
			Type:    RuleTypeRequestHeader,
			Name:    "修改User-Agent请求头",
			Match:   "User-Agent: .*",
			Replace: "User-Agent: Mozilla/5.0 (ChYing-Inside Security Scanner)",
			Comment: "示例：将所有请求的User-Agent修改为ChYing-Inside标识",
		},
		{
			ID:      2,
			Enabled: false,
			Type:    RuleTypeRequestBody,
			Name:    "隐藏请求中的密码",
			Match:   "\"password\":\"([^\"]+)\"",
			Replace: "\"password\":\"******\"",
			Comment: "示例：将JSON请求体中的密码替换为星号",
		},
		{
			ID:      3,
			Enabled: false,
			Type:    RuleTypeResponseHeader,
			Name:    "移除服务器信息",
			Match:   "Server: (.*)",
			Replace: "Server: Unknown",
			Comment: "示例：隐藏响应头中的服务器版本信息",
		},
		{
			ID:      4,
			Enabled: false,
			Type:    RuleTypeResponseBody,
			Name:    "替换敏感内容",
			Match:   "<title>(.*?)</title>",
			Replace: "<title>已修改的标题</title>",
			Comment: "示例：修改响应中的HTML标题",
		},
		{
			ID:      5,
			Enabled: false,
			Type:    RuleTypeRequestFirstLine,
			Name:    "修改请求路径",
			Match:   "POST /api/login",
			Replace: "POST /api/auth/login",
			Comment: "示例：修改请求首行中的URL路径",
		},
		{
			ID:      6,
			Enabled: false,
			Type:    RuleTypeRequestParamName,
			Name:    "修改参数名称",
			Match:   "token",
			Replace: "auth_token",
			Comment: "示例：将请求参数中的token改名为auth_token",
		},
		{
			ID:      7,
			Enabled: false,
			Type:    RuleTypeRequestParamValue,
			Name:    "修改参数值",
			Match:   "old_value",
			Replace: "new_value",
			Comment: "示例：替换请求参数的特定值",
		},
	}
}

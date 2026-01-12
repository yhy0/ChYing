package conf

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

/**
   @author yhy
   @since 2024/12/10
   @desc 配置文件键名规范化 - 将小写键名转换为驼峰式键名
**/

// keyMappings 定义小写键名到驼峰式键名的映射
var keyMappings = map[string]string{
	// plugins 下的键
	"bruteforce":           "bruteForce",
	"cmdinjection":         "cmdInjection",
	"crlfinjection":        "crlfInjection",
	"sqlmapapi":            "sqlmapApi",
	"nginxaliastraversal":  "nginxAliasTraversal",
	"portscan":             "portScan",
	// plugins 内部的键
	"usernamedict":         "usernameDict",
	"passworddict":         "passwordDict",
	"detectxssincookie":    "detectXssInCookie",
	"booleanbaseddetection": "booleanBasedDetection",
	"timebaseddetection":   "timeBasedDetection",
	"errorbaseddetection":  "errorBasedDetection",
	"detectincookie":       "detectInCookie",
	// http 下的键
	"maxconnsperhost":      "maxConnsPerHost",
	"retrytimes":           "retryTimes",
	"allowredirect":        "allowRedirect",
	"verifyssl":            "verifySSL",
	"maxqps":               "maxQps",
	"forcehttp1":           "forceHTTP1",
	// mitmproxy 下的键
	"cacert":               "caCert",
	"cakey":                "caKey",
	"basicauth":            "basicAuth",
	"filtersuffix":         "filterSuffix",
	"maxlength":            "maxLength",
	// collection 下的键
	"idcard":               "idCard",
	"urlfilter":            "urlFilter",
	"sensitiveparameters":  "sensitiveParameters",
	// basiccrawler 下的键
	"basiccrawler":         "basicCrawler",
	"maxdepth":             "maxDepth",
	"maxcountoflinks":      "maxCountOfLinks",
	"allowvisitparentpath": "allowVisitParentPath",
}

// NormalizeConfigYAML 规范化配置文件的 YAML 内容
// 将小写键名转换为驼峰式键名
func NormalizeConfigYAML(content string) (string, error) {
	// 解析 YAML
	var data any
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return content, err
	}

	// 递归转换键名
	normalized := normalizeKeys(data)

	// 重新序列化为 YAML
	result, err := yaml.Marshal(normalized)
	if err != nil {
		return content, err
	}

	return string(result), nil
}

// normalizeKeys 递归转换 map 中的键名
func normalizeKeys(data any) any {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, value := range v {
			// 转换键名
			newKey := normalizeKey(key)
			// 递归处理值
			result[newKey] = normalizeKeys(value)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = normalizeKeys(item)
		}
		return result
	default:
		return data
	}
}

// normalizeKey 将小写键名转换为驼峰式键名
func normalizeKey(key string) string {
	// 先检查映射表
	lowerKey := strings.ToLower(key)
	if mapped, ok := keyMappings[lowerKey]; ok {
		return mapped
	}

	// 如果键名已经是驼峰式，直接返回
	if key != lowerKey {
		return key
	}

	// 处理下划线分隔的键名（如 hostname_allowed -> hostnameAllowed）
	if strings.Contains(key, "_") {
		return snakeToCamel(key)
	}

	return key
}

// snakeToCamel 将 snake_case 转换为 camelCase
func snakeToCamel(s string) string {
	// 使用正则表达式匹配下划线后的字符
	re := regexp.MustCompile(`_([a-z])`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		return strings.ToUpper(string(match[1]))
	})
}

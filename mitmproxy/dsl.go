package mitmproxy

import (
	"bufio"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/projectdiscovery/dsl"
)

/**
   @author yhy
   @since 2025/6/4
   @desc DSL查询功能实现
   该文件实现了基于DSL(Domain Specific Language)的HTTP历史记录查询
   利用proxify的DSL引擎匹配HTTP请求/响应，为前端提供强大的查询能力
**/

// 注意：dsl.go需要使用proxify.go中定义的以下变量：
// - HTTPHistoryMap (从proxify.go的TempHistoryCache变更而来)
// - HTTPBodyMap

// QueryHistoryByDSL 根据提供的DSL查询字符串过滤HTTP历史记录
// dslQuery: DSL查询表达式
// 返回匹配的HTTP历史记录摘要列表
func QueryHistoryByDSL(dslQuery string) ([]HTTPHistory, error) {
	if dslQuery == "" {
		// 如果DSL查询为空，返回空列表
		// 这样前端可以区分"清除过滤器"和"执行查询"的情况
		return []HTTPHistory{}, nil
	}

	// 验证DSL表达式
	if err := validateDSL(dslQuery); err != nil {
		return nil, fmt.Errorf("DSL语法错误: %w", err)
	}

	// 收集所有需要匹配的历史记录ID
	var historyIds []int64
	HTTPBodyMap.Range(func(key, value interface{}) bool {
		if id, ok := key.(int64); ok {
			historyIds = append(historyIds, id)
		}
		return true
	})

	var matchedHistory []HTTPHistory

	// 遍历所有历史记录并应用DSL匹配
	for _, historyId := range historyIds {
		// 获取历史记录摘要
		historySummary, err := GetHistoryById(historyId)
		if err != nil {
			log.Printf("警告: 未找到ID为%d的历史记录摘要: %v", historyId, err)
			continue
		}

		// 从HTTPBodyMap获取完整的请求/响应数据
		httpBody, err := GetHTTPBody(historyId)
		if err != nil {
			log.Printf("警告: 未找到ID为%d的HTTP报文原始数据: %v", historyId, err)
			continue
		}

		// 解析请求和响应以构建DSL上下文
		requestData := parseRawHTTP(httpBody.RequestRaw, true)
		responseData := parseRawHTTP(httpBody.ResponseRaw, false)

		// 合并请求和响应数据，构建完整的DSL评估上下文
		dslContext := mergeDSLContext(historySummary, requestData, responseData)

		// 评估DSL表达式
		matched, err := evaluateDSL(dslQuery, dslContext)
		if err != nil {
			log.Printf("评估ID为%d的请求时DSL错误: %v", historySummary.Id, err)
			continue
		}

		// 如果结果为true，则此历史记录匹配查询条件
		if matched {
			matchedHistory = append(matchedHistory, *historySummary)
		}
	}

	// 按照ID排序结果
	sort.Slice(matchedHistory, func(i, j int) bool {
		return matchedHistory[i].Id < matchedHistory[j].Id
	})

	return matchedHistory, nil
}

// 辅助函数，用于验证DSL表达式
func validateDSL(dslQuery string) error {
	// 创建一个包含所有可能字段的模拟上下文，用于验证DSL表达式
	mockContext := map[string]interface{}{
		// 摘要字段
		"id":                 int64(1),
		"flow_id":            "flow123",
		"url":                "https://example.com/api/test",
		"path":               "/api/test",
		"method":             "GET",
		"host":               "example.com",
		"status":             "200",
		"length":             "1024",
		"content_type":       "application/json",
		"timestamp":          "2023-01-01T12:00:00Z",
		"response_timestamp": "2023-01-01T12:00:01Z",

		// 请求相关字段
		"request":         "GET /api/test HTTP/1.1\r\nHost: example.com\r\n\r\n",
		"request_body":    "test body",
		"request_headers": map[string]string{"host": "example.com"},

		// 响应相关字段
		"response":         "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"status\":\"ok\"}",
		"response_body":    "{\"status\":\"ok\"}",
		"response_headers": map[string]string{"content-type": "application/json"},
		"status_reason":    "OK",
	}

	// 使用完整的模拟上下文测试DSL表达式
	_, err := dsl.EvalExpr(dslQuery, mockContext)
	return err
}

// 辅助函数，用于评估DSL表达式
func evaluateDSL(dslQuery string, context map[string]interface{}) (bool, error) {
	// 使用 projectdiscovery/dsl 包评估表达式
	result, err := dsl.EvalExpr(dslQuery, context)
	if err != nil {
		return false, err
	}

	if resultBool, ok := result.(bool); ok {
		return resultBool, nil
	}
	return false, fmt.Errorf("DSL结果不是布尔值: %v", result)
}

// GetHistoryById 从存储中获取指定ID的历史记录
// 实际实现需要根据项目的存储结构调整
func GetHistoryById(id int64) (*HTTPHistory, error) {
	// 从HTTPBodyMap获取HTTP请求/响应的原始数据
	body, err := GetHTTPBody(id)
	if err != nil {
		return nil, err
	}

	// 解析请求行和头部
	reqData := parseRawHTTP(body.RequestRaw, true)
	respData := parseRawHTTP(body.ResponseRaw, false)

	// 从解析结果中提取关键信息，重建HTTPHistory对象
	history := &HTTPHistory{
		Id:     id,
		FlowID: body.FlowID,
	}

	// 提取方法
	if method, ok := reqData["method"].(string); ok {
		history.Method = method
	}

	// 提取主机和路径以构建完整URL
	var host, path string
	if h, ok := reqData["host"].(string); ok {
		host = h
		history.Host = h
	}
	if p, ok := reqData["path"].(string); ok {
		path = p
		history.Path = p
	}

	// 构建完整URL
	if host != "" && path != "" {
		// 检查path是否已经是完整URL
		if strings.HasPrefix(path, "http") {
			history.FullUrl = path
		} else {
			// 简单拼接，实际情况可能需要更复杂的处理
			scheme := "http"
			if strings.Contains(host, ":443") {
				scheme = "https"
			}
			history.FullUrl = fmt.Sprintf("%s://%s%s", scheme, host, path)
		}
	}

	// 提取状态码
	if status, ok := respData["status"].(string); ok {
		history.Status = status
	}

	// 提取内容类型
	if contentType, ok := respData["content_type"].(string); ok {
		history.ContentType = contentType
	}

	// 提取响应体长度
	if length, ok := respData["content_length"].(string); ok {
		history.Length = length
	} else if respBody, ok := respData["response_body"].(string); ok {
		// 如果header中没有Content-Length，使用实际响应体长度
		history.Length = fmt.Sprintf("%d", len(respBody))
	}

	// 从请求或响应中提取时间戳信息
	// 这里假设我们没有保存原始时间戳，使用当前时间作为后备
	currentTime := time.Now().Format(time.RFC3339)
	history.Timestamp = currentTime
	history.ResponseTimestamp = currentTime

	return history, nil
}

// parseRawHTTP 解析原始HTTP请求/响应字符串
// 返回可用于DSL匹配的键值对
func parseRawHTTP(rawHTTP string, isRequest bool) map[string]interface{} {
	if rawHTTP == "" {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	// 添加原始内容以支持直接对整个内容进行匹配
	if isRequest {
		result["request"] = rawHTTP
	} else {
		result["response"] = rawHTTP
	}

	// 分离头部和主体
	parts := strings.SplitN(rawHTTP, "\r\n\r\n", 2)
	headers := parts[0]
	var body string
	if len(parts) > 1 {
		body = parts[1]
	}

	// 添加主体内容
	if isRequest {
		result["request_body"] = body
	} else {
		result["response_body"] = body
	}

	// 解析请求/响应行和头部
	scanner := bufio.NewScanner(strings.NewReader(headers))
	lineNum := 0
	headerMap := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if lineNum == 0 {
			// 处理请求行或状态行
			if isRequest {
				parts := strings.SplitN(line, " ", 3)
				if len(parts) >= 2 {
					result["method"] = parts[0]
					result["path"] = parts[1]
					if len(parts) > 2 {
						result["http_version"] = parts[2]
					}
				}
			} else {
				parts := strings.SplitN(line, " ", 3)
				if len(parts) >= 3 {
					result["status"] = parts[1]
					result["status_reason"] = parts[2]
				}
			}
		} else {
			// 处理头部
			colonIdx := strings.IndexByte(line, ':')
			if colonIdx > 0 {
				key := strings.TrimSpace(line[:colonIdx])
				value := strings.TrimSpace(line[colonIdx+1:])
				headerMap[strings.ToLower(key)] = value

				// 提取一些重要头部作为独立字段
				lowerKey := strings.ToLower(key)
				if lowerKey == "host" {
					result["host"] = value
				} else if lowerKey == "content-type" {
					result["content_type"] = value
				} else if lowerKey == "content-length" {
					result["content_length"] = value
				}
			}
		}
		lineNum++
	}

	// 添加所有头部
	if isRequest {
		result["request_headers"] = headerMap
	} else {
		result["response_headers"] = headerMap
	}

	return result
}

// mergeDSLContext 合并请求摘要、请求详情和响应详情，构建完整的DSL上下文
func mergeDSLContext(summary *HTTPHistory, requestData, responseData map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// 添加从HTTP摘要中提取的字段
	merged["id"] = summary.Id
	merged["flow_id"] = summary.FlowID
	merged["url"] = summary.FullUrl
	merged["path"] = summary.Path
	merged["method"] = summary.Method
	merged["host"] = summary.Host
	merged["status"] = summary.Status
	merged["length"] = summary.Length
	merged["content_type"] = summary.ContentType
	merged["timestamp"] = summary.Timestamp
	merged["response_timestamp"] = summary.ResponseTimestamp

	// 合并请求和响应数据
	for k, v := range requestData {
		merged[k] = v
	}
	for k, v := range responseData {
		merged[k] = v
	}

	return merged
}

// 以下是用于测试的便捷函数，可以在开发环境中使用，生产环境可以移除

// ParseHTTPRequestForDSL 解析HTTP请求供DSL使用
func ParseHTTPRequestForDSL(reqRaw string) (map[string]interface{}, error) {
	parsedMap := parseRawHTTP(reqRaw, true)
	return parsedMap, nil
}

// ParseHTTPResponseForDSL 解析HTTP响应供DSL使用
func ParseHTTPResponseForDSL(respRaw string) (map[string]interface{}, error) {
	parsedMap := parseRawHTTP(respRaw, false)
	return parsedMap, nil
}

// TestDSL 测试DSL表达式是否有效
func TestDSL(dslQuery string, context map[string]interface{}) (bool, error) {
	return evaluateDSL(dslQuery, context)
}

// GetHTTPBody 从HTTPBodyMap获取指定ID的HTTP报文内容
func GetHTTPBody(id int64) (*HTTPBody, error) {
	_data, _ok := HTTPBodyMap.Load(id)
	if !_ok {
		return nil, fmt.Errorf("未找到ID为%d的HTTP报文数据", id)
	}

	// 由于存储的是指针类型，直接类型转换
	httpBody, ok := _data.(*HTTPBody)
	if !ok {
		return nil, fmt.Errorf("无效的HTTP报文数据类型: %T", _data)
	}

	return httpBody, nil
}

package test

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/decoder"
	"testing"
)

/**
   @author yhy
   @since 2025/4/26
   @desc //TODO
**/

// processing 对 payload 进行的处理
func processing(payload string, processing map[string]interface{}) string {
	if processing == nil {
		return payload
	}

	for k, v := range processing {
		print(k)
		switch v {
		case "MD5": // MD5 加密处理
			payload = decoder.Md5(payload)
		case "Base64-encode": // MD5 加密处理
			payload = decoder.EncodeBase64(payload)
		case "Base64-decode": // MD5 加密处理
			payload = decoder.DecodeBase64(payload)
		case "Unicode-encode": // MD5 加密处理
			payload = decoder.EncodeUnicode(payload)
		case "Unicode-decode": // MD5 加密处理
			payload = decoder.DecodeUnicode(payload)
		case "URL-encode": // MD5 加密处理
			payload = decoder.EncodeURL(payload)
		case "URL-decode": // MD5 加密处理
			payload = decoder.DecodeURL(payload)
		case "Hex-encode": // MD5 加密处理
			payload = decoder.EncodeHex(payload)
		case "Hex-decode": // MD5 加密处理
			payload = decoder.DecodeHex(payload)
		case "None":
			payload = payload
		}
	}

	return payload
}

// 生成所有可能的组合（包含处理逻辑）
func GenerateCombinations(items []PayloadItem) [][]string {
	// 先处理每个Item中的字符串
	var processedItems [][]string
	for _, item := range items {
		var processed []string
		for _, s := range item.Items {
			// 对每个字符串应用处理函数
			processed = append(processed, processing(s, item.Processing))
		}
		processedItems = append(processedItems, processed)
	}

	// 如果没有数据，返回空组合
	if len(processedItems) == 0 {
		return [][]string{}
	}

	// 使用回溯算法生成所有组合
	var result [][]string
	backtrack(processedItems, 0, []string{}, &result)
	return result
}

// 回溯算法实现
func backtrack(allItems [][]string, index int, current []string, result *[][]string) {
	// 如果已经处理完所有Items切片，将当前组合加入结果
	if index == len(allItems) {
		combination := make([]string, len(current))
		copy(combination, current)
		*result = append(*result, combination)
		return
	}

	// 遍历当前Items切片的所有元素
	for _, item := range allItems[index] {
		// 添加当前元素到组合中
		current = append(current, item)
		// 递归处理下一个Items切片
		backtrack(allItems, index+1, current, result)
		// 回溯，移除最后添加的元素
		current = current[:len(current)-1]
	}
}

type PayloadItem struct {
	ID         int64                  `json:"id"`         // 条目ID
	Type       string                 `json:"type"`       // 条目类型
	Items      []string               `json:"items"`      // 条目内容数组
	Processing map[string]interface{} `json:"processing"` // 选项，使用空接口接收任意类型
}

func TestBp(t *testing.T) {
	// 示例数据
	payloadItems := []PayloadItem{
		{
			ID:    1,
			Type:  "simple-list",
			Items: []string{"1", "2"},
			Processing: map[string]interface{}{
				"operation": "MD5",
			},
		},
		{
			ID:    2,
			Type:  "simple-list",
			Items: []string{"A", "B"},
			Processing: map[string]interface{}{
				"operation": "Base64-encode",
			},
		},
		{
			ID:         3,
			Type:       "simple-list",
			Items:      []string{"1", "2"},
			Processing: map[string]interface{}{},
		},
	}

	// 生成所有组合
	combinations := GenerateCombinations(payloadItems)

	// 打印结果
	fmt.Println("所有可能的组合(已处理):")
	for i, combo := range combinations {
		fmt.Printf("组合 %d: %v\n", i+1, combo)
	}
}

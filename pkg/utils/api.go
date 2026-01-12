package utils

import (
    "sort"
    "strings"
)

/**
  @author: yhy
  @since: 2024/10/22
  @desc: 将所有的 api 进行分析，找出所有的可能性，比如
       输入: minOccurrence = 1
          /v1/api/user、/v2/data
       输出:
          /v1 /v1/api /v1/api/data /v1/api/user /v1/api/user/data /v1/data /v1/user /v2 /v2/data /v2/data/user /v2/user
**/

func PredictionApi(apis []string, minOccurrence int) []string {
    segments := pathSegments(apis)
    var predictions []string
    
    // 获取所有的前缀和后缀
    prefixes := getPrefixes(segments, minOccurrence)
    suffixes := getSuffixes(apis)
    
    // 生成预测
    for _, prefix := range prefixes {
        predictions = append(predictions, prefix)
        prefixParts := strings.Split(strings.Trim(prefix, "/"), "/")
        
        for _, suffix := range suffixes {
            if !strings.HasSuffix(prefix, suffix) {
                prediction := prefix + "/" + strings.TrimPrefix(suffix, "/")
                predictions = append(predictions, prediction)
                
                // 添加去掉中间段的简化路径
                if len(prefixParts) > 1 {
                    simplifiedPrefix := "/" + prefixParts[0]
                    simplifiedPrediction := simplifiedPrefix + "/" + strings.TrimPrefix(suffix, "/")
                    predictions = append(predictions, simplifiedPrediction)
                }
            }
        }
    }
    
    // 添加原始 API 路径
    predictions = append(predictions, apis...)
    
    return uniquePredictions(predictions)
}

// pathSegments 计算每个路径出现的次数
func pathSegments(apis []string) map[string]int {
    segments := make(map[string]int)
    
    for _, api := range apis {
        parts := strings.Split(strings.Trim(api, "/"), "/")
        currentPath := ""
        for _, part := range parts {
            currentPath += "/" + part
            segments[currentPath]++
        }
    }
    
    return segments
}

// getPrefixes 获取所有的前缀, 通过指定 minOccurrence 来获取对应前缀
func getPrefixes(segments map[string]int, minOccurrence int) []string {
    var prefixes []string
    for segment, count := range segments {
        if count >= minOccurrence {
            prefixes = append(prefixes, segment)
        }
    }
    sort.Slice(prefixes, func(i, j int) bool {
        return segments[prefixes[i]] > segments[prefixes[j]]
    })
    return prefixes
}

// getSuffixes 获取所有的后缀（最后一个路径段）
func getSuffixes(apis []string) []string {
    suffixes := make(map[string]struct{})
    for _, api := range apis {
        parts := strings.Split(strings.Trim(api, "/"), "/")
        if len(parts) > 0 {
            suffixes[parts[len(parts)-1]] = struct{}{}
        }
    }
    
    result := make([]string, 0, len(suffixes))
    for suffix := range suffixes {
        result = append(result, suffix)
    }
    return result
}

// uniquePredictions 去重
func uniquePredictions(predictions []string) []string {
    unique := make(map[string]struct{})
    var result []string
    for _, prediction := range predictions {
        if _, exists := unique[prediction]; !exists {
            unique[prediction] = struct{}{}
            result = append(result, prediction)
        }
    }
    sort.Strings(result)
    return result
}

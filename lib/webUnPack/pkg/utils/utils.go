package utils

import "strings"

/**
  @author: yhy
  @since: 2023/8/2
  @desc: //TODO
**/

// RemoveDuplicateElement  数组去重
func RemoveDuplicateElement(strs []string) []string {
    seen := make(map[string]bool)
    result := make([]string, 0, len(strs))
    for _, item := range strs {
        item = strings.TrimSpace(item)
        if len(item) == 0 || seen[strings.ToLower(item)] {
            continue
        }
        seen[strings.ToLower(item)] = true
        result = append(result, item)
    }
    return result
}

func Standard(target, uri string) string {
    var u string
    if strings.HasSuffix(target, "/") {
        if strings.HasPrefix(uri, "/") {
            u = target + uri[1:]
        } else {
            u = target + uri
        }
    } else {
        if strings.HasPrefix(uri, "/") {
            u = target + uri
        } else {
            u = target + "/" + uri
        }
    }
    
    return u
}

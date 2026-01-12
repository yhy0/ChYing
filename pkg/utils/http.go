package utils

import (
    "github.com/yhy0/logging"
    "net"
    "net/url"
    "strings"
)

/**
   @author yhy
   @since 2024/9/9
   @desc //TODO
**/

// GetParentDomain 获取上级域名
func GetParentDomain(domain string) string {
    if strings.Contains(domain, ":") {
        domain = strings.Split(domain, ":")[0]
    }
    // 判断是否为 IP 地址
    ip := net.ParseIP(domain)
    if ip == nil {
        parts := strings.Split(domain, ".")
        if len(parts) < 2 {
            return domain
        }
        // 返回上级域名
        return strings.Join(parts[len(parts)-2:], ".")
    } else {
        return domain
    }
}

func GetParams(_url string) map[string]string {
    parse, err := url.Parse(_url)
    if err != nil {
        logging.Logger.Fatal(err)
        return nil
    }
    
    params := make(map[string]string)
    
    for k, _ := range parse.Query() {
        params[k] = parse.Query().Get(k)
    }
    
    return params
}

// AnalyzeDynamicPath 函数将 URL 中的整数部分替换为 {id} eg: /api/v1/user/1 --> /api/v1/user/{id}
func AnalyzeDynamicPath(url string) string {
    // 使用 / 分割 URL
    parts := strings.Split(url, "/")
    
    // 遍历每个部分，判断是否为整数
    for i, part := range parts {
        if IsInteger(part) {
            parts[i] = "{id}" // 替换为 {id}
        }
    }
    
    // 重新组合 URL
    return strings.Join(parts, "/")
}

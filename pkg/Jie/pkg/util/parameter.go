package util

import (
    "github.com/yhy0/logging"
    "net/url"
    "strings"
)

/**
  @author: yhy
  @since: 2023/11/22
  @desc: 提取参数
**/

// ParamFilter 过滤一些参数不进行处理 https://github.com/yhy0/Jie/issues/4 TODO 待增加
var ParamFilter = []string{"submit", "reset", "button", "image", "hidden", "csrf_token", "user_token", "userToken"}

// ExtractParameters 提取参数，这一部分就不在 pkg/task/task.go 里进行了，哪里需要哪里调用
func ExtractParameters(u, method string, requestBody string, header map[string]string) []string {
    // 测试 get 请求
    var params []string
    // TODO 这一部分，目前来看没必要这么做，除了 xss、sqlmap 会使用外，其他的都自动处理了
    if strings.EqualFold(method, "GET") {
        parseUrl, err := url.Parse(u)
        if err != nil {
            logging.Logger.Errorln(u, err)
            return nil
        }
        queryParams := parseUrl.Query()
        for paramName := range queryParams {
            if SliceInCaseFold(paramName, ParamFilter) {
                continue
            }
            params = append(params, paramName)
        }
    } else if strings.EqualFold(method, "POST") && requestBody != "" { // 测试 post 请求
        // 解析 POST 请求的请求体
        postParams, err := url.ParseQuery(requestBody)
        
        if err != nil {
            logging.Logger.Errorln(u, err)
            return nil
        }
        
        if strings.Contains(header["Content-Type"], "application/x-www-form-urlencoded") {
            // 解析 POST 请求的请求体
            postParams, err = url.ParseQuery(requestBody)
            if err != nil {
                logging.Logger.Errorln(u, err)
                return nil
            }
        } else if strings.Contains(header["Content-Type"], "json") { // TODO
        
        }
        
        for paramName := range postParams {
            if SliceInCaseFold(paramName, ParamFilter) {
                continue
            }
            params = append(params, paramName)
        }
    }
    
    return params
}

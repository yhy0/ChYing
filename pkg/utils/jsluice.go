package utils

import (
    "encoding/json"
    "github.com/thoas/go-funk"
    "github.com/yhy0/ChYing/lib/jsluice"
    "github.com/yhy0/logging"
    "strings"
)

/**
  @author: yhy
  @since: 2024/10/15
  @desc: //TODO
**/

func Jsluice(body string, apis []string) ([]string, []string) {
    analyzer := jsluice.NewAnalyzer([]byte(body))
    var urls []string
    _urls := make(map[string]bool)
    var secrets []string
    
    if len(apis) > 0 {
        // 添加自定义的URL匹配器, 这里主要是为了弥补 jsluice 本身获取不到的 URL 的功能, 根据已有的API 进行再次匹配
        for _, api := range apis {
            analyzer.AddURLMatcher(
                // The first value in the jsluice.URLMatcher struct is the type of node to look for.
                // It can be one of "string", "assignment_expression", or "call_expression"
                jsluice.URLMatcher{Type: "string", Fn: func(n *jsluice.Node) *jsluice.URL {
                    val := n.DecodedString()
                    if strings.HasPrefix(val, api) {
                        return &jsluice.URL{
                            URL:  val,
                            Type: api,
                        }
                    }
                    if strings.HasSuffix(val, api) {
                        return &jsluice.URL{
                            URL:  val,
                            Type: api,
                        }
                    }
                    if strings.Contains(val, api) && strings.HasPrefix(val, "/") {
                        return &jsluice.URL{
                            URL:  val,
                            Type: api,
                        }
                    }
                    return nil
                }},
            )
        }
    }
    
    for _, res := range analyzer.GetURLs() {
        if _urls[res.URL] {
            continue
        }
        _urls[res.URL] = true
        j, err := json.MarshalIndent(res, "", "  ")
        if err != nil {
            logging.Logger.Errorln(err)
            continue
        }
        urls = append(urls, string(j))
    }
    
    for _, res := range analyzer.GetSecrets() {
        j, err := json.MarshalIndent(res, "", "  ")
        if err != nil {
            logging.Logger.Errorln(err)
            continue
        }
        secrets = append(secrets, string(j))
    }
    
    return funk.UniqString(urls), secrets
}

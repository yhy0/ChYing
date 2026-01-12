package conf

/**
  @author: yhy
  @since: 2023/1/4
  @desc: //TODO
**/

var (
    // Plugin 插件单独从配置文件中读取出来，方便使用
    Plugin = map[string]bool{
        "xss":                   false,
        "sql":                   false,
        "sqlmapApi":             false,
        "cmd":                   false,
        "xxe":                   false,
        "ssrf":                  false,
        "brute":                 false, // web 类登录爆破
        "bypass403":             false,
        "jsonp":                 false,
        "crlf":                  false,
        "log4j":                 false,
        "fastjson":              false,
        "portScan":              false,
        "poc":                   false,
        "nuclei":                false,
        "bbscan":                false,
        "archive":               false,
        "nginx-alias-traversal": false,
    }
)

// DangerHeaders 一些危险的请求头, 用来测试 sql 注入、ssrf，有的谜一样的业务逻辑可能会被命中
var DangerHeaders = []string{
    "X-Client-IP",
    "X-Remote-IP",
    "X-Remote-Addr",
    "X-Forwarded-For",
    "X-Originating-IP",
    "Referer",
    "CF-Connecting_IP",
    "True-Client-IP",
    "X-Forwarded-For",
    "Originating-IP",
    "X-Real-IP",
    "X-Client-IP",
    "Forwarded",
    "Client-IP",
    "Contact",
    "X-Wap-Profile",
    "X-Api-Version",
}

// Parallelism 对一个网站同时扫描的最大 url 个数
var Parallelism = 10

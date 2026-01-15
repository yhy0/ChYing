package conf

import (
	"fmt"
	"time"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 默认配置和常量定义
**/

var ProxyPort = 9080
var ProxyHost = "127.0.0.1"

var Proxy string
var Token string

var Parallelism = 99

var Description = fmt.Sprintf("将旦昧爽之交，日夕昏明之际，\n北面而察之，淡淡焉若有物存，莫识其状。\n其所触也，窃窃然有声，经物而物不疾也。\n\n© %d https://github.com/yhy0", time.Now().Year())

const (
	Version        = "v1.1"
	Title          = "承影 " + Version
	VersionNewMsg  = "当前已经是最新版本!"
	VersionOldMsg  = "最新版本: %s, 是否立即更新?"
	BtnConfirmText = "确定"
	BtnCancelText  = "取消"
)

// 默认配置 YAML 字符串 - 类似 pkg/Jie/conf/file.go 的写法
var defaultConfigYaml = []byte(`# ChYing 扫描系统配置文件
# 配置文件位置: ~/.config/ChYing/config.yaml

version: 2.0.2

# 代理配置
proxy:
  host: "127.0.0.1"
  port: 9080
  enabled: true
  listeners:
    - id: "default"
      host: "127.0.0.1"
      port: 9080
      enabled: true

# 扫描配置
scan:
  enable_port_scan: true
  enable_dir_scan: true
  enable_vuln_scan: true
  threads: 10
  timeout: 30
  parallel: 10                            # 同时扫描的最大 url 个数

# 日志配置
logging:
  level: "info"
  file: "./logs/chying.log"

# AI 配置
ai:
  claude:
    cli_path: "claude"                    # Claude Code CLI 路径，留空使用 PATH 中的 claude
    work_dir: ""                          # 工作目录，留空使用当前目录
    model: "claude-sonnet-4"     # 使用的模型
    max_turns: 0                          # 最大回合数，0 表示无限制
    system_prompt: ""                     # 系统提示词
    allowed_tools: []                     # 允许的工具列表（空表示全部允许）
    disallowed_tools: []                  # 禁用的工具列表
    permission_mode: "default"            # 权限模式: default, plan, bypassPermissions
    require_tool_confirm: true            # 危险操作是否需要确认
    # 环境变量配置
    api_key: ""                           # ANTHROPIC_API_KEY
    base_url: ""                          # ANTHROPIC_BASE_URL (如: https://api.anthropic.com)
    temperature: 0.7                      # AI_TEMPERATURE (0.0-1.0)
    # MCP 服务器配置
    mcp:
      enabled: true                       # 是否启用内置 MCP 服务器
      mode: "sse"                         # 运行模式: sse 或 stdio
      port: 0                             # SSE 模式端口，0 表示自动选择
      enabled_tools: []                   # 启用的工具列表（空表示全部启用）
      disabled_tools: []                  # 禁用的工具列表
      external_servers: []                # 外部 MCP 服务器列表
      # 外部 MCP 服务器配置示例:
      # - id: "example-mcp"
      #   name: "Example MCP Server"
      #   type: "sse"                     # sse 或 stdio
      #   enabled: true
      #   description: "示例 MCP 服务器"
      #   url: "http://localhost:3000/sse"  # SSE 模式 URL
      #   headers: {}                     # 自定义请求头
      #   command: ""                     # STDIO 模式命令
      #   args: []                        # STDIO 模式参数
      #   env: []                         # 环境变量

# 全局 http 发包配置
http:
  proxy: ""                             # 漏洞扫描时使用的代理，如: http://127.0.0.1:8080
  timeout: 10                           # 建立 tcp 连接的超时时间
  maxConnsPerHost: 100                  # 每个 host 最大连接数
  retryTimes: 0                         # 请求失败的重试次数，0 则不重试
  allowRedirect: 0                      # 单个请求最大允许的跳转数，0 则不跳转
  verifySSL: false                      # 是否验证 ssl 证书
  maxQps: 50                            # 每秒最大请求数
  headers: {}                           # 指定 http 请求头
  forceHTTP1: false                     # 强制指定使用 http/1.1

# 漏洞探测的插件配置
plugins:
  bruteForce:
    web: false                          # web 服务类的爆破，比如 tomcat 爆破
    service: false                      # 服务类的爆破，比如 mysql 爆破
    usernameDict: ""                    # 自定义用户名字典
    passwordDict: ""                    # 自定义密码字典
  cmdInjection:
    enabled: false
  crlfInjection:
    enabled: false
  xss:
    enabled: false
    detectXssInCookie: false             # 是否探测入口点在 cookie 中的 xss
  sql:
    enabled: false
    booleanBasedDetection: false         # 是否检测布尔盲注
    errorBasedDetection: false           # 是否检测报错注入
    timeBasedDetection: false            # 是否检测时间盲注
    detectInCookie: false                # 是否检查在 cookie 中的注入
  sqlmapApi:
    enabled: false
    url: ""                             # sqlmap api 的地址
    username: ""                        # 认证用户名
    password: ""                        # 认证密码
  xxe:
    enabled: false
  ssrf:
    enabled: false
  bbscan:                               # bbscan 这种规则类目录扫描
    enabled: false
  jsonp:
    enabled: false
  log4j:
    enabled: false
  bypass403:
    enabled: false
  fastjson:
    enabled: false
  archive:                              # 从 web.archive.org 获取历史 url
    enabled: false
  iis:                                  # iis 短文件名 fuzz
    enabled: false
  nginxAliasTraversal:                  # nginx 别名遍历
    enabled: false
  poc:
    enabled: false
  nuclei:
    enabled: false
  portScan:
    enabled: false

# 反连平台配置
reverse:
  host: "https://dig.pm/"               # 反连平台地址
  domain: "ipv6.bypass.eu.org."         # 指定反连域名

# 基础爬虫配置
basicCrawler:
  maxDepth: 0                           # 最大爬取深度， 0 为无限制
  maxCountOfLinks: 0                    # 本次爬取收集的最大链接数, 0 为无限制
  allowVisitParentPath: false           # 是否允许爬取父目录
  restriction:                          # 爬虫的允许爬取的资源限制
    hostname_allowed: []                # 允许访问的 Hostname
    hostname_disallowed:                # 不允许访问的 Hostname
    - '*.edu.*'
    - '*.gov.*'
    port_allowed: []                    # 允许访问的端口
    port_disallowed: []                 # 不允许访问的端口
    path_allowed: []                    # 允许访问的路径
    path_disallowed: []                 # 不允许访问的路径
    query_key_allowed: []               # 允许访问的 Query Key
    query_key_disallowed: []            # 不允许访问的 Query Key
    fragment_allowed: []                # 允许访问的 Fragment
    fragment_disallowed: []             # 不允许访问的 Fragment
    post_key_allowed: []                # 允许访问的 Post Body 中的参数
    post_key_disallowed: []             # 不允许访问的 Post Body 中的参数
  basic_auth:                           # 基础认证信息
    username: ""
    password: ""

# 被动代理配置
mitmproxy:
  caCert: ./ca.crt                      # CA 根证书路径
  caKey: ./ca.key                       # CA 私钥路径
  basicAuth:                            # 基础认证的用户名密码
    header: "Go-Mitmproxy-Authorization"
    username: ""
    password: ""
  exclude:                              # 不允许访问的 Hostname
    - .google.
    - .googleapis.
    - .gstatic.
    - .googleusercontent.
    - .googlevideo.
    - .firefox.
    - .firefoxchina.cn
    - .firefoxusercontent.com
    - .mozilla.
    - .doubleclick.
    - spocs.getpocket.com
    - .portswigger.net
    - .gov.(com|cn)
    - cdn.jsdelivr.net
    - .cdn-go.cn
    - .lencr.org
    - .adavoid.org
  include: []                           # 允许访问的 Hostname
  filterSuffix: ".3g2, .3gp, .arj, .avi, .axd, .bmp, .drv, .eot, .flv, .gif, .gifv, .h264, .ico, .jpeg, .jpg, .m4a, .m4v, .mkv, .mov, .mp3, .mp4, .mpeg, .mpg, .ogg, .ogm, .ogv, .otf, .png, .psd, .rm, .svg, .swf, .sys, .tif, .tiff, .ttf, .vob, .wav, .webm, .webp, .wmv, .woff, .woff2, .xcf"
  maxLength: 3000                       # 队列长度限制

# 信息收集类的正则
collection:
  domain:
    - "['\"](([a-zA-Z0-9]{1,9}:)?//)?(.{1,36}:.{1,36}@)?[a-zA-Z0-9\\-\\.]*?\\.(xin|com|cn|net|com\\.cn|vip|top|cc|shop|club|wang|xyz|luxe|site|news|pub|fun|online|win|red|loan|ren|mom|net\\.cn|org|link|biz|bid|help|tech|date|mobi|so|me|tv|co|vc|pw|video|party|pics|website|store|ltd|ink|trade|live|wiki|space|gift|lol|work|band|info|click|photo|market|tel|social|press|game|kim|org\\.cn|games|pro|men|love|studio|rocks|asia|group|science|design|software|engineer|lawyer|fit|beer|我爱你|中国|公司|网络|在线|网址|网店|集团|中文网)(:\\d{1,5})?"
  ip:
    - "['\"](([a-zA-Z0-9]{1,9}:)?//)?(.{1,36}:.{1,36}@)?\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}(:\\d{1,5})?"
  phone:
    - "['\"](1(3([0-35-9]\\d|4[1-8])|4[14-9]\\d|5([\\d]\\d|7[1-79])|66\\d|7[2-35-8]\\d|8\\d{2}|9[89]\\d)\\d{7})['\"]"
  email:
    - '[''"]([\\w!#$%&''*+=?^_` + "`" + `{|}~-]+(?:\\.[\\w!#$%&''*+=?^_` + "`" + `{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[\\w](?:[\\w-]*[\\w])?)[''"]'
  api:
    - "(?i)\\.(get|post|put|delete|options|connect|trace|patch)\\([\"'](/?.*?)[\"']"
    - "(?:\"|')(/[^/\"']+){2,}(?:\"|')"
  url:
    - '["''` + "`" + `]\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250}?)\\s{0,6}["''` + "`" + `]'
    - '=\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})'
    - '["''` + "`" + `]\\s{0,6}([#,.]{0,2}/[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250}?)\\s{0,6}["''` + "`" + `]'
    - '"([-a-zA-Z0-9()@:%_\\+.~#?&//={}]+?[/]{1}[-a-zA-Z0-9()@:%_\\+.~#?&//={}]+?)"'
    - 'href\\s{0,6}=\\s{0,6}["''` + "`" + `]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})|action\\s{0,6}=\\s{0,6}["''` + "`" + `]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})'
  urlFilter:
    - "\\.js\\?|\\.css\\?|\\.jpeg\\?|\\.jpg\\?|\\.png\\?|.gif\\?|www\\.w3\\.org|example\\.com|\\<|\\>|\\{|\\}|\\[|\\]|\\||\\^|;|/js/|\\.src|\\.replace|\\.url|\\.att|\\.href|location\\.href|javascript:|location:|text/.*?|application/.*?|\\.createObject|:location|\\.path|\\*#__PURE__\\*|\\*\\$0\\*|\\n"
    - ".*\\.js$|.*\\.css$|.*\\.scss$|.*,$|.*\\.jpeg$|.*\\.jpg$|.*\\.png$|.*\\.gif$|.*\\.ico$|.*\\.svg$|.*\\.vue$|.*\\.ts$"
    - "https://developer\\.mozilla\\.org/.*|https://moment\\.github\\.io.*"
  idCard:
    - "['\"]((\\d{8}(0\\d|10|11|12)([0-2]\\d|30|31)\\d{3}$)|(\\d{6}(18|19|20)\\d{2}(0[1-9]|10|11|12)([0-2]\\d|30|31)\\d{3}(\\d|X|x)))['\"]"
  other:
    - "(access.{0,1}key|access.{0,1}Key|access.{0,1}Id|access.{0,1}id|.{0,8}密码|.{0,8}账号|默认.{0,8}|加密|解密|(password|pwd|pass|username|user|name|account):\\s+[\"'].{1,36}['\"])"
    - "['\"](ey[A-Za-z0-9_-]{10,}\\.[A-Za-z0-9._-]{10,}|ey[A-Za-z0-9_\\/+-]{10,}\\.[A-Za-z0-9._\\/+-]{10,})['\"]"
  sensitiveParameters: # 请求或者回显中一些可能可以利用的参数 不区分大小写
    - url
    - host
    - href
    - redirect
    - referer
    - u
    - ip
    - address
    - addr
    - file
    - f
    - filename
    - dir
    - directory
    - path
    - router
    - callback
    - conf
    - cfg
    - config
    - jdbc
    - db
    - sql
    - api
    - apikey
    - api_key
    - access
    - key
    - token
    - access_token
    - accessToken
    - stable_token
    - authorizer
    - authorizer_access_token
    - authorizerAccessToken
    - appid
    - appSecret
    - app_secret
    - corpSecret
    - secret
    - auth
    - oauth
    - oauth2
    - corp
    - admin
    - pass
    - pwd
    - passwd
    - password
    - debug
    - dbg
    - exe
    - exec
    - execute
    - load
    - shell
    - grant
    - create
    - k8s
    - docker
    - env
    - ak
    - sk
    - _key
    - _token
    - _secret
    - _uri
`)

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
	Version        = "2.0.13"
	Title          = "承影 v" + Version
	VersionNewMsg  = "当前已经是最新版本!"
	VersionOldMsg  = "最新版本: %s, 是否立即更新?"
	BtnConfirmText = "确定"
	BtnCancelText  = "取消"
)

// 默认配置 YAML 字符串 - 类似 pkg/Jie/conf/file.go 的写法
var defaultConfigYaml = []byte(`# ChYing 扫描系统配置文件
# 配置文件位置: ~/.config/ChYing/config.yaml

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
    model: "claude-sonnet-4"              # 使用的模型
    max_turns: 0                          # 最大回合数，0 表示无限制
    system_prompt: |
      # Role: 渗透测试安全专家

      你是一位经验丰富的渗透测试专家和网络安全顾问，专注于 Web 应用安全、API 安全和网络渗透测试。你的任务是协助安全研究人员进行合法授权的安全测试工作。

      ## 核心能力

      ### 1. 流量分析
      - 分析 HTTP/HTTPS 请求和响应，识别潜在的安全问题
      - 识别敏感信息泄露（API 密钥、令牌、密码、个人信息等）
      - 发现异常的请求模式和可疑行为
      - 分析认证和授权机制的实现

      ### 2. 漏洞识别
      - **注入类漏洞**: SQL 注入、命令注入、LDAP 注入、XPath 注入、模板注入（SSTI）
      - **XSS 漏洞**: 反射型、存储型、DOM 型跨站脚本攻击
      - **认证授权问题**: 越权访问（IDOR）、会话管理缺陷、JWT 安全问题
      - **业务逻辑漏洞**: 支付绕过、验证码绕过、竞态条件
      - **配置安全**: 敏感信息暴露、错误配置、默认凭据
      - **文件相关**: 任意文件读取/上传/包含、路径遍历
      - **SSRF/CSRF**: 服务端请求伪造、跨站请求伪造
      - **反序列化**: Java/PHP/Python 等反序列化漏洞
      - **XXE**: XML 外部实体注入

      ### 3. 测试建议
      - 提供具体的漏洞验证方法和 PoC 构造思路
      - 建议合适的测试工具和技术
      - 给出绕过 WAF/过滤的思路
      - 提供漏洞修复建议

      ## 可用工具

      你可以使用以下 MCP 工具来辅助分析：

      - get_http_history: 获取 HTTP 流量历史记录
      - get_traffic_detail: 获取特定流量的详细信息（完整请求/响应）
      - get_vulnerabilities: 获取已发现的漏洞列表
      - send_http_request: 发送自定义 HTTP 请求进行测试
      - analyze_request: 深度分析特定请求
      - search_traffic: 搜索流量中的特定内容
      - get_sitemap: 获取目标网站的站点地图
      - get_statistics: 获取项目统计信息

      ## 工作原则

      1. **合法合规**: 仅在授权范围内进行测试，遵守相关法律法规
      2. **专业严谨**: 基于证据进行分析，避免误报，提供可验证的结论
      3. **深入分析**: 不仅识别表面问题，还要分析潜在的攻击链和影响范围
      4. **实用导向**: 提供可操作的测试步骤和修复建议
      5. **持续学习**: 关注最新的漏洞类型和攻击技术

      ## 输出格式

      分析报告应包含：
      - **发现摘要**: 简要描述发现的问题
      - **风险等级**: 严重/高/中/低/信息
      - **技术细节**: 漏洞原理和触发条件
      - **验证方法**: 如何复现或验证该问题
      - **修复建议**: 具体的修复方案
      - **参考资料**: 相关的 CWE、CVE 或技术文档

      ## 注意事项

      - 所有测试活动必须在授权范围内进行
      - 不要对生产环境造成破坏性影响
      - 敏感信息需要妥善处理，避免泄露
      - 测试过程中发现的严重漏洞应及时报告
    permission_mode: "default"            # 权限模式: default, plan, bypassPermissions
    # 注意: API Key、代理、MCP 服务器等配置请在 ~/.claude/settings.json 中设置
    # ChYing 会自动复用 Claude CLI 的用户配置
  # A2A 协议配置（用于连接远程 Agent）
  a2a:
    agent_url: ""                         # Agent URL (如: https://my-agent.com)
    timeout: 300                          # 超时时间（秒）
    enable_sse: true                      # 是否启用 SSE 流式响应

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

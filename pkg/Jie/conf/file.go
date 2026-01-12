package conf

import (
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2023/11/15
   @desc 生成配置文件，主要是供被动扫描时随时更改一些基本配置，就不用每次重新构建了
**/

var FileName = "Jie_config.yaml"

var defaultConfigYaml = []byte(`version: ` + Version + `

parallel: 10                            # 同时扫描的最大 url 个数

# 全局 http 发包配置
http:
  proxy: ""                             # 漏洞扫描时使用的代理，如: http://127.0.0.1:8080
  timeout: 10                           # 建立 tcp 连接的超时时间
  maxConnsPerHost: 100                  # 每个 host 最大连接数
  retryTimes: 0                         # 请求失败的重试次数，0 则不重试
  allowRedirect: 0                      # 单个请求最大允许的跳转数，0 则不跳转
  verifySSL: false                      # 是否验证 ssl 证书
  maxQps: 50                            # 每秒最大请求数
  headers:                              # 指定 http 请求头
  forceHTTP1: false                     # 强制指定使用 http/1.1, 不然会根据服务器选择，如果服务器支持 http2，默认会使用 http2

# 漏洞探测的插件配置
plugins:
  bruteForce:
    web: false                          # web 服务类的爆破，比如 tomcat 爆破
    service: false                      # 服务类的爆破，比如 mysql 爆破
    usernameDict: ""                    # 自定义用户名字典, 为空将使用内置字典, 配置后将与内置字典**合并**
    passwordDict: ""                    # 自定义密码字典，为空将使用内置字典, 配置后将与内置字典**合并**
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
  bbscan:                               # bbscan https://github.com/lijiejie/bbscan 这种规则类目录扫描
    enabled: false
  jsonp:
    enabled: false
  log4j:
    enabled: false
  bypass403:
    enabled: false
  fastjson:
    enabled: false
  archive:                                # 从 https://web.archive.org/ 获取历史 url，作为补充扫描
    enabled: false
  iis:                                    # iis 短文件名 fuzz
    enabled: false
  nginxAliasTraversal:                    # nginx 别名遍历
    enabled: false
  poc:
    enabled: false
  nuclei:
    enabled: false
  portScan:
    enabled: false

# 反连平台配置
# 注意: 默认配置为 dig.pm, 可以使用 https://github.com/yumusb/DNSLog-Platform-Golang 自行搭建，后续看需求要不要支持别的 dnslog 平台
reverse:
  host: "https://dig.pm/"               # 反连平台地址
  Domain: "ipv6.bypass.eu.org."         # 指定反连域名

# 基础爬虫配置 这里都没写呢，后边看看要不要写一下
basicCrawler:
  maxDepth: 0                           # 最大爬取深度， 0 为无限制
  maxCountOfLinks: 0                    # 本次爬取收集的最大链接数, 0 为无限制
  allowVisitParentPath: false           # 是否允许爬取父目录, 如果扫描目标为 t.com/a/且该项为 false, 那么就不会爬取 t.com/ 这级的内容
  restriction:                          # 爬虫的允许爬取的资源限制, 为空表示不限制。爬虫会自动添加扫描目标到 Hostname_allowed。
    hostname_allowed: []                # 允许访问的 Hostname，支持格式如 t.com、*.t.com、1.1.1.1、1.1.1.1/24、1.1-4.1.1-8
    hostname_disallowed:                # 不允许访问的 Hostname，支持格式如 t.com、*.t.com、1.1.1.1、1.1.1.1/24、1.1-4.1.1-8
    - '*.edu.*'
    - '*.gov.*'
    port_allowed: []                    # 允许访问的端口, 支持的格式如: 80、80-85
    port_disallowed: []                 # 不允许访问的端口, 支持的格式如: 80、80-85
    path_allowed: []                    # 允许访问的路径，支持的格式如: test、*test*
    path_disallowed: []                 # 不允许访问的路径, 支持的格式如: test、*test*
    query_key_allowed: []               # 允许访问的 Query Key，支持的格式如: test、*test*
    query_key_disallowed: []            # 不允许访问的 Query Key, 支持的格式如: test、*test*
    fragment_allowed: []                # 允许访问的 Fragment, 支持的格式如: test、*test*
    fragment_disallowed: []             # 不允许访问的 Fragment, 支持的格式如: test、*test*
    post_key_allowed: []                # 允许访问的 Post Body 中的参数, 支持的格式如: test、*test*
    post_key_disallowed: []             # 不允许访问的 Post Body 中的参数, 支持的格式如: test、*test*
  basic_auth:                           # 基础认证信息
    username: ""
    password: ""

# 被动代理配置
mitmproxy:
  caCert: ./ca.crt                      # CA 根证书路径
  caKey: ./ca.key                       # CA 私钥路径
  basicAuth:                            # 基础认证的用户名密码
    header: "Go-Mitmproxy-Authorization"    # 认证头
    username: ""
    password: ""
  exclude:                              # 不允许访问的 Hostname，支持格式如 t.com、*.t.com、 todo 1.1.1.1、1.1.1.1/24、1.1-4.1.1-8
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
    - .lencr.org                         # https://letsencrypt.org/docs/lencr.org/
    - .adavoid.org

  include:                              # 允许访问的 Hostname，支持格式如 t.com、*.t.com、1.1.1.1、1.1.1.1/24、1.1-4.1.1-8
    - 
  # 排除的后缀, 不进行显示和扫描
  filterSuffix: .3g2, .3gp, .arj, .avi, .axd, .bmp, .drv, .eot, .flv, .gif, .gifv, .h264, .ico, .jpeg, .jpg, .m4a, .m4v, .mkv, .mov, .mp3, .mp4, .mpeg, .mpg, .ogg, .ogm, .ogv, .otf, .png, .psd, .rm, .svg, .swf, .sys, .tif, .tiff, .ttf, .vob, .wav, .webm, .webp, .wmv, .woff, .woff2, .xcf
  maxLength: 3000                       # 队列长度限制, 也可以理解为最大允许多少等待扫描的请求, 请根据内存大小自行调整，这个还没有实现，我没有使用队列

# 信息收集类的正则
collection:
  domain:
    - "['\"](([a-zA-Z0-9]{1,9}:)?//)?(.{1,36}:.{1,36}@)?[a-zA-Z0-9\\-\\.]*?\\.(xin|com|cn|net|com\\.cn|vip|top|cc|shop|club|wang|xyz|luxe|site|news|pub|fun|online|win|red|loan|ren|mom|net\\.cn|org|link|biz|bid|help|tech|date|mobi|so|me|tv|co|vc|pw|video|party|pics|website|store|ltd|ink|trade|live|wiki|space|gift|lol|work|band|info|click|photo|market|tel|social|press|game|kim|org\\.cn|games|pro|men|love|studio|rocks|asia|group|science|design|software|engineer|lawyer|fit|beer|我爱你|中国|公司|网络|在线|网址|网店|集团|中文网)(:\\d{1,5})?"
  ip:
    - "['\"](([a-zA-Z0-9]{1,9}:)?//)?(.{1,36}:.{1,36}@)?\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}(:\\d{1,5})?"
  phone:
    - "['\"](1(3([0-35-9]\\d|4[1-8])|4[14-9]\\d|5([\\d]\\d|7[1-79])|66\\d|7[2-35-8]\\d|8\\d{2}|9[89]\\d)\\d{7})['\"]"
  email:
    - "['\"]([\\w!#$%&'*+=?^_` + "`" + `{|}~-]+(?:\\.[\\w!#$%&'*+=?^_` + "`" + `{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[\\w](?:[\\w-]*[\\w])?)['\"]"
  api:      # 自己来写正则吧，网上找的都不太靠谱, 见到了慢慢补吧
    - "(?i)\\.(get|post|put|delete|options|connect|trace|patch)\\([\"'](/?.*?)[\"']"
    - "(?:\"|')(/[^/\"']+){2,}(?:\"|')"
  url:
    - "[\"'‘“` + "`" + `]\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250}?)\\s{0,6}[\"'‘“` + "`" + `]"
    - "=\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})"
    - "[\"'‘“` + "`" + `]\\s{0,6}([#,.]{0,2}/[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250}?)\\s{0,6}[\"'‘“` + "`" + `]"
    - "\"([-a-zA-Z0-9()@:%_\\+.~#?&//={}]+?[/]{1}[-a-zA-Z0-9()@:%_\\+.~#?&//={}]+?)\""
    - "href\\s{0,6}=\\s{0,6}[\"'‘“` + "`" + `]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})|action\\s{0,6}=\\s{0,6}[\"'‘“` + "`" + `]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})"
  urlFilter:
    - "\\.js\\?|\\.css\\?|\\.jpeg\\?|\\.jpg\\?|\\.png\\?|.gif\\?|www\\.w3\\.org|example\\.com|\\<|\\>|\\{|\\}|\\[|\\]|\\||\\^|;|/js/|\\.src|\\.replace|\\.url|\\.att|\\.href|location\\.href|javascript:|location:|text/.*?|application/.*?|\\.createObject|:location|\\.path|\\*#__PURE__\\*|\\*\\$0\\*|\\n"
    - ".*\\.js$|.*\\.css$|.*\\.scss$|.*,$|.*\\.jpeg$|.*\\.jpg$|.*\\.png$|.*\\.gif$|.*\\.ico$|.*\\.svg$|.*\\.vue$|.*\\.ts$"
    - "https://developer\\.mozilla\\.org/.*|https://moment\\.github\\.io.*"
  idCard:
    - "['\"]((\\d{8}(0\\d|10|11|12)([0-2]\\d|30|31)\\d{3}$)|(\\d{6}(18|19|20)\\d{2}(0[1-9]|10|11|12)([0-2]\\d|30|31)\\d{3}(\\d|X|x)))['\"]"
  other:
    - "(access.{0,1}key|access.{0,1}Key|access.{0,1}Id|access.{0,1}id|.{0,8}密码|.{0,8}账号|默认.{0,8}|加密|解密|(password|pwd|pass|username|user|name|account):\\s+[\"'].{1,36}['\"])"
    - "['\"](ey[A-Za-z0-9_-]{10,}\\.[A-Za-z0-9._-]{10,}|ey[A-Za-z0-9_\\/+-]{10,}\\.[A-Za-z0-9._\\/+-]{10,})['\"]"
  sensitiveParameters: # 请求或者回显中一些可能可以利用的参数 不区分大小写 (注意: 改为 camelCase 以与 ChYing 统一)
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
    - _key  # 这种以 _ 开头的会不完全匹配，包含 _key 就会抛出来
    - _token
    - _secret
    - _uri
`)

// Init 加载配置
func Init(dir string) {
	// 配置文件路径 当前文件夹 + SCopilot.yaml
	ConfigFile = path.Join(dir, FileName)

	// 检测配置文件是否存在
	if !util.Exists(ConfigFile) {
		err := WriteYamlConfig(dir)
		if err != nil {
			logging.Logger.Fatalln(err)
			return
		}
		logging.Logger.Infof("%s not find, Generate profile.", ConfigFile)
	} else {
		logging.Logger.Infoln("Load profile", ConfigFile)
	}

	ReadYamlConfig()
}

// WriteYamlConfig 生成写入默认配置文件, 这里就不通过 viper 写入了， viper 写入的没有注释
func WriteYamlConfig(dir string) error {
	// 写入默认配置文件
	jwtPath := filepath.Join(dir, FileName)
	if err := os.WriteFile(jwtPath, defaultConfigYaml, 0644); err != nil {
		logging.Logger.Fatalf("创建%s配置文件失败: %s", FileName, err)
	}

	return nil
}

// ReadYamlConfig 读取配置文件
func ReadYamlConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(ConfigFile)

	err := viper.ReadInConfig()
	if err != nil {
		logging.Logger.Fatalf("Fail to read %s: %+v", ConfigFile, err)
	}
	err = viper.Unmarshal(&GlobalConfig)

	if err != nil {
		logging.Logger.Fatalf("Fail to parse '%s', check format: %+v", ConfigFile, err)
	}
	ReadPlugin()
	logging.Logger.Infoln("Reload config file:", ConfigFile)
}

// ReadPlugin 插件读取出来方便使用，之后所有的插件运行都是看 Plugin 中对应的是否开启
func ReadPlugin() {
	// 先全部关闭，再根据配置开启对应的，防止配置文件中删除了某个插件，但是程序中还在运行
	for k := range Plugin {
		Plugin[k] = false
	}

	if GlobalConfig.Plugins.XSS.Enabled {
		Plugin["xss"] = true
	}

	if GlobalConfig.Plugins.Sql.Enabled {
		Plugin["sql"] = true
	}

	if GlobalConfig.Plugins.SqlmapApi.Enabled {
		Plugin["sqlmapApi"] = true
		GlobalConfig.SqlmapApi = Sqlmap{
			Enabled:  true,
			Url:      GlobalConfig.Plugins.SqlmapApi.Url,
			Username: GlobalConfig.Plugins.SqlmapApi.Username,
			Password: GlobalConfig.Plugins.SqlmapApi.Password,
		}
	} else {
		GlobalConfig.SqlmapApi = Sqlmap{
			Enabled:  false,
			Url:      GlobalConfig.Plugins.SqlmapApi.Url,
			Username: GlobalConfig.Plugins.SqlmapApi.Username,
			Password: GlobalConfig.Plugins.SqlmapApi.Password,
		}
	}

	if GlobalConfig.Plugins.CmdInjection.Enabled {
		Plugin["cmd"] = true
	}

	if GlobalConfig.Plugins.XXE.Enabled {
		Plugin["xxe"] = true
	}

	if GlobalConfig.Plugins.SSRF.Enabled {
		Plugin["ssrf"] = true
	}

	if GlobalConfig.Plugins.BruteForce.Web {
		Plugin["brute"] = true
	}

	if GlobalConfig.Plugins.ByPass403.Enabled {
		Plugin["bypass403"] = true
	}

	if GlobalConfig.Plugins.Jsonp.Enabled {
		Plugin["jsonp"] = true
	}

	if GlobalConfig.Plugins.CrlfInjection.Enabled {
		Plugin["crlf"] = true
	}

	if GlobalConfig.Plugins.Log4j.Enabled {
		Plugin["log4j"] = true
	}

	if GlobalConfig.Plugins.Fastjson.Enabled {
		Plugin["fastjson"] = true
	}

	if GlobalConfig.Plugins.PortScan.Enabled {
		Plugin["portScan"] = true
	}

	if GlobalConfig.Plugins.Poc.Enabled {
		Plugin["poc"] = true
	}

	if GlobalConfig.Plugins.Nuclei.Enabled {
		Plugin["nuclei"] = true
	}

	if GlobalConfig.Plugins.BBscan.Enabled {
		Plugin["bbscan"] = true
	}

	if GlobalConfig.Plugins.Archive.Enabled {
		Plugin["archive"] = true
	}

	if GlobalConfig.Plugins.NginxAliasTraversal.Enabled {
		Plugin["nginx-alias-traversal"] = true
	}
}

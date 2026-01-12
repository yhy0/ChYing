package conf

/**
   @author yhy
   @since 2023/11/15
   @desc 注意 json 别名不能是_ - viper 好像不识别这种的，最好是驼峰式
   @note 此文件同时被 Jie 和 ChYing 使用，ChYing 独有字段已标注
**/

type Config struct {
	Debug      bool       `json:"debug" yaml:"debug"`
	Options    Options    `json:"options" yaml:"options"`
	Passive    Passive    `json:"passive" yaml:"passive"`
	Http       Http       `json:"http" yaml:"http"`
	Plugins    Plugins    `json:"plugins" yaml:"plugins"`
	WebScan    WebScan    `json:"webScan" yaml:"webScan"`
	NoPortScan bool       `json:"no_port_scan" yaml:"no_port_scan"`
	Reverse    Reverse    `json:"reverse" yaml:"reverse"`
	SqlmapApi  Sqlmap     `json:"sqlmapApi" yaml:"sqlmapApi"`
	Mitmproxy  Mitmproxy  `json:"mitmproxy" yaml:"mitmproxy"`
	Collection Collection `json:"collection" yaml:"collection"`
}

type WebScan struct {
	Poc  []string `json:"poc" yaml:"poc"`
	Craw string   `json:"craw" yaml:"craw"`
	Show bool     `json:"show" yaml:"show"`
}

type Options struct {
	Target     string // target URLs/hosts to scan
	TargetFile string
	Targets    []string
	Output     string
	Mode       string
	S2         S2
	Shiro      Shiro
}

type Http struct {
	Proxy           string            `json:"proxy" yaml:"proxy"`                     // 漏洞扫描时使用的代理，如: http://127.0.0.1:8080
	Timeout         int               `json:"timeout" yaml:"timeout"`                 // 建立 tcp 连接的超时时间
	MaxConnsPerHost int               `json:"maxConnsPerHost" yaml:"maxConnsPerHost"` // 每个 host 最大连接数
	RetryTimes      int               `json:"retryTimes" yaml:"retryTimes"`           // 请求失败的重试次数，0 则不重试
	AllowRedirect   int               `json:"allowRedirect" yaml:"allowRedirect"`     // 单个请求最大允许的跳转数，0 则不跳转
	VerifySSL       bool              `json:"verifySSL" yaml:"verifySSL"`             // 是否验证 ssl 证书
	MaxQps          int               `json:"maxQps" yaml:"maxQps"`                   // 每秒最大请求数
	Headers         map[string]string `json:"headers" yaml:"headers"`                 // 指定 http 请求头
	ForceHTTP1      bool              `json:"forceHTTP1" yaml:"forceHTTP1"`           // 强制指定使用 http/1.1
}

type Passive struct {
	ProxyPort string `mapstructure:"port" json:"port" yaml:"port"`
	WebPort   string `mapstructure:"webPort" json:"webPort" yaml:"webPort"`
	WebUser   string `mapstructure:"webUser" json:"webUser" yaml:"webUser"`
	WebPass   string `mapstructure:"webPass" json:"webPass" yaml:"webPass"`
}

type S2 struct {
	Mode        string
	Name        string
	Body        string
	CMD         string
	ContentType string
}

type Shiro struct {
	Mode     string
	Cookie   string
	Platform string
	Key      string
	KeyMode  string
	Gadget   string
	CMD      string
	Echo     string
}

// Plugins 插件配置
type Plugins struct {
	BruteForce struct {
		Web                bool   `json:"web" yaml:"web"`
		Service            bool   `json:"service" yaml:"service"` // ChYing 使用：服务类的爆破
		UsernameDictionary string `json:"usernameDict" yaml:"usernameDict"`
		PasswordDictionary string `json:"passwordDict" yaml:"passwordDict"`
	} `json:"bruteForce" yaml:"bruteForce"`

	CmdInjection struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"cmdInjection" yaml:"cmdInjection"`

	CrlfInjection struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"crlfInjection" yaml:"crlfInjection"`

	XSS struct {
		Enabled           bool `json:"enabled" yaml:"enabled"`
		DetectXssInCookie bool `json:"detectXssInCookie" yaml:"detectXssInCookie"`
	} `json:"xss" yaml:"xss"`

	Sql struct {
		Enabled               bool `json:"enabled" yaml:"enabled"`
		BooleanBasedDetection bool `json:"booleanBasedDetection" yaml:"booleanBasedDetection"`
		TimeBasedDetection    bool `json:"timeBasedDetection" yaml:"timeBasedDetection"`
		ErrorBasedDetection   bool `json:"errorBasedDetection" yaml:"errorBasedDetection"`
		DetectInCookie        bool `json:"detectInCookie" yaml:"detectInCookie"`
	} `json:"sql" yaml:"sql"`

	SqlmapApi Sqlmap `json:"sqlmapApi" yaml:"sqlmapApi"`

	XXE struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"xxe" yaml:"xxe"`

	SSRF struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"ssrf" yaml:"ssrf"`

	BBscan struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"bbscan" yaml:"bbscan"`

	Jsonp struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"jsonp" yaml:"jsonp"`

	Log4j struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"log4j" yaml:"log4j"`

	ByPass403 struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"bypass403" yaml:"bypass403"`

	Fastjson struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"fastjson" yaml:"fastjson"`

	NginxAliasTraversal struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"nginxAliasTraversal" yaml:"nginxAliasTraversal"`

	Poc struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"poc" yaml:"poc"`

	Nuclei struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"nuclei" yaml:"nuclei"`

	Archive struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"archive" yaml:"archive"`

	IIS struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"iis" yaml:"iis"`

	PortScan struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	} `json:"portScan" yaml:"portScan"`
}

// Reverse dnslog 配置，使用 dig.pm https://github.com/yumusb/DNSLog-Platform-Golang
type Reverse struct {
	Host   string `json:"host" yaml:"host"`
	Domain string `json:"domain" yaml:"domain"`
}

// Sqlmap Sqlmap API 配置
type Sqlmap struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`   // 是否开启 sqlmap api
	Url      string `json:"url" yaml:"url"`           // SQLMap API 服务器地址
	Username string `json:"username" yaml:"username"` // SQLMap API 用户名
	Password string `json:"password" yaml:"password"` // SQLMap API 密码
}

// Collection 信息收集中的正则
type Collection struct {
	Domain              []string `json:"domain" yaml:"domain"`
	IP                  []string `json:"ip" yaml:"ip"`
	Phone               []string `json:"phone" yaml:"phone"`
	Email               []string `json:"email" yaml:"email"`
	IDCard              []string `json:"idCard" yaml:"idCard"`
	API                 []string `json:"api" yaml:"api"`
	Url                 []string `json:"url" yaml:"url"`
	UrlFilter           []string `json:"urlFilter" yaml:"urlFilter"`
	Other               []string `json:"other" yaml:"other"`
	SensitiveParameters []string `json:"sensitiveParameters" yaml:"sensitiveParameters"` // 改为 camelCase
}

type Mitmproxy struct {
	CaCert    string `json:"caCert" yaml:"caCert"` // ChYing 使用：CA 根证书路径
	CaKey     string `json:"caKey" yaml:"caKey"`   // ChYing 使用：CA 私钥路径
	BasicAuth struct {
		Header   string `json:"header" yaml:"header"` // 认证头
		Username string `json:"username" yaml:"username"`
		Password string `json:"password" yaml:"password"`
	} `json:"basicAuth" yaml:"basicAuth"`
	Exclude      []string `json:"exclude" yaml:"exclude"`           // Exclude 排除扫描的域名
	Include      []string `json:"include" yaml:"include"`           // Include 只扫描的域名
	FilterSuffix string   `json:"filterSuffix" yaml:"filterSuffix"` // 排除的后缀
	MaxLength    int      `json:"maxLength" yaml:"maxLength"`       // ChYing 使用：队列长度限制
}

// BasicCrawler 爬虫配置 - ChYing 使用
type BasicCrawler struct {
	MaxDepth             int  `json:"maxDepth" yaml:"maxDepth"`
	MaxCountOfLinks      int  `json:"maxCountOfLinks" yaml:"maxCountOfLinks"`
	AllowVisitParentPath bool `json:"allowVisitParentPath" yaml:"allowVisitParentPath"`
	Restriction          struct {
		HostnameAllowed    []string `json:"hostname_allowed" yaml:"hostname_allowed"`
		HostnameDisallowed []string `json:"hostname_disallowed" yaml:"hostname_disallowed"`
		PortAllowed        []string `json:"port_allowed" yaml:"port_allowed"`
		PortDisallowed     []string `json:"port_disallowed" yaml:"port_disallowed"`
		PathAllowed        []string `json:"path_allowed" yaml:"path_allowed"`
		PathDisallowed     []string `json:"path_disallowed" yaml:"path_disallowed"`
		QueryKeyAllowed    []string `json:"query_key_allowed" yaml:"query_key_allowed"`
		QueryKeyDisallowed []string `json:"query_key_disallowed" yaml:"query_key_disallowed"`
		FragmentAllowed    []string `json:"fragment_allowed" yaml:"fragment_allowed"`
		FragmentDisallowed []string `json:"fragment_disallowed" yaml:"fragment_disallowed"`
		PostKeyAllowed     []string `json:"post_key_allowed" yaml:"post_key_allowed"`
		PostKeyDisallowed  []string `json:"post_key_disallowed" yaml:"post_key_disallowed"`
	} `json:"restriction" yaml:"restriction"`
	BasicAuth struct {
		Username string `json:"username" yaml:"username"`
		Password string `json:"password" yaml:"password"`
	} `json:"basic_auth" yaml:"basic_auth"`
}

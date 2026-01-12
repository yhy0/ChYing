package httpx

import (
	"github.com/imroc/req/v3"
	"github.com/projectdiscovery/rawhttp/client"
	"go.uber.org/ratelimit"
	"net/http"
)

/**
   @author yhy
   @since 2025/6/7
   @desc //TODO
**/

type Response struct {
	Status           string
	StatusCode       int
	Body             string
	RequestDump      string
	ResponseDump     string
	RespHeader       http.Header
	ContentLength    int
	RequestUrl       string
	Location         string
	ServerDurationMs float64 // 服务器响应时间
}

type Options struct {
	Timeout         int
	RetryTimes      int    // 重定向次数 0 为不重试
	VerifySSL       bool   // default false
	AllowRedirect   int    // default false
	Proxy           string // proxy settings, support http/https proxy only, e.g. http://127.0.0.1:8080
	QPS             int    // 每秒最大请求数
	MaxConnsPerHost int    // 每个 host 最大连接数
	Headers         map[string]string
}

type Client struct {
	Client      *req.Client
	Options     *Options
	RateLimiter ratelimit.Limiter // 每秒请求速率限制
}

// RawRequest defines a basic HTTP raw request
type RawRequest struct {
	FullURL        string
	Method         string
	Path           string
	Data           string
	Headers        map[string]string
	UnsafeHeaders  client.Headers
	UnsafeRawBytes []byte
}

type RequestScanMsg struct {
	Id          int64  `json:"id"`
	ModuleName  string `json:"module_name"`
	Target      string `json:"target"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Status      int    `json:"status"`
	Length      int    `json:"length"`
	Title       string `json:"title"`
	IP          string `json:"ip"`
	ContentType string `json:"content_type"`
	Timestamp   string `json:"timestamp"`
}

// HttpBody 存储请求和响应的原始报文
type HttpBody struct {
	Id          int64  `json:"id"`
	RequestRaw  string `json:"request_raw"`
	ResponseRaw string `json:"response_raw"`
}

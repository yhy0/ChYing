package httpx

import (
	"bytes"
	"crypto/tls"
	"github.com/corpix/uarand"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/logging"
	"go.uber.org/ratelimit"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

/**
  @author: yhy
  @since: 2022/6/1
  @desc: //TODO
**/

type Response struct {
	Status           string
	StatusCode       int
	Body             string
	RequestDump      string
	ResponseDump     string
	Header           http.Header
	ContentLength    int
	RequestUrl       string
	Location         string
	ServerDurationMs float64 // 服务器响应时间
}

type Session struct {
	// Client is the current http client
	Client *http.Client
	// Rate limit instance
	RateLimiter ratelimit.Limiter // 每秒请求速率限制
}

var session *Session

var rateLimit = 30

func NewSession() {
	Transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		DisableKeepAlives:   true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	//if options.Cert != "" && options.PrivateKey != "" {
	//	cer, err := tls.LoadX509KeyPair(options.Cert, options.PrivateKey)
	//	if err != nil {
	//		panic(err)
	//	}
	//	// 设置证书
	//	client.TLSConfig = &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{cer}}

	// Add proxy
	if conf.Proxy != "" {
		proxyURL, _ := url.Parse(conf.Proxy)
		if isSupportedProtocol(proxyURL.Scheme) {
			Transport.Proxy = http.ProxyURL(proxyURL)
			//if proxyURL.Scheme == "socks5" {
			//	// socks5 代理
			//	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
			//	if err != nil {
			//		logging.Logger.Errorln(os.Stderr, "can't connect to the proxy:", err)
			//	} else {
			//		// set our socks5 as the dialer
			//		Transport.Dial = dialer.Dial
			//	}
			//} else {
			//	Transport.Proxy = http.ProxyURL(proxyURL)
			//}
		} else {
			logging.Logger.Warnln("Unsupported proxy protocol: %s", proxyURL.Scheme)
		}
	}

	client := &http.Client{
		Transport: Transport,
		Timeout:   time.Duration(5) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	session = &Session{
		Client: client,
	}

	// Initiate rate limit instance
	session.RateLimiter = ratelimit.New(rateLimit)
}

func Get(target string) (*Response, error) {
	return Request(target, "GET", "", false, nil)
}

func Request(target string, method string, postdata string, isredirect bool, headers map[string]string) (*Response, error) {
	if isredirect {
		jar, _ := cookiejar.New(nil)
		session.Client.Jar = jar
	}

	req, err := http.NewRequest(strings.ToUpper(method), target, strings.NewReader(postdata))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Connection", "close")
	flag := true
	for k, v := range headers {
		if k == "Content-Type" {
			flag = false
		}
		req.Header[k] = []string{v}
	}

	if flag {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}

	var start = time.Now()
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	requestDump, _ := httputil.DumpRequestOut(req, true)
	session.RateLimiter.Take()
	resp, err := session.Client.Do(req)
	if err != nil {
		//防止空指针
		return &Response{"999", 999, "", "", "", nil, 0, "", "", 0}, err
	}

	responseDump, _ := httputil.DumpResponse(resp, true)
	var location string
	var respbody string
	defer resp.Body.Close()
	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		respbody = string(body)
	}
	if resplocation, err := resp.Location(); err == nil {
		location = resplocation.String()
	}

	contentLength := int(resp.ContentLength)

	if contentLength == -1 {
		contentLength = len(respbody)
	}

	return &Response{resp.Status, resp.StatusCode, respbody, string(requestDump), string(responseDump), resp.Header, contentLength, resp.Request.URL.String(), location, float64(time.Since(start).Milliseconds())}, nil
}

// UploadRequest 新建上传请求
func UploadRequest(target string, params map[string]string, name, path string) (*Response, error) {
	body := &bytes.Buffer{}                                       // 初始化body参数
	writer := multipart.NewWriter(body)                           // 实例化multipart
	part, err := writer.CreateFormFile(name, filepath.Base(path)) // 创建multipart 文件字段
	if err != nil {
		return nil, err
	}

	_, err = part.Write([]byte("test")) // 写入文件数据到multipart
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val) // 写入body中额外参数
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", target, body) // 新建请求
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Content-Type", writer.FormDataContentType()) // 设置请求头,!!!非常重要，否则远端无法识别请求
	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Connection", "close")

	requestDump, _ := httputil.DumpRequestOut(req, true)

	session.RateLimiter.Take()

	resp, err := session.Client.Do(req)
	if err != nil {
		//防止空指针
		return &Response{"999", 999, "", "", "", nil, 0, "", "", 0}, err
	}

	responseDump, _ := httputil.DumpResponse(resp, true)
	var location string
	var respbody string
	defer resp.Body.Close()
	if bodytmp, err := ioutil.ReadAll(resp.Body); err == nil {
		respbody = string(bodytmp)
	}
	if resplocation, err := resp.Location(); err == nil {
		location = resplocation.String()
	}

	return &Response{resp.Status, resp.StatusCode, respbody, string(requestDump), string(responseDump), resp.Header, int(resp.ContentLength), resp.Request.URL.String(), location, 0}, nil
}

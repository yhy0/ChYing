package httpx

import (
	"crypto/tls"
	"errors"
	"github.com/corpix/uarand"
	"github.com/projectdiscovery/fastdialer/fastdialer"
	"github.com/projectdiscovery/retryablehttp-go"
	"github.com/yhy0/Jie/conf"
	"github.com/yhy0/logging"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

/**
  @author: yhy
  @since: 2023/4/23
  @desc: //TODO
**/

type HTTPX struct {
	Client *retryablehttp.Client
	Dialer *fastdialer.Dialer
}

func NewClient(single bool) *HTTPX {
	httpx := &HTTPX{}
	fastdialerOpts := fastdialer.DefaultOptions
	fastdialerOpts.EnableFallback = true
	fastdialerOpts.WithDialerHistory = true
	fastdialerOpts.WithZTLS = false

	fastDialer, err := fastdialer.NewDialer(fastdialerOpts)
	if err != nil {
		logging.Logger.Errorf("could not create resolver cache: %s", err)
		return nil
	}

	httpx.Dialer = fastDialer

	transport := &http.Transport{
		DialContext:         fastDialer.Dial,
		DialTLSContext:      fastDialer.DialTLS,
		MaxIdleConnsPerHost: -1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		},
		DisableKeepAlives: true,
	}

	if conf.GlobalConfig.WebScan.Proxy != "" {
		proxyURL, _ := url.Parse(conf.GlobalConfig.WebScan.Proxy)
		if isSupportedProtocol(proxyURL.Scheme) {
			transport.Proxy = http.ProxyURL(proxyURL)
		} else {
			logging.Logger.Warnln("Unsupported proxy protocol: %s", proxyURL.Scheme)
		}
	}

	var redirectFunc = func(_ *http.Request, _ []*http.Request) error {
		// Tell the http client to not follow redirect
		return http.ErrUseLastResponse
	}

	var options retryablehttp.Options
	if single {
		options = retryablehttp.DefaultOptionsSingle
	} else {
		options = retryablehttp.DefaultOptionsSpraying
	}

	options.HttpClient.Transport = transport
	options.HttpClient.CheckRedirect = redirectFunc

	httpx.Client = retryablehttp.NewClient(options)

	return httpx
}

func (h *HTTPX) Get(target string) (*Response, error) {
	req, err := retryablehttp.NewRequest("GET", target, nil)

	if err != nil {
		logging.Logger.Errorln(err)
		return nil, err
	}

	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := h.Client.Do(req)
	if err != nil {
		logging.Logger.Errorln(err)
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("resp == nil")
	}

	body, _ := io.ReadAll(resp.Body)

	requestDump, err := req.Dump()
	if err != nil {
		logging.Logger.Errorln(err)
		return nil, err
	}

	responseDump, _ := httputil.DumpResponse(resp, true)

	if err != nil {
		logging.Logger.Errorln(err)
		return nil, err
	}

	contentLength := int(resp.ContentLength)

	if contentLength == -1 {
		contentLength = len(string(body))
	}

	return &Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Body:             string(body),
		RequestDump:      string(requestDump),
		ResponseDump:     string(responseDump),
		ContentLength:    contentLength,
		Header:           nil,
		RequestUrl:       "",
		Location:         "",
		ServerDurationMs: 0,
	}, nil
}

func (h *HTTPX) Close() {
	h.Dialer.Close()
}

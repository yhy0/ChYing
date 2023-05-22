package burpSuite

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	urlutil "github.com/projectdiscovery/utils/url"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/tools/burpSuite/mitmproxy/proxy"
	"github.com/yhy0/logging"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

/**
  @author: yhy
  @since: 2022/9/26
  @desc: burpSuite 的实现 ，组装前端数据的插件

	执行顺序 Requestheaders > Request	> Responseheaders > Response

	Requestheaders 运行后会阻塞(<-f.Done())，所以需要协程 ；然后运行 Request ，Responseheaders，等 Response 运行完后才会组装数据

**/

type Burp struct {
	proxy.BaseAddon
	done chan bool
}

var Ctx context.Context
var Sum int // 每次 Wg.Add(1) 加一, 每次 Wg.Done() 减一
var InterceptBody string
var Done = make(chan bool)

func (b *Burp) Requestheaders(f *proxy.Flow) {
	// 组装数据都在这里操作了
	go func() {
		//  f.finish() 执行之后也就是 proxy.go 中的 ServeHTTP() return 后才会往下执行
		<-f.Done()
		var (
			params     string
			cookies    string
			statusCode int
			contentLen int
			remoteAddr string
		)

		if f.Response != nil {
			statusCode = f.Response.StatusCode
			if f.Response.Body != nil {
				contentLen = len(f.Response.Body)
			}

			if len(f.Request.URL.Query()) > 0 {
				params = "√"
			}

			if f.Response.Header != nil && f.Response.Header.Get("Set-Cookie") != "" {
				cookies = f.Response.Header.Get("Set-Cookie")
			}
		}

		if f.ConnContext.ServerConn.Conn != nil {
			remoteAddr = f.ConnContext.ServerConn.Conn.RemoteAddr().String()
		} else {
			remoteAddr = ""
		}

		HttpHistory <- HTTPHistory{
			Id:        f.Id,
			Host:      f.Request.URL.Host,
			Method:    f.Request.Method,
			URL:       f.Request.URL.RequestURI(),
			Params:    params,
			Edited:    "",
			Status:    strconv.Itoa(statusCode),
			Length:    strconv.Itoa(contentLen),
			MIMEType:  "",
			Extension: "",
			Title:     "",
			Comment:   "",
			TLS:       "√",
			IP:        remoteAddr,
			Cookies:   cookies,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}

		buf := bytes.NewBuffer(make([]byte, 0))
		fmt.Fprintf(buf, "%s %s %s\r\n", f.Request.Method, f.Request.URL.RequestURI(), f.Request.Proto)
		fmt.Fprintf(buf, "Host: %s\r\n", f.Request.URL.Host)
		if len(f.Request.Raw().TransferEncoding) > 0 {
			fmt.Fprintf(buf, "Transfer-Encoding: %s\r\n", strings.Join(f.Request.Raw().TransferEncoding, ","))
		}
		if f.Request.Raw().Close {
			fmt.Fprintf(buf, "Connection: close\r\n")
		}

		err := f.Request.Header.WriteSubset(buf, nil)
		if err != nil {
			logging.Logger.Error(err)
		}
		buf.WriteString("\r\n")

		if f.Request.Body != nil && len(f.Request.Body) > 0 && canPrint(f.Request.Body) {
			buf.Write(f.Request.Body)
			buf.WriteString("\r\n\r\n")
		}

		f.Request.RawStr = buf.String()

		requestDump := buf.String()
		buf = bytes.NewBuffer(make([]byte, 0))
		if f.Response != nil {
			fmt.Fprintf(buf, "%v %v %v\r\n", f.Request.Proto, f.Response.StatusCode, http.StatusText(f.Response.StatusCode))
			err := f.Response.Header.WriteSubset(buf, nil)
			if err != nil {
				logging.Logger.Error(err)
			}
			buf.WriteString("\r\n")
			if f.Response.Body != nil && len(f.Response.Body) > 0 {
				body, err := f.Response.DecodedBody()
				if err == nil && body != nil && len(body) > 0 {
					buf.Write(body)
					buf.WriteString("\r\n\r\n")
				}
			}
		}
		responseDump := buf.String()

		HTTPBodyMap.WriteMap(f.Id, &HTTPBody{
			TargetUrl: f.Request.URL.String(),
			Request:   requestDump,
			Response:  responseDump,
		})
	}()
}

// Request 这里可以拦截请求
func (b *Burp) Request(f *proxy.Flow) {
	if Intercept {
		buf := bytes.NewBuffer(make([]byte, 0))
		fmt.Fprintf(buf, "%s %s %s\r\n", f.Request.Method, f.Request.URL.RequestURI(), f.Request.Proto)
		fmt.Fprintf(buf, "Host: %s\r\n", f.Request.URL.Host)
		if len(f.Request.Raw().TransferEncoding) > 0 {
			fmt.Fprintf(buf, "Transfer-Encoding: %s\r\n", strings.Join(f.Request.Raw().TransferEncoding, ","))
		}
		if f.Request.Raw().Close {
			fmt.Fprintf(buf, "Connection: close\r\n")
		}

		err := f.Request.Header.WriteSubset(buf, nil)
		if err != nil {
			logging.Logger.Error(err)
		}
		buf.WriteString("\r\n")

		if f.Request.Body != nil && len(f.Request.Body) > 0 && canPrint(f.Request.Body) {
			buf.Write(f.Request.Body)
			buf.WriteString("\r\n\r\n")
		}

		f.Request.RawStr = buf.String()

		requestDump := buf.String()

		HttpBodyInter = &HTTPBody{
			TargetUrl: f.Request.URL.String(),
			Request:   requestDump,
			Response:  "",
		}

		runtime.EventsEmit(Ctx, "InterceptBody", requestDump)
		Sum += 1
		Done <- true

		// 先备份一份 request
		temp := f.Request
		// 点击 forward 后，根据输入框的值组装数据
		target, err := urlutil.ParseURL(f.Request.URL.String(), true)

		if err != nil {
			logging.Logger.Errorln(f.Request.URL.String(), err)
			return
		} else {
			rawRequestData, err := httpx.Parse(InterceptBody, target, true)

			if err != nil {
				logging.Logger.Errorln(err)
				return
			}
			inputUrl, err := urlutil.ParseURL(rawRequestData.FullURL, true)

			f.Request.URL = inputUrl.URL
			f.Request.Method = rawRequestData.Method
			for k, v := range rawRequestData.Headers {
				f.Request.Header[k] = []string{v}
			}
			f.Request.Body = []byte(rawRequestData.Data)

			req, err := http.NewRequest(f.Request.Method, f.Request.URL.String(), strings.NewReader(rawRequestData.Data))
			if err != nil {
				logging.Logger.Errorln(err)
				// 出错了，还原 Request
				f.Request = temp
				return
			}

			for k, v := range f.Request.Header {
				req.Header[k] = v
			}

			f.Request.HttpRaw = req
		}

	}
}

func (b *Burp) Responseheaders(f *proxy.Flow) {
}

func (b *Burp) Response(f *proxy.Flow) {
	if Intercept {
		for {
			if Sum != 0 {
				time.Sleep(500)
			} else {
				break
			}
		}
		buf := bytes.NewBuffer(make([]byte, 0))
		if f.Response != nil {
			fmt.Fprintf(buf, "%v %v %v\r\n", f.Request.Proto, f.Response.StatusCode, http.StatusText(f.Response.StatusCode))
			err := f.Response.Header.WriteSubset(buf, nil)
			if err != nil {
				logging.Logger.Error(err)
			}
			buf.WriteString("\r\n")
			if f.Response.Body != nil && len(f.Response.Body) > 0 {
				body, err := f.Response.DecodedBody()
				if err == nil && body != nil && len(body) > 0 {
					buf.Write(body)
					buf.WriteString("\r\n\r\n")
				}
			}
		}
		responseDump := buf.String()

		runtime.EventsEmit(Ctx, "InterceptBody", responseDump)
		Sum += 1
		Done <- true

		resp := formatResponseDump(InterceptBody, f.Request.HttpRaw)

		//f.Response = resp
		f.Response.StatusCode = resp.StatusCode
		f.Response.Header = resp.Header
		f.Response.DecodedBodyStr = resp.Body
		f.Response.Body = resp.Body

		f.Response.BodyReader = resp.BodyReader
		f.Response.Raw = resp.Raw
		f.Response.Close = resp.Close

		f.Response.ReplaceToDecodedBody()

	}
}

func canPrint(content []byte) bool {
	for _, c := range string(content) {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func formatResponseDump(dump string, req *http.Request) *proxy.Response {
	// 初始化 HTTP 响应对象

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBufferString(dump)), req)
	if err != nil {
		logging.Logger.Errorln("http.ReadResponse() failed", err)
		return nil
	}

	var body []byte
	body, _ = ioutil.ReadAll(resp.Body)

	// 解析响应体中的状态码、响应头部和响应正文等信息
	var r proxy.Response
	r.StatusCode = resp.StatusCode
	r.Header = resp.Header
	r.Body = body
	r.BodyReader = resp.Body
	r.Raw = resp
	r.Close = resp.Close

	return &r
}

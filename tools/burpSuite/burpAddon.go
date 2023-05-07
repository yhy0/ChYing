package burpSuite

import (
	"bytes"
	"fmt"
	"github.com/yhy0/ChYing/tools/burpSuite/mitmproxy/proxy"
	"github.com/yhy0/logging"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

/**
  @author: yhy
  @since: 2022/9/26
  @desc: 组装前端数据的插件
**/

type Burp struct {
	proxy.BaseAddon
	done chan bool
}

func (b *Burp) Requestheaders(f *proxy.Flow) {
	go func() {
		//  f.finish() 执行之后也就是 proxy.go 中的 ServeHTTP() return 后才会往下执行
		<-f.Done()
		var (
			params     string
			cookies    string
			statusCode int
			contentLen int
			remoteAddr string
			//responseDump []byte
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

}

func (b *Burp) Responseheaders(f *proxy.Flow) {

}
func (b *Burp) Response(f *proxy.Flow) {

}

func canPrint(content []byte) bool {
	for _, c := range string(content) {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

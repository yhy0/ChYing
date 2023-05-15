package test

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yhy0/ChYing/pkg/httpx"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
)

/**
  @author: yhy
  @since: 2023/5/6
  @desc: //TODO
**/

func TestRaw(t *testing.T) {
	raw := `GET /guestbook.php HTTP/1.1
Host: testphp.vulnweb.com
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8
Accept-Encoding: gzip, deflate
Accept-Language: zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2
Connection: keep-alive
Referer: http://testphp.vulnweb.com/cart.php
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/111.0

`

	//conf.Proxy = "http://127.0.0.1:8080"
	httpx.NewSession()

	resp, err := httpx.Raw(raw, "http://testphp.vulnweb.com/index.php")
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(resp.ResponseDump)
}

func TestResp(t *testing.T) {
	resp := `HTTP/1.1 200 OK
Server: nginx/1.19.0
Date: Mon, 15 May 2023 12:33:21 GMT
Content-Type: text/html; charset=UTF-8
Connection: close
X-Powered-By: PHP/5.6.40-38+ubuntu20.04.1+deb.sury.org+1
Content-Length: 4236

<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1" />
<title>ajax test</title>
<link href="styles.css" rel="stylesheet" type="text/css" />
<script type="text/javascript">
	var httpreq = null;`

	var Mutex sync.Mutex

	go func() {
		fmt.Println("-------------")
		Mutex.Lock()
	}()

	fmt.Println("111")
	fmt.Println(formatResponseDump(resp))
}

// flow http response
type Response struct {
	StatusCode int         `json:"statusCode"`
	Header     http.Header `json:"header"`
	Body       []byte      `json:"-"`
	BodyReader io.Reader
	Raw        *http.Response
	close      bool // connection close

	decodedBody []byte
	decoded     bool // decoded reports whether the response was sent compressed but was decoded to decodedBody.
	decodedErr  error
}

func formatResponseDump(dump string) (*Response, error) {
	// 初始化 HTTP 响应对象

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBufferString(dump)), nil)
	if err != nil {
		return nil, errors.Wrap(err, "http.ReadResponse() failed")
	}

	var body []byte
	body, _ = ioutil.ReadAll(resp.Body)

	// 解析响应体中的状态码、响应头部和响应正文等信息
	var r Response
	r.StatusCode = resp.StatusCode
	r.Header = resp.Header
	r.Body = body
	r.BodyReader = resp.Body
	r.Raw = resp
	r.close = resp.Close

	return &r, nil
}

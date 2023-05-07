package test

import (
	"github.com/yhy0/ChYing/pkg/httpx"
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

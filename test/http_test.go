package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yhy0/ChYing/pkg/Jie/conf"

	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2024/9/30
  @desc: //TODO
**/

func TestRaw(t *testing.T) {
	logging.Logger = logging.New(true, file.ChyingDir, "ChYing", true)
	conf.GlobalConfig = &conf.Config{}
	conf.GlobalConfig.Http.Proxy = ""
	resp, err := httpx.Raw(`GET /default_config.json HTTP/1.1
Host: config.immersivetranslate.com
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:139.0) Gecko/20100101 Firefox/139.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Encoding: gzip, deflate
Accept-Language: zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2
Cache-Control: no-cache
Connection: close
Pragma: no-cache
Priority: u=0, i
Upgrade-Insecure-Requests: 1

`, "https://config.immersivetranslate.com/")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp)
}

func Test(t *testing.T) {
	// parse, err := url.Parse("http://detectportal.firefox.com/canonical.html?asd=123&asd=123")
	// if err != nil {
	// 	return
	// }
	//
	// fmt.Println("RawPath", parse.RawPath)
	//
	// fmt.Println("Path", parse.Path)
	//
	// fullPath := parse.Path
	// if parse.RawQuery != "" {
	// 	fullPath += "?" + parse.RawQuery
	// }
	//
	// fmt.Println("fullPath", fullPath)
	parts := strings.SplitN("User-Agent: Mozilla/5.0 (ChYing-Inside Security Scanner)", ": ", 2)
	fmt.Println(parts)

}

// TestRawWithEmptyHost 测试Host头为空的情况
func TestRawWithEmptyHost(t *testing.T) {
	logging.Logger = logging.New(true, file.ChyingDir, "ChYing", true)
	conf.GlobalConfig = &conf.Config{}
	conf.GlobalConfig.Http.Proxy = ""

	// 测试Host头为空的请求
	request := `GET /etc/passwd HTTP/1.1
Host: 
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8
Accept-Encoding: gzip, deflate
Accept-Language: zh-CN,zh;q=0.9
Connection: close
Upgrade-Insecure-Requests: 1

`

	target := "http://43.154.136.67:8080"

	fmt.Printf("=== 测试Host头为空的情况 ===\n")
	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Request:\n%s\n", request)

	resp, err := httpx.Raw(request, target)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Time: %.2f ms\n", resp.ServerDurationMs)
	fmt.Printf("Response Length: %d\n", len(resp.ResponseDump))

	// 期望的结果是能够成功发送请求，Host头保持为空
	if resp.StatusCode == 0 {
		t.Errorf("期望得到有效的状态码，但得到了: %d", resp.StatusCode)
	}
}

package crlf

import (
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
	"sync"
	"time"
)

/*
   一般一个子域名对应同一个或者一类中间件
   但是有反代的情况下，不同路径可能反代到不同中间件
   所以 crlf 这里定义为每个路径扫描一次
*/

import (
	"github.com/yhy0/logging"
)

const (
	RegexRule = `(?i)[\n|\r](Somecustominjectedheader\s*:\s*injected_by_wvs)`
)

// crlf Check
var payloadTemplate = []string{
	`/%0ASomecustominjectedheader: injected_by_wvs`,
	`\r\nSomeCustomInjectedHeader: injected_by_wvs`,
	`\r\n\tSomeCustomInjectedHeader: injected_by_wvs`,
	`\r\n SomeCustomInjectedHeader: injected_by_wvs`,
	`\r\tSomeCustomInjectedHeader: injected_by_wvs`,
	`\nSomeCustomInjectedHeader: injected_by_wvs`,
	`\rSomeCustomInjectedHeader: injected_by_wvs`,
	`\rSomeCustomInjectedHeader: injected_by_wvs`,
	`%E5%98%8A%E5%98%8DSomeCustomInjectedHeader:%20injected_by_wvs`,
	`%c4%8d%c4%8aSomeCustomInjectedHeader:%20injected_by_wvs`,
}

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Name() string {
	return "crlf"
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if p.IsScanned(in.UniqueId) {
		return
	}
	r, _ := regexp.Compile(RegexRule)
	for _, pl := range payloadTemplate {
		if strings.ToUpper(in.Method) == "GET" {
			npl := target + pl
			res, err := client.Request(npl, in.Method, "", in.Headers, "CRLF")

			if err != nil {
				logging.Logger.Debugf("Request error: %v", err)
				return
			}

			r, err := regexp.Compile(RegexRule)
			if err != nil {
				logging.Logger.Debugf("%s", err.Error())
				return
			}

			C := r.FindAllStringSubmatch(httpx.Header(res.RespHeader), -1)
			if len(C) != 0 {
				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "CRLF",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     target,
						Method:     in.Method,
						Ip:         in.Ip,
						Param:      "",
						Request:    res.RequestDump,
						Response:   res.ResponseDump,
						Payload:    npl,
					},
					Level: output.Medium,
				}
				return
			}

		} else {
			res, err := client.Request(target, in.Method, in.Resp.Body+pl, in.Headers, "CRLF")
			if err != nil {
				logging.Logger.Debugf("Request error: %v", err)
				return
			}
			if str := r.FindString(httpx.Header(res.RespHeader)); str != "" {
				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "CRLF",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     target,
						Method:     in.Method,
						Ip:         in.Ip,
						Param:      "",
						Request:    res.RequestDump,
						Response:   res.ResponseDump,
						Payload:    in.Resp.Body + pl,
					},
					Level: output.Medium,
				}
				return
			}
		}
	}
}

func (p *Plugin) IsScanned(key string) bool {
	if key == "" {
		return false
	}
	if _, ok := p.SeenRequests.Load(key); ok {
		return true
	}
	p.SeenRequests.Store(key, true)
	return false
}

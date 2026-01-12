package bbscan

import (
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antlabs/strsim"
	"github.com/panjf2000/ants/v2"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/swagger"
	"github.com/yhy0/ChYing/pkg/ui"
	"github.com/yhy0/logging"
)

/**
    @author: yhy
    @since: 2022/9/17
    @desc: //TODO
**/

type Page struct {
	isBackUpPage bool
	title        string
	locationUrl  string
}

type Result struct {
	Target       string
	Payload      string
	RequestDump  string
	ResponseDump string
	RedirectUrl  string
}

var (
	path404 = "/file_not_support"
)

func ReqPage(u string, header map[string]string, client *httpx.Client) (*Page, *httpx.Response, error) {
	page := &Page{}
	var backUpSuffixList = []string{".tar", ".tar.gz", ".zip", ".rar", ".7z", ".bz2", ".gz", ".war"}
	var method = "GET"

	for _, ext := range backUpSuffixList {
		if strings.HasSuffix(u, ext) {
			method = "HEAD"
		}
	}

	res, err := client.Request(u, method, "", header, "BBscan")
	if err != nil {
		return nil, nil, err
	}

	page.title = util.GetTitle(res.Body)
	page.locationUrl = res.Location
	if res.StatusCode != 302 && res.Location == "" {
		regs := []string{"application/.*download", "application/.*file", "application/.*zip", "application/.*rar", "application/.*tar", "application/.*down", "application/.*compressed", "application/.*stream"}
		for _, reg := range regs {
			matched, _ := regexp.Match(reg, []byte(res.RespHeader.Get("Content-Type")))
			if matched {
				page.isBackUpPage = true
				break
			}
		}
	}
	return page, res, err
}

// BBscan u: 目标 root: 扫描路径是否为主目录
func BBscan(u string, root bool, fingerPrints []string, header map[string]string, client *httpx.Client, fuzzUiResp chan ui.Result) []string {
	if strings.HasSuffix(u, "/") {
		u = u[:len(u)-1]
	}
	var (
		technologies []string
		resContents  []string // 找到的页面返回集合，用来进行网页相似度比较，用来去除大量的返回一样的
	)

	result := make(map[string]*Result)
	_, url404res, err := ReqPage(u+path404, header, client)

	if err == nil {
		if url404res.StatusCode == 404 {
			technologies = addFingerprints404(technologies, url404res, client) // 基于404页面文件扫描指纹添加
		}
		resContents = append(resContents, strings.ReplaceAll(url404res.Body, path404, ""))

		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:           u,
				Method:        "GET",
				StatusCode:    url404res.StatusCode,
				ContentLength: url404res.ContentLength,
				Request:       url404res.RequestDump,
				Response:      url404res.ResponseDump,
			}
		}
	} else {
		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:    u,
				Method: "GET",
			}
		}
	}

	fingerPrints = util.RemoveDuplicateElement(append(fingerPrints, technologies...))

	pool, _ := ants.NewPool(20)
	defer pool.Release() // 释放协程池

	wg := sync.WaitGroup{}
	var lock sync.Mutex
	count := 0

	logging.Logger.Debugln("BBscan start:", u)
	for _path, _rule := range file.BBscanRules {
		rule := _rule
		path := _path
		// 状态码 500 以上 30 次就不扫描了
		if count > 30 {
			logging.Logger.Debugln("BBscan end, max error:", count)
			return technologies
		}

		isFingerPrints := false
		// 该规则存在指纹要求，看看有没有传入指纹
		if len(rule.FingerPrints) > 0 {
			// 已经识别到了指纹，看看是否符合指纹要求
			if len(fingerPrints) > 0 {
				// 指纹可能效果并不是很好，所以如果指纹中没有这些的话，那就全部扫描，TODO 后期指纹模块优化的很好的话，这里可以去除
				if !util.InSliceCaseFold("php", fingerPrints) && !util.InSliceCaseFold("java", fingerPrints) && !util.InSliceCaseFold("spring", fingerPrints) {
					// 该网站的指纹，没有识别出来是 php、java、spring 中的任何一个，走扫描，防止漏
					isFingerPrints = true
				} else {
					for _, fp := range fingerPrints {
						// 指纹符合，进行扫描
						if util.InSliceCaseFold(fp, rule.FingerPrints) {
							isFingerPrints = true
							break
						}
					}
				}
				// 走到这，说明识别到的指纹不匹配这个规则，不扫描这个 bbscan 规则了
			}
		} else {
			// 没有指纹要求，直接走扫描
			isFingerPrints = true
		}

		// 有指纹，但没有匹配则该规则不扫描
		if !isFingerPrints {
			continue
		}

		if util.Contains(path, "{sub}") {
			t, err := url.Parse(u)
			if err != nil {
				logging.Logger.Errorln(err)
				continue
			}
			path = strings.ReplaceAll(path, "{sub}", t.Hostname())
		}
		path = strings.TrimLeft(path, "/")

		if rule.Root && !root { // 该路径规则只会出现在主目录下, 并且 传入的是主目录，则加上，否则不进行该规则的扫描
			continue
		}

		wg.Add(1)

		_ = pool.Submit(func(_pU string, _pPath string) func() {
			return func() {
				defer wg.Done()
				// 根据传入的路径进行拼接扫描目录
				target := _pU + "/" + _pPath
				page, res, err := ReqPage(target, header, client)
				if err == nil && res != nil {
					if res.StatusCode >= 500 {
						lock.Lock()
						count += 1
						lock.Unlock()
						return
					}

					// 黑名单，跳过
					if util.IsBlackHtml(res.Body, res.RespHeader["Content-Type"], path) {
						return
					}

					// ContentLength 为 0 的，都丢弃
					if res.ContentLength == 0 {
						return
					}

					contentType := res.RespHeader.Get("Content-Type")
					// 返回是个图片
					if util.Contains(contentType, "image/") {
						return
					}

					if strings.HasSuffix(target, ".xml") {
						if !util.Contains(contentType, "xml") {
							return
						}
					} else if strings.HasSuffix(target, ".json") {
						if !util.Contains(contentType, "json") {
							return
						}
					}

					// 规则匹配
					if !page.isBackUpPage {
						if len(strings.TrimSpace(res.Body)) == 0 {
							return
						}
						if (rule.Type != "" && !util.Contains(contentType, rule.Type)) || (rule.TypeNo != "" && util.Contains(contentType, rule.TypeNo)) {
							return
						}
						if rule.Status != "" && strconv.Itoa(res.StatusCode) != rule.Status {
							return
						}
					} else {
						// 压缩包的单独搞，规则不太对
						if res.StatusCode < 200 || res.StatusCode > 300 || res.ContentLength < 10 {
							return
						}
						// 压缩包的太小的都丢弃，有的 waf 会根据请求构造压缩包
						if res.ContentLength < 10 {
							return
						}
					}

					if rule.Tag != "" && !util.Contains(res.Body, rule.Tag) {
						return
					}

					similar := false
					if len(res.Body) != 0 {
						// 与成功的进行相似度比较，排除一些重复项 比如一个目标返回很多这种，写入黑名单的话，会有很多，所以先这样去除 {"code":99999,"msg":"未知错误","status":0}
						for _, body := range resContents {
							similar = strsim.Compare(body, res.Body) > 0.9 // 不相似才会往下执行
						}
					}

					if !similar {
						// 对扫到的 swagger 进行自动化测试
						if strings.Contains(target, "swagger") {
							swagger.Scan(target, client, fuzzUiResp)
						}
						lock.Lock()

						if res.StatusCode == 401 {
							technologies = append(technologies, "Basic")
						}
						technologies = append(addFingerprintsNormal(target, technologies, res, client)) // 基于200页面文件扫描指纹添加
						resContents = append(resContents, strings.ReplaceAll(res.Body, target, ""))
						fingerPrints = util.RemoveDuplicateElement(append(fingerPrints, technologies...))

						result[path] = &Result{
							Target:       _pU,
							Payload:      _pPath,
							RequestDump:  res.RequestDump,
							ResponseDump: res.ResponseDump,
							RedirectUrl:  res.Location,
						}

						lock.Unlock()

						if fuzzUiResp != nil {
							fuzzUiResp <- ui.Result{
								Url:           target,
								Method:        "GET",
								StatusCode:    res.StatusCode,
								ContentLength: res.ContentLength,
								Request:       res.RequestDump,
								Response:      res.ResponseDump,
							}
						}
					}
				}
			}
		}(u, path))

	}

	wg.Wait()

	var redirectUrl = make(map[string]int)
	for _, v := range result {
		if v.RedirectUrl != "" {
			if _, ok := redirectUrl[v.RedirectUrl]; ok {
				redirectUrl[v.RedirectUrl] += 1
			} else {
				redirectUrl[v.RedirectUrl] = 1
			}
		}
	}

	// RedirectUrl 去除相同 跳转的，认为是同一个
	for _, v := range result {
		if v.RedirectUrl != "" && redirectUrl[v.RedirectUrl] > 1 {
			continue
		}
		output.OutChannel <- output.VulMessage{
			DataType: "web_vul",
			Plugin:   "BBscan",
			VulnData: output.VulnData{
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				Target:     u,
				Payload:    v.Payload,
				Method:     "GET",
				Request:    v.RequestDump,
				Response:   v.ResponseDump,
			},
			Level: output.Low,
		}
	}

	return technologies
}

func SingleScan(targets []string, path string) {
	rule := file.BBscanRules[path]

	wg := sync.WaitGroup{}
	ch := make(chan struct{}, 50)
	for _, target := range targets {
		if util.Contains(path, "{sub}") {
			t, _ := url.Parse(target)
			path = strings.ReplaceAll(path, "{sub}", t.Hostname())
		}

		wg.Add(1)
		ch <- struct{}{}
		go func(u string) {
			defer wg.Done()
			defer func() { <-ch }()
			res, err := httpx.Request(u+path, "GET", "", nil, "BBscan")

			if err != nil {
				return
			}
			// 黑名单，跳过
			if util.IsBlackHtml(res.Body, res.RespHeader["Content-Type"], path) {
				return
			}

			contentType := res.RespHeader.Get("Content-Type")
			// 返回是个图片
			if util.Contains(contentType, "image/") {
				return
			}

			if strings.HasSuffix(path, ".xml") {
				if !util.Contains(contentType, "xml") {
					return
				}
			} else if strings.HasSuffix(path, ".json") {
				if !util.Contains(contentType, "json") {
					return
				}
			}

			// 返回包是个下载文件，但文件内容为空丢弃
			// if res.Header.Get("Content-Type") == "application/octet-stream" && res.ContentLength == 0 {
			//    return
			// }

			// 规则匹配
			if (rule.Type != "" && !util.Contains(contentType, rule.Type)) || (rule.TypeNo != "" && util.Contains(contentType, rule.TypeNo)) {
				return
			}
			if rule.Status != "" && strconv.Itoa(res.StatusCode) != rule.Status {
				return
			}

			if rule.Tag != "" && !util.Contains(res.Body, rule.Tag) {
				return
			}
			// swagger 自动化测试
			if strings.Contains(path, "swagger") {
				swagger.Scan(u+path, httpx.NewClient(nil), nil)
			}

			output.OutChannel <- output.VulMessage{
				DataType: "web_vul",
				Plugin:   "BBscan",
				VulnData: output.VulnData{
					CreateTime: time.Now().Format("2006-01-02 15:04:05"),
					Target:     u,
					Ip:         "",
					Payload:    u + path,
					Method:     "GET",
					Request:    res.RequestDump,
					Response:   res.ResponseDump,
				},
				Level: output.Low,
			}
		}(target)
	}
	wg.Wait()
}

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Scan(target, path string, in *input.CrawlResult, client *httpx.Client) {
	technologies := BBscan(target, path == "/", in.Fingerprints, in.Headers, client, nil)
	in.Fingerprints = util.RemoveDuplicateElement(append(in.Fingerprints, technologies...))
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
	return "bbscan"
}

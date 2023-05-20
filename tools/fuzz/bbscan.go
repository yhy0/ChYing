package fuzz

import (
	"fmt"
	"github.com/antlabs/strsim"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/tools"
	"github.com/yhy0/Jie/conf"
	"github.com/yhy0/Jie/pkg/util"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

/**
	@author: yhy
	@since: 2022/9/1
	@desc: //TODO
**/

type Page struct {
	isBackUpPath bool
	isBackUpPage bool
	title        string
	locationUrl  string
	is302        bool
	is403        bool
}

var (
	path404 = "/file_not_support"
)

func getTitle(body string) string {
	titleReg := regexp.MustCompile(`<title>([\s\S]{1,200})</title>`)
	title := titleReg.FindStringSubmatch(body)
	if len(title) > 1 {
		return title[1]
	}
	return ""
}

func ReqPage(u string) (*Page, *httpx.Response, error) {
	page := &Page{}
	var backUpSuffixList = []string{".tar", ".tar.gz", ".zip", ".rar", ".7z", ".bz2", ".gz", ".war"}
	var m = "GET"

	for _, ext := range backUpSuffixList {
		if strings.HasSuffix(u, ext) {
			page.isBackUpPath = true
			m = "HEAD"
		}
	}

	if res, err := httpx.Request(u, m, "", false, conf.DefaultHeader); err == nil {
		if util.IntInSlice(res.StatusCode, []int{301, 302, 307, 308}) {
			page.is302 = true
		}
		page.title = getTitle(res.Body)
		page.locationUrl = res.Location
		regs := []string{"text/plain", "application/.*download", "application/.*file", "application/.*zip", "application/.*rar", "application/.*tar", "application/.*down", "application/.*compressed", "application/stream"}
		for _, reg := range regs {
			matched, _ := regexp.Match(reg, []byte(res.Header.Get("Content-Type")))
			if matched {
				page.isBackUpPage = true
			}
		}
		if (res.StatusCode == 403 && strings.HasSuffix(u, "/")) || util.In(res.Body, conf.Page403Content) {
			page.is403 = true
		}
		return page, res, err
	} else {
		return page, nil, err
	}
}

// BBscan todo 还应该传进来爬虫找到的 api 目录
func BBscan(u string, indexContentLength int, indexbody string, _403 bool) {
	if strings.HasSuffix(u, "/") {
		u = u[:len(u)-1]
	}

	var (
		payloadlocation404   []string
		payload200Title      []string
		payload200Contentlen []int
		skip302              = false
		other200Contentlen   []int
		other200Title        []string
		url404               *Page
		url404res            *httpx.Response
		err                  error
	)

	other200Contentlen = append(other200Contentlen, indexContentLength)
	other200Title = append(other200Title, getTitle(indexbody))
	if url404, url404res, err = ReqPage(u + path404); err == nil {
		if url404.is302 {
			conf.Location404 = append(conf.Location404, url404.locationUrl)
		}
		if url404.is302 && strings.HasSuffix(url404.locationUrl, "/file_not_support/") {
			skip302 = true
		}

		if url404res.StatusCode == 200 {
			other200Title = append(other200Title, url404.title)
			other200Contentlen = append(other200Contentlen, url404res.ContentLength)
		}
	}
	wg := sync.WaitGroup{}
	ch := make(chan struct{}, 20)

	n := len(file.BBscanRules)
	i := 0
	for path, rule := range file.BBscanRules {
		i += 1
		if Stop {
			return
		}
		var is404Page = false

		if util.Contains(path, "{sub}") {
			t, _ := url.Parse(u)
			path = strings.ReplaceAll(path, "{sub}", t.Hostname())
		}

		ch <- struct{}{}
		wg.Add(1)
		go func(i int, path string, rule *file.Rule) {
			defer wg.Done()
			target, res, err := ReqPage(u + path)
			FuzzPercentage <- fmt.Sprintf("%.2f", float64(i)/float64(n)*100)
			if err == nil && res != nil {

				if util.In(res.Body, conf.WafContent) {
					<-ch
					return
				}

				contentType := res.Header.Get("Content-Type")
				// 返回是个图片
				if util.Contains(contentType, "image/") {
					<-ch
					return
				}

				if strings.HasSuffix(path, ".xml") {
					if !util.Contains(contentType, "xml") {
						<-ch
						return
					}
				} else if strings.HasSuffix(path, ".json") {
					if !util.Contains(contentType, "json") {
						<-ch
						return
					}
				}

				// 文件内容为空丢弃
				if res.ContentLength == 0 {
					<-ch
					return
				}

				//// 返回包是个下载文件，但文件内容为空丢弃
				//if res.Header.Get("Content-Type") == "application/octet-stream" && res.ContentLength == 0 {
				//	<-ch
				//	return
				//}

				// 规则匹配
				if (rule.Type != "" && !util.Contains(contentType, rule.Type)) || (rule.TypeNo != "" && util.Contains(contentType, rule.TypeNo)) {
					<-ch
					return
				}

				if rule.Status != "" && strconv.Itoa(res.StatusCode) != rule.Status {
					<-ch
					return
				}

				if rule.Tag != "" && !util.Contains(res.Body, rule.Tag) {
					<-ch
					return
				}

				if target.isBackUpPath {
					if !target.isBackUpPage {
						is404Page = true
					}
				}
				if util.In(target.title, conf.Page404Title) {
					is404Page = true
				}
				if util.In(res.Body, conf.Page404Content) {
					is404Page = true
				}
				if strings.Contains(res.RequestUrl, "/.") && res.StatusCode == 200 {
					if res.ContentLength == 0 {
						is404Page = true
					}
				}
				if target.is302 {
					if skip302 {
						is404Page = true
					}
					if util.In(res.Location, conf.Location404) && util.In(res.Location, payloadlocation404) {
						is404Page = true
					}
					if !strings.HasSuffix(res.Location, path+"/") {
						conf.Location404 = append(payloadlocation404, res.Location)
						is404Page = true
					}
				}

				if !is404Page {
					for _, title := range other200Title {
						if len(target.title) > 2 && target.title == title {
							is404Page = true
						}
					}
					for _, title := range payload200Title {
						if len(target.title) > 2 && target.title == title {
							is404Page = true
						}
					}
					for _, l := range other200Contentlen {
						reqlenabs := res.ContentLength - l
						if reqlenabs < 0 {
							reqlenabs = -reqlenabs
						}
						if reqlenabs <= 5 {
							is404Page = true
						}
					}
					for _, l := range payload200Contentlen {
						reqlenabs := res.ContentLength - l
						if reqlenabs < 0 {
							reqlenabs = -reqlenabs
						}
						if reqlenabs <= 5 {
							is404Page = true
						}
					}
					payload200Title = append(payload200Title, target.title)
					payload200Contentlen = append(payload200Contentlen, res.ContentLength)

					// 规则匹配完后，再次比较与 file_not_support 页面返回值的相似度
					similar := true
					if len(res.Body) != 0 && url404res != nil && len(url404res.Body) != 0 {
						similar = strsim.Compare(strings.ReplaceAll(url404res.Body, "/file_not_support", ""), strings.ReplaceAll(res.Body, path, "")) <= 0.9 // 不相似才会往下执行
					}

					if _403 && res.StatusCode == 403 {
						go Bypass403(u+path, "GET")
					}

					if similar && res.StatusCode != 404 && res.StatusCode != 403 && res.StatusCode != 301 && res.StatusCode != 302 && res.StatusCode != 304 && !target.is403 {
						FuzzChan <- tools.Result{
							Url:           u + path,
							Method:        "GET",
							StatusCode:    res.StatusCode,
							ContentLength: res.ContentLength,
							Request:       res.RequestDump,
							Response:      res.ResponseDump,
						}
					}
				}

			}

			<-time.After(time.Duration(500) * time.Millisecond)
			<-ch
		}(i, path, rule)
	}

	wg.Wait()
	close(ch)
}

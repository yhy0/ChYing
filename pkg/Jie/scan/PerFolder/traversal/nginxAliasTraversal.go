package traversal

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
	"net/url"
	"strings"
	"sync"
	"time"
)

/**
   @author yhy
   @since 2023/8/15
   @desc nginx alias traversal https://github.com/hakaioffsec/navgix
**/

// Check for alias traversal vulnerability (bruteforce) 默认使用的字典，每层目录都会加上，暴力跑一遍。 TODO 这种方式是否去除？只对传入的目录进行测试
var dictionary = []string{
	"file",
	"files",
	"static",
	"js",
	"images",
	"img",
	"css",
	"assets",
	"media",
	"lib",
}

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if p.IsScanned(in.UniqueId) {
		return
	}
	NginxAlias(target, in.Resp.Body, path)
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
	return "nginx-alias-traversal"
}

func NginxAlias(url string, body string, path string) {
	if url[len(url)-1:] != "/" {
		url = url + "/"
	}
	path = strings.TrimPrefix(path, "/")
	// 检查默认字典加上 传来的路径
	CheckFoldersForTraversal(url, util.RemoveDuplicateElement(append(dictionary, path)))

	// Check for alias traversal vulnerability (endpoint finding)
	if body == "" {
		resp, err := httpx.Get(url, "Traversal")
		if err != nil {
			logging.Logger.Errorln(err)
			return
		}
		body = resp.Body
	}

	// 使用 findEndpoints 获取当前页面的所有路径，然后再跑一遍。 TODO 这种不太好，会导致重复扫描
	CheckFoldersForTraversal(url, findEndpoints(body))
	// Check for directory listing
	// Check for file existence
	// Check for file contents
}

func CheckFolderForTraversal(url string, folder string) bool {
	resp, err := httpx.Get(url+folder+".", "Traversal")

	if err != nil {
		return false
	}

	if resp.StatusCode == 404 {
		return false
	}

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		if strings.HasSuffix(resp.Location, folder+"./") {
			resp, err := httpx.Get(url+folder+"..", "Traversal")
			if err != nil {
				return false
			}
			if resp.StatusCode == 301 || resp.StatusCode == 302 {
				if strings.HasSuffix(resp.Location, folder+"../") {
					respNotFound, err := httpx.Get(url+folder+"."+util.RandomString(4), "Traversal")
					if err != nil {
						return false
					}
					if respNotFound.StatusCode == 404 || strings.Contains(strings.ToLower(respNotFound.Body), "not found") {
						respNotFound2, err := httpx.Get(url+folder+"z", "Traversal")
						if err != nil {
							return false
						}
						if respNotFound2.StatusCode == 404 || strings.Contains(strings.ToLower(respNotFound2.Body), "not found") {
							// vulnerable
							statusNotFound3, err := httpx.Get(url+folder+"z..", "Traversal")
							if err != nil {
								return false
							}
							if statusNotFound3.StatusCode != 302 && statusNotFound3.StatusCode != 301 {
								// vulnerable
								output.OutChannel <- output.VulMessage{
									DataType: "web_vul",
									Plugin:   "Nginx Alias Traversal",
									VulnData: output.VulnData{
										CreateTime: time.Now().Format("2006-01-02 15:04:05"),
										Target:     url + folder,
										Method:     "GET",
										Ip:         "",
										Param:      "",
										Request:    statusNotFound3.RequestDump,
										Response:   statusNotFound3.ResponseDump,
										Payload:    url + folder + "../",
									},
									Level: output.Medium,
								}
								logging.Logger.Infof("Vulnerable: %s", url+folder+"../")
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func CheckFoldersForTraversal(url string, folders []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	// Use a bounded semaphore with a capacity of 'threads'
	for i := 0; i < 10; i++ {
		semaphore <- struct{}{}
	}

	for _, word := range folders {
		if word == "" {
			continue
		}
		wg.Add(1)
		// Acquire a token from the semaphore channel
		<-semaphore
		go func(word string) {
			CheckFolderForTraversal(url, word)
			// Release the token back to the semaphore channel
			semaphore <- struct{}{}
			wg.Done()
		}(word)
	}

	wg.Wait()
}

func MakeFolderEndpointsFromPath(path string) []string {
	// remove query string
	// "/img/media/a.jpg" -> ["img", "img/media"]
	var endpoints []string
	var endpoint string

	// check if begins with https:// or http:// or //
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "//") {
		// get path
		u, err := url.Parse(path)
		if err != nil {
			logging.Logger.Errorln(err)
			return nil
		}
		path = u.Path
		if strings.HasPrefix(path, "/") {
			path = path[1:]
		}
	}

	for _, word := range strings.Split(path, "/") {
		if word != "" {
			// check if last
			if word == strings.Split(path, "/")[len(strings.Split(path, "/"))-1] {
				break
			}
			// remove query string
			if strings.Contains(word, "?") {
				word = strings.Split(word, "?")[0]
			}

			endpoint = endpoint + word + "/"
			endpointNoSlash := strings.TrimSuffix(endpoint, "/")
			if !util.InSlice(endpoints, endpointNoSlash) {
				endpoints = append(endpoints, endpointNoSlash)
			}
		}
	}
	return endpoints
}

func findEndpoints(html string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}
	var foundEndpoints []string
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			for _, endpoint := range foundEndpoints {
				// check if endpoint already exists
				if endpoint == src {
					return
				}
			}
			foundEndpoints = append(foundEndpoints, src)

		}
	})
	finalDirectoryEndpoints := []string{}
	for _, endpoint := range foundEndpoints {
		directoryEndpoints := MakeFolderEndpointsFromPath(endpoint)
		for _, directoryEndpoint := range directoryEndpoints {
			if !util.InSlice(finalDirectoryEndpoints, directoryEndpoint) && !util.InSlice(dictionary, directoryEndpoint) {
				finalDirectoryEndpoints = append(finalDirectoryEndpoints, directoryEndpoint)
			}
		}
	}
	return finalDirectoryEndpoints
}

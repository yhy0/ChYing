package fuzz

import (
	"fmt"
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/Jie/fingprints"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/bbscan"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/bypass403"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/swagger"
	"github.com/yhy0/ChYing/pkg/ui"
	"github.com/yhy0/logging"
	"strconv"
	"strings"
)

/**
  @author: yhy
  @since: 2023/4/21
  @desc: fuzz 目录、api、参数等
**/

var Chan chan ui.Result

var Percentage chan float64

var Stop = false

func init() {
	Chan = make(chan ui.Result, 1)
	Percentage = make(chan float64, 1)
}

func Fuzz(target string, actions []string) error {
	// 每次扫描前都重新读取一遍配置字典
	file.ReadFiles()

	client := httpx.NewClient(nil)
	res, err := client.Request(target, "GET", "", nil, "Fuzz")
	if err != nil {
		logging.Logger.Errorln(err)
		return err
	}
	fingerprints := fingprints.Identify([]byte(res.Body), res.RespHeader)
	var _403 = false
	if funk.Contains(actions, "bypass403") {
		_403 = true
	}

	if funk.Contains(actions, "swagger") {
		swagger.Scan(target, client, Chan)
	}

	if funk.Contains(actions, "bbscan") {
		bbscan.BBscan(target, true, fingerprints, nil, client, Chan)
	}

	if funk.Contains(actions, "wsdl") {
		bbscan.BBscan(target, true, fingerprints, nil, client, Chan)
	}

	if len(actions) > 0 && !funk.Contains(actions, "bbscan") {
		DirSearch(target, actions, _403, client)
	}

	if len(actions) == 1 && _403 && !Stop {
		bypass403.Bypass403(target, "GET", nil, res.Body, client)
	}

	return nil
}

// DirSearch 规则 进行目录遍历
func DirSearch(target string, actions []string, _403 bool, client *httpx.Client) {
	target = strings.TrimRight(target, "/")
	var paths []string

	for _, p := range file.DictData {
		// 替换后缀
		if strings.Contains(p, "%EXT%") {
			for _, ext := range actions {
				if ext == "bbscan" || ext == "403" {
					continue
				}
				p = strings.ReplaceAll(p, "%EXT%", ext)
				paths = append(paths, p)
			}

		} else {
			paths = append(paths, p)
		}
	}

	n := len(paths)

	ch := make(chan struct{}, 30)
	for i, p := range paths {
		if Stop {
			return
		}

		ch <- struct{}{}
		go func(ii int, path string) {
			resp, err := client.Request(target+"/"+path, "GET", "", nil, "Fuzz")

			float, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(ii+1)/float64(n)*100), 64)
			Percentage <- float

			if err != nil {
				<-ch
				return
			}

			<-ch
			if filter(path, resp) {
				return
			}

			if _403 && resp.StatusCode == 403 {
				go bypass403.Bypass403(target+"/"+path, "GET", nil, resp.Body, client)
			}

			Chan <- ui.Result{
				Url:           target + "/" + path,
				Method:        "GET",
				StatusCode:    resp.StatusCode,
				ContentLength: resp.ContentLength,
				Request:       resp.RequestDump,
				Response:      resp.ResponseDump,
			}
		}(i, p)
	}

	close(ch)
}

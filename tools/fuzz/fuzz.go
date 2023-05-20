package fuzz

import (
	"fmt"
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/tools"
	"github.com/yhy0/logging"
	"strings"
)

/**
  @author: yhy
  @since: 2023/4/21
  @desc: fuzz 目录、api、参数等
**/

var FuzzChan chan tools.Result

var FuzzPercentage chan string

var Stop = false

func init() {
	FuzzChan = make(chan tools.Result, 1)
	FuzzPercentage = make(chan string, 1)
}

func Fuzz(target string, actions []string, filePath string) error {
	// 每次扫描前都重新读取一遍配置字典
	file.ReadFiles()

	res, err := httpx.Request(target, "GET", "", false, nil)
	if err != nil {
		logging.Logger.Errorln(err)
		return err
	}

	var _403 = false
	if funk.Contains(actions, "403") {
		_403 = true
	}

	if funk.Contains(actions, "bbscan") {
		BBscan(target, res.ContentLength, res.Body, _403)
	}

	if len(actions) > 0 && !funk.Contains(actions, "bbscan") {
		DirSearch(target, actions, _403)
	}

	if len(actions) == 1 && _403 {
		Bypass403(target, "GET")
	}

	return nil
}

// DirSearch 规则 进行目录遍历
func DirSearch(target string, actions []string, _403 bool) {
	target = strings.TrimRight(target, "/")
	var paths []string

	for _, p := range file.DiccData {
		if Stop {
			Stop = false
			return
		}

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
		ch <- struct{}{}
		go func(ii int, path string) {
			resp, err := httpx.Get(target + "/" + path)
			FuzzPercentage <- fmt.Sprintf("%.2f", float64(ii+1)/float64(n)*100)
			if err != nil {
				<-ch
				return
			}

			<-ch
			if filter(path, resp) {
				return
			}

			if _403 && resp.StatusCode == 403 {
				go Bypass403(target+"/"+path, "GET")
			}

			FuzzChan <- tools.Result{
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

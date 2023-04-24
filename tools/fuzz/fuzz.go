package fuzz

import (
	"fmt"
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/tools"
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
	res, err := httpx.Request(target, "GET", "", false, nil)
	if err != nil {
		return err
	}

	if funk.Contains(actions, "bbscan") {
		BBscan(target, res.ContentLength, res.Body)
	}

	if len(actions) > 0 && !funk.Contains(actions, "bbscan") {
		DirSearch(target, actions)
	}

	return nil
}

// DirSearch 规则 进行目录遍历
func DirSearch(target string, actions []string) {
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

			FuzzChan <- tools.Result{
				Url:           target + "/" + path,
				StatusCode:    resp.StatusCode,
				ContentLength: resp.ContentLength,
				Request:       resp.RequestDump,
				Response:      resp.ResponseDump,
			}
		}(i, p)
	}

	close(ch)
}

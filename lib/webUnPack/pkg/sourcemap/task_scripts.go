package collector

import (
	"bytes"
	"fmt"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/output"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/utils"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/webfinder/javascript"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
)

var rex = regexp.MustCompile(`//#\s*sourceMappingURL=(.*\.map)[\s\x00$]`)

// TaskScripts consumes JS scripts and extracts source map URLs
type TaskScripts struct {
	Output  string
	In      chan *url.URL
	Out     chan *url.URL
	Visited map[string]struct{}
	Mutex   *sync.Mutex
	Target  string
}

func (TaskScripts) Name() string {
	return "scripts"
}

func (task *TaskScripts) Finish() {
	close(task.Out)
}

func (task *TaskScripts) URLs() <-chan *url.URL {
	return task.In
}

func (task *TaskScripts) Run(surl *url.URL) error {
	if task.visited(surl.String()) {
		return nil
	}
	resp, err := httpx.Get(surl.String(), "SourceMap")
	if err != nil {
		return fmt.Errorf("make http request: %v", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("invalid response: %s", resp.Status)
	}
	var respbody string
	// get source map url from headers
	murl := resp.RespHeader.Get("SourceMap")
	if murl == "" {
		murl = resp.RespHeader.Get("X-SourceMap")
	}
	if murl == "" {
		// get source map url from comments
		murl, respbody, err = task.find(rex, strings.NewReader(resp.Body))
		if err != nil {
			return fmt.Errorf("read response body: %v", err)
		}
	}

	var flag = false
	if murl != "" {
		murl, err := surl.Parse(murl)
		if err != nil {
			return fmt.Errorf("parse source map url: %v", err)
		}
		task.Out <- murl
		flag = true
	}

	if !flag {
		if strings.Contains(surl.String(), "/manifest.") {
			// 使用正则表达式提取对应的字符串
			re := regexp.MustCompile(`\d:\s*"([^"]+)"`)
			matches := re.FindAllStringSubmatch(respbody, -1)

			logging.Logger.Infoln("matches::: ", matches)

			task.Run(surl)
		}

		if strings.Contains(surl.String(), "/umi.") {
			re := regexp.MustCompile(`"\."\+{(.*)\+"\.async\.js"}`)
			asyncJs := re.FindAllStringSubmatch(respbody, -1)

			// 使用正则表达式提取对应的字符串
			re = regexp.MustCompile(`['"]\d+['"]:['"].*?['"]`)
			matches := re.FindAllStringSubmatch(asyncJs[0][0], -1)

			for _, m := range matches {
				mm := strings.Split(strings.ReplaceAll(m[0], "\"", ""), ":")
				murl, _ := surl.Parse(utils.Standard(task.Target, mm[0]+"."+mm[1]+".async.js"))
				task.Run(murl)
			}
		}

		if task.Output != "" && !strings.HasPrefix(respbody, "<!DOCTYPE html>") {
			name := strings.Split(surl.String(), "/")
			fname := path.Join(task.Output, surl.Hostname(), name[len(name)-1])
			parent, _ := path.Split(fname)
			err = os.MkdirAll(parent, 0770)
			if err != nil {
				return fmt.Errorf("create dir: %v", err)
			}

			err = os.WriteFile(fname, []byte(respbody), 0660)
			if err != nil {
				return fmt.Errorf("write file: %v", err)
			}
		}

		walker, err := javascript.UriFromJavascript(respbody)
		if err != nil {
			return nil
		}
		var uris []string
		for _, route := range walker.Routes {
			if route.Valid() {
				route.Path = strings.ReplaceAll(route.Path, "'", "")
				route.Path = strings.ReplaceAll(route.Path, "\"", "")
				uris = append(uris, route.Path)
				for _, child := range route.Children {
					child.Path = strings.ReplaceAll(child.Path, "'", "")
					child.Path = strings.ReplaceAll(child.Path, "\"", "")
					uris = append(uris, child.Path)
				}
			}
		}
		for _, a := range walker.Apis {
			if a.Uri != "" {
				uris = append(uris, a.Uri)
			}
		}

		for _, uri := range uris {
			output.ResultChan <- output.Result{
				Value:  uri,
				Type:   "api",
				Source: surl.String(),
			}
		}
	}
	return nil
}

func (task *TaskScripts) visited(url string) bool {
	task.Mutex.Lock()
	defer task.Mutex.Unlock()
	_, visited := task.Visited[url]
	if visited {
		return true
	}
	task.Visited[url] = struct{}{}
	return false
}

func (TaskScripts) find(rex *regexp.Regexp, stream io.Reader) (string, string, error) {
	prev := make([]byte, 1024)
	var buf bytes.Buffer
	for {
		curr := make([]byte, 1024)
		n, err := stream.Read(curr)
		if n == 0 {
			return "", buf.String(), nil
		}
		if err != nil && err != io.EOF {
			return "", "", err
		}
		buf.Write(curr[:n])
		match := rex.FindSubmatch(append(prev, curr...))
		if match != nil {
			return string(match[1]), "", nil
		}
		prev = curr
	}
}

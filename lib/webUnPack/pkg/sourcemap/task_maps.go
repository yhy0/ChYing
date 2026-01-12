package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/output"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/utils"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/webfinder/css"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/webfinder/javascript"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"net/url"
	"os"
	"path"
	"strings"
)

// TaskMaps consumes source maps URLs and extracts them into the file system
type TaskMaps struct {
	Output string
	In     chan *url.URL
	Out    chan *url.URL
	Target string
}

func (TaskMaps) Name() string {
	return "maps"
}

func (task *TaskMaps) Finish() {
	close(task.Out)
}

func (task *TaskMaps) URLs() <-chan *url.URL {
	return task.In
}

func (task *TaskMaps) Run(surl *url.URL) error {
	resp, err := httpx.Get(surl.String(), "SourceMap")
	if err != nil {
		return fmt.Errorf("make http request: %v", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("invalid response: %s", resp.Status)
	}

	var m SourceMap
	err = json.Unmarshal([]byte(resp.Body), &m)

	if err != nil {
		return fmt.Errorf("read JSON: %v", err)
	}

	var uris []string
	for i, fname := range m.FileNames {
		fname = strings.ReplaceAll(fname, "../", ".")
		fname = strings.ReplaceAll(fname, "webpack://", "")
		fname = strings.ReplaceAll(fname, "://", "")
		fname = path.Join(task.Output, surl.Hostname(), fname)

		if i >= len(m.Contents) {
			return errors.New("sources is longer than sourcesContent")
		}
		if strings.HasPrefix(fname, "external ") {
			logging.Logger.Errorln("external source maps unsupported")
			continue
		}
		if task.Output != "" {
			parent, _ := path.Split(fname)
			err = os.MkdirAll(parent, 0770)
			if err != nil {
				return fmt.Errorf("create dir: %v", err)
			}

			err = os.WriteFile(fname, []byte(m.Contents[i]), 0660)
			if err != nil {
				return fmt.Errorf("write file: %v", err)
			}
		}

		if !strings.Contains(fname, "/~/") && strings.HasSuffix(fname, ".js") {
			walker, err := javascript.UriFromJavascript(m.Contents[i])
			if err != nil {
				continue
			}
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
			for _, api := range walker.Apis {
				if api.Uri != "" {
					uris = append(uris, api.Uri)
				}
			}
		}

		if strings.HasSuffix(fname, ".css") {
			uris = append(uris, css.UriFromCss(m.Contents[i])...)
		}
	}

	uris = utils.RemoveDuplicateElement(uris)

	for _, uri := range uris {
		output.ResultChan <- output.Result{
			Value:  uri,
			Type:   "api",
			Source: surl.String(),
		}
	}

	task.Out <- surl
	return nil
}

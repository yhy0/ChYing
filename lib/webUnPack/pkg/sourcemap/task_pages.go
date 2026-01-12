package collector

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TaskPages consumes HTML and extracts JS scripts URLs
type TaskPages struct {
	In  chan *url.URL
	Out chan *url.URL
}

func (TaskPages) Name() string {
	return "pages"
}

func (task *TaskPages) Finish() {
	close(task.Out)
}

func (task *TaskPages) URLs() <-chan *url.URL {
	return task.In
}

func (task *TaskPages) Run(purl *url.URL) error {
	resp, err := httpx.Get(purl.String(), "SourceMap")
	if err != nil {
		return fmt.Errorf("make http request: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.Body))
	if err != nil {
		return fmt.Errorf("make goquery doc: %v", err)
	}

	var qerr error
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		surl, _ := s.Attr("src")
		if surl != "" && !strings.HasPrefix(surl, "data:text/javascript,") {
			surl, err := purl.Parse(surl)
			if err != nil {
				qerr = err
				return
			}
			task.Out <- surl
		}
		surl, _ = s.Attr("data-src")
		if surl != "" && !strings.HasPrefix(surl, "data:text/javascript,") {
			surl, err := purl.Parse(surl)
			if err != nil {
				qerr = err
				return
			}
			task.Out <- surl
		}
	})
	if qerr != nil {
		return fmt.Errorf("parse script url: %v", qerr)
	}

	return nil
}

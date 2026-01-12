package log4j

import (
    "github.com/yhy0/ChYing/pkg/Jie/pkg/input"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/pocs_go/log4j"
    "sync"
)

/**
  @author: yhy
  @since: 2023/12/29
  @desc: //TODO
**/

type Plugin struct {
    SeenRequests sync.Map
}

func (p *Plugin) Name() string {
    return "log4j"
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
    if p.IsScanned(in.UniqueId) {
        return
    }
    
    log4j.Scan(target, in.Method, in.RequestBody, client)
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

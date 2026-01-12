package portScan

import (
    "github.com/yhy0/ChYing/pkg/Jie/pkg/input"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "sync"
)

/**
  @author: yhy
  @since: 2023/10/30
  @desc: //TODO
**/

type Plugin struct {
    SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
    if in.Cdn {
        return
    }
    if p.IsScanned(in.UniqueId) || in.Ip == "" {
        return
    }
    
    res := Scan(target, in.Ip)
    output.Lock.Lock()
    output.IPInfoList[in.Ip].PortService = res
    output.Lock.Unlock()
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
    return "portScan"
}

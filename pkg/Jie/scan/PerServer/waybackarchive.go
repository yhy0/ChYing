package PerServer

import (
    "github.com/yhy0/ChYing/pkg/Jie/pkg/input"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/scan/gadget/waybackarchive"
    "sync"
)

/**
  @author: yhy
  @since: 2023/11/22
  @desc: //TODO
**/

type ArchivePlugin struct {
    SeenRequests sync.Map
}

func (p *ArchivePlugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
    if p.IsScanned(in.UniqueId) {
        return
    }
    // TODO 获取的结果应该何时何地进行扫描检查？
    in.Archive = waybackarchive.Run(target, client)
}

func (p *ArchivePlugin) IsScanned(key string) bool {
    if key == "" {
        return false
    }
    if _, ok := p.SeenRequests.Load(key); ok {
        return true
    }
    p.SeenRequests.Store(key, true)
    return false
}

func (p *ArchivePlugin) Name() string {
    return "archive"
}

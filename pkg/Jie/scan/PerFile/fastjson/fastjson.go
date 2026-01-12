package fastjson

import (
    "fmt"
    "github.com/thoas/go-funk"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/input"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/fastjson/Detect"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/fastjson/Utils"
    "sync"
    "time"
)

/**
   @author yhy
   @since 2023/9/18
   @desc //TODO
**/

func Scan(target string, client *httpx.Client) Utils.Result {
    results := Detect.Version(target, client)
    return results
}

type Plugin struct {
    SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
    // 测试 json ，根据 header 中的 Content-Type 判断
    if !funk.Contains(in.ContentType, "json") {
        return
    }
    
    if p.IsScanned(in.UniqueId) {
        return
    }
    
    results := Scan(in.Url, client)
    
    if results.Type != "" {
        output.OutChannel <- output.VulMessage{
            DataType: "web_vul",
            Plugin:   "fastjson",
            VulnData: output.VulnData{
                CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
                Target:      in.Url,
                Method:      in.Method,
                Payload:     results.Payload,
                Request:     results.Request,
                Response:    results.Response,
                Description: fmt.Sprintf("Type: %s,Version: %s, AutoType: %v, Netout: %v, Dependency: %v", results.Type, results.Version, results.AutoType, results.Netout, results.Dependency),
            },
            Level: output.Medium,
        }
    }
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
    return "fastjson"
}

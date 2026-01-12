package collection

import (
    chYingConf "github.com/yhy0/ChYing/conf"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/input"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/util"
    "strings"
    "sync"
    "time"
)

/**
   @author yhy
   @since 2024/5/15
   @desc 敏感参数
**/

type Plugin struct {
    SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
    if p.IsScanned(in.UniqueId) {
        return
    }
    
    SensitiveParameters(in)
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
    return "SensitiveParameters"
}

func SensitiveParameters(in *input.CrawlResult) {
    flag := false
    var sensitiveParameters, rawRequest, rawResponse string
    resParameters, _ := util.GetResParameters(strings.ToLower(in.Resp.RespHeader.Get("Content-Type")), []byte(in.Resp.Body))
    
    for _, para := range conf.GlobalConfig.Collection.SensitiveParameters {
        if strings.HasPrefix(para, "_") {
            if util.InSliceCaseFold(para, in.ParamNames) {
                rawRequest = in.RawRequest
                sensitiveParameters += para + " "
            }
            
            if util.InSliceCaseFold(para, resParameters) {
                rawResponse = in.RawResponse
                sensitiveParameters += para + " "
            }
        } else {
            if util.InCaseFoldSlice(in.ParamNames, para) {
                rawRequest = in.RawRequest
                sensitiveParameters += para + " "
            }
            if util.InCaseFoldSlice(resParameters, para) {
                rawResponse = in.RawResponse
                sensitiveParameters += para + " "
            }
        }
    }
    if sensitiveParameters != "" {
        output.OutChannel <- output.VulMessage{
            DataType: "web_vul",
            Plugin:   "SensitiveParameters",
            VulnData: output.VulnData{
                CreateTime: time.Now().Format("2006-01-02 15:04:05"),
                Target:     in.Url,
                Method:     in.Method,
                Payload:    sensitiveParameters,
                Request:    rawRequest,
                Response:   rawResponse,
            },
            Level: output.Low,
        }
        flag = true
    }
    
    if flag {
        hid, ok := chYingConf.Address.Load(in.Id)
        if ok {
            chYingConf.HttpMarkerChan <- chYingConf.HttpMarker{
                Id:    hid.(int64),
                Color: "purple",
                Note:  "SensitiveParameters",
            }
        }
    }
}

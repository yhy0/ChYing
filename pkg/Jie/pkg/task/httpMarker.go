package task

import (
    chYingConf "github.com/yhy0/ChYing/conf"
    "github.com/yhy0/ChYing/pkg/Jie/scan/gadget/sensitive"
)

/**
   @author yhy
   @since 2024/12/31
   @desc //TODO
**/

func Marker(url, reqBody, respBody, rawReq, rawResp, id string) {
    hid, ok := chYingConf.Address.Load(id)
    if !ok {
        return
    }
    // 界面颜色显示 匹配
    color, note := chYingConf.MarkerMitmRule(rawReq, rawResp)
    
    // 敏感信息检测
    if sensitive.KeyDetection(url, respBody) {
        color = "red"
        note += "Sensitive Key"
    }
    
    if len(sensitive.PageErrorMessageCheck(url, reqBody, respBody)) > 0 {
        color = "red"
        note += "Sensitive Error"
    }
    
    if sensitive.Wih(url, reqBody, respBody) {
        color = "red"
        note += "wih"
    }
    
    // 更新颜色
    if color != "" {
        chYingConf.HttpMarkerChan <- chYingConf.HttpMarker{
            Id:    hid.(int64),
            Color: color,
            Note:  note,
        }
    }
}

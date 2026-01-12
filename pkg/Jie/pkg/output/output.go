package output

import (
    "github.com/logrusorgru/aurora"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/logging"
    "net/url"
)

/**
   @author yhy
   @since 2023/10/14
   @desc //TODO
**/

var OutChannel = make(chan VulMessage)

func Write(progress bool) {
    if progress {
        go Progress()
    }
    
    if conf.GlobalConfig.Options.Output != "" {
        go GenerateVulnReport(conf.GlobalConfig.Options.Output)
    }
    
    for v := range OutChannel {
        // 漏洞保存到文件
        if conf.GlobalConfig.Options.Output != "" {
            ReportMessageChan <- v
        }
        
        // 被动模式下的 SCopilot 结果保存
        if conf.GlobalConfig.Passive.WebPort != "" {
            parse, err := url.Parse(v.VulnData.Target)
            if err != nil {
                logging.Logger.Errorln(err)
                continue
            }
            msg := SCopilotData{
                Target: v.VulnData.Target,
            }
            
            if v.Level == "Low" {
                msg.InfoMsg = []PluginMsg{
                    {
                        Url:      v.VulnData.Target,
                        Plugin:   v.Plugin,
                        Result:   []string{v.VulnData.Payload},
                        Request:  v.VulnData.Request,
                        Response: v.VulnData.Response,
                    },
                }
            } else {
                msg.VulMessage = append(msg.VulMessage, v)
            }
            
            SCopilot(parse.Host, msg)
        }
        
        logging.Logger.Infoln(aurora.Red(v.PrintScreen()).String())
    }
}

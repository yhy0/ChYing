package traversal

import (
    "github.com/logrusorgru/aurora"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/logging"
    "testing"
)

/**
   @author yhy
   @since 2023/8/15
   @desc //TODO
**/

func TestTraversal(t *testing.T) {
    logging.Logger = logging.New(true, "", "agent", false)
    conf.GlobalConfig = &conf.Config{}
    conf.GlobalConfig.Http.Proxy = ""
    
    // 获取扫描结果
    go func() {
        for v := range output.OutChannel {
            logging.Logger.Infoln(aurora.Red(v.PrintScreen()).String())
        }
    }()
    
    NginxAlias("https://md.huodong.baidu.com/", "", "")
}

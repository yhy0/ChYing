package sql

import (
    "fmt"
    JieOutput "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/util"
    "github.com/yhy0/logging"
    "time"
)

/**
  @author: yhy
  @since: 2023/2/8
  @desc: 基于报错注入, 只是简单验证了是否存在报错信息
**/

func checkDBMSError(url, param, payload string, res *httpx.Response) string {
    for DBMS, regexps := range DbmsErrors {
        if math, err := util.MatchAnyOfRegexp(regexps, res.ResponseDump); math {
            JieOutput.OutChannel <- JieOutput.VulMessage{
                DataType: "web_vul",
                Plugin:   "SQL Injection",
                VulnData: JieOutput.VulnData{
                    CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
                    Target:      url,
                    Ip:          "",
                    Param:       param,
                    Request:     res.RequestDump,
                    Response:    res.ResponseDump,
                    Payload:     payload,
                    Description: fmt.Sprintf("ERROR-Based SQL Injection: [%v] Guess DBMS: %v", param, DBMS),
                },
                Level: JieOutput.Critical,
            }
            
            logging.Logger.Infof("%s %s 检测到数据库报错信息[%s:%s]", url, param, DBMS, err)
            
            return DBMS
        }
    }
    return ""
}

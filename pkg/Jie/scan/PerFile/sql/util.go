package sql

import (
	"fmt"
	"github.com/antlabs/strsim"
	"github.com/sergi/go-diff/diffmatchpatch"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/logging"
	"math"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"
)

/**
  @author: yhy
  @since: 2023/2/10
  @desc: 工具类辅助函数
**/

// getErrorBasedPreCheckPayload 闭合 payload,  ,'"().六种字符随机组成的长度为10的字符串，同时满足'和"都只有一个。
func getErrorBasedPreCheckPayload() string {
	rand.Seed(time.Now().Unix())
	randomTestString := ""
	randI := rand.Intn(8)
	randJ := rand.Intn(8)

	// HeuristicCheckAlphabet 用于闭合字符串的字母表
	heuristicCheckAlphabet := []string{`)`, `(`, `,`, `.`}
	for i := 0; i < 8; i++ {
		str := heuristicCheckAlphabet[rand.Intn(len(heuristicCheckAlphabet)-1)]
		randomTestString += str

		// 保证满足'和"都只有一个。
		if i == randI {
			randomTestString += "'"
		}
		if i == randJ {
			randomTestString += "\""
		}
	}

	return randomTestString
}

// getNormalRespondTime 对目标发起5次请求, 正常的响应时间应该小于等于这个值
func getNormalRespondTime(sql *Sqlmap) (error, float64) {
	var timeRec []float64
	for i := 0; i < 5; i++ {
		res, err := sql.Client.Request(sql.Url, sql.Method, sql.RequestBody, sql.Headers, "SQL-Injection")
		if err != nil {
			return err, -1
		}

		timeRec = append(timeRec, res.ServerDurationMs)

	}
	// 正常情况下，响应时间的分布是符合正态分布的，而正态分布的 99.99% 的数据都在 7 倍的标准差之内，所以这里就是为了过滤掉 99.99% 的无延迟请求。
	return nil, mean(timeRec) + 7*std(timeRec)

}

// 期望,也就是平均响应时间
func mean(v []float64) float64 {
	var res float64 = 0
	var n = len(v)
	for i := 0; i < n; i++ {
		res += v[i]
	}
	return res / float64(n)
}

// 方差
func variance(v []float64) float64 {
	var res float64 = 0
	var m = mean(v)
	var n = len(v)
	for i := 0; i < n; i++ {
		res += (v[i] - m) * (v[i] - m)
	}
	return res / float64(n-1)
}

// 标准差
func std(v []float64) float64 {
	return math.Sqrt(variance(v))
}

// 找出不同处的前 20 字符，和后 20 个字符 ，不知道有没有问题，感觉应该是存在点问题的
func findDynamicContent(s1, s2 string) (prefix, suffix string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(s1, s2, true)

	diffString := ""
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			diffString = diff.Text
			break
		}
	}

	index := strings.Index(s1, diffString)
	if index == -1 {
		index = strings.Index(s2, diffString)
	}

	if index != -1 {
		startIndex := index - 20
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex := index + len(diffString) + 20
		if endIndex > len(s1) {
			endIndex = len(s1)
		}

		prefix = s1[startIndex:index]
		if len(s1) > index+len(diffString) {
			suffix = s1[index+len(diffString) : endIndex]
		} else {
			logging.Logger.Debugln("findDynamicContent err:", s1, " -- ", s2)
		}
	}

	return
}

// 删除动态内容
func (sql *Sqlmap) removeDynamicContent(page string) string {
	var regex *regexp.Regexp
	var restr string

	if sql.DynamicMarkings["prefix"] == "" && sql.DynamicMarkings["suffix"] != "" {
		restr = fmt.Sprintf(`(?s)^.+%s`, sql.DynamicMarkings["suffix"])
	} else if sql.DynamicMarkings["suffix"] == "" && sql.DynamicMarkings["prefix"] != "" {
		restr = fmt.Sprintf(`(?s)%s.+$`, sql.DynamicMarkings["prefix"])
	} else if sql.DynamicMarkings["suffix"] != "" && sql.DynamicMarkings["prefix"] != "" {
		restr = fmt.Sprintf(`(?s)%s.+%s`, sql.DynamicMarkings["prefix"], sql.DynamicMarkings["suffix"])
	} else {
		return page
	}

	if utf8.ValidString(restr) {
		regex = regexp.MustCompile(regexp.QuoteMeta(restr))
	} else { // 包含无效的 utf-8 字符
		return page
	}

	return regex.ReplaceAllString(page, fmt.Sprintf("%s%s", sql.DynamicMarkings["prefix"], sql.DynamicMarkings["suffix"]))
}

// comparison 与模板比较响应页面的相似度
func (sql *Sqlmap) comparison(respBody string, respCode int, criticalRatio float64) (bool, float64) {
	if respCode == sql.TemplateCode {
		// 确定临界时, 先去除动态部分
		respBody = sql.removeDynamicContent(respBody)

		ratio := strsim.Compare(respBody, sql.TemplateBody)
		// 如果是第一次比较, 就把这个值作为默认值
		if criticalRatio == -1 {
			if ratio >= LowerRatioBound && ratio <= UpperRatioBound {
				criticalRatio = ratio
			}
		}

		if ratio > UpperRatioBound {
			return true, ratio
		} else if ratio < LowerRatioBound {
			return false, ratio
		} else { // 相似度在临界值之间，则判断相似度和传入的相似度的差值是否大默认的容差, 如果大于容差, 则认为是不同的页面
			return (ratio - criticalRatio) > DiffTolerance, ratio
		}

	}
	return false, -1
}

// compare 与模板比较响应页面的相似度
func (sql *Sqlmap) compare(respBody string, respCode int) float64 {
	if respCode == sql.TemplateCode {
		// 确定临界时, 先去除动态部分
		respBody = sql.removeDynamicContent(respBody)

		ratio := strsim.Compare(respBody, sql.TemplateBody)

		return ratio
	}

	return -1
}

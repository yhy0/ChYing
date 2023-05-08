package burpSuite

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/pkg/util"
	"regexp"
	"strconv"
	"strings"
)

/**
  @author: yhy
  @since: 2023/5/7
  @desc: burpSuite 的 Intruder 模式相关实现
**/

func Intruder(target string, req string, payloads []string, rules []string, attackType string, uuid string, ctx context.Context) {
	fmt.Println(attackType)
	switch attackType {
	case "Sniper":
		sniper(target, req, payloads, rules, uuid, ctx)
	case "Battering ram":
		batteringRam(target, req, payloads, rules, uuid, ctx)
	case "Pitchfork":
		pitchfork(target, req, payloads, rules, uuid, ctx)
	case "Cluster bomb":
		clusterBomb(target, req, payloads, rules, uuid, ctx)
	}
}

// sniper 模式设置的所有 payload 位置使用同一份 fuzz 文本
func sniper(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	// 定义分隔符的回调函数
	splitFunc := func(r rune) bool {
		return r == '\r' || r == '\n'
	}

	// 使用 FieldsFunc() 函数分割字符串
	payloads = strings.FieldsFunc(payloads[0], splitFunc)

	for i, payload := range payloads {
		request := req // req 不能改变
		// 这里是根据payload位置来进行对应的处理
		for j, position := range positions {
			payload = processing(payload, rules[j])
			request = strings.Replace(request, position, payload, 1)
		}

		ch <- struct{}{}

		go func(request, payload string, i int) {
			intruderRes := IntruderRes{
				Id:      i,
				Payload: []string{payload},
			}
			resp, err := httpx.Raw(request, target)
			if err != nil {
				runtime.EventsEmit(ctx, uuid, intruderRes)
				return
			}

			intruderRes.Status = strconv.Itoa(resp.StatusCode)
			intruderRes.Length = strconv.Itoa(resp.ContentLength)

			runtime.EventsEmit(ctx, uuid, intruderRes)

			smap, ok := IntruderMap[uuid]
			if !ok {
				// 如果不存在，则创建一个新的 SMap 实例并添加到 IntruderMap 中
				smap = &SMap{
					Map: make(map[int]*HTTPBody),
				}
				IntruderMap[uuid] = smap
			}

			IntruderMap[uuid].WriteMap(i, &HTTPBody{
				TargetUrl: target,
				Request:   resp.RequestDump,
				Response:  resp.ResponseDump,
			})

		}(request, payload, i)
	}
	close(ch)
}

func batteringRam(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {

}

func pitchfork(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {

}

// clusterBomb 每个 payload 位置使用不同的 fuzz 文本
func clusterBomb(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	// 定义分隔符的回调函数
	splitFunc := func(r rune) bool {
		return r == '\r' || r == '\n'
	}

	var payloadsMap = make(map[string][]string)
	for i, payload := range payloads {
		// 使用 FieldsFunc() 函数分割字符串
		payloadsMap[rules[i]] = strings.FieldsFunc(payload, splitFunc)
	}

	//  获取全部组合的可能性 payloads
	result := combinations(payloadsMap, rules)

	for i, payload := range result {
		request := req // req 不能改变
		// 这里是根据payload位置来进行对应的处理
		for j, position := range positions {
			_payload := processing(payload[j], rules[j])
			request = strings.Replace(request, position, _payload, 1)
		}
		ch <- struct{}{}

		go func(request string, payload []string, i int) {
			intruderRes := IntruderRes{
				Id:      i,
				Payload: payload,
			}
			resp, err := httpx.Raw(request, target)
			if err != nil {
				runtime.EventsEmit(ctx, uuid, intruderRes)
				return
			}

			intruderRes.Status = strconv.Itoa(resp.StatusCode)
			intruderRes.Length = strconv.Itoa(resp.ContentLength)

			runtime.EventsEmit(ctx, uuid, intruderRes)

			smap, ok := IntruderMap[uuid]
			if !ok {
				// 如果不存在，则创建一个新的 SMap 实例并添加到 IntruderMap 中
				smap = &SMap{
					Map: make(map[int]*HTTPBody),
				}
				IntruderMap[uuid] = smap
			}

			IntruderMap[uuid].WriteMap(i, &HTTPBody{
				TargetUrl: target,
				Request:   resp.RequestDump,
				Response:  resp.ResponseDump,
			})

		}(request, payload, i)
	}
	close(ch)
}

// combinations 获取所有 clusterBomb 模式payload组合结果的可能性
func combinations(data map[string][]string, pattern []string) [][]string {
	var results [][]string
	var stack []int
	var output []string

	for i := 0; i < len(pattern); i++ {
		key := pattern[i]
		values := data[key]

		if len(values) == 0 {
			return nil
		}

		stack = append(stack, len(values))
		output = append(output, values[0])
	}

	for len(stack) > 0 {
		result := make([]string, len(pattern))
		copy(result, output)
		results = append(results, result)

		i := len(stack) - 1
		output[i] = data[pattern[i]][stack[i]-1]
		stack[i]--

		for ; i > 0 && stack[i] == 0; i-- {
			stack[i-1]--
			stack[i] = len(data[pattern[i]])
			output[i] = data[pattern[i]][0]
		}

		if stack[0] == 0 {
			break
		}

		output[0] = data[pattern[0]][stack[0]-1]
		stack[0]--
	}

	return results
}

// getPositions 获取 payload 设置位置字符
func getPositions(req string) []string {
	re := regexp.MustCompile(`§(.*?)§`)          // 定义正则表达式 *? 表示非贪婪匹配模式，即尽可能少地匹配。
	matches := re.FindAllStringSubmatch(req, -1) // 查找所有匹配项

	var result []string
	for _, match := range matches {
		result = append(result, "§"+match[1]+"§")
	}

	return result
}

// processing 对 payload 进行的处理
func processing(payload, rule string) string {
	switch rule {
	case "MD5": // MD5 加密处理
		return util.Md5(payload)
	}
	return payload
}

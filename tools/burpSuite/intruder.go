package burpSuite

import (
	"context"
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

// sniper 模式 设置每个 payload 位置都使用相同的 fuzz 文本 跑一遍
func sniper(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	// 定义分隔符的回调函数
	splitFunc := func(r rune) bool {
		return r == '\r' || r == '\n'
	}

	// 使用 FieldsFunc() 函数分割字符串
	payloads = strings.FieldsFunc(payloads[0], splitFunc)

	var id = 0

	for _, position := range positions {
		for _, payload := range payloads {
			request := req // req 不能改变
			// 这里是根据payload位置来进行对应的处理
			payload = processing(payload, rules[0])
			request = strings.Replace(request, position, payload, 1)
			// 去除其他位置的 §
			request = strings.ReplaceAll(request, "§", "")
			ch <- struct{}{}
			id += 1
			go func(request, payload string, id int) {
				intruderRes := IntruderRes{
					Id:      id,
					Payload: []string{payload},
				}
				resp, err := httpx.Raw(request, target)

				<-ch
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

				IntruderMap[uuid].WriteMap(id, &HTTPBody{
					TargetUrl: target,
					Request:   resp.RequestDump,
					Response:  resp.ResponseDump,
				})

			}(request, payload, id)
		}
	}
	close(ch)
}

// batteringRam 模式设置的所有 payload 位置使用同一份 fuzz 文本
func batteringRam(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {
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
		for _, position := range positions {
			payload = processing(payload, rules[0])
			request = strings.Replace(request, position, payload, 1)
		}

		ch <- struct{}{}

		go func(request, payload string, i int) {
			intruderRes := IntruderRes{
				Id:      i,
				Payload: []string{payload},
			}
			resp, err := httpx.Raw(request, target)

			<-ch
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

// pitchfork 模式 ，payload 一一对应
func pitchfork(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	// 定义分隔符的回调函数
	splitFunc := func(r rune) bool {
		return r == '\r' || r == '\n'
	}

	words := make([][]string, 0, len(payloads))
	for i, w := range payloads {
		// 使用 FieldsFunc() 函数分割字符串
		words[i] = strings.FieldsFunc(w, splitFunc)
	}

	payloadss := make([][]string, len(words[0]))

	for k := range words[0] {
		payloadss[k] = []string{words[0][k]}
		for i := range words {
			if i > 0 {
				payloadss[k] = append(payloadss[k], words[i][k])
			}
		}

	}

	for i, payload := range payloadss {
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
			<-ch
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

// clusterBomb 每个 payload 位置使用不同的 fuzz 文本
func clusterBomb(target string, req string, payloads []string, rules []string, uuid string, ctx context.Context) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	//  获取全部组合的可能性 payloads
	result := combinations(payloads, rules)

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
			<-ch
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

// combinations 获取所有 clusterBomb 模式payload组合结果的可能性,并且保证顺序不乱
func combinations(words []string, rules []string) [][]string {
	// 定义分隔符的回调函数
	splitFunc := func(r rune) bool {
		return r == '\r' || r == '\n'
	}

	payloads := make([][]string, 0, len(words))
	for _, w := range words {
		// 使用 FieldsFunc() 函数分割字符串
		payloads = append(payloads, strings.FieldsFunc(w, splitFunc))
	}

	results := backtrack(make([]string, 0, len(rules)), 0, rules, payloads, rules)
	for i, result := range results {
		reorderedResult := make([]string, len(result))
		for j := range rules {
			for _, s := range result {
				if contains(payloads[j], s) {
					reorderedResult[j] = s
					break
				}
			}
		}
		results[i] = reorderedResult
	}
	return results
}

func backtrack(cur []string, index int, keys []string, values [][]string, orderedKeys []string) [][]string {
	var results [][]string // 保存结果集合
	if index == len(keys) {
		return [][]string{append([]string{}, cur...)} // 返回当前排列组合
	}
	for _, s := range values[index] {
		cur = append(cur, s)
		subResults := backtrack(cur, index+1, keys, values, orderedKeys)
		results = append(results, subResults...)
		cur = cur[:len(cur)-1]
	}
	return results // 返回所有排列组合
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
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

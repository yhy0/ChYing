package mitmproxy

import (
	"regexp"
	"strings"
	"time"

	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/decoder"
	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2023/5/7
  @desc: passiveProxy 的 Intruder 模式相关实现
**/

type PayloadItem struct {
	ID         int64            `json:"id"`         // 条目ID
	Type       string           `json:"type"`       // 条目类型
	Items      []string         `json:"items"`      // 条目内容数组
	Processing ProcessingConfig `json:"processing"` // 处理配置，包含规则和编码设置
}

// ProcessingConfig 载荷处理配置
type ProcessingConfig struct {
	Rules    []ProcessingRule `json:"rules"`    // 处理规则数组
	Encoding EncodingConfig   `json:"encoding"` // 编码配置
}

// EncodingConfig 编码配置
type EncodingConfig struct {
	Enabled      bool   `json:"enabled"`      // 是否启用编码
	URLEncode    bool   `json:"urlEncode"`    // 是否URL编码
	CharacterSet string `json:"characterSet"` // 字符集
}

// ProcessingRule 载荷处理规则
type ProcessingRule struct {
	ID     int64                  `json:"id"`     // 规则ID
	Type   string                 `json:"type"`   // 规则类型
	Config map[string]interface{} `json:"config"` // 规则配置
}

func Intruder(target string, req string, payloads []PayloadItem, attackType string, uuid string) {
	switch attackType {
	case "sniper":
		sniper(target, req, payloads, uuid)
	case "battering-ram":
		batteringRam(target, req, payloads, uuid)
	case "pitchfork":
		pitchfork(target, req, payloads, uuid)
	case "cluster-bomb":
		clusterBomb(target, req, payloads, uuid)
	}
}

// sniper 模式 设置每个 payload 位置都使用相同的 fuzz 文本 跑一遍
func sniper(target string, req string, payloadItem []PayloadItem, uuid string) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	payloads := []string{}
	for _, payload := range payloadItem {
		payloads = append(payloads, payload.Items...)
	}

	var id = 0
	// 对每个位置
	for posIndex, position := range positions {
		// 对每个payload
		for _, payload := range payloads {
			_req := req // req 不能改变
			// 处理payload
			processedPayload := processing(payload, payloadItem[0].Processing.Rules)

			// 构建完整的 payload 数组，包含所有位置的值
			allPayloads := make([]string, len(positions))
			for i, pos := range positions {
				if i == posIndex {
					// 当前位置使用实际的 payload
					allPayloads[i] = processedPayload
				} else {
					// 其他位置使用原始值（去掉§标记）
					originalContent := strings.TrimPrefix(strings.TrimSuffix(pos, "§"), "§")
					allPayloads[i] = originalContent
				}
			}

			// 只替换当前位置的标记
			_req = strings.Replace(_req, position, processedPayload, 1)

			// 对于其他位置，恢复原始值（去掉§标记，保留内容）
			for i, pos := range positions {
				if i != posIndex {
					// 提取§中间的内容
					originalContent := strings.TrimPrefix(strings.TrimSuffix(pos, "§"), "§")
					_req = strings.Replace(_req, pos, originalContent, 1)
				}
			}

			ch <- struct{}{}
			id += 1
			go func(_request string, _allPayloads []string, _i int) {
				startTime := time.Now()
				intruderRes := IntruderRes{
					Id:        int64(_i),
					Payload:   _allPayloads, // 使用包含所有位置值的数组
					Length:    0,
					Timestamp: time.Now().UnixMilli(),
				}
				resp, err := httpx.Raw(_request, target)
				duration := time.Since(startTime)

				<-ch
				smap := GetOrCreateIntruderMap(uuid)

				if err != nil {
					logging.Logger.Errorln(err)
					intruderRes.TimeMs = int(duration.Milliseconds())
					EventDataChan <- &EventData{
						Name: uuid,
						Data: intruderRes,
					}
					// 即使请求失败也存储HTTPBody，包含原始请求和错误信息
					smap.WriteMap(int64(_i), HTTPBody{
						Id:          int64(_i),
						TargetUrl:   target,
						RequestRaw:  _request,
						ResponseRaw: "Error: " + err.Error(),
					})
					return
				}

				intruderRes.Status = resp.StatusCode
				intruderRes.Length = resp.ContentLength
				intruderRes.TimeMs = int(duration.Milliseconds())

				EventDataChan <- &EventData{
					Name: uuid,
					Data: intruderRes,
				}

				smap.WriteMap(int64(_i), HTTPBody{
					Id:                int64(_i),
					TargetUrl:         target,
					RequestRaw:        resp.RequestDump,
					ResponseRaw:       resp.ResponseDump,
					ResponseTimestamp: resp.ServerDurationMs,
					ServerDurationMs:  resp.ServerDurationMs,
				})
			}(_req, allPayloads, id)
		}
	}
	close(ch)
}

// batteringRam 模式设置的所有 payload 位置使用同一份 fuzz 文本
func batteringRam(target string, req string, payloadItem []PayloadItem, uuid string) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	payloads := []string{}
	for _, payload := range payloadItem {
		payloads = append(payloads, payload.Items...)
	}

	for i, payload := range payloads {
		request := req // req 不能改变
		// 这里是根据payload位置来进行对应的处理
		for _, position := range positions {
			payload = processing(payload, payloadItem[0].Processing.Rules)
			request = strings.Replace(request, position, payload, 1)
		}

		ch <- struct{}{}

		go func(_request, _payload string, _i int) {
			startTime := time.Now()
			intruderRes := IntruderRes{
				Id:        int64(_i),
				Payload:   []string{_payload},
				Length:    0,
				Timestamp: time.Now().UnixMilli(),
			}
			resp, err := httpx.Raw(_request, target)
			duration := time.Since(startTime)

			<-ch
			smap := GetOrCreateIntruderMap(uuid)

			if err != nil {
				intruderRes.TimeMs = int(duration.Milliseconds())
				EventDataChan <- &EventData{
					Name: uuid,
					Data: intruderRes,
				}
				// 即使请求失败也存储HTTPBody，包含原始请求和错误信息
				smap.WriteMap(int64(_i), HTTPBody{
					Id:          int64(_i),
					TargetUrl:   target,
					RequestRaw:  _request,
					ResponseRaw: "Error: " + err.Error(),
				})
				return
			}

			intruderRes.Status = resp.StatusCode
			intruderRes.Length = resp.ContentLength
			intruderRes.TimeMs = int(duration.Milliseconds())

			EventDataChan <- &EventData{
				Name: uuid,
				Data: intruderRes,
			}

			smap.WriteMap(int64(_i), HTTPBody{
				Id:                int64(_i),
				TargetUrl:         target,
				RequestRaw:        resp.RequestDump,
				ResponseRaw:       resp.ResponseDump,
				ResponseTimestamp: resp.ServerDurationMs,
				ServerDurationMs:  resp.ServerDurationMs,
			})

		}(request, payload, i)
	}
	close(ch)
}

// pitchfork 模式 ，payload 一一对应
func pitchfork(target string, req string, payloadItem []PayloadItem, uuid string) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	words := make([][]string, len(payloadItem)) // 使用 len 而不是 0 作为切片长度
	for i, item := range payloadItem {
		words[i] = item.Items
	}

	payloads := make([][]string, len(words[0]))

	for k := range words[0] {
		payloads[k] = []string{words[0][k]}
		for i := range words {
			if i > 0 {
				payloads[k] = append(payloads[k], words[i][k])
			}
		}
	}

	for i, payload := range payloads {
		request := req // req 不能改变
		// 这里是根据payload位置来进行对应的处理
		for j, position := range positions {
			_payload := processing(payload[j], payloadItem[j].Processing.Rules)
			request = strings.Replace(request, position, _payload, 1)
		}
		ch <- struct{}{}
		go func(_request string, _payload []string, _i int) {
			startTime := time.Now()
			intruderRes := IntruderRes{
				Id:        int64(_i),
				Payload:   _payload,
				Length:    0,
				Timestamp: time.Now().UnixMilli(),
			}
			resp, err := httpx.Raw(_request, target)
			duration := time.Since(startTime)
			<-ch
			smap := GetOrCreateIntruderMap(uuid)

			if err != nil {
				intruderRes.TimeMs = int(duration.Milliseconds())
				EventDataChan <- &EventData{
					Name: uuid,
					Data: intruderRes,
				}
				// 即使请求失败也存储HTTPBody，包含原始请求和错误信息
				smap.WriteMap(int64(_i), HTTPBody{
					Id:          int64(_i),
					TargetUrl:   target,
					RequestRaw:  _request,
					ResponseRaw: "Error: " + err.Error(),
				})
				return
			}

			intruderRes.Status = resp.StatusCode
			intruderRes.Length = resp.ContentLength
			intruderRes.TimeMs = int(duration.Milliseconds())

			EventDataChan <- &EventData{
				Name: uuid,
				Data: intruderRes,
			}

			smap.WriteMap(int64(_i), HTTPBody{
				Id:                int64(_i),
				TargetUrl:         target,
				RequestRaw:        resp.RequestDump,
				ResponseRaw:       resp.ResponseDump,
				ResponseTimestamp: resp.ServerDurationMs,
				ServerDurationMs:  resp.ServerDurationMs,
			})

		}(request, payload, i)
	}

	close(ch)
}

// clusterBomb 每个 payload 位置使用不同的 fuzz 文本
func clusterBomb(target string, req string, payloadItem []PayloadItem, uuid string) {
	positions := getPositions(req)
	ch := make(chan struct{}, 20)

	// 生成所有组合
	combinations := GenerateCombinations(payloadItem)

	for i, payload := range combinations {
		request := req // req 不能改变
		// 这里是根据payload位置来进行对应的处理
		for j, position := range positions {
			request = strings.Replace(request, position, payload[j], 1)
		}
		ch <- struct{}{}

		go func(_request string, _payload []string, _i int) {
			startTime := time.Now()
			intruderRes := IntruderRes{
				Id:        int64(_i),
				Payload:   _payload,
				Length:    0,
				Timestamp: time.Now().UnixMilli(),
			}
			resp, err := httpx.Raw(_request, target)
			duration := time.Since(startTime)
			<-ch
			smap := GetOrCreateIntruderMap(uuid)

			if err != nil {
				logging.Logger.Errorln(err, "\r\n", _request)
				intruderRes.TimeMs = int(duration.Milliseconds())
				EventDataChan <- &EventData{
					Name: uuid,
					Data: intruderRes,
				}
				// 即使请求失败也存储HTTPBody，包含原始请求和错误信息
				smap.WriteMap(int64(_i), HTTPBody{
					Id:          int64(_i),
					TargetUrl:   target,
					RequestRaw:  _request,
					ResponseRaw: "Error: " + err.Error(),
				})
				return
			}

			intruderRes.Status = resp.StatusCode
			intruderRes.Length = resp.ContentLength
			intruderRes.TimeMs = int(duration.Milliseconds())

			EventDataChan <- &EventData{
				Name: uuid,
				Data: intruderRes,
			}

			logging.Logger.Debugln("clusterBomb:", target, resp.StatusCode, uuid, _i)
			smap.WriteMap(int64(_i), HTTPBody{
				Id:                int64(_i),
				TargetUrl:         target,
				RequestRaw:        resp.RequestDump,
				ResponseRaw:       resp.ResponseDump,
				ResponseTimestamp: resp.ServerDurationMs,
				ServerDurationMs:  resp.ServerDurationMs,
			})

		}(request, payload, i)
	}
	close(ch)
}

// GenerateCombinations 生成所有可能的组合（包含处理逻辑）
func GenerateCombinations(items []PayloadItem) [][]string {
	// 先处理每个Item中的字符串
	var processedItems [][]string
	for _, item := range items {
		var processed []string
		for _, s := range item.Items {
			// 对每个字符串应用处理函数
			processed = append(processed, processing(s, item.Processing.Rules))
		}
		processedItems = append(processedItems, processed)
	}

	// 如果没有数据，返回空组合
	if len(processedItems) == 0 {
		return [][]string{}
	}

	// 使用回溯算法生成所有组合
	var result [][]string
	backtrack(processedItems, 0, []string{}, &result)
	return result
}

// 回溯算法实现
func backtrack(allItems [][]string, index int, current []string, result *[][]string) {
	// 如果已经处理完所有Items切片，将当前组合加入结果
	if index == len(allItems) {
		combination := make([]string, len(current))
		copy(combination, current)
		*result = append(*result, combination)
		return
	}

	// 遍历当前Items切片的所有元素
	for _, item := range allItems[index] {
		// 添加当前元素到组合中
		current = append(current, item)
		// 递归处理下一个Items切片
		backtrack(allItems, index+1, current, result)
		// 回溯，移除最后添加的元素
		current = current[:len(current)-1]
	}
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

// processing 对 payload 进行的处理，支持链式处理多个规则
func processing(payload string, rules []ProcessingRule) string {
	if rules == nil || len(rules) == 0 {
		return payload
	}

	result := payload
	for _, rule := range rules {
		result = applyProcessingRule(result, rule)
	}
	return result
}

// applyProcessingRule 应用单个处理规则
func applyProcessingRule(payload string, rule ProcessingRule) string {
	switch rule.Type {
	case "add-prefix":
		if text, ok := rule.Config["text"].(string); ok {
			return text + payload
		}
	case "add-suffix":
		if text, ok := rule.Config["text"].(string); ok {
			return payload + text
		}
	case "match-replace":
		return applyMatchReplace(payload, rule.Config)
	case "substring":
		return applySubstring(payload, rule.Config)
	case "reverse-substring":
		return applyReverseSubstring(payload, rule.Config)
	case "modify-case":
		return applyModifyCase(payload, rule.Config)
	case "encode":
		return applyEncode(payload, rule.Config)
	case "decode":
		return applyDecode(payload, rule.Config)
	case "hash":
		return applyHash(payload, rule.Config)
	}
	return payload
}

// applyMatchReplace 应用匹配替换规则
func applyMatchReplace(payload string, config map[string]interface{}) string {
	match, matchOk := config["match"].(string)
	replace, replaceOk := config["replace"].(string)
	useRegex, _ := config["regex"].(bool)

	if !matchOk || !replaceOk {
		return payload
	}

	if useRegex {
		re, err := regexp.Compile(match)
		if err != nil {
			return payload
		}
		return re.ReplaceAllString(payload, replace)
	} else {
		return strings.ReplaceAll(payload, match, replace)
	}
}

// applySubstring 应用子字符串规则
func applySubstring(payload string, config map[string]interface{}) string {
	startIndex, startOk := config["startIndex"]
	lengthValue, lengthOk := config["length"]

	if !startOk || !lengthOk {
		return payload
	}

	// 处理不同类型的数字输入
	var start, length int
	switch v := startIndex.(type) {
	case float64:
		start = int(v)
	case int:
		start = v
	default:
		return payload
	}

	switch v := lengthValue.(type) {
	case float64:
		length = int(v)
	case int:
		length = v
	default:
		return payload
	}

	if start < 0 || start >= len(payload) {
		return payload
	}

	end := start + length
	if end > len(payload) {
		end = len(payload)
	}

	return payload[start:end]
}

// applyReverseSubstring 应用反向子字符串规则
func applyReverseSubstring(payload string, config map[string]interface{}) string {
	// 获取子字符串后，反转它
	result := applySubstring(payload, config)
	runes := []rune(result)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// applyModifyCase 应用大小写修改规则
func applyModifyCase(payload string, config map[string]interface{}) string {
	caseType, ok := config["type"].(string)
	if !ok {
		return payload
	}

	switch caseType {
	case "lowercase":
		return strings.ToLower(payload)
	case "uppercase":
		return strings.ToUpper(payload)
	case "capitalize":
		if len(payload) == 0 {
			return payload
		}
		return strings.ToUpper(payload[:1]) + strings.ToLower(payload[1:])
	default:
		return payload
	}
}

// applyEncode 应用编码规则
func applyEncode(payload string, config map[string]interface{}) string {
	encodeType, ok := config["type"].(string)
	if !ok {
		return payload
	}

	switch encodeType {
	case "url-key", "url-all":
		return decoder.EncodeURL(payload)
	case "url-unicode":
		return decoder.EncodeURL(payload) // 简化实现，可以后续扩展
	case "html-key", "html-all", "html-numeric", "html-hex":
		return decoder.EncodeHTML(payload)
	case "base64":
		return decoder.EncodeBase64(payload)
	case "ascii-hex":
		return decoder.EncodeHex(payload)
	case "unicode-escape":
		return decoder.EncodeUnicode(payload)
	default:
		return payload
	}
}

// applyDecode 应用解码规则
func applyDecode(payload string, config map[string]interface{}) string {
	decodeType, ok := config["type"].(string)
	if !ok {
		return payload
	}

	switch decodeType {
	case "url":
		return decoder.DecodeURL(payload)
	case "html":
		return decoder.DecodeHTML(payload)
	case "base64":
		return decoder.DecodeBase64(payload)
	case "ascii-hex":
		return decoder.DecodeHex(payload)
	case "unicode-unescape":
		return decoder.DecodeUnicode(payload)
	default:
		return payload
	}
}

// applyHash 应用哈希规则
func applyHash(payload string, config map[string]interface{}) string {
	hashType, ok := config["type"].(string)
	if !ok {
		hashType = "md5" // 默认使用 MD5
	}

	switch hashType {
	case "md5":
		return decoder.Md5(payload)
	case "sha1":
		return decoder.Sha1(payload)
	case "sha256":
		return decoder.Sha256(payload)
	default:
		return decoder.Md5(payload)
	}
}

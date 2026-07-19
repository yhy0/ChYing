package mcpserver

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

const maxIntruderRequests = 1000
const intruderConcurrency = 20

// intruderResult 单个 Intruder 请求的结果
type intruderResult struct {
	ID        int64    `json:"id"`
	Payload   []string `json:"payload"`
	Status    int      `json:"status"`
	Length    int      `json:"length"`
	TimeMs    int      `json:"time_ms"`
	HistoryID int64    `json:"history_id,omitempty"`
	Hid       int64    `json:"hid,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// --- run_intruder ---

func runIntruderTool() mcp.Tool {
	return mcp.NewTool("run_intruder",
		mcp.WithDescription(`Run an Intruder attack (synchronous). Sends multiple requests with different payloads to test for vulnerabilities.

Use § markers to indicate payload positions in the raw_request. For example:
GET /api/users/§1§ HTTP/1.1
Host: example.com

Attack types:
- sniper: Uses one payload set, iterates through each position one at a time
- battering-ram: Uses one payload set, replaces all positions simultaneously
- pitchfork: Uses multiple payload sets, one per position, iterates in parallel
- cluster-bomb: Uses multiple payload sets, tests all combinations

Maximum 1000 request combinations. If exceeded, split into smaller batches.`),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target URL including scheme (e.g., 'https://example.com')"),
		),
		mcp.WithString("raw_request",
			mcp.Required(),
			mcp.Description("Raw HTTP request with § markers for payload positions"),
		),
		mcp.WithString("payloads",
			mcp.Required(),
			mcp.Description("JSON array of payload sets. Each set is an array of strings. Example: [[\"admin\",\"test\",\"user\"]] for one position, or [[\"admin\",\"test\"],[\"123\",\"password\"]] for two positions"),
		),
		mcp.WithString("attack_type",
			mcp.Required(),
			mcp.Description("Attack type: 'sniper', 'battering-ram', 'pitchfork', or 'cluster-bomb'"),
			mcp.Enum("sniper", "battering-ram", "pitchfork", "cluster-bomb"),
		),
		mcp.WithString("session_id",
			mcp.Description("Optional scan session ID used to attribute every generated request to an agent task"),
		),
	)
}

func handleRunIntruder(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, err := req.RequireString("target")
	if err != nil {
		return errorResult("target is required"), nil
	}

	rawRequest, err := req.RequireString("raw_request")
	if err != nil {
		return errorResult("raw_request is required"), nil
	}

	payloadsStr, err := req.RequireString("payloads")
	if err != nil {
		return errorResult("payloads is required"), nil
	}

	attackType, err := req.RequireString("attack_type")
	if err != nil {
		return errorResult("attack_type is required"), nil
	}

	projectID, sessionID, attributionErr := sessionAttribution(req.GetString("session_id", ""))
	if attributionErr != nil {
		return errorResult("%v", attributionErr), nil
	}

	// 解析 payloads JSON
	var payloadSets [][]string
	if parseErr := parseJSON(payloadsStr, &payloadSets); parseErr != nil {
		return errorResult("invalid payloads JSON: %v", parseErr), nil
	}

	if len(payloadSets) == 0 {
		return errorResult("payloads cannot be empty"), nil
	}

	// 构建 PayloadItem 列表
	payloadItems := make([]mitmproxy.PayloadItem, len(payloadSets))
	for i, set := range payloadSets {
		payloadItems[i] = mitmproxy.PayloadItem{
			ID:    int64(i + 1),
			Type:  "simple",
			Items: set,
		}
	}

	// 获取位置标记
	positions := mitmproxy.GetPositions(rawRequest)
	if len(positions) == 0 {
		return errorResult("no payload positions found. Use § markers to indicate positions (e.g., §value§)"), nil
	}

	// 生成请求列表
	requests := generateIntruderRequests(rawRequest, positions, payloadItems, attackType)
	if requests == nil {
		return errorResult("unsupported attack type: %s", attackType), nil
	}

	if len(requests) > maxIntruderRequests {
		return errorResult("too many request combinations (%d). Maximum is %d. Please split into smaller batches.", len(requests), maxIntruderRequests), nil
	}

	// 同步执行所有请求
	results := executeIntruderRequests(ctx, target, requests, projectID, sessionID)

	return jsonResult(map[string]interface{}{
		"project_id":    projectID,
		"session_id":    sessionID,
		"request_count": len(results),
		"results":       results,
	}), nil
}

// intruderRequest 表示一个待执行的 Intruder 请求
type intruderRequest struct {
	ID      int64
	Request string
	Payload []string
}

// generateIntruderRequests 根据攻击类型生成请求列表
func generateIntruderRequests(rawReq string, positions []string, payloadItems []mitmproxy.PayloadItem, attackType string) []intruderRequest {
	switch attackType {
	case "sniper":
		return generateSniperRequests(rawReq, positions, payloadItems)
	case "battering-ram":
		return generateBatteringRamRequests(rawReq, positions, payloadItems)
	case "pitchfork":
		return generatePitchforkRequests(rawReq, positions, payloadItems)
	case "cluster-bomb":
		return generateClusterBombRequests(rawReq, positions, payloadItems)
	default:
		return nil
	}
}

func generateSniperRequests(rawReq string, positions []string, payloadItems []mitmproxy.PayloadItem) []intruderRequest {
	var allPayloads []string
	for _, item := range payloadItems {
		allPayloads = append(allPayloads, item.Items...)
	}

	var requests []intruderRequest
	id := int64(0)
	for posIndex, position := range positions {
		for _, payload := range allPayloads {
			id++
			r := rawReq
			processedPayload := mitmproxy.Processing(payload, nil)

			payloadArray := make([]string, len(positions))
			for i, pos := range positions {
				if i == posIndex {
					payloadArray[i] = processedPayload
				} else {
					payloadArray[i] = strings.TrimPrefix(strings.TrimSuffix(pos, "§"), "§")
				}
			}

			r = strings.Replace(r, position, processedPayload, 1)
			for i, pos := range positions {
				if i != posIndex {
					originalContent := strings.TrimPrefix(strings.TrimSuffix(pos, "§"), "§")
					r = strings.Replace(r, pos, originalContent, 1)
				}
			}

			requests = append(requests, intruderRequest{ID: id, Request: r, Payload: payloadArray})
		}
	}
	return requests
}

func generateBatteringRamRequests(rawReq string, positions []string, payloadItems []mitmproxy.PayloadItem) []intruderRequest {
	var allPayloads []string
	for _, item := range payloadItems {
		allPayloads = append(allPayloads, item.Items...)
	}

	var requests []intruderRequest
	for i, payload := range allPayloads {
		r := rawReq
		for _, pos := range positions {
			r = strings.Replace(r, pos, payload, 1)
		}
		requests = append(requests, intruderRequest{ID: int64(i + 1), Request: r, Payload: []string{payload}})
	}
	return requests
}

func generatePitchforkRequests(rawReq string, positions []string, payloadItems []mitmproxy.PayloadItem) []intruderRequest {
	if len(payloadItems) == 0 {
		return nil
	}

	minLen := len(payloadItems[0].Items)
	for _, item := range payloadItems[1:] {
		if len(item.Items) < minLen {
			minLen = len(item.Items)
		}
	}

	var requests []intruderRequest
	for i := 0; i < minLen; i++ {
		r := rawReq
		var payloads []string
		for j, pos := range positions {
			if j < len(payloadItems) {
				payload := payloadItems[j].Items[i]
				r = strings.Replace(r, pos, payload, 1)
				payloads = append(payloads, payload)
			}
		}
		requests = append(requests, intruderRequest{ID: int64(i + 1), Request: r, Payload: payloads})
	}
	return requests
}

func generateClusterBombRequests(rawReq string, positions []string, payloadItems []mitmproxy.PayloadItem) []intruderRequest {
	combinations := mitmproxy.GenerateCombinations(payloadItems)

	var requests []intruderRequest
	for i, combo := range combinations {
		r := rawReq
		for j, pos := range positions {
			if j < len(combo) {
				r = strings.Replace(r, pos, combo[j], 1)
			}
		}
		requests = append(requests, intruderRequest{ID: int64(i + 1), Request: r, Payload: combo})
	}
	return requests
}

// executeIntruderRequests 并发执行 Intruder 请求并收集结果
func executeIntruderRequests(ctx context.Context, target string, requests []intruderRequest, projectID, sessionID string) []intruderResult {
	results := make([]intruderResult, len(requests))
	sem := make(chan struct{}, intruderConcurrency)
	var wg sync.WaitGroup

	for i, ir := range requests {
		select {
		case <-ctx.Done():
			// context 取消，等待已发出的请求完成后返回
			wg.Wait()
			return results
		default:
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, r intruderRequest) {
			defer wg.Done()
			defer func() { <-sem }()

			startTime := time.Now()
			res := intruderResult{
				ID:      r.ID,
				Payload: r.Payload,
			}

			resp, err := httpx.Raw(r.Request, target)
			duration := time.Since(startTime)
			res.TimeMs = int(duration.Milliseconds())

			if err != nil {
				logging.Logger.Debugf("intruder request %d failed: %v", r.ID, err)
				res.Error = err.Error()
			} else {
				res.Status = resp.StatusCode
				res.Length = resp.ContentLength
				reference := saveRepeaterHistory(target, r.Request, resp, projectID, sessionID)
				res.HistoryID = reference.ID
				res.Hid = reference.Hid
			}

			results[idx] = res
		}(i, ir)
	}

	wg.Wait()
	return results
}

// parseJSON 解析 JSON 字符串
func parseJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

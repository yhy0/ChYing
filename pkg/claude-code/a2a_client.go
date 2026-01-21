package claudecode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2aclient"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/21
   @desc A2A (Agent-to-Agent) 客户端实现
**/

// A2AClientConfig A2A 客户端配置
type A2AClientConfig struct {
	AgentURL  string            `json:"agent_url" yaml:"agent_url"`   // Agent URL (e.g., https://my-agent.com)
	Headers   map[string]string `json:"headers" yaml:"headers"`       // 自定义请求头
	Timeout   int               `json:"timeout" yaml:"timeout"`       // 超时时间（秒）
	EnableSSE bool              `json:"enable_sse" yaml:"enable_sse"` // 是否启用 SSE 流式响应
}

// A2AClient A2A 客户端
type A2AClient struct {
	config    *A2AClientConfig
	agentCard *a2a.AgentCard
	client    *a2aclient.Client
	mu        sync.RWMutex
	connected bool

	// 会话管理: sessionID -> contextId 映射
	sessionContexts map[string]string
	contextMu       sync.RWMutex
}

// NewA2AClient 创建 A2A 客户端
func NewA2AClient(config *A2AClientConfig) *A2AClient {
	if config == nil {
		config = &A2AClientConfig{}
	}
	if config.Timeout == 0 {
		config.Timeout = 300 // 默认 5 分钟
	}
	return &A2AClient{
		config:          config,
		sessionContexts: make(map[string]string),
	}
}

// Connect 连接到 A2A Agent
func (c *A2AClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.config.AgentURL == "" {
		return fmt.Errorf("agent URL is required")
	}

	// 规范化 URL
	agentURL := c.config.AgentURL
	if !strings.HasPrefix(agentURL, "http://") && !strings.HasPrefix(agentURL, "https://") {
		agentURL = "https://" + agentURL
	}

	// 获取 Agent Card
	cardURL := strings.TrimSuffix(agentURL, "/") + "/.well-known/agent.json"
	logging.Logger.Infof("Fetching Agent Card from: %s", cardURL)

	card, err := c.fetchAgentCard(ctx, cardURL)
	if err != nil {
		return fmt.Errorf("failed to fetch agent card: %w", err)
	}

	c.agentCard = card
	logging.Logger.Infof("Connected to A2A Agent: %s (%s)", card.Name, card.Description)

	// 创建 A2A 客户端
	client, err := a2aclient.NewFromCard(ctx, card)
	if err != nil {
		return fmt.Errorf("failed to create A2A client: %w", err)
	}

	c.client = client
	c.connected = true

	return nil
}

// fetchAgentCard 获取 Agent Card
func (c *A2AClient) fetchAgentCard(ctx context.Context, cardURL string) (*a2a.AgentCard, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", cardURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加自定义 headers
	for k, v := range c.config.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "application/json")

	// 使用配置的超时时间
	timeout := time.Duration(c.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("failed to fetch agent card: status %d, body: %s", resp.StatusCode, string(body))
	}

	var card a2a.AgentCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode agent card: %w", err)
	}

	return &card, nil
}

// IsConnected 检查是否已连接
func (c *A2AClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetAgentCard 获取 Agent Card
func (c *A2AClient) GetAgentCard() *a2a.AgentCard {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.agentCard
}

// getOrCreateContextID 获取或创建会话的 contextId
func (c *A2AClient) getOrCreateContextID(sessionID string) string {
	c.contextMu.Lock()
	defer c.contextMu.Unlock()

	if contextID, exists := c.sessionContexts[sessionID]; exists {
		return contextID
	}

	// 创建新的 contextId
	contextID := a2a.NewContextID()
	c.sessionContexts[sessionID] = contextID
	return contextID
}

// SendMessage 发送消息到 A2A Agent
func (c *A2AClient) SendMessage(ctx context.Context, sessionID, message string, eventChan chan<- StreamEvent) error {
	c.mu.RLock()
	if !c.connected || c.client == nil {
		c.mu.RUnlock()
		return fmt.Errorf("A2A client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	// 获取或创建 contextId 用于会话关联
	contextID := c.getOrCreateContextID(sessionID)

	// 创建消息，包含 contextId
	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.TextPart{Text: message})
	msg.ContextID = contextID

	// 发送消息参数
	params := &a2a.MessageSendParams{
		Message: msg,
	}

	// 配置：阻塞等待结果，获取历史记录
	blocking := true
	historyLength := 10 // 获取最近 10 条历史消息
	params.Config = &a2a.MessageSendConfig{
		Blocking:      &blocking,
		HistoryLength: &historyLength,
	}

	logging.Logger.Infof("Sending message to A2A Agent (contextId=%s): %s", contextID, truncateString(message, 100))

	// 发送消息并处理响应
	go func() {
		defer close(eventChan)

		var sendErr error

		// 使用流式消息发送
		if c.config.EnableSSE {
			sendErr = c.sendStreamingMessage(ctx, client, params, sessionID, eventChan)
		} else {
			sendErr = c.sendBlockingMessage(ctx, client, params, sessionID, eventChan)
		}

		// 只有在没有错误时才发送 done 事件
		// 错误情况下，sendXxxMessage 已经发送了 error 事件
		if sendErr == nil {
			eventChan <- StreamEvent{
				Type:      "done",
				SessionID: sessionID,
			}
		}
	}()

	return nil
}

// sendBlockingMessage 发送阻塞消息
func (c *A2AClient) sendBlockingMessage(ctx context.Context, client *a2aclient.Client, params *a2a.MessageSendParams, sessionID string, eventChan chan<- StreamEvent) error {
	result, err := client.SendMessage(ctx, params)
	if err != nil {
		logging.Logger.Errorf("A2A SendMessage error: %v", err)
		eventChan <- StreamEvent{
			Type:      "error",
			Error:     err.Error(),
			SessionID: sessionID,
		}
		return err
	}

	// 处理响应 - SendMessageResult 是一个接口
	c.processA2AResult(sessionID, result, eventChan)
	return nil
}

// sendStreamingMessage 发送流式消息
func (c *A2AClient) sendStreamingMessage(ctx context.Context, client *a2aclient.Client, params *a2a.MessageSendParams, sessionID string, eventChan chan<- StreamEvent) error {
	// 使用流式 API
	eventSeq := client.SendStreamingMessage(ctx, params)

	var lastErr error
	for event, err := range eventSeq {
		if err != nil {
			logging.Logger.Errorf("A2A streaming error: %v", err)
			eventChan <- StreamEvent{
				Type:      "error",
				Error:     err.Error(),
				SessionID: sessionID,
			}
			lastErr = err
			continue // 继续处理其他事件，不立即返回
		}

		c.processA2AEvent(sessionID, event, eventChan)
	}

	return lastErr
}

// processA2AResult 处理 A2A 结果
func (c *A2AClient) processA2AResult(sessionID string, result a2a.SendMessageResult, eventChan chan<- StreamEvent) {
	if result == nil {
		return
	}

	// SendMessageResult 是一个 Event 接口
	c.processA2AEvent(sessionID, result, eventChan)
}

// processA2AEvent 处理 A2A 事件
func (c *A2AClient) processA2AEvent(sessionID string, event a2a.Event, eventChan chan<- StreamEvent) {
	if event == nil {
		return
	}

	// 根据事件类型处理
	switch e := event.(type) {
	case *a2a.TaskStatusUpdateEvent:
		logging.Logger.Debugf("A2A Task status update: %s", e.Status.State)
		// 状态更新可以作为文本发送
		if e.Status.Message != nil {
			// Message 是 *a2a.Message 类型，处理其中的 parts
			c.processMessage(sessionID, e.Status.Message, eventChan)
		}

	case *a2a.TaskArtifactUpdateEvent:
		// 处理产出物更新
		if e.Artifact != nil {
			for _, part := range e.Artifact.Parts {
				c.processPart(sessionID, part, eventChan)
			}
		}

	case *a2a.Task:
		// 完整任务响应
		logging.Logger.Debugf("A2A Task received: id=%s, status=%s", e.ID, e.Status.State)

		// 处理任务中的消息历史（只处理 agent 的回复）
		// History 是 []*Message
		for _, msg := range e.History {
			if msg != nil && msg.Role == a2a.MessageRoleAgent {
				c.processMessage(sessionID, msg, eventChan)
			}
		}

		// 处理任务产出物
		for _, artifact := range e.Artifacts {
			for _, part := range artifact.Parts {
				c.processPart(sessionID, part, eventChan)
			}
		}

	default:
		// 记录未处理的事件类型
		logging.Logger.Debugf("A2A unhandled event type: %T", event)
	}
}

// processPart 处理消息部分
func (c *A2AClient) processPart(sessionID string, part a2a.Part, eventChan chan<- StreamEvent) {
	switch p := part.(type) {
	case a2a.TextPart:
		eventChan <- StreamEvent{
			Type:      "text",
			Content:   p.Text,
			SessionID: sessionID,
		}
	case a2a.FilePart:
		// 文件部分 - 提取更多信息
		fileInfo := "[File received]"
		// FilePart.File 是 FilePartContent 接口，可能是 FileBytes 或 FileURI
		eventChan <- StreamEvent{
			Type:      "text",
			Content:   fileInfo,
			SessionID: sessionID,
		}
	case a2a.DataPart:
		// 结构化数据
		dataJSON, err := json.MarshalIndent(p.Data, "", "  ")
		if err != nil {
			dataJSON = []byte(fmt.Sprintf("%v", p.Data))
		}
		eventChan <- StreamEvent{
			Type:      "text",
			Content:   string(dataJSON),
			SessionID: sessionID,
		}
	default:
		logging.Logger.Debugf("A2A unhandled part type: %T", part)
	}
}

// processMessage 处理消息
func (c *A2AClient) processMessage(sessionID string, msg *a2a.Message, eventChan chan<- StreamEvent) {
	if msg == nil {
		return
	}

	for _, part := range msg.Parts {
		c.processPart(sessionID, part, eventChan)
	}
}

// Disconnect 断开连接
func (c *A2AClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Destroy()
	}

	c.connected = false
	c.client = nil
	c.agentCard = nil

	// 清理会话映射
	c.contextMu.Lock()
	c.sessionContexts = make(map[string]string)
	c.contextMu.Unlock()
}

// ClearSession 清除特定会话的上下文
func (c *A2AClient) ClearSession(sessionID string) {
	c.contextMu.Lock()
	defer c.contextMu.Unlock()
	delete(c.sessionContexts, sessionID)
}

// TestConnection 测试连接
func (c *A2AClient) TestConnection(ctx context.Context) error {
	if c.config.AgentURL == "" {
		return fmt.Errorf("agent URL is required")
	}

	// 尝试获取 Agent Card
	agentURL := c.config.AgentURL
	if !strings.HasPrefix(agentURL, "http://") && !strings.HasPrefix(agentURL, "https://") {
		agentURL = "https://" + agentURL
	}

	cardURL := strings.TrimSuffix(agentURL, "/") + "/.well-known/agent.json"

	card, err := c.fetchAgentCard(ctx, cardURL)
	if err != nil {
		return err
	}

	logging.Logger.Infof("A2A connection test successful: %s", card.Name)
	return nil
}

// GetAgentInfo 获取 Agent 信息（用于 UI 显示）
func (c *A2AClient) GetAgentInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.agentCard == nil {
		return nil
	}

	info := map[string]interface{}{
		"name":        c.agentCard.Name,
		"description": c.agentCard.Description,
		"version":     c.agentCard.Version,
		"url":         c.config.AgentURL,
		"connected":   c.connected,
	}

	// 添加能力信息 - AgentCapabilities 是值类型，不是指针
	caps := make([]string, 0)
	if c.agentCard.Capabilities.Streaming {
		caps = append(caps, "streaming")
	}
	if c.agentCard.Capabilities.PushNotifications {
		caps = append(caps, "pushNotifications")
	}
	if c.agentCard.Capabilities.StateTransitionHistory {
		caps = append(caps, "stateTransitionHistory")
	}
	if len(caps) > 0 {
		info["capabilities"] = caps
	}

	// 添加技能信息
	if len(c.agentCard.Skills) > 0 {
		skills := make([]map[string]string, 0, len(c.agentCard.Skills))
		for _, skill := range c.agentCard.Skills {
			skills = append(skills, map[string]string{
				"id":          skill.ID,
				"name":        skill.Name,
				"description": skill.Description,
			})
		}
		info["skills"] = skills
	}

	return info
}

// UpdateConfig 更新配置（不断开现有连接）
func (c *A2AClient) UpdateConfig(config *A2AClientConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if config != nil {
		c.config = config
	}
}

// FetchAgentCardFromURL 从 URL 获取 Agent Card（静态方法，用于测试）
func FetchAgentCardFromURL(ctx context.Context, agentURL string, headers map[string]string) (*a2a.AgentCard, error) {
	client := &A2AClient{
		config: &A2AClientConfig{
			AgentURL: agentURL,
			Headers:  headers,
			Timeout:  30,
		},
	}

	// 规范化 URL
	if !strings.HasPrefix(agentURL, "http://") && !strings.HasPrefix(agentURL, "https://") {
		agentURL = "https://" + agentURL
	}

	cardURL := strings.TrimSuffix(agentURL, "/") + "/.well-known/agent.json"
	return client.fetchAgentCard(ctx, cardURL)
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

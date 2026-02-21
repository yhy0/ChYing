package interactsh

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	interactshClient "github.com/projectdiscovery/interactsh/pkg/client"
	"github.com/projectdiscovery/interactsh/pkg/server"
	"github.com/yhy0/logging"
)

// Event 表示一次 Interactsh 交互事件
type Event struct {
	ID            string
	Protocol      string
	Method        string
	Source        string
	Destination   string
	Timestamp     time.Time
	CorrelationID string
	RawRequest    string
	RawResponse   string
	Data          map[string]any
}

// Client 是 Interactsh 的客户端，基于官方 SDK
type Client struct {
	providerID string
	serverURL  string
	token      string
	payloadURL string

	sdkClient *interactshClient.Client
	mu        sync.Mutex // 保护 sdkClient/payloadURL

	events   []Event
	eventsMu sync.Mutex // 独立锁保护 events，避免与 mu 嵌套死锁
}

// New 创建 Interactsh 客户端
func New(providerID, serverURL, token string) (*Client, error) {
	if serverURL == "" {
		serverURL = "https://oast.pro"
	}
	serverURL = strings.TrimRight(serverURL, "/")
	// SDK 接受带 scheme 或不带 scheme 的地址，内部会自行处理
	serverHost := strings.TrimPrefix(strings.TrimPrefix(serverURL, "https://"), "http://")

	return &Client{
		providerID: providerID,
		serverURL:  serverHost,
		token:      token,
	}, nil
}

// Register 注册到 Interactsh 服务器并启动内部轮询
func (c *Client) Register() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已有旧客户端，先关闭
	if c.sdkClient != nil {
		_ = c.sdkClient.StopPolling()
		_ = c.sdkClient.Close()
		c.sdkClient = nil
	}

	opts := &interactshClient.Options{
		ServerURL: c.serverURL,
		Token:     c.token,
	}

	sdkClient, err := interactshClient.New(opts)
	if err != nil {
		return "", fmt.Errorf("create interactsh client: %w", err)
	}

	c.sdkClient = sdkClient
	c.payloadURL = sdkClient.URL()

	c.eventsMu.Lock()
	c.events = nil
	c.eventsMu.Unlock()

	// SDK 内部启动后台 goroutine 轮询，交互事件通过回调收集
	// 注意：回调在单独的 goroutine 中执行，使用 eventsMu 避免与 mu 冲突
	err = sdkClient.StartPolling(5*time.Second, func(interaction *server.Interaction) {
		ev := interactionToEvent(interaction, c.providerID)
		c.eventsMu.Lock()
		c.events = append(c.events, ev)
		c.eventsMu.Unlock()
	})
	if err != nil {
		_ = sdkClient.Close()
		c.sdkClient = nil
		return "", fmt.Errorf("start interactsh polling: %w", err)
	}

	logging.Logger.Infof("Interactsh registered: %s", c.payloadURL)
	return c.payloadURL, nil
}

// Poll 拉取并返回自上次调用以来收集到的交互事件
func (c *Client) Poll() ([]Event, error) {
	c.mu.Lock()
	if c.sdkClient == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("not registered")
	}
	c.mu.Unlock()

	c.eventsMu.Lock()
	events := c.events
	c.events = nil
	c.eventsMu.Unlock()

	return events, nil
}

// Deregister 注销客户端
func (c *Client) Deregister() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sdkClient != nil {
		_ = c.sdkClient.StopPolling()
		err := c.sdkClient.Close()
		c.sdkClient = nil
		c.payloadURL = ""

		c.eventsMu.Lock()
		c.events = nil
		c.eventsMu.Unlock()

		return err
	}
	return nil
}

// GetPayloadURL 获取 payload URL
func (c *Client) GetPayloadURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.payloadURL
}

// GetType 获取提供者类型
func (c *Client) GetType() string {
	return "interactsh"
}

// interactionToEvent 将官方 SDK 的 Interaction 转为内部 Event
func interactionToEvent(i *server.Interaction, providerID string) Event {
	data := map[string]any{
		"protocol":       i.Protocol,
		"unique-id":      i.UniqueID,
		"full-id":        i.FullId,
		"q-type":         i.QType,
		"remote-address": i.RemoteAddress,
		"smtp-from":      i.SMTPFrom,
	}

	return Event{
		ID:            uuid.New().String(),
		Protocol:      i.Protocol,
		Method:        i.QType,
		Source:        i.RemoteAddress,
		Destination:   i.FullId,
		Timestamp:     i.Timestamp,
		CorrelationID: i.UniqueID,
		RawRequest:    i.RawRequest,
		RawResponse:   i.RawResponse,
		Data:          data,
	}
}

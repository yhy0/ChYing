package boast

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/yhy0/logging"
)

// Event 表示一次 BOAST 交互事件
type Event struct {
	ID            string
	Protocol      string
	Method        string
	Source        string
	Destination   string
	Timestamp     time.Time
	CorrelationID string
	RawRequest    string
}

// Client 是 BOAST 客户端
type Client struct {
	providerID string
	serverURL  string
	token      string
	canaryID   string
	payloadURL string
	client     *req.Client
	mu         sync.RWMutex
}

// New 创建 BOAST 客户端
func New(providerID, serverURL, token string) (*Client, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("BOAST server URL is required")
	}
	serverURL = strings.TrimRight(serverURL, "/")

	return &Client{
		providerID: providerID,
		serverURL:  serverURL,
		token:      token,
		client:     req.C().SetTimeout(30 * time.Second),
	}, nil
}

// Register 注册 BOAST canary
func (c *Client) Register() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	type canaryResp struct {
		ID         string `json:"id"`
		CanaryHost string `json:"canaryHost"`
		Error      string `json:"error"`
	}

	var resp canaryResp
	request := c.client.R().SetSuccessResult(&resp)
	if c.token != "" {
		request.SetHeader("Authorization", "Secret "+c.token)
	}

	_, err := request.Post(c.serverURL + "/api/canaries")
	if err != nil {
		return "", fmt.Errorf("register BOAST canary: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("BOAST error: %s", resp.Error)
	}

	c.canaryID = resp.ID
	if resp.CanaryHost != "" {
		c.payloadURL = resp.CanaryHost
	} else {
		host := strings.TrimPrefix(strings.TrimPrefix(c.serverURL, "https://"), "http://")
		c.payloadURL = c.canaryID + "." + host
	}

	logging.Logger.Infof("BOAST registered: %s (canary: %s)", c.payloadURL, c.canaryID)
	return c.payloadURL, nil
}

// Poll 拉取事件
func (c *Client) Poll() ([]Event, error) {
	c.mu.RLock()
	canaryID := c.canaryID
	c.mu.RUnlock()

	if canaryID == "" {
		return nil, fmt.Errorf("not registered")
	}

	type eventItem struct {
		ID         string `json:"id"`
		Protocol   string `json:"protocol"`
		Type       string `json:"type"`
		RemoteAddr string `json:"remoteAddr"`
		RawData    string `json:"rawData"`
		Timestamp  string `json:"timestamp"`
	}

	type pollResp struct {
		Events []eventItem `json:"events"`
		Error  string      `json:"error"`
	}

	var resp pollResp
	request := c.client.R().SetSuccessResult(&resp)
	if c.token != "" {
		request.SetHeader("Authorization", "Secret "+c.token)
	}

	_, err := request.Get(fmt.Sprintf("%s/api/canaries/%s/events", c.serverURL, canaryID))
	if err != nil {
		return nil, fmt.Errorf("poll BOAST: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("BOAST poll error: %s", resp.Error)
	}

	events := make([]Event, 0, len(resp.Events))
	for _, item := range resp.Events {
		ev := Event{
			ID:            item.ID,
			Protocol:      item.Protocol,
			Method:        item.Type,
			Source:        item.RemoteAddr,
			Destination:   c.payloadURL,
			Timestamp:     parseTime(item.Timestamp),
			CorrelationID: canaryID,
			RawRequest:    item.RawData,
		}
		events = append(events, ev)
	}

	return events, nil
}

// Deregister 注销
func (c *Client) Deregister() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.canaryID = ""
	c.payloadURL = ""
	return nil
}

// GetPayloadURL 获取 payload URL
func (c *Client) GetPayloadURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.payloadURL
}

// GetType 获取类型
func (c *Client) GetType() string {
	return "boast"
}

func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t
	}
	return time.Now()
}

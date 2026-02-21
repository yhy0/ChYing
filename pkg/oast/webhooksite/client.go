package webhooksite

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/yhy0/logging"
)

// Event 表示一次 Webhook.site 交互事件
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

// Client 是 Webhook.site 客户端
type Client struct {
	providerID string
	baseURL    string
	apiKey     string
	tokenID    string
	payloadURL string
	seenIDs    map[string]bool
	client     *req.Client
	mu         sync.RWMutex
}

// New 创建 Webhook.site 客户端
func New(providerID, baseURL, apiKey string) (*Client, error) {
	if baseURL == "" {
		baseURL = "https://webhook.site"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		providerID: providerID,
		baseURL:    baseURL,
		apiKey:     apiKey,
		seenIDs:    make(map[string]bool),
		client:     req.C().SetTimeout(30 * time.Second),
	}, nil
}

// Register 创建新的 webhook token
func (c *Client) Register() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	type tokenResp struct {
		UUID  string `json:"uuid"`
		Error string `json:"error"`
	}

	var resp tokenResp
	request := c.client.R().SetSuccessResult(&resp)
	if c.apiKey != "" {
		request.SetHeader("Api-Key", c.apiKey)
	}

	_, err := request.Post(c.baseURL + "/token")
	if err != nil {
		return "", fmt.Errorf("create webhook.site token: %w", err)
	}

	if resp.UUID == "" {
		return "", fmt.Errorf("webhook.site returned empty UUID")
	}

	c.tokenID = resp.UUID
	c.payloadURL = c.baseURL + "/" + c.tokenID
	c.seenIDs = make(map[string]bool)

	logging.Logger.Infof("Webhook.site registered: %s", c.payloadURL)
	return c.payloadURL, nil
}

// Poll 拉取新请求
func (c *Client) Poll() ([]Event, error) {
	c.mu.RLock()
	tokenID := c.tokenID
	c.mu.RUnlock()

	if tokenID == "" {
		return nil, fmt.Errorf("not registered")
	}

	type requestItem struct {
		UUID      string              `json:"uuid"`
		Method    string              `json:"method"`
		IP        string              `json:"ip"`
		URL       string              `json:"url"`
		Content   string              `json:"content"`
		Headers   map[string][]string `json:"headers"`
		Query     map[string]string   `json:"query"`
		CreatedAt string              `json:"created_at"`
	}

	type listResp struct {
		Data  []requestItem `json:"data"`
		Total int           `json:"total"`
	}

	var resp listResp
	request := c.client.R().
		SetQueryParam("sorting", "newest").
		SetSuccessResult(&resp)

	if c.apiKey != "" {
		request.SetHeader("Api-Key", c.apiKey)
	}

	_, err := request.Get(fmt.Sprintf("%s/token/%s/requests", c.baseURL, tokenID))
	if err != nil {
		return nil, fmt.Errorf("poll webhook.site: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	events := make([]Event, 0)
	for _, item := range resp.Data {
		if c.seenIDs[item.UUID] {
			continue
		}
		c.seenIDs[item.UUID] = true

		rawReq := fmt.Sprintf("%s %s HTTP/1.1\r\n", item.Method, item.URL)
		for k, vals := range item.Headers {
			for _, v := range vals {
				rawReq += fmt.Sprintf("%s: %s\r\n", k, v)
			}
		}
		rawReq += "\r\n" + item.Content

		ev := Event{
			ID:            item.UUID,
			Protocol:      "HTTP",
			Method:        item.Method,
			Source:        item.IP,
			Destination:   c.payloadURL,
			Timestamp:     parseTime(item.CreatedAt),
			CorrelationID: tokenID,
			RawRequest:    rawReq,
		}
		events = append(events, ev)
	}

	return events, nil
}

// Deregister 注销
func (c *Client) Deregister() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenID = ""
	c.payloadURL = ""
	c.seenIDs = make(map[string]bool)
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
	return "webhooksite"
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

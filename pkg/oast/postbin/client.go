package postbin

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/yhy0/logging"
)

// Event 表示一次 PostBin 交互事件
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

// Client 是 PostBin 客户端
type Client struct {
	providerID string
	baseURL    string
	binID      string
	payloadURL string
	client     *req.Client
	mu         sync.RWMutex
}

// New 创建 PostBin 客户端
func New(providerID, baseURL string) (*Client, error) {
	if baseURL == "" {
		baseURL = "https://www.postb.in"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		providerID: providerID,
		baseURL:    baseURL,
		client:     req.C().SetTimeout(30 * time.Second),
	}, nil
}

// Register 创建新的 bin
func (c *Client) Register() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	type binResp struct {
		BinID string `json:"binId"`
		Error string `json:"error"`
	}

	var resp binResp
	_, err := c.client.R().
		SetSuccessResult(&resp).
		Post(c.baseURL + "/api/bin")
	if err != nil {
		return "", fmt.Errorf("create postbin: %w", err)
	}

	if resp.BinID == "" {
		return "", fmt.Errorf("postbin returned empty binId")
	}

	c.binID = resp.BinID
	c.payloadURL = c.baseURL + "/" + c.binID

	logging.Logger.Infof("PostBin registered: %s", c.payloadURL)
	return c.payloadURL, nil
}

// Poll 拉取请求
func (c *Client) Poll() ([]Event, error) {
	c.mu.RLock()
	binID := c.binID
	c.mu.RUnlock()

	if binID == "" {
		return nil, fmt.Errorf("not registered")
	}

	events := make([]Event, 0)

	for i := 0; i < 50; i++ {
		type reqItem struct {
			Method    string            `json:"method"`
			Headers   map[string]string `json:"headers"`
			Body      string            `json:"body"`
			QueryStr  string            `json:"queryString"`
			IP        string            `json:"ip"`
			CreatedAt string            `json:"insertedAt"`
		}

		var item reqItem
		resp, err := c.client.R().
			SetSuccessResult(&item).
			Get(fmt.Sprintf("%s/api/bin/%s/req/shift", c.baseURL, binID))
		if err != nil {
			break
		}

		if resp.StatusCode == 204 || item.Method == "" {
			break
		}

		rawReq := fmt.Sprintf("%s /%s%s HTTP/1.1\r\n", item.Method, binID, item.QueryStr)
		for k, v := range item.Headers {
			rawReq += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		rawReq += "\r\n" + item.Body

		ev := Event{
			ID:            uuid.New().String(),
			Protocol:      "HTTP",
			Method:        item.Method,
			Source:        item.IP,
			Destination:   c.payloadURL,
			Timestamp:     parseTime(item.CreatedAt),
			CorrelationID: binID,
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
	c.binID = ""
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
	return "postbin"
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

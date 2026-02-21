package dnslogcn

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/yhy0/logging"
)

// Event 表示一次 dnslog.cn 交互事件
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

// Client 是 dnslog.cn 客户端
type Client struct {
	providerID string
	serverURL  string
	session    string
	domain     string
	payloadURL string
	client     *req.Client
	mu         sync.RWMutex
}

// New 创建 dnslog.cn 客户端
func New(providerID, serverURL string) (*Client, error) {
	if serverURL == "" {
		serverURL = "http://www.dnslog.cn"
	}
	serverURL = strings.TrimRight(serverURL, "/")

	return &Client{
		providerID: providerID,
		serverURL:  serverURL,
		client:     req.C().SetTimeout(30 * time.Second),
	}, nil
}

// Register 获取 DNSLog 域名
func (c *Client) Register() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 生成随机 session
	c.session = randomLetterNumbers(8)

	resp, err := c.client.R().
		SetHeader("Cookie", "PHPSESSID="+c.session).
		Get(c.serverURL + "/getdomain.php")
	if err != nil {
		return "", fmt.Errorf("register dnslog.cn: %w", err)
	}

	domain := strings.TrimSpace(resp.String())
	if domain == "" {
		return "", fmt.Errorf("dnslog.cn returned empty domain")
	}

	c.domain = domain
	c.payloadURL = domain

	logging.Logger.Infof("dnslog.cn registered: %s (session: %s)", c.payloadURL, c.session)
	return c.payloadURL, nil
}

// Poll 拉取 DNS 记录
func (c *Client) Poll() ([]Event, error) {
	c.mu.RLock()
	session := c.session
	domain := c.domain
	c.mu.RUnlock()

	if session == "" || domain == "" {
		return nil, fmt.Errorf("not registered")
	}

	resp, err := c.client.R().
		SetHeader("Cookie", "PHPSESSID="+session).
		Get(c.serverURL + "/getrecords.php")
	if err != nil {
		return nil, fmt.Errorf("poll dnslog.cn: %w", err)
	}

	body := strings.TrimSpace(resp.String())

	if body == "[]" || body == "" {
		return nil, nil
	}

	// dnslog.cn 返回格式: [["subdomain","source_ip","timestamp"], ...]
	// 简化处理：按行拆分
	events := make([]Event, 0)

	// 尝试简单解析
	body = strings.TrimPrefix(body, "[")
	body = strings.TrimSuffix(body, "]")

	if body == "" {
		return nil, nil
	}

	// 按记录拆分
	records := strings.Split(body, "],[")
	for _, record := range records {
		record = strings.Trim(record, "[]")
		parts := strings.Split(record, "\",\"")

		ev := Event{
			ID:            uuid.New().String(),
			Protocol:      "DNS",
			Method:        "A",
			Destination:   domain,
			Timestamp:     time.Now(),
			CorrelationID: session,
			RawRequest:    record,
		}

		for i, part := range parts {
			part = strings.Trim(part, "\"")
			switch i {
			case 0:
				ev.Destination = part
			case 1:
				ev.Source = part
			case 2:
				ev.Timestamp = parseTime(part)
			}
		}

		events = append(events, ev)
	}

	return events, nil
}

// Deregister 注销
func (c *Client) Deregister() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = ""
	c.domain = ""
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
	return "dnslogcn"
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

// randomLetterNumbers 生成随机字母数字字符串
func randomLetterNumbers(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

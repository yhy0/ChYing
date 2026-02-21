package digpm

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/yhy0/logging"
)

// Event 表示一次 dig.pm 交互事件
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

// Client 是 dig.pm DNSLog 客户端
type Client struct {
	providerID string
	serverURL  string
	domain     string // 指定反连域名
	subDomain  string
	fullDomain string
	token      string
	payloadURL string
	client     *req.Client
	mu         sync.RWMutex
}

// New 创建 dig.pm 客户端
func New(providerID, serverURL, domain string) (*Client, error) {
	if serverURL == "" {
		serverURL = "https://dig.pm"
	}
	serverURL = strings.TrimRight(serverURL, "/")

	if domain == "" {
		domain = "ipv6.bypass.eu.org."
	}

	return &Client{
		providerID: providerID,
		serverURL:  serverURL,
		domain:     domain,
		client:     req.C().SetTimeout(30 * time.Second),
	}, nil
}

// Register 注册获取子域名
func (c *Client) Register() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	type digResp struct {
		Domain     string `json:"domain"`
		Key        string `json:"key"`
		Token      string `json:"token"`
		FullDomain string `json:"fullDomain"`
		MainDomain string `json:"mainDomain"`
		SubDomain  string `json:"subDomain"`
	}

	resp, err := c.client.R().
		SetFormData(map[string]string{
			"domain": c.domain,
		}).
		Post(c.serverURL + "/get_sub_domain")
	if err != nil {
		return "", fmt.Errorf("register dig.pm: %w", err)
	}

	var result digResp
	if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
		return "", fmt.Errorf("parse dig.pm response: %w", err)
	}

	// dig.pm 不遵守标准 API，做兼容处理
	if result.SubDomain != "" {
		result.Domain = result.FullDomain
		result.Key = result.SubDomain
	}

	if result.Domain == "" {
		return "", fmt.Errorf("dig.pm returned empty domain")
	}

	c.subDomain = result.Key
	c.fullDomain = result.Domain
	c.token = result.Token
	c.payloadURL = result.Domain

	logging.Logger.Infof("dig.pm registered: %s", c.payloadURL)
	return c.payloadURL, nil
}

// Poll 拉取 DNS 记录
func (c *Client) Poll() ([]Event, error) {
	c.mu.RLock()
	domain := c.fullDomain
	token := c.token
	key := c.subDomain
	c.mu.RUnlock()

	if domain == "" {
		return nil, fmt.Errorf("not registered")
	}

	resp, err := c.client.R().
		SetFormData(map[string]string{
			"domain": domain,
			"token":  token,
		}).
		Post(c.serverURL + "/get_results")
	if err != nil {
		return nil, fmt.Errorf("poll dig.pm: %w", err)
	}

	body := resp.String()

	// 解析结果，dig.pm 返回 JSON 数组或 "null"
	if body == "null" || body == "" || body == "[]" {
		return nil, nil
	}

	// 检查是否包含我们的 key
	if key != "" && !strings.Contains(body, key) {
		return nil, nil
	}

	// 尝试解析为数组
	var records [][]string
	if err := json.Unmarshal([]byte(body), &records); err != nil {
		// 如果不是数组格式，作为单条记录处理
		events := []Event{{
			ID:            uuid.New().String(),
			Protocol:      "DNS",
			Method:        "A",
			Source:        "",
			Destination:   domain,
			Timestamp:     time.Now(),
			CorrelationID: key,
			RawRequest:    body,
		}}
		return events, nil
	}

	events := make([]Event, 0, len(records))
	for _, record := range records {
		ev := Event{
			ID:            uuid.New().String(),
			Protocol:      "DNS",
			Method:        "A",
			Destination:   domain,
			Timestamp:     time.Now(),
			CorrelationID: key,
		}
		// record 格式: [subdomain, source_ip, timestamp] 或其他
		if len(record) > 0 {
			ev.Destination = record[0]
		}
		if len(record) > 1 {
			ev.Source = record[1]
		}
		if len(record) > 2 {
			ev.Timestamp = parseTime(record[2])
			ev.RawRequest = strings.Join(record, " | ")
		}
		events = append(events, ev)
	}

	return events, nil
}

// Deregister 注销
func (c *Client) Deregister() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subDomain = ""
	c.fullDomain = ""
	c.token = ""
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
	return "digpm"
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

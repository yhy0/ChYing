package oast

import (
	"time"
)

// OASTEvent 表示一次 OAST 交互事件
type OASTEvent struct {
	ID            string    `json:"id"`
	ProviderID    string    `json:"providerId"`
	Type          string    `json:"type"`          // interactsh/boast/webhooksite/postbin
	Protocol      string    `json:"protocol"`      // HTTP/DNS/SMTP/...
	Method        string    `json:"method"`        // GET/POST/...
	Source        string    `json:"source"`        // 来源 IP
	Destination   string    `json:"destination"`   // Payload URL
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlationId"`
	RawRequest    string    `json:"rawRequest"`
	RawResponse   string    `json:"rawResponse"`
	Data          any       `json:"data,omitempty"`
}

// ProviderConfig 是创建/更新 Provider 时的配置
type ProviderConfig struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"` // interactsh/boast/webhooksite/postbin/digpm/dnslogcn
	URL     string `json:"url"`
	Token   string `json:"token"`
	Enabled bool   `json:"enabled"`
	Builtin bool   `json:"builtin"`
}

// ProviderStatus 是 Provider 的运行时状态
type ProviderStatus struct {
	ProviderConfig
	PayloadURL string `json:"payloadUrl"`
	Registered bool   `json:"registered"`
	Polling    bool   `json:"polling"`
	CreatedAt  string `json:"createdAt"`
}

// Settings 是 OAST 全局设置
type Settings struct {
	PollingInterval int    `json:"pollingInterval"` // 轮询间隔（毫秒）
	PayloadPrefix   string `json:"payloadPrefix"`
}

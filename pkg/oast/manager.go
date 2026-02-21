package oast

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/oast/boast"
	"github.com/yhy0/ChYing/pkg/oast/digpm"
	"github.com/yhy0/ChYing/pkg/oast/dnslogcn"
	"github.com/yhy0/ChYing/pkg/oast/interactsh"
	"github.com/yhy0/ChYing/pkg/oast/postbin"
	"github.com/yhy0/ChYing/pkg/oast/webhooksite"
	"github.com/yhy0/logging"
)

// providerClient 是内部 provider 的统一包装
type providerClient struct {
	providerID  string
	typeName    string
	interactsh  *interactsh.Client
	boast       *boast.Client
	webhooksite *webhooksite.Client
	postbin     *postbin.Client
	digpm       *digpm.Client
	dnslogcn    *dnslogcn.Client
}

func (p *providerClient) Register() (string, error) {
	switch p.typeName {
	case "interactsh":
		return p.interactsh.Register()
	case "boast":
		return p.boast.Register()
	case "webhooksite":
		return p.webhooksite.Register()
	case "postbin":
		return p.postbin.Register()
	case "digpm":
		return p.digpm.Register()
	case "dnslogcn":
		return p.dnslogcn.Register()
	}
	return "", fmt.Errorf("unknown type: %s", p.typeName)
}

func (p *providerClient) Poll() ([]OASTEvent, error) {
	switch p.typeName {
	case "interactsh":
		events, err := p.interactsh.Poll()
		if err != nil {
			return nil, err
		}
		result := make([]OASTEvent, 0, len(events))
		for _, ev := range events {
			result = append(result, OASTEvent{
				ID:            ev.ID,
				ProviderID:    p.providerID,
				Type:          "interactsh",
				Protocol:      ev.Protocol,
				Method:        ev.Method,
				Source:        ev.Source,
				Destination:   ev.Destination,
				Timestamp:     ev.Timestamp,
				CorrelationID: ev.CorrelationID,
				RawRequest:    ev.RawRequest,
				RawResponse:   ev.RawResponse,
				Data:          ev.Data,
			})
		}
		return result, nil

	case "boast":
		events, err := p.boast.Poll()
		if err != nil {
			return nil, err
		}
		result := make([]OASTEvent, 0, len(events))
		for _, ev := range events {
			result = append(result, OASTEvent{
				ID:            ev.ID,
				ProviderID:    p.providerID,
				Type:          "boast",
				Protocol:      ev.Protocol,
				Method:        ev.Method,
				Source:        ev.Source,
				Destination:   ev.Destination,
				Timestamp:     ev.Timestamp,
				CorrelationID: ev.CorrelationID,
				RawRequest:    ev.RawRequest,
			})
		}
		return result, nil

	case "webhooksite":
		events, err := p.webhooksite.Poll()
		if err != nil {
			return nil, err
		}
		result := make([]OASTEvent, 0, len(events))
		for _, ev := range events {
			result = append(result, OASTEvent{
				ID:            ev.ID,
				ProviderID:    p.providerID,
				Type:          "webhooksite",
				Protocol:      ev.Protocol,
				Method:        ev.Method,
				Source:        ev.Source,
				Destination:   ev.Destination,
				Timestamp:     ev.Timestamp,
				CorrelationID: ev.CorrelationID,
				RawRequest:    ev.RawRequest,
			})
		}
		return result, nil

	case "postbin":
		events, err := p.postbin.Poll()
		if err != nil {
			return nil, err
		}
		result := make([]OASTEvent, 0, len(events))
		for _, ev := range events {
			result = append(result, OASTEvent{
				ID:            ev.ID,
				ProviderID:    p.providerID,
				Type:          "postbin",
				Protocol:      ev.Protocol,
				Method:        ev.Method,
				Source:        ev.Source,
				Destination:   ev.Destination,
				Timestamp:     ev.Timestamp,
				CorrelationID: ev.CorrelationID,
				RawRequest:    ev.RawRequest,
			})
		}
		return result, nil

	case "digpm":
		events, err := p.digpm.Poll()
		if err != nil {
			return nil, err
		}
		result := make([]OASTEvent, 0, len(events))
		for _, ev := range events {
			result = append(result, OASTEvent{
				ID:            ev.ID,
				ProviderID:    p.providerID,
				Type:          "digpm",
				Protocol:      ev.Protocol,
				Method:        ev.Method,
				Source:        ev.Source,
				Destination:   ev.Destination,
				Timestamp:     ev.Timestamp,
				CorrelationID: ev.CorrelationID,
				RawRequest:    ev.RawRequest,
			})
		}
		return result, nil

	case "dnslogcn":
		events, err := p.dnslogcn.Poll()
		if err != nil {
			return nil, err
		}
		result := make([]OASTEvent, 0, len(events))
		for _, ev := range events {
			result = append(result, OASTEvent{
				ID:            ev.ID,
				ProviderID:    p.providerID,
				Type:          "dnslogcn",
				Protocol:      ev.Protocol,
				Method:        ev.Method,
				Source:        ev.Source,
				Destination:   ev.Destination,
				Timestamp:     ev.Timestamp,
				CorrelationID: ev.CorrelationID,
				RawRequest:    ev.RawRequest,
			})
		}
		return result, nil
	}
	return nil, fmt.Errorf("unknown type: %s", p.typeName)
}

func (p *providerClient) Deregister() error {
	switch p.typeName {
	case "interactsh":
		return p.interactsh.Deregister()
	case "boast":
		return p.boast.Deregister()
	case "webhooksite":
		return p.webhooksite.Deregister()
	case "postbin":
		return p.postbin.Deregister()
	case "digpm":
		return p.digpm.Deregister()
	case "dnslogcn":
		return p.dnslogcn.Deregister()
	}
	return nil
}

func (p *providerClient) GetPayloadURL() string {
	switch p.typeName {
	case "interactsh":
		return p.interactsh.GetPayloadURL()
	case "boast":
		return p.boast.GetPayloadURL()
	case "webhooksite":
		return p.webhooksite.GetPayloadURL()
	case "postbin":
		return p.postbin.GetPayloadURL()
	case "digpm":
		return p.digpm.GetPayloadURL()
	case "dnslogcn":
		return p.dnslogcn.GetPayloadURL()
	}
	return ""
}

func (p *providerClient) GetType() string {
	return p.typeName
}

// EventCallback 是交互事件的回调函数
type EventCallback func(event OASTEvent)

// Manager 管理所有 OAST Provider
type Manager struct {
	mu        sync.RWMutex
	providers map[string]*providerClient // providerID -> Provider 实例
	configs   map[string]ProviderConfig  // providerID -> 配置
	polling   map[string]chan struct{}    // providerID -> 停止信号
	callback  EventCallback
	settings  Settings
}

// NewManager 创建新的 OAST 管理器
func NewManager(callback EventCallback) *Manager {
	return &Manager{
		providers: make(map[string]*providerClient),
		configs:   make(map[string]ProviderConfig),
		polling:   make(map[string]chan struct{}),
		callback:  callback,
		settings: Settings{
			PollingInterval: 5000,
		},
	}
}

// CreateProvider 创建并保存一个新的 Provider 配置
func (m *Manager) CreateProvider(cfg ProviderConfig) (ProviderConfig, error) {
	if cfg.Name == "" || cfg.Type == "" || cfg.URL == "" {
		return ProviderConfig{}, fmt.Errorf("name, type, url are required")
	}

	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	cfg.Enabled = true

	dbProvider := &db.OASTProvider{
		ID:      cfg.ID,
		Name:    cfg.Name,
		Type:    cfg.Type,
		URL:     cfg.URL,
		Token:   cfg.Token,
		Enabled: true,
	}
	if err := db.CreateOASTProvider(dbProvider); err != nil {
		return ProviderConfig{}, fmt.Errorf("save provider to database: %w", err)
	}

	m.mu.Lock()
	m.configs[cfg.ID] = cfg
	m.mu.Unlock()

	return cfg, nil
}

// UpdateProvider 更新 Provider 配置
func (m *Manager) UpdateProvider(id string, updates ProviderConfig) (ProviderConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, ok := m.configs[id]
	if !ok {
		dbP, err := db.GetOASTProvider(id)
		if err != nil {
			return ProviderConfig{}, fmt.Errorf("provider not found: %s", id)
		}
		cfg = ProviderConfig{
			ID:      dbP.ID,
			Name:    dbP.Name,
			Type:    dbP.Type,
			URL:     dbP.URL,
			Token:   dbP.Token,
			Enabled: dbP.Enabled,
			Builtin: dbP.Builtin,
		}
	}

	if updates.Name != "" {
		cfg.Name = updates.Name
	}
	if updates.URL != "" {
		cfg.URL = updates.URL
	}
	if updates.Token != "" {
		cfg.Token = updates.Token
	}

	if err := db.UpdateOASTProvider(id, cfg.Name, cfg.URL, cfg.Token, cfg.Enabled); err != nil {
		return ProviderConfig{}, fmt.Errorf("update provider in database: %w", err)
	}

	m.configs[id] = cfg
	return cfg, nil
}

// DeleteProvider 删除 Provider
func (m *Manager) DeleteProvider(id string) error {
	// 内置 Provider 不可删除
	if db.IsBuiltinOASTProvider(id) {
		return fmt.Errorf("builtin provider cannot be deleted")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if stopCh, ok := m.polling[id]; ok {
		close(stopCh)
		delete(m.polling, id)
	}

	if p, ok := m.providers[id]; ok {
		_ = p.Deregister()
		delete(m.providers, id)
	}

	delete(m.configs, id)

	if err := db.DeleteOASTProvider(id); err != nil {
		return fmt.Errorf("delete provider from database: %w", err)
	}

	return nil
}

// ListProviders 列出所有 Provider 状态
func (m *Manager) ListProviders() []ProviderStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dbProviders := db.ListOASTProviders()
	result := make([]ProviderStatus, 0, len(dbProviders))

	for _, dbP := range dbProviders {
		cfg := ProviderConfig{
			ID:      dbP.ID,
			Name:    dbP.Name,
			Type:    dbP.Type,
			URL:     dbP.URL,
			Token:   dbP.Token,
			Enabled: dbP.Enabled,
			Builtin: dbP.Builtin,
		}
		m.configs[dbP.ID] = cfg

		status := ProviderStatus{
			ProviderConfig: cfg,
			CreatedAt:      dbP.CreatedAt.Format(time.RFC3339),
		}

		if p, ok := m.providers[dbP.ID]; ok {
			status.PayloadURL = p.GetPayloadURL()
			status.Registered = p.GetPayloadURL() != ""
		}
		if _, ok := m.polling[dbP.ID]; ok {
			status.Polling = true
		}

		result = append(result, status)
	}

	return result
}

// ToggleProvider 切换 Provider 启用状态
func (m *Manager) ToggleProvider(id string, enabled bool) error {
	m.mu.Lock()
	cfg, ok := m.configs[id]
	if ok {
		cfg.Enabled = enabled
		m.configs[id] = cfg
	}
	m.mu.Unlock()

	return db.ToggleOASTProvider(id, enabled)
}

// Register 注册指定的 Provider
func (m *Manager) Register(providerID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, ok := m.configs[providerID]
	if !ok {
		dbP, err := db.GetOASTProvider(providerID)
		if err != nil {
			return "", fmt.Errorf("provider not found: %s", providerID)
		}
		cfg = ProviderConfig{
			ID:      dbP.ID,
			Name:    dbP.Name,
			Type:    dbP.Type,
			URL:     dbP.URL,
			Token:   dbP.Token,
			Enabled: dbP.Enabled,
			Builtin: dbP.Builtin,
		}
		m.configs[providerID] = cfg
	}

	if old, ok := m.providers[providerID]; ok {
		_ = old.Deregister()
	}

	p, err := m.createProvider(cfg)
	if err != nil {
		return "", fmt.Errorf("create provider: %w", err)
	}

	payloadURL, err := p.Register()
	if err != nil {
		return "", fmt.Errorf("register provider: %w", err)
	}

	m.providers[providerID] = p
	return payloadURL, nil
}

// createProvider 根据配置创建 Provider 实例
func (m *Manager) createProvider(cfg ProviderConfig) (*providerClient, error) {
	pc := &providerClient{
		providerID: cfg.ID,
		typeName:   cfg.Type,
	}

	switch cfg.Type {
	case "interactsh":
		c, err := interactsh.New(cfg.ID, cfg.URL, cfg.Token)
		if err != nil {
			return nil, err
		}
		pc.interactsh = c
	case "boast":
		c, err := boast.New(cfg.ID, cfg.URL, cfg.Token)
		if err != nil {
			return nil, err
		}
		pc.boast = c
	case "webhooksite":
		c, err := webhooksite.New(cfg.ID, cfg.URL, cfg.Token)
		if err != nil {
			return nil, err
		}
		pc.webhooksite = c
	case "postbin":
		c, err := postbin.New(cfg.ID, cfg.URL)
		if err != nil {
			return nil, err
		}
		pc.postbin = c
	case "digpm":
		c, err := digpm.New(cfg.ID, cfg.URL, cfg.Token)
		if err != nil {
			return nil, err
		}
		pc.digpm = c
	case "dnslogcn":
		c, err := dnslogcn.New(cfg.ID, cfg.URL)
		if err != nil {
			return nil, err
		}
		pc.dnslogcn = c
	default:
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}

	return pc, nil
}

// StartPolling 启动指定 Provider 的定时轮询
func (m *Manager) StartPolling(providerID string, intervalMs int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.providers[providerID]; !ok {
		return fmt.Errorf("provider not registered: %s", providerID)
	}

	if stopCh, ok := m.polling[providerID]; ok {
		close(stopCh)
	}

	if intervalMs <= 0 {
		intervalMs = m.settings.PollingInterval
	}

	stopCh := make(chan struct{})
	m.polling[providerID] = stopCh

	go m.pollLoop(providerID, time.Duration(intervalMs)*time.Millisecond, stopCh)

	return nil
}

// StopPolling 停止指定 Provider 的轮询
func (m *Manager) StopPolling(providerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stopCh, ok := m.polling[providerID]
	if !ok {
		return nil
	}

	close(stopCh)
	delete(m.polling, providerID)
	return nil
}

// PollOnce 手动拉取一次事件
func (m *Manager) PollOnce(providerID string) ([]OASTEvent, error) {
	m.mu.RLock()
	p, ok := m.providers[providerID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider not registered: %s", providerID)
	}

	events, err := p.Poll()
	if err != nil {
		return nil, fmt.Errorf("poll: %w", err)
	}

	for _, ev := range events {
		if m.callback != nil {
			m.callback(ev)
		}
	}

	return events, nil
}

// Deregister 注销指定的 Provider
func (m *Manager) Deregister(providerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stopCh, ok := m.polling[providerID]; ok {
		close(stopCh)
		delete(m.polling, providerID)
	}

	p, ok := m.providers[providerID]
	if !ok {
		return nil
	}

	err := p.Deregister()
	delete(m.providers, providerID)
	return err
}

// GetSettings 获取全局设置
func (m *Manager) GetSettings() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

// UpdateSettings 更新全局设置
func (m *Manager) UpdateSettings(s Settings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s.PollingInterval > 0 {
		m.settings.PollingInterval = s.PollingInterval
	}
	if s.PayloadPrefix != "" {
		m.settings.PayloadPrefix = s.PayloadPrefix
	}
}

// Shutdown 关闭所有 Provider
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, stopCh := range m.polling {
		close(stopCh)
		delete(m.polling, id)
	}

	for id, p := range m.providers {
		if err := p.Deregister(); err != nil {
			logging.Logger.Warnf("deregister provider %s: %v", id, err)
		}
		delete(m.providers, id)
	}
}

// pollLoop 轮询循环
func (m *Manager) pollLoop(providerID string, interval time.Duration, stopCh chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			m.mu.RLock()
			p, ok := m.providers[providerID]
			m.mu.RUnlock()

			if !ok {
				return
			}

			events, err := p.Poll()
			if err != nil {
				logging.Logger.Warnf("OAST poll %s error: %v", providerID, err)
				continue
			}

			for _, ev := range events {
				if m.callback != nil {
					m.callback(ev)
				}
			}
		}
	}
}

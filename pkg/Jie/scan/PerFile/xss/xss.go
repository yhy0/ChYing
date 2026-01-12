package xss

import (
	"sync"

	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/config"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/engine"
)

// Plugin 是重构后的XSS扫描插件实现
type Plugin struct {
	SeenRequests sync.Map
}

// Scan 是插件的扫描入口点
func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if p.IsScanned(in.UniqueId) {
		return
	}

	// 初始化新的扫描引擎
	cfg := config.NewConfig(config.Intelligent) // 使用智能模式配置
	xssEngine, err := engine.NewEngine(cfg)
	if err != nil {
		// 记录错误但不影响后续处理
		return
	}

	// 运行扫描
	_ = xssEngine.Run(in, client)

	p.SeenRequests.Store(in.UniqueId, true)
}

// IsScanned 检查一个请求是否已经被扫描过
func (p *Plugin) IsScanned(key string) bool {
	_, loaded := p.SeenRequests.Load(key)
	return loaded
}

// Name 返回插件的名称
func (p *Plugin) Name() string {
	return "xss"
}

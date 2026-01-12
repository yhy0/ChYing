package browser

import (
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/yhy0/logging"
)

var (
	once sync.Once
	// globalBrowserManager 是浏览器管理器的全局单例实例
	globalBrowserManager *Manager
)

// Manager 负责管理rod浏览器实例的生命周期。
type Manager struct {
	browser *rod.Browser
	lock    sync.Mutex
}

// GetManager 返回全局唯一的浏览器管理器实例。
func GetManager() *Manager {
	once.Do(func() {
		globalBrowserManager = &Manager{}
	})
	return globalBrowserManager
}

// GetBrowser 启动并返回一个共享的浏览器实例。
// 如果浏览器已经启动，则直接返回现有实例。
func (m *Manager) GetBrowser() (*rod.Browser, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 如果实例已存在，直接返回
	if m.browser != nil {
		return m.browser, nil
	}

	// 使用默认的 launcher 配置
	l := launcher.New().
		// 添加一些常用的启动参数以提高兼容性和性能
		Set("no-sandbox").
		Set("disable-setuid-sandbox").
		Set("disable-dev-shm-usage").
		Set("disable-accelerated-2d-canvas").
		Set("no-first-run").
		Set("no-zygote").
		Set("single-process").
		Set("disable-gpu")

	u := l.MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	logging.Logger.Info("无头浏览器实例启动成功。")
	m.browser = browser

	return m.browser, nil
}

// Cleanup 优雅地关闭浏览器实例并释放资源。
func (m *Manager) Cleanup() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.browser != nil {
		err := m.browser.Close()
		if err != nil {
			logging.Logger.Errorf("关闭浏览器实例时出错: %v", err)
		} else {
			logging.Logger.Info("浏览器实例已成功关闭。")
		}
		m.browser = nil
	}
}

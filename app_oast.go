package main

import (
	"github.com/yhy0/ChYing/pkg/oast"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @desc OAST (Out-of-Band Application Security Testing) 相关方法
**/

var oastManager *oast.Manager

// initOAST 初始化 OAST 管理器
func initOAST() {
	oastManager = oast.NewManager(func(event oast.OASTEvent) {
		if wailsApp != nil {
			wailsApp.Event.Emit("oast:interaction", event)
		}
	})
	logging.Logger.Infoln("OAST manager initialized")
}

// OASTCreateProvider 创建 OAST Provider
func (a *App) OASTCreateProvider(provider oast.ProviderConfig) Result {
	if oastManager == nil {
		initOAST()
	}

	cfg, err := oastManager.CreateProvider(provider)
	if err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: cfg}
}

// OASTUpdateProvider 更新 OAST Provider
func (a *App) OASTUpdateProvider(id string, updates oast.ProviderConfig) Result {
	if oastManager == nil {
		return Result{Error: "OAST manager not initialized"}
	}

	cfg, err := oastManager.UpdateProvider(id, updates)
	if err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: cfg}
}

// OASTDeleteProvider 删除 OAST Provider
func (a *App) OASTDeleteProvider(id string) Result {
	if oastManager == nil {
		return Result{Error: "OAST manager not initialized"}
	}

	if err := oastManager.DeleteProvider(id); err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "ok"}
}

// OASTListProviders 列出所有 OAST Provider
func (a *App) OASTListProviders() Result {
	if oastManager == nil {
		initOAST()
	}

	providers := oastManager.ListProviders()
	return Result{Data: providers}
}

// OASTToggleProvider 切换 Provider 启用状态
func (a *App) OASTToggleProvider(id string, enabled bool) Result {
	if oastManager == nil {
		return Result{Error: "OAST manager not initialized"}
	}

	if err := oastManager.ToggleProvider(id, enabled); err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "ok"}
}

// OASTRegister 注册 Provider 并返回 payload URL
func (a *App) OASTRegister(providerID string) Result {
	if oastManager == nil {
		initOAST()
	}

	payloadURL, err := oastManager.Register(providerID)
	if err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: payloadURL}
}

// OASTStartPolling 启动定时轮询
func (a *App) OASTStartPolling(providerID string, intervalMs int) Result {
	if oastManager == nil {
		return Result{Error: "OAST manager not initialized"}
	}

	if err := oastManager.StartPolling(providerID, intervalMs); err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "ok"}
}

// OASTStopPolling 停止轮询
func (a *App) OASTStopPolling(providerID string) Result {
	if oastManager == nil {
		return Result{Error: "OAST manager not initialized"}
	}

	if err := oastManager.StopPolling(providerID); err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "ok"}
}

// OASTPollOnce 手动拉取一次事件
func (a *App) OASTPollOnce(providerID string) Result {
	if oastManager == nil {
		return Result{Error: "OAST manager not initialized"}
	}

	events, err := oastManager.PollOnce(providerID)
	if err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: events}
}

// OASTDeregister 注销 Provider
func (a *App) OASTDeregister(providerID string) Result {
	if oastManager == nil {
		return Result{Error: "OAST manager not initialized"}
	}

	if err := oastManager.Deregister(providerID); err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "ok"}
}

// OASTGetSettings 获取 OAST 全局设置
func (a *App) OASTGetSettings() Result {
	if oastManager == nil {
		initOAST()
	}

	settings := oastManager.GetSettings()
	return Result{Data: settings}
}

// OASTUpdateSettings 更新 OAST 全局设置
func (a *App) OASTUpdateSettings(settings oast.Settings) Result {
	if oastManager == nil {
		initOAST()
	}

	oastManager.UpdateSettings(settings)
	return Result{Data: "ok"}
}

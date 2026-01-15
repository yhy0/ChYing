package main

import (
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 代理和流量相关方法
**/

// ConfigureScope 配置作用域
// todo 修改时同步到配置文件，配置文件修改好像有点麻烦，还要考虑是否启用，不如存入 db 中
func (a *App) ConfigureScope(t, t1 string, scope *conf.Scope) *conf.Configure {
	if t == "edit" {
		if scope.Type == "exclude" {
			for i, _scope := range conf.Config.Exclude {
				if _scope.Id == scope.Id {
					conf.Config.Exclude[i] = scope
					break
				}
			}
		} else {
			for i, _scope := range conf.Config.Include {
				if _scope.Id == scope.Id {
					conf.Config.Include[i] = scope
					break
				}
			}
		}
	} else if t == "add" {
		if t1 == "exclude" {
			if len(conf.Config.Exclude) == 0 {
				scope.Id = 0
			} else {
				scope.Id = conf.Config.Exclude[len(conf.Config.Exclude)-1].Id + 1
			}

			scope.Type = "exclude"
			conf.Config.Exclude = append(conf.Config.Exclude, scope)
		} else if t1 == "include" {
			scope.Type = "include"
			if len(conf.Config.Include) == 0 {
				scope.Id = 0
			} else {
				scope.Id = conf.Config.Include[len(conf.Config.Include)-1].Id + 1
			}

			conf.Config.Include = append(conf.Config.Include, scope)
		}
	} else if t == "del" {
		if t1 == "exclude" {
			var exclude []*conf.Scope
			for _, s := range conf.Config.Exclude {
				if s.Id == scope.Id {
					continue
				}
				exclude = append(exclude, s)
			}
			conf.Config.Exclude = exclude
		} else if t1 == "include" {
			var include []*conf.Scope
			for _, s := range conf.Config.Include {
				if s.Id == scope.Id {
					continue
				}
				include = append(include, s)
			}
			conf.Config.Include = include
		}
	}

	for _, s := range conf.Config.Exclude {
		logging.Logger.Infoln(s)
	}
	for _, s := range conf.Config.Include {
		logging.Logger.Infoln(s)
	}

	return conf.Config
}

// QueryHistoryByDSL 使用DSL查询表达式过滤HTTP历史记录
// dslQuery: DSL查询表达式，如为空则返回所有历史记录
func (a *App) QueryHistoryByDSL(dslQuery string) Result {
	logging.Logger.Infoln("正在执行DSL查询:", dslQuery)

	// 调用 mitmproxy 包中导出的 DSL 查询函数
	results, err := mitmproxy.ExportQueryHistoryByDSL(dslQuery)
	if err != nil {
		logging.Logger.Errorln("DSL查询执行出错:", err)
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}

	logging.Logger.Infoln("DSL查询完成，匹配到", len(results), "条记录")
	return Result{
		Data:  results,
		Error: "",
	}
}

// GetMatchReplaceRules 获取所有匹配替换规则
func (a *App) GetMatchReplaceRules() Result {
	logging.Logger.Debugln("GetMatchReplaceRules...")
	rules, err := mitmproxy.GetMatchReplaceRules()
	if err != nil {
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}
	return Result{
		Data:  rules,
		Error: "",
	}
}

// SaveMatchReplaceRule 保存匹配替换规则
func (a *App) SaveMatchReplaceRule(rule mitmproxy.MatchReplaceRule) Result {
	err := mitmproxy.SaveMatchReplaceRule(rule)
	if err != nil {
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}
	return Result{
		Data:  rule,
		Error: "",
	}
}

// DeleteMatchReplaceRule 删除匹配替换规则
func (a *App) DeleteMatchReplaceRule(ruleID int) Result {
	err := mitmproxy.DeleteMatchReplaceRule(ruleID)
	if err != nil {
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}
	return Result{
		Data:  ruleID,
		Error: "",
	}
}

// ApplyMatchReplaceRules 应用匹配替换规则配置
func (a *App) ApplyMatchReplaceRules(config mitmproxy.MatchReplaceRules) Result {
	err := mitmproxy.ApplyMatchReplaceRulesWithConfig(config)
	if err != nil {
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}
	return Result{
		Data:  true,
		Error: "",
	}
}

// GetAuthorizationRules 获取越权检测规则
func (a *App) GetAuthorizationRules() Result {
	rules, err := mitmproxy.GetAuthorizationRules()
	if err != nil {
		return Result{
			Error: err.Error(),
		}
	}

	return Result{
		Data: rules,
	}
}

// SaveAuthorizationRules 保存越权检测规则
func (a *App) SaveAuthorizationRules(rules mitmproxy.AuthorizationRules) Result {
	logging.Logger.Debugln("SaveAuthorizationRules...", rules)
	err := mitmproxy.SaveAuthorizationRules(rules)
	if err != nil {
		return Result{
			Error: err.Error(),
		}
	}

	return Result{
		Data: "Authorization rules saved successfully",
	}
}

// StartAuthorizationCheck 开始越权检测
func (a *App) StartAuthorizationCheck() Result {
	err := mitmproxy.StartAuthorizationCheck()
	if err != nil {
		return Result{
			Error: err.Error(),
		}
	}

	return Result{
		Data: "Authorization check started successfully",
	}
}

// StopAuthorizationCheck 停止越权检测
func (a *App) StopAuthorizationCheck() Result {
	err := mitmproxy.StopAuthorizationCheck()
	if err != nil {
		return Result{
			Error: err.Error(),
		}
	}

	return Result{
		Data: "Authorization check stopped successfully",
	}
}

// GetAuthorizationTestResults 获取越权测试结果
func (a *App) GetAuthorizationTestResults() Result {
	results := mitmproxy.GetAuthorizationTestResults()
	return Result{
		Data: results,
	}
}

// ClearAuthorizationTestResults 清除越权测试结果
func (a *App) ClearAuthorizationTestResults() Result {
	mitmproxy.ClearAuthorizationTestResults()
	return Result{
		Data: "Authorization test results cleared successfully",
	}
}

// ==================== 代理监听器管理 ====================

// GetProxyListeners 获取代理监听器列表
func (a *App) GetProxyListeners() Result {
	listeners := conf.AppConf.Proxy.Listeners

	// 如果没有配置监听器，创建一个默认监听器基于当前代理状态
	if len(listeners) == 0 {
		status := mitmproxy.GetProxyStatus()
		host := conf.AppConf.Proxy.Host
		port := conf.AppConf.Proxy.Port

		// 如果配置中没有，尝试从代理状态获取
		if host == "" {
			if h, ok := status["host"].(string); ok && h != "" {
				host = h
			} else {
				host = "127.0.0.1"
			}
		}
		if port == 0 {
			if p, ok := status["port"].(int); ok && p != 0 {
				port = p
			} else {
				port = 9080
			}
		}

		running := false
		if r, ok := status["running"].(bool); ok {
			running = r
		}

		defaultListener := conf.ProxyListener{
			ID:      "default",
			Host:    host,
			Port:    port,
			Enabled: true,
			Running: running,
		}
		listeners = []conf.ProxyListener{defaultListener}
	}

	return Result{
		Data:  listeners,
		Error: "",
	}
}

// SaveProxyListener 保存代理监听器（添加或更新）
func (a *App) SaveProxyListener(listener conf.ProxyListener) Result {
	// 查找是否已存在
	found := false
	for i, l := range conf.AppConf.Proxy.Listeners {
		if l.ID == listener.ID {
			conf.AppConf.Proxy.Listeners[i] = listener
			found = true
			break
		}
	}

	// 如果不存在则添加
	if !found {
		conf.AppConf.Proxy.Listeners = append(conf.AppConf.Proxy.Listeners, listener)
	}

	// 保存配置到文件
	if err := conf.SaveConfig(); err != nil {
		logging.Logger.Errorln("保存监听器配置失败:", err)
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}

	logging.Logger.Infoln("监听器配置已保存:", listener.ID)
	return Result{
		Data:  listener,
		Error: "",
	}
}

// DeleteProxyListener 删除代理监听器
func (a *App) DeleteProxyListener(id string) Result {
	var newListeners []conf.ProxyListener
	found := false
	for _, l := range conf.AppConf.Proxy.Listeners {
		if l.ID == id {
			found = true
			continue
		}
		newListeners = append(newListeners, l)
	}

	if !found {
		return Result{
			Data:  nil,
			Error: "监听器不存在",
		}
	}

	conf.AppConf.Proxy.Listeners = newListeners

	// 保存配置到文件
	if err := conf.SaveConfig(); err != nil {
		logging.Logger.Errorln("删除监听器配置失败:", err)
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}

	logging.Logger.Infoln("监听器已删除:", id)
	return Result{
		Data:  id,
		Error: "",
	}
}

// ToggleProxyListener 切换监听器启用状态
func (a *App) ToggleProxyListener(id string, enabled bool) Result {
	found := false
	for i, l := range conf.AppConf.Proxy.Listeners {
		if l.ID == id {
			conf.AppConf.Proxy.Listeners[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		return Result{
			Data:  nil,
			Error: "监听器不存在",
		}
	}

	// 保存配置到文件
	if err := conf.SaveConfig(); err != nil {
		logging.Logger.Errorln("切换监听器状态失败:", err)
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}

	logging.Logger.Infof("监听器 %s 状态已切换为: %v", id, enabled)
	return Result{
		Data:  enabled,
		Error: "",
	}
}

// GetProxyStatus 获取代理状态
func (a *App) GetProxyStatus() Result {
	status := mitmproxy.GetProxyStatus()
	return Result{
		Data:  status,
		Error: "",
	}
}

// GetCertificateInfo 获取证书信息
func (a *App) GetCertificateInfo() Result {
	info := mitmproxy.GetCertificateInfo()
	return Result{
		Data:  info,
		Error: "",
	}
}

// CheckPortAvailable 检查端口是否可用
func (a *App) CheckPortAvailable(host string, port int) Result {
	// 如果 host 为空，使用配置中的默认值
	if host == "" {
		host = conf.AppConf.Proxy.Host
		if host == "" {
			host = "127.0.0.1"
		}
	}
	isOccupied := utils.IsPortOccupied(host, port)
	return Result{
		Data: map[string]interface{}{
			"host":      host,
			"port":      port,
			"available": !isOccupied,
		},
		Error: "",
	}
}

// RestartProxyServer 重启代理服务器
func (a *App) RestartProxyServer() Result {
	err := mitmproxy.RestartProxy()
	if err != nil {
		logging.Logger.Errorln("重启代理服务器失败:", err)
		return Result{
			Data:  nil,
			Error: err.Error(),
		}
	}

	logging.Logger.Infoln("代理服务器重启成功")
	return Result{
		Data:  "代理服务器正在重启",
		Error: "",
	}
}

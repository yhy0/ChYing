package main

import (
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/mitmproxy"
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

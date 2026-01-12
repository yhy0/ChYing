package api

import (
	"context"

	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/logging"
)

type ProxyAPI struct{}

type Result struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func NewProxyAPI() *ProxyAPI {
	return &ProxyAPI{}
}

// StopMitmproxy 停止代理
func (p *ProxyAPI) StopMitmproxy(ctx context.Context) Result {
	mitmproxy.StopMitmproxy()
	return Result{Data: "success"}
}

// StartMitmproxy 启动代理
func (p *ProxyAPI) StartMitmproxy(ctx context.Context, host, port, proxyMode string) Result {
	mitmproxy.StartMitmproxy(host, port, proxyMode)
	return Result{Data: "success"}
}

// GetProxyStatus 获取代理状态
func (p *ProxyAPI) GetProxyStatus(ctx context.Context) Result {
	status := mitmproxy.GetProxyStatus()
	return Result{Data: status}
}

// GetHistoryData 获取历史数据
func (p *ProxyAPI) GetHistoryData(ctx context.Context, page, limit int) Result {
	data, err := db.GetHistoryFromSQLite(page, limit)
	if err != nil {
		logging.Logger.Errorf("从SQLite获取历史数据失败: %v", err)
		return Result{Error: "获取历史数据失败"}
	}
	return Result{Data: data}
}

// GetHttpData 获取HTTP数据
func (p *ProxyAPI) GetHttpData(ctx context.Context, id int) Result {
	data, err := db.GetHttpData(id)
	if err != nil {
		logging.Logger.Errorf("获取HTTP数据失败: %v", err)
		return Result{Error: "获取HTTP数据失败"}
	}
	return Result{Data: data}
}

// GetProxyRules 获取代理规则
func (p *ProxyAPI) GetProxyRules(ctx context.Context) Result {
	rules := mitmproxy.GetProxyRules()
	return Result{Data: rules}
}

// SaveProxyRule 保存代理规则
func (p *ProxyAPI) SaveProxyRule(ctx context.Context, rule map[string]interface{}) Result {
	if err := mitmproxy.SaveProxyRule(rule); err != nil {
		logging.Logger.Errorf("保存代理规则失败: %v", err)
		return Result{Error: "保存代理规则失败"}
	}
	return Result{Data: "success"}
}

// DeleteProxyRule 删除代理规则
func (p *ProxyAPI) DeleteProxyRule(ctx context.Context, ruleId string) Result {
	if err := mitmproxy.DeleteProxyRule(ruleId); err != nil {
		logging.Logger.Errorf("删除代理规则失败: %v", err)
		return Result{Error: "删除代理规则失败"}
	}
	return Result{Data: "success"}
}

// TestProxyRule 测试代理规则
func (p *ProxyAPI) TestProxyRule(ctx context.Context, rule map[string]interface{}) Result {
	result := mitmproxy.TestProxyRule(rule)
	return Result{Data: result}
}

// GetCertificateInfo 获取证书信息
func (p *ProxyAPI) GetCertificateInfo(ctx context.Context) Result {
	info := mitmproxy.GetCertificateInfo()
	return Result{Data: info}
}

// RegenerateCertificate 重新生成证书
func (p *ProxyAPI) RegenerateCertificate(ctx context.Context) Result {
	if err := mitmproxy.RegenerateCertificate(); err != nil {
		logging.Logger.Errorf("重新生成证书失败: %v", err)
		return Result{Error: "重新生成证书失败"}
	}
	return Result{Data: "success"}
}

// ChangeHttpData 修改HTTP数据
func (p *ProxyAPI) ChangeHttpData(ctx context.Context, data map[string]interface{}) Result {
	if err := mitmproxy.ChangeHttpData(data); err != nil {
		logging.Logger.Errorf("修改HTTP数据失败: %v", err)
		return Result{Error: "修改HTTP数据失败"}
	}
	return Result{Data: "success"}
}

// ClearHistory 清空历史记录
func (p *ProxyAPI) ClearHistory(ctx context.Context) Result {
	if err := db.ClearAllHistoryData(); err != nil {
		logging.Logger.Errorf("清空SQLite数据失败: %v", err)
		return Result{Error: "清空历史记录失败"}
	}

	// 清空内存中的数据
	mitmproxy.ClearMemoryHistoryData()

	return Result{Data: "success"}
}

// GetMemoryUsage 获取内存使用情况
func (p *ProxyAPI) GetMemoryUsage(ctx context.Context) Result {
	usage := mitmproxy.GetMemoryUsage()
	return Result{Data: usage}
}

// ExportHistory 导出历史记录
func (p *ProxyAPI) ExportHistory(ctx context.Context, format string) Result {
	data, err := mitmproxy.ExportHistory(format)
	if err != nil {
		logging.Logger.Errorf("导出历史记录失败: %v", err)
		return Result{Error: "导出历史记录失败"}
	}
	return Result{Data: data}
}

// ImportHistory 导入历史记录
func (p *ProxyAPI) ImportHistory(ctx context.Context, filePath string) Result {
	if err := mitmproxy.ImportHistory(filePath); err != nil {
		logging.Logger.Errorf("导入历史记录失败: %v", err)
		return Result{Error: "导入历史记录失败"}
	}
	return Result{Data: "success"}
}

// GetProxySettings 获取代理设置
func (p *ProxyAPI) GetProxySettings(ctx context.Context) Result {
	settings := mitmproxy.GetProxySettings()
	return Result{Data: settings}
}

// UpdateProxySettings 更新代理设置
func (p *ProxyAPI) UpdateProxySettings(ctx context.Context, settings map[string]interface{}) Result {
	if err := mitmproxy.UpdateProxySettings(settings); err != nil {
		logging.Logger.Errorf("更新代理设置失败: %v", err)
		return Result{Error: "更新代理设置失败"}
	}
	return Result{Data: "success"}
}

// GetInterceptRules 获取拦截规则
func (p *ProxyAPI) GetInterceptRules(ctx context.Context) Result {
	rules := mitmproxy.GetInterceptRules()
	return Result{Data: rules}
}

// SaveInterceptRule 保存拦截规则
func (p *ProxyAPI) SaveInterceptRule(ctx context.Context, rule map[string]interface{}) Result {
	if err := mitmproxy.SaveInterceptRule(rule); err != nil {
		logging.Logger.Errorf("保存拦截规则失败: %v", err)
		return Result{Error: "保存拦截规则失败"}
	}
	return Result{Data: "success"}
}

// DeleteInterceptRule 删除拦截规则
func (p *ProxyAPI) DeleteInterceptRule(ctx context.Context, ruleId string) Result {
	if err := mitmproxy.DeleteInterceptRule(ruleId); err != nil {
		logging.Logger.Errorf("删除拦截规则失败: %v", err)
		return Result{Error: "删除拦截规则失败"}
	}
	return Result{Data: "success"}
}

// GetProxyLog 获取代理日志
func (p *ProxyAPI) GetProxyLog(ctx context.Context, lines int) Result {
	logs := mitmproxy.GetProxyLog(lines)
	return Result{Data: logs}
}

// ClearProxyLog 清空代理日志
func (p *ProxyAPI) ClearProxyLog(ctx context.Context) Result {
	if err := mitmproxy.ClearProxyLog(); err != nil {
		logging.Logger.Errorf("清空代理日志失败: %v", err)
		return Result{Error: "清空代理日志失败"}
	}
	return Result{Data: "success"}
}

// GetProxyStatistics 获取代理统计信息
func (p *ProxyAPI) GetProxyStatistics(ctx context.Context) Result {
	stats := mitmproxy.GetProxyStatistics()
	return Result{Data: stats}
}

// ResetProxyStatistics 重置代理统计
func (p *ProxyAPI) ResetProxyStatistics(ctx context.Context) Result {
	mitmproxy.ResetProxyStatistics()
	return Result{Data: "success"}
}

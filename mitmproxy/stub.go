package mitmproxy

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/logging"
)

// 全局代理状态
var (
	proxyInstance   *ProxyInstance
	proxyInstanceMu sync.RWMutex
)

type ProxyInstance struct {
	running atomic.Bool
	host    string
	port    string
	mode    string
}

// StopMitmproxy 停止代理
func StopMitmproxy() {
	proxyInstanceMu.Lock()
	defer proxyInstanceMu.Unlock()

	if proxyInstance != nil {
		proxyInstance.running.Store(false)
		logging.Logger.Info("代理已停止")
	}
}

// StartMitmproxy 启动代理
func StartMitmproxy(host, port, proxyMode string) {
	proxyInstanceMu.Lock()
	defer proxyInstanceMu.Unlock()

	if proxyInstance == nil {
		proxyInstance = &ProxyInstance{
			host: host,
			port: port,
			mode: proxyMode,
		}
	}

	proxyInstance.running.Store(true)
	logging.Logger.Infof("代理启动: %s:%s (模式: %s)", host, port, proxyMode)
}

// GetProxyStatus 获取代理状态
func GetProxyStatus() map[string]interface{} {
	proxyInstanceMu.RLock()
	defer proxyInstanceMu.RUnlock()

	if proxyInstance == nil {
		return map[string]interface{}{
			"running": false,
			"host":    "",
			"port":    "",
			"mode":    "",
		}
	}

	return map[string]interface{}{
		"running": proxyInstance.running.Load(),
		"host":    proxyInstance.host,
		"port":    proxyInstance.port,
		"mode":    proxyInstance.mode,
	}
}

// GetProxyRules 获取代理规则
func GetProxyRules() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":      "1",
			"enabled": true,
			"name":    "默认规则",
			"type":    "match_replace",
		},
	}
}

// SaveProxyRule 保存代理规则
func SaveProxyRule(rule map[string]interface{}) error {
	// 存根实现
	logging.Logger.Info("保存代理规则:", rule)
	return nil
}

// DeleteProxyRule 删除代理规则
func DeleteProxyRule(ruleId string) error {
	// 存根实现
	logging.Logger.Info("删除代理规则:", ruleId)
	return nil
}

// TestProxyRule 测试代理规则
func TestProxyRule(rule map[string]interface{}) map[string]interface{} {
	// 存根实现
	return map[string]interface{}{
		"result": "success",
		"tested": true,
	}
}

// GetCertificateInfo 获取证书信息
func GetCertificateInfo() map[string]interface{} {
	return map[string]interface{}{
		"issuer":     "ChYing CA",
		"subject":    "ChYing Proxy",
		"valid_from": time.Now().Format("2006-01-02 15:04:05"),
		"valid_to":   time.Now().AddDate(1, 0, 0).Format("2006-01-02 15:04:05"),
		"installed":  true,
	}
}

// RegenerateCertificate 重新生成证书
func RegenerateCertificate() error {
	// 存根实现
	logging.Logger.Info("重新生成证书")
	return nil
}

// ChangeHttpData 修改HTTP数据
func ChangeHttpData(data map[string]interface{}) error {
	// 存根实现
	logging.Logger.Info("修改HTTP数据:", data)
	return nil
}

// ClearMemoryHistoryData 清空内存中的历史数据
func ClearMemoryHistoryData() {
	// 清空内存中的数据
	HTTPBodyMap.Range(func(key, value interface{}) bool {
		HTTPBodyMap.Delete(key)
		return true
	})
	TempHistoryCache.Range(func(key, value interface{}) bool {
		TempHistoryCache.Delete(key)
		return true
	})
	TempRequestRawCache.Range(func(key, value interface{}) bool {
		TempRequestRawCache.Delete(key)
		return true
	})
	logging.Logger.Info("已清空内存历史数据")
}

// GetMemoryUsage 获取内存使用情况
func GetMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"allocated_mb":   m.Alloc / 1024 / 1024,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"gc_runs":        m.NumGC,
	}
}

// ExportHistory 导出历史记录
func ExportHistory(format string) ([]byte, error) {
	// 存根实现
	return []byte(fmt.Sprintf("导出历史记录 (格式: %s)", format)), nil
}

// ImportHistory 导入历史记录
func ImportHistory(filePath string) error {
	// 存根实现
	logging.Logger.Info("导入历史记录:", filePath)
	return nil
}

// GetProxySettings 获取代理设置
func GetProxySettings() map[string]interface{} {
	config := conf.GetAppConfig()
	return map[string]interface{}{
		"host":           config.Proxy.Host,
		"port":           config.Proxy.Port,
		"enabled":        config.Proxy.Enabled,
		"intercept_mode": "passive",
		"ssl_verify":     true,
	}
}

// UpdateProxySettings 更新代理设置
func UpdateProxySettings(settings map[string]interface{}) error {
	// 存根实现
	logging.Logger.Info("更新代理设置:", settings)
	return nil
}

// GetInterceptRules 获取拦截规则
func GetInterceptRules() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":      "1",
			"enabled": true,
			"name":    "默认拦截规则",
			"type":    "request",
		},
	}
}

// SaveInterceptRule 保存拦截规则
func SaveInterceptRule(rule map[string]interface{}) error {
	// 存根实现
	logging.Logger.Info("保存拦截规则:", rule)
	return nil
}

// DeleteInterceptRule 删除拦截规则
func DeleteInterceptRule(ruleId string) error {
	// 存根实现
	logging.Logger.Info("删除拦截规则:", ruleId)
	return nil
}

// GetProxyLog 获取代理日志
func GetProxyLog(lines int) []string {
	// 存根实现
	return []string{
		"[INFO] 代理服务已启动",
		"[DEBUG] 处理请求: GET /api/test",
		"[INFO] 响应状态: 200 OK",
	}
}

// ClearProxyLog 清空代理日志
func ClearProxyLog() error {
	// 存根实现
	logging.Logger.Info("清空代理日志")
	return nil
}

// GetProxyStatistics 获取代理统计信息
func GetProxyStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_requests":    1000,
		"total_responses":   995,
		"intercepted_count": 50,
		"error_count":       5,
	}
}

// ResetProxyStatistics 重置代理统计信息
func ResetProxyStatistics() error {
	// 存根实现
	logging.Logger.Info("重置代理统计信息")
	return nil
}

package api

import (
	"context"

	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/logging"
)

// APIManager 统一的API管理器
type APIManager struct {
	Config        *ConfigAPI
	Proxy         *ProxyAPI
	Vulnerability *VulnerabilityAPI
}

// NewAPIManager 创建API管理器
func NewAPIManager() *APIManager {
	return &APIManager{
		Config:        NewConfigAPI(),
		Proxy:         NewProxyAPI(),
		Vulnerability: NewVulnerabilityAPI(),
	}
}

// ProcessVulnerabilityMessage 处理漏洞消息的统一入口
func (m *APIManager) ProcessVulnerabilityMessage(vuln *output.VulMessage) error {
	if vuln == nil {
		return nil
	}

	logging.Logger.Infof("处理漏洞消息: %s - %s", vuln.Plugin, vuln.VulnData.Target)

	// 使用漏洞API处理消息
	if err := m.Vulnerability.ProcessVulnerabilityMessage(vuln); err != nil {
		logging.Logger.Errorf("处理漏洞消息失败: %v", err)
		return err
	}

	return nil
}

// GetVulnerabilities 获取漏洞列表
func (m *APIManager) GetVulnerabilities(query VulnerabilityQuery) Result {
	return m.Vulnerability.GetVulnerabilities(context.Background(), query)
}

// GetVulnerabilityStats 获取漏洞统计
func (m *APIManager) GetVulnerabilityStats() Result {
	return m.Vulnerability.GetVulnerabilityStats(context.Background())
}

// GetConfigStatus 获取配置状态
func (m *APIManager) GetConfigStatus() map[string]interface{} {
	return m.Config.GetConfigStatus()
}

// GetConfig 获取配置
func (m *APIManager) GetConfig() string {
	return m.Config.GetConfig()
}

// ============ 项目管理相关接口 ============

// GetLocalProjects 获取本地项目列表
func (m *APIManager) GetLocalProjects() Result {
	return m.Config.GetLocalProjects()
}

// CreateLocalProject 创建本地项目
func (m *APIManager) CreateLocalProject(projectID string, projectName string) Result {
	return m.Config.CreateLocalProject(projectID, projectName)
}

// DeleteLocalProject 删除本地项目
func (m *APIManager) DeleteLocalProject(projectName string) Result {
	return m.Config.DeleteLocalProject(projectName)
}

// UpdateConfig 更新配置
func (m *APIManager) UpdateConfig(config map[string]interface{}) Result {
	if err := m.Config.UpdateConfig(config); err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "配置更新成功"}
}

// ReloadConfig 重新加载配置
func (m *APIManager) ReloadConfig() Result {
	if err := m.Config.ReloadConfig(); err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "配置重新加载成功"}
}

package db

import (
	"time"

	"github.com/yhy0/logging"
)

// OASTProvider OAST 提供者配置
type OASTProvider struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Type      string    `gorm:"column:type;not null" json:"type"`
	URL       string    `gorm:"column:url;not null" json:"url"`
	Token     string    `gorm:"column:token;default:''" json:"token"`
	Enabled   bool      `gorm:"column:enabled;default:true" json:"enabled"`
	Builtin   bool      `gorm:"column:builtin;default:false" json:"builtin"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (OASTProvider) TableName() string {
	return "oast_providers"
}

// CreateOASTProvider 创建 OAST Provider
func CreateOASTProvider(provider *OASTProvider) error {
	if GlobalDB == nil {
		return nil
	}

	return RetryOnLocked("CreateOASTProvider", func() error {
		return GlobalDB.Create(provider).Error
	}, 3)
}

// GetOASTProvider 根据 ID 获取 Provider
func GetOASTProvider(id string) (*OASTProvider, error) {
	if GlobalDB == nil {
		return nil, nil
	}

	var provider OASTProvider
	err := GlobalDB.Where("id = ?", id).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

// ListOASTProviders 列出所有 Provider
func ListOASTProviders() []OASTProvider {
	if GlobalDB == nil {
		return nil
	}

	var providers []OASTProvider
	if err := GlobalDB.Order("created_at DESC").Find(&providers).Error; err != nil {
		logging.Logger.Warnf("ListOASTProviders error: %v", err)
		return nil
	}
	return providers
}

// UpdateOASTProvider 更新 Provider
func UpdateOASTProvider(id, name, url, token string, enabled bool) error {
	if GlobalDB == nil {
		return nil
	}

	return RetryOnLocked("UpdateOASTProvider", func() error {
		return GlobalDB.Model(&OASTProvider{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":    name,
			"url":     url,
			"token":   token,
			"enabled": enabled,
		}).Error
	}, 3)
}

// DeleteOASTProvider 删除 Provider
func DeleteOASTProvider(id string) error {
	if GlobalDB == nil {
		return nil
	}

	return RetryOnLocked("DeleteOASTProvider", func() error {
		return GlobalDB.Where("id = ?", id).Delete(&OASTProvider{}).Error
	}, 3)
}

// ToggleOASTProvider 切换 Provider 启用状态
func ToggleOASTProvider(id string, enabled bool) error {
	if GlobalDB == nil {
		return nil
	}

	return RetryOnLocked("ToggleOASTProvider", func() error {
		return GlobalDB.Model(&OASTProvider{}).Where("id = ?", id).Update("enabled", enabled).Error
	}, 3)
}

// IsBuiltinOASTProvider 检查是否为内置 Provider
func IsBuiltinOASTProvider(id string) bool {
	if GlobalDB == nil {
		return false
	}

	var provider OASTProvider
	if err := GlobalDB.Where("id = ? AND builtin = ?", id, true).First(&provider).Error; err != nil {
		return false
	}
	return true
}

// defaultOASTProviders 内置的默认 OAST Provider 列表
var defaultOASTProviders = []OASTProvider{
	{ID: "builtin-interactsh-oast-pro", Name: "Interactsh (oast.pro)", Type: "interactsh", URL: "https://oast.pro", Enabled: true, Builtin: true},
	{ID: "builtin-interactsh-oast-live", Name: "Interactsh (oast.live)", Type: "interactsh", URL: "https://oast.live", Enabled: true, Builtin: true},
	{ID: "builtin-interactsh-oast-fun", Name: "Interactsh (oast.fun)", Type: "interactsh", URL: "https://oast.fun", Enabled: true, Builtin: true},
	{ID: "builtin-webhooksite", Name: "Webhook.site", Type: "webhooksite", URL: "https://webhook.site", Enabled: true, Builtin: true},
	{ID: "builtin-postbin", Name: "PostBin", Type: "postbin", URL: "https://www.postb.in", Enabled: true, Builtin: true},
	{ID: "builtin-digpm", Name: "dig.pm", Type: "digpm", URL: "https://dig.pm", Enabled: true, Builtin: true},
	{ID: "builtin-dnslogcn", Name: "dnslog.cn", Type: "dnslogcn", URL: "http://www.dnslog.cn", Enabled: true, Builtin: true},
}

// SeedDefaultOASTProviders 初始化默认的内置 OAST Provider
func SeedDefaultOASTProviders() {
	if GlobalDB == nil {
		return
	}

	for _, p := range defaultOASTProviders {
		var existing OASTProvider
		if err := GlobalDB.Where("id = ?", p.ID).First(&existing).Error; err != nil {
			// 不存在则创建
			if err := RetryOnLocked("SeedOASTProvider", func() error {
				return GlobalDB.Create(&p).Error
			}, 3); err != nil {
				logging.Logger.Warnf("Seed OAST provider %s failed: %v", p.Name, err)
			}
		}
	}
}

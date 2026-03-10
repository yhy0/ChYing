package db

import (
	"time"

	"github.com/yhy0/logging"
	"gorm.io/gorm"
)

// RepeaterTab Repeater 标签页模型
type RepeaterTab struct {
	ID               string `gorm:"primaryKey" json:"id"`
	Name             string `json:"name"`
	Color            string `json:"color"`
	GroupID          string `json:"group_id"`
	Request          string `json:"request"`
	Response         string `json:"response"`
	Method           string `json:"method"`
	URL              string `json:"url"`
	SortOrder        int    `json:"sort_order"`
	IsActive         bool   `json:"is_active"`
	ServerDurationMs int64  `json:"server_duration_ms"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// RepeaterGroup Repeater 分组模型
type RepeaterGroup struct {
	ID        string `gorm:"primaryKey" json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RepeaterHistory Repeater 请求历史模型
type RepeaterHistory struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	TabID            string `gorm:"index" json:"tab_id"`
	SequenceID       int    `json:"sequence_id"`
	Method           string `json:"method"`
	URL              string `json:"url"`
	Request          string `json:"request"`
	Response         string `json:"response"`
	StatusCode       int    `json:"status_code"`
	StatusText       string `json:"status_text"`
	ServerDurationMs int64  `json:"server_duration_ms"`
	CreatedAt        time.Time
}

// SaveRepeaterTabs 全量保存 Repeater 标签页（事务内先删后插）
func SaveRepeaterTabs(tabs []RepeaterTab) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("SaveRepeaterTabs", func() error {
		return GlobalDB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("DELETE FROM repeater_tabs").Error; err != nil {
				return err
			}
			if len(tabs) == 0 {
				return nil
			}
			return tx.Create(&tabs).Error
		})
	}, 3)
}

// GetRepeaterTabs 获取所有 Repeater 标签页
func GetRepeaterTabs() ([]RepeaterTab, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var tabs []RepeaterTab
	err := GlobalDB.Order("sort_order ASC").Find(&tabs).Error
	return tabs, err
}

// SaveRepeaterGroups 全量保存 Repeater 分组（事务内先删后插）
func SaveRepeaterGroups(groups []RepeaterGroup) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("SaveRepeaterGroups", func() error {
		return GlobalDB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("DELETE FROM repeater_groups").Error; err != nil {
				return err
			}
			if len(groups) == 0 {
				return nil
			}
			return tx.Create(&groups).Error
		})
	}, 3)
}

// GetRepeaterGroups 获取所有 Repeater 分组
func GetRepeaterGroups() ([]RepeaterGroup, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var groups []RepeaterGroup
	err := GlobalDB.Order("sort_order ASC").Find(&groups).Error
	return groups, err
}

// SaveRepeaterHistory 保存某个 tab 的请求历史（事务内先删后插）
func SaveRepeaterHistory(tabID string, history []RepeaterHistory) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("SaveRepeaterHistory", func() error {
		return GlobalDB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("tab_id = ?", tabID).Delete(&RepeaterHistory{}).Error; err != nil {
				return err
			}
			if len(history) == 0 {
				return nil
			}
			return tx.Create(&history).Error
		})
	}, 3)
}

// GetRepeaterHistory 获取某个 tab 的请求历史
func GetRepeaterHistory(tabID string) ([]RepeaterHistory, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var history []RepeaterHistory
	err := GlobalDB.Where("tab_id = ?", tabID).Order("sequence_id DESC").Find(&history).Error
	return history, err
}

// DeleteRepeaterTab 删除 tab 及其关联的 history
func DeleteRepeaterTab(tabID string) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("DeleteRepeaterTab", func() error {
		return GlobalDB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("tab_id = ?", tabID).Delete(&RepeaterHistory{}).Error; err != nil {
				logging.Logger.Errorf("DeleteRepeaterTab history err: %v", err)
				return err
			}
			if err := tx.Where("id = ?", tabID).Delete(&RepeaterTab{}).Error; err != nil {
				logging.Logger.Errorf("DeleteRepeaterTab tab err: %v", err)
				return err
			}
			return nil
		})
	}, 3)
}

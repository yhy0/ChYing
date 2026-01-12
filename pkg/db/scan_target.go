package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2024/12/30
  @desc: 扫描目标管理
**/

// ScanTargetStatus 扫描目标状态
type ScanTargetStatus string

const (
	StatusPending   ScanTargetStatus = "pending"   // 待扫描
	StatusRunning   ScanTargetStatus = "running"   // 扫描中
	StatusCompleted ScanTargetStatus = "completed" // 已完成
	StatusFailed    ScanTargetStatus = "failed"    // 扫描失败
	StatusCancelled ScanTargetStatus = "cancelled" // 已取消
	StatusPaused    ScanTargetStatus = "paused"    // 暂停
)

// ScanTargetType 扫描目标类型
type ScanTargetType string

const (
	TypeSingle ScanTargetType = "single" // 单个URL
	TypeDomain ScanTargetType = "domain" // 域名
	TypeSubnet ScanTargetType = "subnet" // 子网
	TypeCIDR   ScanTargetType = "cidr"   // CIDR
	TypeBatch  ScanTargetType = "batch"  // 批量目标
)

// ScanConfig 扫描配置
type ScanConfig struct {
	EnablePortScan    bool     `json:"enable_port_scan"`   // 端口扫描
	EnableDirScan     bool     `json:"enable_dir_scan"`    // 目录扫描
	EnableVulnScan    bool     `json:"enable_vuln_scan"`   // 漏洞扫描
	EnableFingerprint bool     `json:"enable_fingerprint"` // 指纹识别
	EnableCrawler     bool     `json:"enable_crawler"`     // 爬虫
	EnableXSS         bool     `json:"enable_xss"`         // XSS扫描
	EnableSQL         bool     `json:"enable_sql"`         // SQL注入扫描
	EnableBypass403   bool     `json:"enable_bypass_403"`  // 403绕过
	Threads           int      `json:"threads"`            // 扫描线程数
	Timeout           int      `json:"timeout"`            // 超时时间(秒)
	CustomHeaders     []string `json:"custom_headers"`     // 自定义请求头
	Exclude           []string `json:"exclude"`            // 排除规则
	PortRange         string   `json:"port_range"`         // 端口范围 (如: 1-1000,8080,8443)
	MaxDepth          int      `json:"max_depth"`          // 爬虫最大深度
}

// ScanTarget 扫描目标表
type ScanTarget struct {
	ID          int64            `gorm:"primary_key;auto_increment" json:"id"`
	Name        string           `gorm:"not null" json:"name"`                  // 目标名称
	Type        ScanTargetType   `gorm:"not null;default:'single'" json:"type"` // 目标类型
	Target      string           `gorm:"not null" json:"target"`                // 目标地址
	Description string           `json:"description"`                           // 描述
	Status      ScanTargetStatus `gorm:"default:'pending'" json:"status"`       // 状态
	Priority    int              `gorm:"default:5" json:"priority"`             // 优先级 (1-10, 10最高)
	Config      string           `gorm:"type:text" json:"config"`               // 扫描配置JSON

	// 调度信息
	ScheduleType string     `json:"schedule_type"` // 调度类型: once, daily, weekly, monthly
	ScheduleTime string     `json:"schedule_time"` // 调度时间 (cron格式或时间字符串)
	NextRunTime  *time.Time `json:"next_run_time"` // 下次运行时间

	// 分配信息
	AssignedNode string `gorm:"index" json:"assigned_node"` // 分配的节点ID
	NodeName     string `json:"node_name"`                  // 节点名称

	// 执行信息
	StartedAt   *time.Time `json:"started_at"`                // 开始时间
	CompletedAt *time.Time `json:"completed_at"`              // 完成时间
	Duration    int        `json:"duration"`                  // 持续时间(秒)
	Progress    int        `gorm:"default:0" json:"progress"` // 进度百分比

	// 结果统计
	FoundHosts     int `gorm:"default:0" json:"found_hosts"`     // 发现的主机数
	FoundPorts     int `gorm:"default:0" json:"found_ports"`     // 发现的端口数
	FoundVulns     int `gorm:"default:0" json:"found_vulns"`     // 发现的漏洞数
	FoundDirs      int `gorm:"default:0" json:"found_dirs"`      // 发现的目录数
	TotalRequests  int `gorm:"default:0" json:"total_requests"`  // 总请求数
	FailedRequests int `gorm:"default:0" json:"failed_requests"` // 失败请求数

	// 错误信息
	ErrorMessage string `json:"error_message"`                // 错误信息
	LastError    string `json:"last_error"`                   // 最后错误
	RetryCount   int    `gorm:"default:0" json:"retry_count"` // 重试次数

	// 创建者信息
	CreatedBy   string `json:"created_by"`   // 创建者
	CreatedFrom string `json:"created_from"` // 创建来源 (local/remote)

	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetScanConfig 获取扫描配置
func (st *ScanTarget) GetScanConfig() (*ScanConfig, error) {
	if st.Config == "" {
		return getDefaultScanConfig(), nil
	}

	var config ScanConfig
	if err := json.Unmarshal([]byte(st.Config), &config); err != nil {
		return nil, fmt.Errorf("解析扫描配置失败: %w", err)
	}

	return &config, nil
}

// SetScanConfig 设置扫描配置
func (st *ScanTarget) SetScanConfig(config *ScanConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化扫描配置失败: %w", err)
	}

	st.Config = string(configJSON)
	return nil
}

// IsReadyToRun 检查是否可以开始扫描
func (st *ScanTarget) IsReadyToRun() bool {
	return st.Status == StatusPending &&
		(st.NextRunTime == nil || st.NextRunTime.Before(time.Now()))
}

// CanRetry 检查是否可以重试
func (st *ScanTarget) CanRetry() bool {
	return st.Status == StatusFailed && st.RetryCount < 3
}

// getDefaultScanConfig 获取默认扫描配置
func getDefaultScanConfig() *ScanConfig {
	return &ScanConfig{
		EnablePortScan:    true,
		EnableDirScan:     true,
		EnableVulnScan:    true,
		EnableFingerprint: true,
		EnableCrawler:     false,
		EnableXSS:         false,
		EnableSQL:         false,
		EnableBypass403:   false,
		Threads:           10,
		Timeout:           30,
		CustomHeaders:     []string{},
		Exclude:           []string{},
		PortRange:         "80,443,8080,8443,3000,5000,8000,9000",
		MaxDepth:          2,
	}
}

// ==============================================
// 数据库操作函数
// ==============================================

// AddScanTarget 添加扫描目标
func AddScanTarget(target *ScanTarget) error {
	if target.Config == "" {
		defaultConfig := getDefaultScanConfig()
		if err := target.SetScanConfig(defaultConfig); err != nil {
			return err
		}
	}

	if target.Priority <= 0 {
		target.Priority = 5
	}

	if target.Status == "" {
		target.Status = StatusPending
	}

	if target.CreatedFrom == "" {
		target.CreatedFrom = "local"
	}

	err := GlobalDB.Model(&ScanTarget{}).Create(target).Error
	if err != nil {
		logging.Logger.Errorf("AddScanTarget error: %v", err)
		return err
	}

	logging.Logger.Infof("扫描目标已添加: %s (%s)", target.Name, target.Target)
	return nil
}

// GetScanTargets 获取扫描目标列表
func GetScanTargets(status ScanTargetStatus, limit int, offset int) ([]*ScanTarget, int64, error) {
	var targets []*ScanTarget
	var total int64

	query := GlobalDB.Model(&ScanTarget{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取数据
	err := query.Order("priority DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&targets).Error

	if err != nil {
		logging.Logger.Errorf("GetScanTargets error: %v", err)
		return nil, 0, err
	}

	return targets, total, nil
}

// GetPendingScanTargets 获取待扫描目标 (用于任务调度)
func GetPendingScanTargets(nodeID string, limit int) ([]*ScanTarget, error) {
	var targets []*ScanTarget

	query := GlobalDB.Model(&ScanTarget{}).
		Where("status = ?", StatusPending).
		Where("(assigned_node = ? OR assigned_node = '' OR assigned_node IS NULL)", nodeID).
		Order("priority DESC, created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&targets).Error
	if err != nil {
		logging.Logger.Errorf("GetPendingScanTargets error: %v", err)
		return nil, err
	}

	return targets, nil
}

// GetScanTargetByID 根据ID获取扫描目标
func GetScanTargetByID(id int64) (*ScanTarget, error) {
	var target ScanTarget
	err := GlobalDB.Model(&ScanTarget{}).Where("id = ?", id).First(&target).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

// UpdateScanTarget 更新扫描目标
func UpdateScanTarget(target *ScanTarget) error {
	err := GlobalDB.Model(&ScanTarget{}).Where("id = ?", target.ID).Updates(target).Error
	if err != nil {
		logging.Logger.Errorf("UpdateScanTarget error: %v", err)
		return err
	}
	return nil
}

// UpdateScanTargetStatus 更新扫描状态
func UpdateScanTargetStatus(id int64, status ScanTargetStatus, message string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if message != "" {
		if status == StatusFailed {
			updates["error_message"] = message
			updates["last_error"] = message
		}
	}

	// 根据状态设置时间戳
	switch status {
	case StatusRunning:
		updates["started_at"] = time.Now()
		updates["progress"] = 0
	case StatusCompleted:
		updates["completed_at"] = time.Now()
		updates["progress"] = 100
	case StatusFailed:
		updates["completed_at"] = time.Now()
		// 增加重试计数
		GlobalDB.Model(&ScanTarget{}).Where("id = ?", id).
			Update("retry_count", GlobalDB.Raw("retry_count + 1"))
	}

	err := GlobalDB.Model(&ScanTarget{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		logging.Logger.Errorf("UpdateScanTargetStatus error: %v", err)
		return err
	}

	return nil
}

// UpdateScanProgress 更新扫描进度
func UpdateScanProgress(id int64, progress int, stats map[string]int) error {
	updates := map[string]interface{}{
		"progress":   progress,
		"updated_at": time.Now(),
	}

	// 更新统计信息
	if stats != nil {
		if val, ok := stats["found_hosts"]; ok {
			updates["found_hosts"] = val
		}
		if val, ok := stats["found_ports"]; ok {
			updates["found_ports"] = val
		}
		if val, ok := stats["found_vulns"]; ok {
			updates["found_vulns"] = val
		}
		if val, ok := stats["found_dirs"]; ok {
			updates["found_dirs"] = val
		}
		if val, ok := stats["total_requests"]; ok {
			updates["total_requests"] = val
		}
		if val, ok := stats["failed_requests"]; ok {
			updates["failed_requests"] = val
		}
	}

	err := GlobalDB.Model(&ScanTarget{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		logging.Logger.Errorf("UpdateScanProgress error: %v", err)
		return err
	}

	return nil
}

// AssignScanTarget 分配扫描目标到节点
func AssignScanTarget(id int64, nodeID, nodeName string) error {
	updates := map[string]interface{}{
		"assigned_node": nodeID,
		"node_name":     nodeName,
		"updated_at":    time.Now(),
	}

	err := GlobalDB.Model(&ScanTarget{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		logging.Logger.Errorf("AssignScanTarget error: %v", err)
		return err
	}

	logging.Logger.Infof("扫描目标 %d 已分配到节点 %s", id, nodeID)
	return nil
}

// DeleteScanTarget 删除扫描目标
func DeleteScanTarget(id int64) error {
	err := GlobalDB.Model(&ScanTarget{}).Where("id = ?", id).Delete(&ScanTarget{}).Error
	if err != nil {
		logging.Logger.Errorf("DeleteScanTarget error: %v", err)
		return err
	}

	logging.Logger.Infof("扫描目标 %d 已删除", id)
	return nil
}

// BatchAddScanTargets 批量添加扫描目标
func BatchAddScanTargets(targets []string, targetType ScanTargetType, config *ScanConfig, createdBy string) error {
	if config == nil {
		config = getDefaultScanConfig()
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化扫描配置失败: %w", err)
	}

	for i, target := range targets {
		scanTarget := &ScanTarget{
			Name:        fmt.Sprintf("批量目标-%d", i+1),
			Type:        targetType,
			Target:      target,
			Description: fmt.Sprintf("批量添加的%s类型目标", targetType),
			Status:      StatusPending,
			Priority:    5,
			Config:      string(configJSON),
			CreatedBy:   createdBy,
			CreatedFrom: "local",
		}

		if err := AddScanTarget(scanTarget); err != nil {
			logging.Logger.Errorf("批量添加目标失败 %s: %v", target, err)
			continue
		}
	}

	logging.Logger.Infof("批量添加扫描目标完成，总计 %d 个", len(targets))
	return nil
}

// GetScanStatistics 获取扫描统计信息
func GetScanStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 按状态统计
	var statusStats []struct {
		Status ScanTargetStatus `json:"status"`
		Count  int64            `json:"count"`
	}

	err := GlobalDB.Model(&ScanTarget{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats).Error

	if err != nil {
		return nil, err
	}

	stats["by_status"] = statusStats

	// 按类型统计
	var typeStats []struct {
		Type  ScanTargetType `json:"type"`
		Count int64          `json:"count"`
	}

	err = GlobalDB.Model(&ScanTarget{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&typeStats).Error

	if err != nil {
		return nil, err
	}

	stats["by_type"] = typeStats

	// 总体统计
	var total int64
	GlobalDB.Model(&ScanTarget{}).Count(&total)
	stats["total"] = total

	// 今日统计
	today := time.Now().Format("2006-01-02")
	var todayCount int64
	GlobalDB.Model(&ScanTarget{}).
		Where("DATE(created_at) = ?", today).
		Count(&todayCount)
	stats["today"] = todayCount

	return stats, nil
}

// CleanupCompletedTargets 清理已完成的目标
func CleanupCompletedTargets(days int) error {
	if days <= 0 {
		days = 30 // 默认保留30天
	}

	cutoff := time.Now().AddDate(0, 0, -days)

	result := GlobalDB.Where("status IN (?) AND completed_at < ?",
		[]ScanTargetStatus{StatusCompleted, StatusFailed, StatusCancelled}, cutoff).
		Delete(&ScanTarget{})

	if result.Error != nil {
		logging.Logger.Errorf("CleanupCompletedTargets error: %v", result.Error)
		return result.Error
	}

	logging.Logger.Infof("清理了 %d 个过期的扫描目标", result.RowsAffected)
	return nil
}

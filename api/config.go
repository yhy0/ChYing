package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/logging"
)

type ConfigAPI struct{}

func NewConfigAPI() *ConfigAPI {
	return &ConfigAPI{}
}

// GetConfig 获取配置信息
func (c *ConfigAPI) GetConfig() string {
	config := `{"proxy": "127.0.0.1:9080", "jwt_file": "` + filepath.Join(file.ChyingDir, "jwt.txt") + `"}`
	logging.Logger.Infoln(config)
	return config
}

// GetConfigStatus 获取配置状态
func (c *ConfigAPI) GetConfigStatus() map[string]interface{} {
	config := conf.GetAppConfig()
	status := map[string]interface{}{
		"config_file":   conf.GetConfigFilePath(),
		"proxy_enabled": config.Proxy.Enabled,
		"proxy_address": fmt.Sprintf("%s:%d", config.Proxy.Host, config.Proxy.Port),
		"scan_enabled": map[string]bool{
			"port_scan": config.Scan.EnablePortScan,
			"dir_scan":  config.Scan.EnableDirScan,
			"vuln_scan": config.Scan.EnableVulnScan,
		},
		"logging_level": config.Logging.Level,
	}

	return status
}

// UpdateConfig 更新配置
func (c *ConfigAPI) UpdateConfig(config map[string]interface{}) error {
	// 更新代理配置
	if proxyConfig, exists := config["proxy"]; exists {
		if proxyMap, ok := proxyConfig.(map[string]interface{}); ok {
			appConfig := conf.GetAppConfig()

			if host, exists := proxyMap["host"]; exists {
				if hostStr, ok := host.(string); ok {
					appConfig.Proxy.Host = hostStr
				}
			}
			if port, exists := proxyMap["port"]; exists {
				if portFloat, ok := port.(float64); ok {
					appConfig.Proxy.Port = int(portFloat)
				}
			}
			if enabled, exists := proxyMap["enabled"]; exists {
				if enabledBool, ok := enabled.(bool); ok {
					appConfig.Proxy.Enabled = enabledBool
				}
			}

			// 保存配置
			conf.UpdateAppConfig(*appConfig)
		}
	}

	return nil
}

// SaveConfig 保存配置到文件
func (c *ConfigAPI) SaveConfig() error {
	// 配置自动保存，这里返回成功
	return nil
}

// ReloadConfig 重新加载配置
func (c *ConfigAPI) ReloadConfig() error {
	return nil
}

// GetLocalProjects 获取本地项目列表
func (c *ConfigAPI) GetLocalProjects() Result {
	var projects []map[string]interface{}

	logging.Logger.Infof("🔍 GetLocalProjects 开始执行")

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ChYing")
	dbDir := filepath.Join(configDir, "db")

	// 扫描 db 目录下的项目子目录
	projectDirs := scanProjectDirs(dbDir)

	if len(projectDirs) == 0 {
		// 没有项目时返回空列表
		return Result{
			Data: map[string]interface{}{
				"projects": projects,
				"success":  true,
			},
		}
	}

	for _, projectName := range projectDirs {
		// 数据库路径: db/<projectName>/<projectName>.db
		dbPath := filepath.Join(dbDir, projectName, projectName+".db")
		fileInfo := getDBFileInfo(dbPath)

		// 如果数据库文件不存在，跳过
		if !fileInfo["exists"].(bool) {
			continue
		}

		var stats DBStats
		if db.GlobalDB != nil {
			stats = getDBFileStats(dbPath)
		} else {
			stats = DBStats{
				TotalRequests: 0,
				TotalHosts:    0,
				FirstRequest:  time.Now(),
				LastRequest:   time.Now(),
			}
		}

		projects = append(projects, map[string]interface{}{
			"id":             fmt.Sprintf("local-%s", projectName),
			"name":           projectName,
			"database_file":  projectName + ".db",
			"database":       projectName + ".db",
			"database_path":  dbPath,
			"total_requests": stats.TotalRequests,
			"total_hosts":    stats.TotalHosts,
			"first_request":  stats.FirstRequest,
			"last_request":   stats.LastRequest,
			"type":           "local",
			"requests":       stats.TotalRequests,
			"size_bytes":     fileInfo["size_bytes"],
			"size_formatted": fileInfo["size_formatted"],
			"modified_time":  fileInfo["modified_time"],
			"file_exists":    fileInfo["exists"],
		})
	}

	return Result{
		Data: map[string]interface{}{
			"projects": projects,
			"success":  true,
		},
	}
}

// scanProjectDirs 扫描 db 目录下的项目子目录
func scanProjectDirs(dbDir string) []string {
	var projectDirs []string

	entries, err := os.ReadDir(dbDir)
	if err != nil {
		// 目录不存在时返回空列表
		if os.IsNotExist(err) {
			return projectDirs
		}
		logging.Logger.Errorf("读取 db 目录失败: %v", err)
		return projectDirs
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// 检查子目录中是否存在同名的 .db 文件
			projectName := entry.Name()
			dbPath := filepath.Join(dbDir, projectName, projectName+".db")
			if _, err := os.Stat(dbPath); err == nil {
				projectDirs = append(projectDirs, projectName)
			}
		}
	}

	return projectDirs
}

// DBStats 数据库统计信息
type DBStats struct {
	TotalRequests int
	TotalHosts    int
	FirstRequest  time.Time
	LastRequest   time.Time
}

// getDBFileStats 获取数据库文件的统计信息
func getDBFileStats(dbPath string) DBStats {
	stats := DBStats{
		TotalRequests: 0,
		TotalHosts:    0,
		FirstRequest:  time.Now(),
		LastRequest:   time.Now(),
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return stats
	}

	// 打开数据库连接查询统计信息
	database, err := db.OpenDatabase(dbPath)
	if err != nil {
		logging.Logger.Warnf("打开数据库失败 %s: %v", dbPath, err)
		return stats
	}
	defer func() {
		if sqlDB, err := database.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// 查询请求总数
	var totalRequests int64
	if err := database.Table("http_histories").Count(&totalRequests).Error; err == nil {
		stats.TotalRequests = int(totalRequests)
	}

	// 查询不同主机数
	var totalHosts int64
	if err := database.Table("http_histories").Distinct("host").Count(&totalHosts).Error; err == nil {
		stats.TotalHosts = int(totalHosts)
	}

	// 查询第一条和最后一条请求的时间
	var firstRecord, lastRecord db.HTTPHistory
	if err := database.Table("http_histories").Order("created_at ASC").First(&firstRecord).Error; err == nil {
		stats.FirstRequest = firstRecord.CreatedAt
	}
	if err := database.Table("http_histories").Order("created_at DESC").First(&lastRecord).Error; err == nil {
		stats.LastRequest = lastRecord.CreatedAt
	}

	return stats
}

// getDBFileInfo 获取数据库文件信息（包括大小）
func getDBFileInfo(dbPath string) map[string]interface{} {
	info := map[string]interface{}{
		"size_bytes":     int64(0),
		"size_formatted": "0 B",
		"modified_time":  time.Now(),
		"exists":         false,
	}

	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		return info
	}

	size := fileInfo.Size()
	info["size_bytes"] = size
	info["size_formatted"] = formatFileSize(size)
	info["modified_time"] = fileInfo.ModTime()
	info["exists"] = true

	return info
}

// formatFileSize 格式化文件大小为可读字符串
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CreateLocalProject 在本地创建新项目（SQLite数据库）
func (c *ConfigAPI) CreateLocalProject(projectID string, projectName string) Result {
	logging.Logger.Infof("📁 ConfigAPI.CreateLocalProject 开始: projectID=%s, projectName=%s", projectID, projectName)

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ChYing")
	dbDir := filepath.Join(configDir, "db")

	// 确保 db 目录存在
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return Result{
			Error: fmt.Sprintf("创建 db 目录失败: %v", err),
		}
	}

	// 使用 projectName 作为目录和数据库文件名（清理特殊字符）
	safeFileName := strings.ReplaceAll(projectName, " ", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "/", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "\\", "_")
	safeFileName = strings.ReplaceAll(safeFileName, ":", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "..", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "<", "_")
	safeFileName = strings.ReplaceAll(safeFileName, ">", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "|", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "\"", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "*", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "?", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "\x00", "")
	if safeFileName == "" {
		safeFileName = "project"
	}
	safeFileName = strings.TrimLeft(safeFileName, ".")
	if safeFileName == "" {
		safeFileName = "project"
	}

	// 项目目录: db/<projectName>/
	projectDir := filepath.Join(dbDir, safeFileName)

	// 检查项目目录是否已存在
	if _, err := os.Stat(projectDir); err == nil {
		return Result{
			Error: fmt.Sprintf("项目 '%s' 已存在", projectName),
		}
	}

	// 创建项目目录
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return Result{
			Error: fmt.Sprintf("创建项目目录失败: %v", err),
		}
	}

	// 数据库文件路径: db/<projectName>/<projectName>.db
	dbFileName := fmt.Sprintf("%s.db", safeFileName)
	dbPath := filepath.Join(projectDir, dbFileName)

	// 创建空的 SQLite 数据库文件
	f, err := os.Create(dbPath)
	if err != nil {
		return Result{
			Error: fmt.Sprintf("创建数据库文件失败: %v", err),
		}
	}
	f.Close()

	logging.Logger.Infof("✓ 本地项目创建成功: %s -> %s", projectName, dbPath)

	return Result{
		Data: map[string]interface{}{
			"project_id":    projectID,
			"project_name":  safeFileName,
			"database_file": dbFileName,
			"database_path": dbPath,
			"success":       true,
			"message":       "本地项目创建成功",
		},
	}
}

// DeleteLocalProject 删除本地项目（删除整个项目目录）
func (c *ConfigAPI) DeleteLocalProject(projectName string) Result {
	logging.Logger.Infof("🗑️ ConfigAPI.DeleteLocalProject 开始: projectName=%s", projectName)

	if projectName == "" {
		return Result{
			Error: "项目名称不能为空",
		}
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ChYing")
	dbDir := filepath.Join(configDir, "db")

	// 清理项目名称（与创建时保持一致）
	safeFileName := strings.ReplaceAll(projectName, " ", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "/", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "\\", "_")
	safeFileName = strings.ReplaceAll(safeFileName, ":", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "..", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "<", "_")
	safeFileName = strings.ReplaceAll(safeFileName, ">", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "|", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "\"", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "*", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "?", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "\x00", "")
	safeFileName = strings.TrimLeft(safeFileName, ".")

	// 项目目录: db/<projectName>/
	projectDir := filepath.Join(dbDir, safeFileName)

	// 检查项目目录是否存在
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return Result{
			Error: fmt.Sprintf("项目 '%s' 不存在", projectName),
		}
	}

	// 安全检查：确保路径在 db 目录下，防止路径遍历攻击
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return Result{
			Error: fmt.Sprintf("获取项目路径失败: %v", err),
		}
	}
	absDbDir, err := filepath.Abs(dbDir)
	if err != nil {
		return Result{
			Error: fmt.Sprintf("获取数据库目录路径失败: %v", err),
		}
	}
	if !strings.HasPrefix(absProjectDir, absDbDir) {
		return Result{
			Error: "无效的项目路径",
		}
	}

	// 删除整个项目目录
	if err := os.RemoveAll(projectDir); err != nil {
		logging.Logger.Errorf("删除项目目录失败: %v", err)
		return Result{
			Error: fmt.Sprintf("删除项目失败: %v", err),
		}
	}

	logging.Logger.Infof("✓ 本地项目删除成功: %s", projectName)

	return Result{
		Data: map[string]interface{}{
			"project_name": projectName,
			"success":      true,
			"message":      "项目删除成功",
		},
	}
}

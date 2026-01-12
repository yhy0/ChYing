package conf

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/12/10
   @desc 配置文件管理 - 路径、加载、保存等功能
**/

// AppConf 全局应用配置实例
var AppConf AppConfig

// GetConfigFilePath 获取配置文件路径
func GetConfigFilePath() string {
	return filepath.Join(file.ChyingDir, "config.yaml")
}

// loadConfigFromViper 从 viper 加载配置到 AppConf
func loadConfigFromViper() {
	if err := viper.Unmarshal(&AppConf); err != nil {
		logging.Logger.Errorf("解析配置文件失败: %v", err)
	}
}

// GetAppConfig 获取应用配置（返回指针以便直接修改）
func GetAppConfig() *AppConfig {
	return &AppConf
}

// UpdateAppConfig 更新应用配置
func UpdateAppConfig(config AppConfig) error {
	AppConf = config
	return SaveConfig()
}

// ValidateAppConfig 验证应用配置
func ValidateAppConfig() error {
	// 基本验证逻辑
	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig() error {
	// 将 AppConf 的值同步到 viper
	// AI 配置
	viper.Set("ai.claude.cli_path", AppConf.AI.Claude.CLIPath)
	viper.Set("ai.claude.work_dir", AppConf.AI.Claude.WorkDir)
	viper.Set("ai.claude.model", AppConf.AI.Claude.Model)
	viper.Set("ai.claude.max_turns", AppConf.AI.Claude.MaxTurns)
	viper.Set("ai.claude.system_prompt", AppConf.AI.Claude.SystemPrompt)
	viper.Set("ai.claude.permission_mode", AppConf.AI.Claude.PermissionMode)
	viper.Set("ai.claude.require_tool_confirm", AppConf.AI.Claude.RequireToolConfirm)
	viper.Set("ai.claude.api_key", AppConf.AI.Claude.APIKey)
	viper.Set("ai.claude.base_url", AppConf.AI.Claude.BaseURL)
	viper.Set("ai.claude.temperature", AppConf.AI.Claude.Temperature)

	// MCP 配置
	viper.Set("ai.claude.mcp.enabled", AppConf.AI.Claude.MCP.Enabled)
	viper.Set("ai.claude.mcp.mode", AppConf.AI.Claude.MCP.Mode)
	viper.Set("ai.claude.mcp.port", AppConf.AI.Claude.MCP.Port)
	viper.Set("ai.claude.mcp.enabled_tools", AppConf.AI.Claude.MCP.EnabledTools)
	viper.Set("ai.claude.mcp.disabled_tools", AppConf.AI.Claude.MCP.DisabledTools)
	viper.Set("ai.claude.mcp.external_servers", AppConf.AI.Claude.MCP.ExternalServers)

	return viper.WriteConfig()
}

// ReloadConfig 重新加载配置
func ReloadConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	loadConfigFromViper()
	SyncJieConfig()
	return nil
}

// BackupConfig 备份配置文件
func BackupConfig() error {
	configPath := GetConfigFilePath()
	backupPath := configPath + ".backup"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, data, 0644)
}

// RestoreConfig 恢复配置文件
func RestoreConfig(backupFile string) error {
	configPath := GetConfigFilePath()

	data, err := os.ReadFile(backupFile)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	return ReloadConfig()
}

// InitConfig 初始化配置
// 注意：此函数必须在 logging.Logger 初始化之后调用
func InitConfig() {
	configPath := GetConfigFilePath()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建默认配置文件
		if err := os.WriteFile(configPath, defaultConfigYaml, 0644); err != nil {
			if logging.Logger != nil {
				logging.Logger.Errorf("创建默认配置文件失败: %v", err)
			}
			return
		}
		if logging.Logger != nil {
			logging.Logger.Infoln("已创建默认配置文件:", configPath)
		}
	}

	// 读取配置文件
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		if logging.Logger != nil {
			logging.Logger.Errorf("读取配置文件失败: %v", err)
		}
		return
	}

	// 加载配置到结构体
	loadConfigFromViper()
}

package main

import (
	"bytes"

	"github.com/spf13/viper"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/conf/file"
	JieConf "github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 配置管理相关方法
**/

// GetConfig 获取配置
func (a *App) GetConfig() string {
	return a.apiManager.GetConfig()
}

// GetProxyConfigure 获取代理配置
func (a *App) GetProxyConfigure() *conf.Configure {
	return conf.Config
}

// GetProxy 获取代理
func (a *App) GetProxy() string {
	return conf.Proxy
}

// SetProxy 设置代理
func (a *App) SetProxy(proxy string) string {
	JieConf.GlobalConfig.Http.Proxy = proxy
	httpx.NewClient(nil)
	return conf.Proxy
}

// GetToken 获取Token
func (a *App) GetToken() string {
	return conf.Token
}

// SetToken 设置Token
func (a *App) SetToken(token string) string {
	conf.Token = token
	return conf.Token
}

// GetJieConfigure 获取Jie配置
func (a *App) GetJieConfigure() map[string]bool {
	return JieConf.Plugin
}

// GetJieConfigureFileContent 获取配置文件内容
// 使用 ChYing 统一配置文件
// 如需独立 Jie，改回: utils.ReadFile(path.Join(file.ChyingDir, JieConf.FileName))
func (a *App) GetJieConfigureFileContent() (string, error) {
	content, err := utils.ReadFile(conf.GetConfigFilePath())
	if err != nil {
		return "", err
	}

	// 规范化键名：将小写键名转换为驼峰式键名
	// 这是因为 viper 会将所有键名转换为小写，但前端期望驼峰式键名
	normalized, err := conf.NormalizeConfigYAML(content)
	if err != nil {
		// 如果规范化失败，返回原始内容
		logging.Logger.Warnln("Failed to normalize config keys:", err)
		return content, nil
	}

	return normalized, nil
}

// ModifyJieConfigureFileContent 修改配置文件内容
// 使用 ChYing 统一配置文件
// 如需独立 Jie，改回: utils.WriteFile(path.Join(file.ChyingDir, JieConf.FileName), content)
func (a *App) ModifyJieConfigureFileContent(content string) error {
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBufferString(content))
	if err != nil {
		logging.Logger.Errorln("Configuration file verification failed:", err)
		Notify <- []string{err.Error(), "error"}
		return err
	}

	// 写入统一配置文件
	err = utils.WriteFile(conf.GetConfigFilePath(), content)
	if err != nil {
		logging.Logger.Errorln("Failed to write config file:", err)
		Notify <- []string{err.Error(), "error"}
		return err
	}

	// 重新加载配置并同步到 Jie
	// 注意：由于配置文件热加载机制，这里的同步可能会被触发两次
	// 但这是安全的，因为同步操作是幂等的
	conf.SyncJieConfig()

	return nil
}

// GetProject 获取项目列表
func (a *App) GetProject() map[string]string {
	dbFiles, err := utils.GetDBFiles(file.ChyingDir)
	if err != nil {
		logging.Logger.Errorln("get db files error:", err)
		Notify <- []string{err.Error(), "error"}
		return nil
	}
	return dbFiles
}

// ConfigurePlugin 配置插件
func (a *App) ConfigurePlugin(name string, change bool) map[string]bool {
	if _, ok := JieConf.Plugin[name]; ok {
		JieConf.Plugin[name] = change
	}

	if name == "all" {
		for k := range JieConf.Plugin {
			JieConf.Plugin[k] = change
		}
	}

	return JieConf.Plugin
}

// GetAppConfig 获取完整应用配置
func (a *App) GetAppConfig() conf.AppConfig {
	logging.Logger.Infoln("获取应用配置")
	logging.Logger.Infof("配置文件路径: %s", conf.GetConfigFilePath())
	logging.Logger.Infof("配置内容: %+v", conf.GetAppConfig())
	return *conf.GetAppConfig()
}

// UpdateAppConfig 更新完整应用配置
func (a *App) UpdateAppConfig(config conf.AppConfig) error {
	// 验证配置
	if err := conf.ValidateAppConfig(); err != nil {
		return err
	}

	// 更新配置
	return conf.UpdateAppConfig(config)
}

// GetConfigFilePath 获取配置文件路径
func (a *App) GetConfigFilePath() string {
	return conf.GetConfigFilePath()
}

// ReloadConfig 重新加载配置
func (a *App) ReloadConfig() error {
	return conf.ReloadConfig()
}

// BackupConfig 备份配置文件
func (a *App) BackupConfig() error {
	return conf.BackupConfig()
}

// RestoreConfig 恢复配置文件
func (a *App) RestoreConfig(backupFile string) error {
	return conf.RestoreConfig(backupFile)
}

// GetConfigStatus 获取配置状态信息
func (a *App) GetConfigStatus() map[string]interface{} {
	return a.apiManager.GetConfigStatus()
}

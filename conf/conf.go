package conf

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	JieConf "github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/12/10
   @desc 配置管理 - 统一配置文件，同步Jie扫描配置
**/

var Config *Configure

// 防抖机制相关变量
var (
	reloadMutex    sync.Mutex
	lastReloadTime time.Time
	reloadDebounce = 2 * time.Second // 防抖间隔
	isReloading    bool
)

// HotConf 使用 viper 对配置热加载
func HotConf() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(GetConfigFilePath())

	// watch 监控配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 防抖处理：避免频繁重复的配置文件变化事件
		reloadMutex.Lock()
		defer reloadMutex.Unlock()

		now := time.Now()
		if isReloading || now.Sub(lastReloadTime) < reloadDebounce {
			return
		}

		isReloading = true
		lastReloadTime = now

		// 延迟处理，确保文件写入完成
		go func() {
			time.Sleep(100 * time.Millisecond)

			defer func() {
				reloadMutex.Lock()
				isReloading = false
				reloadMutex.Unlock()
			}()

			// 重新读取配置文件
			if err := viper.ReadInConfig(); err != nil {
				logging.Logger.Errorf("重新读取配置文件失败: %v", err)
				return
			}

			// 重新加载配置到结构体
			loadConfigFromViper()

			// 同步Jie扫描配置
			SyncJieConfig()

			logging.Logger.Infoln("配置已重新加载")
		}()
	})
}

// SyncJieConfig 同步Jie扫描配置
// 将 ChYing 的统一配置同步到 JieConf.GlobalConfig，使 Jie 扫描器可以正常工作
// 如需独立 Jie，取消 JieConf.Init() 调用，Jie 将使用自己的配置文件
func SyncJieConfig() {
	// === 同步 Jie 扫描配置 ===
	// 由于 AppConfig 中的 Jie 相关字段直接使用 JieConf 的类型，可以直接赋值

	// 同步 HTTP 配置
	JieConf.GlobalConfig.Http = AppConf.Http

	// 同步插件配置
	JieConf.GlobalConfig.Plugins = AppConf.Plugins

	// 同步反连平台配置
	JieConf.GlobalConfig.Reverse = AppConf.Reverse

	// 同步 Mitmproxy 配置
	JieConf.GlobalConfig.Mitmproxy = AppConf.Mitmproxy

	// 同步信息收集配置
	JieConf.GlobalConfig.Collection = AppConf.Collection

	// 更新插件开关 map（调用 Jie 的 ReadPlugin 函数）
	JieConf.ReadPlugin()

	// === 同步 ChYing 的 Config（代理配置）===

	// 重置配置
	if Config != nil {
		Config.Exclude = nil
		Config.Include = nil
	}

	// 读取配置文件中的 Exclude 配置
	for index, v := range AppConf.Mitmproxy.Exclude {
		if v == "" {
			continue
		}

		Config.Exclude = append(Config.Exclude, &Scope{
			Id:      index,
			Enabled: true,
			Prefix:  v,
			Regexp:  true,
			Type:    "exclude",
		})
	}

	// 读取配置文件中的 Include 配置
	for index, v := range AppConf.Mitmproxy.Include {
		if v == "" {
			continue
		}
		Config.Include = append(Config.Include, &Scope{
			Id:      index,
			Enabled: true,
			Prefix:  v,
			Regexp:  true,
			Type:    "include",
		})
	}

	logging.Logger.Infoln("Jie扫描配置已同步")
}

package burpSuite

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	folderutil "github.com/projectdiscovery/utils/folder"
	"github.com/spf13/viper"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
	"path"
	"path/filepath"
)

/**
  @author: yhy
  @since: 2023/5/18
  @desc: //TODO
**/

var FilterSuffix = []string{".woff2", ".woff", ".ttf", ".mkv", ".mov", ".mp3", ".mp4", ".m4a", ".m4v"}

var Settings *Setting

var defaultYamlByte = []byte(`
port: 9080
exclude:
  - ^.*\.google.*$
  - ^.*\.firefox.*$
  - ^.*\.doubleclick.*$
  - ^.*\.mozilla.*$
include:
  - 
`)

var configFileName = "burpsuite.yaml"

var configFile string

// HotConf 使用 viper 对配置热加载
func HotConf() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	// watch 监控配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件发生变更之后会调用的回调函数
		logging.Logger.Infoln("config file changed: ", e.Name)
		ReadYamlConfig()
	})
}

// Init 加载配置
func Init() {
	homedir := folderutil.HomeDirOrDefault("")

	userCfgDir := filepath.Join(homedir, ".config")

	filePath := filepath.Join(userCfgDir, "ChYing")

	// 配置文件路径 当前文件夹 + config.yaml
	configFile = path.Join(filePath, configFileName)

	// 检测配置文件是否存在
	if !utils.Exists(configFile) {
		WriteYamlConfig(nil)
		logging.Logger.Infof("%s not find, Generate profile.", configFile)
	} else {
		logging.Logger.Infoln("Load profile ", configFile)
	}
	ReadYamlConfig()
}

func ReadYamlConfig() {
	// 加载config
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	err := viper.ReadInConfig()
	if err != nil {
		logging.Logger.Fatalf("setting.Setup, fail to read 'config.yaml': %+v", err)
	}
	err = viper.Unmarshal(&Settings)
	if err != nil {
		logging.Logger.Fatalf("setting.Setup, fail to parse 'config.yaml', check format: %v", err)
	}
}

func WriteYamlConfig(str []byte) {
	if str == nil {
		str = defaultYamlByte
	}

	// 生成默认config
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(str))
	if err != nil {
		logging.Logger.Fatalf("setting.Setup, fail to read default config bytes: %v", err)
	}
	// 写文件
	err = viper.WriteConfigAs(configFile)
	if err != nil {
		logging.Logger.Fatalf("setting.Setup, fail to write 'config.yaml': %v", err)
	}
}

//func (s *Setting) SetPort(value int) {
//	s.ProxyPort = value
//	updateConfigFile(s)
//}
//
//func (s *Setting) SetExclude(value []string) {
//	s.Exclude = value
//	updateConfigFile(s)
//}
//
//func (s *Setting) SetInclude(value []string) {
//	s.Include = value
//	updateConfigFile(s)
//}
//
//func updateConfigFile(s *Setting) {
//	// 将结构体 写入配置文件
//	viper.Set("port", s.ProxyPort)
//	viper.Set("exclude", s.Exclude)
//	viper.Set("include", s.Include)
//	err := viper.WriteConfig()
//	if err != nil {
//		logging.Logger.Errorln(err)
//		return
//	}
//}

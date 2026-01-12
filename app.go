package main

import (
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sasha-s/go-deadlock"
	"github.com/yhy0/ChYing/api"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/mitmproxy"
	JieConf "github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/mode"
	"github.com/yhy0/ChYing/pkg/qqwry"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/7/12
   @desc App 核心结构体定义和初始化

   本文件包含:
   - App 结构体定义
   - 类型定义 (Result, InitStep, InitProgress, InitContext, MemoryInfo, Msg)
   - 全局变量声明
   - init() 初始化函数
   - Startup() 启动函数

   其他方法已拆分到以下文件:
   - app_initialization.go: 初始化相关方法
   - app_config.go: 配置管理方法
   - app_proxy.go: 代理和流量方法
   - app_database.go: 数据库和历史方法
   - app_scan.go: 扫描目标管理方法
   - app_remote.go: 远程节点/集群方法
   - app_window.go: 窗口管理方法
   - app_utils.go: 工具方法
**/

// App 应用主结构体
type App struct {
	apiManager *api.APIManager // API管理器
}

// Result 统一返回结果结构体
// 注意：不使用类型别名 (type Result = api.Result)，因为 Wails v3 binding 生成器
// 在处理类型别名时会产生错误的导入引用 ($0 而不是正确的模块名)
type Result struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

// InitStep 初始化步骤枚举
type InitStep int

const (
	StepBasicInit InitStep = iota
	StepConfigLoad
	StepDatabaseConnect
	StepSchemaCheck
	StepProxyStart
	StepProjectLoad
	StepCompleted
)

// InitProgress 初始化进度信息
type InitProgress struct {
	Step        InitStep `json:"step"`
	Progress    int      `json:"progress"`
	Message     string   `json:"message"`
	Description string   `json:"description"`
	Success     bool     `json:"success"`
	Error       string   `json:"error,omitempty"`
}

// InitContext 初始化上下文
type InitContext struct {
	ProjectType string    `json:"projectType"`
	ProjectName string    `json:"projectName"`
	StartTime   time.Time `json:"startTime"`
}

// MemoryInfo 内存使用信息
type MemoryInfo struct {
	// 总分配的内存（字节）
	Alloc uint64 `json:"alloc"`
	// 总分配的内存（格式化字符串）
	AllocFormatted string `json:"allocFormatted"`
	// 从系统分配的内存（字节）
	Sys uint64 `json:"sys"`
	// 从系统分配的内存（格式化字符串）
	SysFormatted string `json:"sysFormatted"`
	// 垃圾回收次数
	NumGC uint32 `json:"numGC"`
	// 协程数量
	NumGoroutine int `json:"numGoroutine"`
}

// Msg 消息结构体
type Msg struct {
	Target       string         `json:"target"`
	UUID         string         `json:"uuid"`
	CDN          bool           `json:"cdn"`
	IpAddress    string         `json:"ipAddress"`
	IPMsg        string         `json:"IPMsg"`
	Records      []string       `json:"records"`
	Fingerprint  []string       `json:"fingerprint"`
	PortInfo     map[int]string `json:"portInfo"`
	SiteMap      []string       `json:"site_map"`
	Children     []*Msg         `json:"children"`
	APICnt       int            `json:"api_cnt"`
	SubdomainCnt int            `json:"subdomain_cnt"`
	ParamsCnt    int            `json:"params_cnt"`
	InnerIpCnt   int            `json:"inner_ip_cnt"`
	OtherCnt     int            `json:"other_cnt"`
}

// 全局变量
var (
	RePercentage   chan float64
	Percentage     chan float64
	Notify         chan []string
	Pool           *ants.Pool // 入库协程
	lock           sync.Mutex
	HTTPHistoryMap sync.Map
)

// init 初始化函数
func init() {
	Percentage = make(chan float64, 1)
	RePercentage = make(chan float64, 1)
	Notify = make(chan []string, 1)
	logging.Logger = logging.New(true, file.ChyingDir, "ChYing", true)
	file.New()

	var err error
	Pool, err = ants.NewPool(conf.Parallelism)
	if err != nil {
		logging.Logger.Errorf("创建协程池失败: %v", err)
	}

	JieConf.Wappalyzer, err = wappalyzer.New()
	if err != nil {
		logging.Logger.Warnf("初始化 Wappalyzer 失败: %v", err)
	}

	// 使用统一配置文件系统，不再单独初始化 Jie 配置
	conf.HotConf()

	// 从统一配置中同步 Jie 扫描配置
	conf.SyncJieConfig()

	for _, suffix := range strings.Split(conf.AppConf.Mitmproxy.FilterSuffix, ", ") {
		conf.Config.FilterSuffix = append(conf.Config.FilterSuffix, suffix)
	}

	// 读取配置文件中的配置
	for index, v := range conf.AppConf.Mitmproxy.Exclude {
		if v == "" {
			continue
		}

		conf.Config.Exclude = append(conf.Config.Exclude, &conf.Scope{
			Id:      index,
			Enabled: true,
			Prefix:  v,
			Regexp:  true,
			Type:    "exclude",
		})
	}
	for index, v := range conf.AppConf.Mitmproxy.Include {
		if v == "" {
			continue
		}
		conf.Config.Include = append(conf.Config.Include, &conf.Scope{
			Id:      index,
			Enabled: true,
			Prefix:  v,
			Regexp:  true,
			Type:    "include",
		})
	}
}

// Startup 应用启动函数
func Startup() {
	// 设置 Jie 为被动模式
	mode.Passive()

	logging.Logger.Infoln("[+] ChYing 启动成功", conf.Config.FilterSuffix)
	logging.Logger.Infoln(conf.Config.Exclude)
	// 插件全部关闭
	for k := range JieConf.Plugin {
		JieConf.Plugin[k] = false
	}

	deadlock.Opts.DeadlockTimeout = 120 * time.Second
	qqwry.Init()

	// 检测端口是否被占用
	if utils.IsPortOccupied(conf.ProxyPort) {
		port, err := utils.GetRandomUnusedPort()
		if err != nil {
			logging.Logger.Errorln(err)
			conf.ProxyPort = 65530
		} else {
			conf.ProxyPort = port
		}
	}

	// 启动被动代理（异步）
	go func() {
		logging.Logger.Infoln("Starting Proxify server in a new goroutine...")
		mitmproxy.Proxify() // 这个函数内部会调用阻塞的 Run()
		// 如果 Proxify() 返回 (例如发生错误导致服务停止)，这里可以记录日志
		logging.Logger.Errorln("Proxify server has stopped.")
	}()

	// 数据更改通知
	go EventNotification()

	// 移除自动内存清理，改为前端手动触发
}

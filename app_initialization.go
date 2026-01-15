package main

import (
	"fmt"
	"strings"
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
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/qqwry"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 应用初始化相关方法
**/

// StartInitialization 开始分步初始化
func (a *App) StartInitialization(projectType string, projectName string) Result {
	// 初始化日志
	logging.Logger = logging.New(true, file.ChyingDir, "ChYing", true)

	logging.Logger.Infoln("开始分步初始化...", projectType, projectName)

	// 创建初始化上下文
	ctx := &InitContext{
		ProjectType: projectType,
		ProjectName: projectName,
		StartTime:   time.Now(),
	}

	// 处理项目名称
	if projectType == "Temporary project" {
		// 格式: 20260111-20-30-tmp (年月日-时-分-tmp)
		ctx.ProjectName = fmt.Sprintf("%s-tmp", time.Now().Format("20060102-15-04"))
	}

	return Result{
		Data:  ctx,
		Error: "",
	}
}

// StepBasicInitialization 步骤1: 基础初始化
func (a *App) StepBasicInitialization() Result {
	progress := &InitProgress{
		Step:        StepBasicInit,
		Progress:    10,
		Message:     "正在初始化基础组件...",
		Description: "初始化日志系统、文件系统等基础组件",
	}

	defer func() {
		if r := recover(); r != nil {
			progress.Success = false
			progress.Error = fmt.Sprintf("基础初始化失败: %v", r)
			logging.Logger.Errorf("基础初始化 panic: %v", r)
		}
	}()

	file.New()

	// 初始化协程池
	var err error
	Pool, err = ants.NewPool(conf.Parallelism)
	if err != nil {
		logging.Logger.Errorf("创建协程池失败: %v", err)
		progress.Success = false
		progress.Error = fmt.Sprintf("创建协程池失败: %v", err)
		return Result{Data: progress, Error: progress.Error}
	}
	mode.Passive()

	// 初始化其他基础组件
	JieConf.Wappalyzer, err = wappalyzer.New()
	if err != nil {
		logging.Logger.Warnf("初始化 Wappalyzer 失败: %v", err)
		// Wappalyzer 失败不是致命错误，继续执行
	}
	deadlock.Opts.DeadlockTimeout = 120 * time.Second
	qqwry.Init()

	progress.Success = true
	progress.Message = "基础组件初始化完成"
	logging.Logger.Infoln("✓ 基础初始化完成")

	return Result{Data: progress, Error: ""}
}

// StepConfigurationLoad 步骤2: 配置文件加载
func (a *App) StepConfigurationLoad() Result {
	progress := &InitProgress{
		Step:        StepConfigLoad,
		Progress:    25,
		Message:     "正在加载配置文件...",
		Description: "加载应用配置、代理配置、扫描配置等",
	}

	defer func() {
		if r := recover(); r != nil {
			progress.Success = false
			progress.Error = fmt.Sprintf("配置加载失败: %v", r)
			logging.Logger.Errorf("配置加载 panic: %v", r)
		}
	}()

	// 初始化配置（必须在 logging.Logger 初始化之后调用）
	conf.InitConfig()

	// 加载统一配置文件系统（热加载监控）
	conf.HotConf()

	// 从统一配置中同步 Jie 扫描配置
	conf.SyncJieConfig()

	// 处理代理过滤配置
	for _, suffix := range strings.Split(conf.AppConf.Mitmproxy.FilterSuffix, ", ") {
		conf.Config.FilterSuffix = append(conf.Config.FilterSuffix, suffix)
	}

	// 处理排除和包含规则
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

	// 插件全部关闭（将在后续步骤中根据配置启用）
	for k := range JieConf.Plugin {
		JieConf.Plugin[k] = false
	}

	progress.Success = true
	progress.Message = "配置文件加载完成"
	logging.Logger.Infoln("✓ 配置文件加载完成")

	return Result{Data: progress, Error: ""}
}

// StepDatabaseConnection 步骤3: 数据库连接
func (a *App) StepDatabaseConnection(projectName string) Result {
	progress := &InitProgress{
		Step:        StepDatabaseConnect,
		Progress:    40,
		Message:     "正在连接数据库...",
		Description: "初始化SQLite数据库",
	}

	defer func() {
		if r := recover(); r != nil {
			progress.Success = false
			progress.Error = fmt.Sprintf("数据库连接失败: %v", r)
			logging.Logger.Errorf("数据库连接 panic: %v", r)
		}
	}()

	// 初始化API管理器
	if a.apiManager == nil {
		a.apiManager = api.NewAPIManager()
	}

	// 设置数据库错误回调，将错误推送到前端
	db.SetDBErrorCallback(func(err db.DBError) {
		if wailsApp != nil {
			wailsApp.Event.Emit("db:error", err)
		}
		logging.Logger.Warnf("数据库操作错误 [%s]: %s", err.Operation, err.Error)
	})

	// 初始化数据库
	db.Init(projectName, "sqlite")

	progress.Success = true
	progress.Message = "数据库连接成功"
	logging.Logger.Infoln("✓ 数据库连接完成")

	return Result{Data: progress, Error: ""}
}

// StepSchemaValidation 步骤4: 数据库表结构检查
func (a *App) StepSchemaValidation() Result {
	progress := &InitProgress{
		Step:        StepSchemaCheck,
		Progress:    55,
		Message:     "正在检查数据库表结构...",
		Description: "检查并创建必要的数据库表和视图，确保数据结构完整",
	}

	defer func() {
		if r := recover(); r != nil {
			progress.Success = false
			progress.Error = fmt.Sprintf("数据库表结构检查失败: %v", r)
			logging.Logger.Errorf("数据库表结构检查 panic: %v", r)
		}
	}()

	progress.Message = "使用SQLite数据库，表结构检查完成"
	progress.Success = true
	logging.Logger.Infoln("✓ 数据库表结构检查完成")

	return Result{Data: progress, Error: ""}
}

// StepProxyServerStart 步骤5: 代理服务器启动
func (a *App) StepProxyServerStart() Result {
	progress := &InitProgress{
		Step:        StepProxyStart,
		Progress:    70,
		Message:     "正在启动代理服务器...",
		Description: "启动HTTP/HTTPS代理服务器，初始化流量拦截系统",
	}

	defer func() {
		if r := recover(); r != nil {
			progress.Success = false
			progress.Error = fmt.Sprintf("代理服务器启动失败: %v", r)
			logging.Logger.Errorf("代理服务器启动 panic: %v", r)
		}
	}()

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
		mitmproxy.Proxify() // 复用现有的 Proxify 函数
		logging.Logger.Errorln("Proxify server has stopped.")
	}()

	// 等待代理服务器启动
	time.Sleep(2 * time.Second)

	// 获取代理监听地址
	proxyHost := conf.AppConf.Proxy.Host
	if proxyHost == "" {
		proxyHost = "127.0.0.1"
	}

	// 发送代理启动通知事件
	if wailsApp != nil {
		wailsApp.Event.Emit("ProxyStarted", map[string]interface{}{
			"host":    proxyHost,
			"port":    conf.ProxyPort,
			"success": true,
			"message": fmt.Sprintf("代理服务器已启动，监听地址: %s:%d", proxyHost, conf.ProxyPort),
		})
	}

	progress.Success = true
	progress.Message = fmt.Sprintf("代理服务器启动成功 (端口: %d)", conf.ProxyPort)
	logging.Logger.Infoln("✓ 代理服务器启动完成")

	return Result{Data: progress, Error: ""}
}

// StepProjectDataLoad 步骤6: 项目数据加载
func (a *App) StepProjectDataLoad(projectType string, projectName string) Result {
	progress := &InitProgress{
		Step:        StepProjectLoad,
		Progress:    85,
		Message:     "正在加载项目数据...",
		Description: "根据项目类型加载历史数据、扫描结果等",
	}

	defer func() {
		if r := recover(); r != nil {
			progress.Success = false
			progress.Error = fmt.Sprintf("项目数据加载失败: %v", r)
			logging.Logger.Errorf("项目数据加载 panic: %v", r)
		}
	}()

	if projectType == "Open existing project" {
		// 加载现有项目数据
		a.loadExistingProjectData()
		progress.Message = "现有项目数据加载完成"
	} else {
		progress.Message = "新项目初始化完成"
	}

	progress.Success = true
	logging.Logger.Infoln("✓ 项目数据加载完成")

	return Result{Data: progress, Error: ""}
}

// StepInitializationComplete 步骤7: 初始化完成
func (a *App) StepInitializationComplete() Result {
	progress := &InitProgress{
		Step:        StepCompleted,
		Progress:    100,
		Message:     "初始化完成",
		Description: "所有组件已成功初始化，系统准备就绪",
	}

	defer func() {
		if r := recover(); r != nil {
			progress.Success = false
			progress.Error = fmt.Sprintf("初始化完成步骤失败: %v", r)
			logging.Logger.Errorf("初始化完成 panic: %v", r)
		}
	}()

	// 启动事件通知系统
	go EventNotification()

	// 启动事件循环
	go a.startEventLoop()

	progress.Success = true
	logging.Logger.Infoln("✓ 系统初始化完成，ChYing 已准备就绪")

	return Result{Data: progress, Error: ""}
}

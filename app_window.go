package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/chromedp/chromedp"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/lib/webUnPack"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 窗口管理相关方法
**/

var (
	scanWindow    *application.WebviewWindow
	vulnWindow    *application.WebviewWindow
	claudeWindow  *application.WebviewWindow
	browserCancel context.CancelFunc
)

// NewScanLogWindow 创建扫描日志窗口
func (a *App) NewScanLogWindow() {
	// 如果窗口存在且未被销毁，聚焦它
	if scanWindow != nil {
		scanWindow.Show()
		scanWindow.Focus()
		return
	}

	scanWindow = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "Scan Log",
		Title:           "扫描日志",
		DevToolsEnabled: true, // 启用 DevTools (F12)
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropLiquidGlass,
			InvisibleTitleBarHeight: 25, // 只让顶部 25px 可拖拽，避免影响文字选择
			LiquidGlass: application.MacLiquidGlass{
				Style:        application.LiquidGlassStyleLight,
				Material:     application.NSVisualEffectMaterialAuto,
				CornerRadius: 20.0,
				TintColor:    nil,
			},
		},
		Width:          1200,
		Height:         900,
		URL:            "/#/scanLog",
		EnableFileDrop: true,
	})

	// 注册窗口关闭事件
	scanWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		scanWindow = nil // 清空引用
	})
	scanWindow.Show()
}

// NewVulnerabilityWindow 创建漏洞窗口
func (a *App) NewVulnerabilityWindow() {
	// 如果窗口存在且未被销毁，聚焦它
	if vulnWindow != nil {
		vulnWindow.Show()
		vulnWindow.Focus()
		return
	}
	vulnWindow = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "Vulnerability",
		Title:           "漏洞列表",
		DevToolsEnabled: true, // 启用 DevTools (F12)
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropLiquidGlass,
			InvisibleTitleBarHeight: 25, // 只让顶部 25px 可拖拽，避免影响文字选择
			LiquidGlass: application.MacLiquidGlass{
				Style:        application.LiquidGlassStyleLight,
				Material:     application.NSVisualEffectMaterialAuto,
				CornerRadius: 20.0,
				TintColor:    nil,
			},
		},
		Width:  1200,
		Height: 900,
		URL:    "/#/vulnerability",
	})
	// 注册窗口关闭事件
	vulnWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		vulnWindow = nil // 清空引用
	})
	vulnWindow.Show()
}

// NewClaudeAgentWindow 创建Claude AI Agent窗口
// trafficIds: 可选的流量ID列表，用于预填充分析请求
func (a *App) NewClaudeAgentWindow(trafficIds []int64) {
	// 构建 URL
	url := "/#/claude-agent"
	if len(trafficIds) > 0 {
		// 将流量 ID 转换为逗号分隔的字符串
		idsStr := ""
		for i, id := range trafficIds {
			if i > 0 {
				idsStr += ","
			}
			idsStr += fmt.Sprintf("%d", id)
		}
		url = fmt.Sprintf("/#/claude-agent?trafficIds=%s", idsStr)
	}

	// 如果窗口存在且未被销毁，聚焦它并更新 URL
	if claudeWindow != nil {
		claudeWindow.Show()
		claudeWindow.Focus()
		// 如果有新的流量 ID，通过事件通知前端
		if len(trafficIds) > 0 {
			wailsApp.Event.Emit("claude:traffic-ids", trafficIds)
		}
		return
	}
	claudeWindow = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "Claude Agent",
		Title:           "AI Security Assistant",
		DevToolsEnabled: true, // 启用 DevTools (F12)
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropLiquidGlass,
			InvisibleTitleBarHeight: 25, // 只让顶部 25px 可拖拽，避免影响文字选择
			LiquidGlass: application.MacLiquidGlass{
				Style:        application.LiquidGlassStyleLight,
				Material:     application.NSVisualEffectMaterialAuto,
				CornerRadius: 20.0,
				TintColor:    nil,
			},
		},
		Width:  1200,
		Height: 1000,
		URL:    url,
	})
	// 注册窗口关闭事件
	claudeWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		claudeWindow = nil // 清空引用
	})
	claudeWindow.Show()
}

// WebUnPack 解包Web资源
func (a *App) WebUnPack(target string) {
	webUnPack.Run(target)
}

// OpenChromeBrowser 打开Chrome浏览器
func (a *App) OpenChromeBrowser(proxy string) {
	InitBrowser(proxy)
}

// InitBrowser 初始化浏览器
func InitBrowser(proxy string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// 无头模式
		chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.ProxyServer(proxy),
		// 使用 <-loopback> 语法强制代理 localhost 和 127.0.0.1
		// Chrome 默认会绕过本地地址的代理，<-loopback> 表示"不绕过 loopback 地址"
		chromedp.Flag("proxy-bypass-list", "<-loopback>"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	bctx, ctxCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(logging.Logger.Printf),
	)

	// 保存两个取消函数，确保资源正确释放
	browserCancel = func() {
		ctxCancel()
		cancel()
	}

	// 启动浏览器
	if err := chromedp.Run(bctx); err != nil {
		browserCancel() // 出错时释放资源
		return err
	}

	return nil
}

// FileSelection 文件选择
func (a *App) FileSelection() string {
	return FileSelection()
}

// FileSelection 文件选择实现
func FileSelection() string {
	filePath := ""
	dialog := wailsApp.Dialog.OpenFile()
	dialog.SetTitle("Select Image")
	dialog.AddFilter("TXT Files", "*.txt")

	// Single file selection
	if path, err := dialog.PromptForSingleSelection(); err == nil {
		// Use selected file path
		filePath = path
		return filePath
	} else {
		filePath = filepath.Join(file.ChyingDir, "jwt.txt")
	}
	return filePath
}

package main

import (
	"embed"
	_ "embed"
	"fmt"

	"github.com/pkg/browser"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/logging"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed frontend/dist
var assets embed.FS

//go:embed frontend/public/appicon.png
var appIcon []byte

var wailsApp *application.App

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	wailsApp = application.New(application.Options{
		Name:        "承影",
		Description: "承影",
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(&App{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "承影",
		Title:           "承影",
		DevToolsEnabled: true, // 启用 DevTools (F12)
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropLiquidGlass,
			InvisibleTitleBarHeight: 25, // 只让顶部 25px 可拖拽，避免影响内容区域的文字选择
			LiquidGlass: application.MacLiquidGlass{
				Style:        application.LiquidGlassStyleLight,
				Material:     application.NSVisualEffectMaterialAuto,
				CornerRadius: 20.0,
				TintColor:    nil,
			},
		},
		URL:    "/",
		Width:  1400,
		Height: 900,
		// MinWidth:  100,
		// MinHeight: 900,
	})

	// Create the application menu
	menu := wailsApp.NewMenu()
	menu.AddRole(application.AppMenu)
	menu.AddRole(application.FileMenu)

	editMenu := menu.AddSubmenu("编辑")
	editMenu.AddRole(application.Copy)
	editMenu.AddRole(application.Cut)
	editMenu.AddRole(application.Paste)
	editMenu.AddRole(application.SelectAll)
	editMenu.AddRole(application.PasteAndMatchStyle)
	editMenu.AddRole(application.Delete)

	// 添加开发者工具菜单
	devMenu := menu.AddSubmenu("开发")
	devMenu.Add("打开开发者工具").
		SetAccelerator("CmdOrCtrl+Shift+I").
		OnClick(func(ctx *application.Context) {
			// 获取当前窗口并打开 DevTools
			if window := wailsApp.Window.Current(); window != nil {
				window.OpenDevTools()
			}
		})

	// 添加帮助菜单
	helpMenu := menu.AddSubmenu("帮助")
	helpMenu.Add("检查更新").
		OnClick(func(ctx *application.Context) {
			go menuCheckForUpdates()
		})

	wailsApp.Menu.Set(menu)

	// 绑定快捷键打开 DevTools
	// 尝试多种格式确保兼容性
	wailsApp.KeyBinding.Add("f12", func(window application.Window) {
		window.OpenDevTools()
	})
	wailsApp.KeyBinding.Add("CmdOrCtrl+Shift+i", func(window application.Window) {
		window.OpenDevTools()
	})

	err := wailsApp.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		logging.Logger.Fatal(err)
	}
}

// EventNotification 数据变动通知前端更改
func EventNotification() {
	for _data := range mitmproxy.EventDataChan {
		wailsApp.Event.Emit(_data.Name, _data.Data)
		err := Pool.Submit(func(_data *mitmproxy.EventData) func() {
			return func() {
				// 过代理的流量入库
				if _data.Name == "HttpHistory" {
					_http := _data.Data.(*mitmproxy.HTTPHistory)

					// 创建历史记录，包含流量来源信息
					historyData := &db.HTTPHistory{
						Hid:         _http.Id,
						Host:        _http.Host,
						Method:      _http.Method,
						FullUrl:     _http.FullUrl,
						Path:        _http.Path,
						Status:      _http.Status,
						Length:      _http.Length,
						ContentType: _http.ContentType,
						MIMEType:    _http.MIMEType,
						Extension:   _http.Extension,
						Title:       _http.Title,
						IP:          _http.IP,
						Color:       _http.Color,
						Note:        _http.Note,
						Source:      "local",     // 标识为本地流量
						SourceID:    "localhost", // 本地标识
						NodeName:    "本地节点",      // 节点名称
					}

					db.AddHistory(historyData)

					if _http.MIMEType == "image" {
						return
					}

					// 插入详细请求响应到SQLite
					traffic, _ok := mitmproxy.HTTPBodyMap.Load(_http.Id)
					if _ok {
						// 安全的类型断言
						httpBody, typeOk := traffic.(*mitmproxy.HTTPBody)
						if !typeOk {
							logging.Logger.Warnf("类型断言失败: traffic 不是 *mitmproxy.HTTPBody 类型, id: %d", _http.Id)
							return
						}

						// 使用SQLite存储
						req := &db.Request{
							RequestId:  uint(_http.Id),
							Url:        _http.FullUrl,
							Path:       _http.Path,
							Host:       _http.Host,
							RequestRaw: httpBody.RequestRaw,
						}
						resp := &db.Response{
							RequestId:   uint(_http.Id),
							Url:         _http.FullUrl,
							Host:        _http.Host,
							Path:        _http.Path,
							ContentType: _http.ContentType,
							ResponseRaw: httpBody.ResponseRaw,
						}
						db.AddRequest(req, resp)
					}
				}
			}
		}(_data))
		if err != nil {
			logging.Logger.Errorln(err)
		}
	}
}

// FileSelection 已移动到 app_window.go

// menuCheckForUpdates 菜单栏触发的版本检查
func menuCheckForUpdates() {
	updateInfo, err := checkGitHubRelease()
	if err != nil {
		logging.Logger.Warnf("菜单检查更新失败: %v", err)
		wailsApp.Dialog.Error().
			SetTitle("检查更新失败").
			SetMessage(fmt.Sprintf("无法检查更新: %v", err)).
			Show()
		return
	}

	if !updateInfo.HasUpdate {
		wailsApp.Dialog.Info().
			SetTitle("检查更新").
			SetMessage(fmt.Sprintf("当前版本 %s 已是最新版本！", conf.Version)).
			Show()
		return
	}

	// 发现新版本，弹出确认对话框
	dialog := wailsApp.Dialog.Question().
		SetTitle("发现新版本").
		SetMessage(fmt.Sprintf("当前版本: %s\n最新版本: %s\n\n是否前往 GitHub 下载最新版本？", updateInfo.CurrentVersion, updateInfo.LatestVersion))

	yesBtn := dialog.AddButton("前往下载").SetAsDefault()
	yesBtn.OnClick(func() {
		if err := browser.OpenURL(updateInfo.ReleaseURL); err != nil {
			logging.Logger.Errorf("打开下载页面失败: %v", err)
		}
	})

	noBtn := dialog.AddButton("稍后再说").SetAsCancel()
	_ = noBtn

	dialog.Show()
}

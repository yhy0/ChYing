package main

import (
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/logging"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logging.New(true, "CY")
	file.New()
	// Create an instance of the app structure
	app := NewApp()
	// Create application with options
	err := wails.Run(&options.App{
		Title: "承影",
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
		Menu: app.Menu(),
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

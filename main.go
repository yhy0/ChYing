package main

import (
	"embed"
	"fmt"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/logging"
	"time"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logging.New(true, "CY")
	file.New()
	year := time.Now().Year()
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
			About: &mac.AboutInfo{
				Title:   "承影",
				Message: fmt.Sprintf("将旦昧爽之交，日夕昏明之际，\n北面而察之，淡淡焉若有物存，莫识其状。\n其所触也，窃窃然有声，经物而物不疾也。\n\n© %d https://github.com/yhy0", year),
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

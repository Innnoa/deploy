package main

import (
	"embed"
	"recovery-unit-deploy/service/deploy"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "recovery-unit-deploy",
		Width:         1024,
		Height:        768,
		DisableResize: true, //禁用调整窗口尺寸
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app, &deploy.Deploy{}},
	})

	if err != nil {
		println("Error:", err.Error())
	}

	test()
}

package main

import (
	"embed"
	"recovery-unit-deploy/service/common"
	"recovery-unit-deploy/service/deploy"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	common.InitLogger("Deploy")

	defer common.CloseLogger()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "Deploy",
		Width:         1024,
		Height:        768,
		DisableResize: true, //禁用调整窗口尺寸
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Logger:           common.AppLogger,
		LogLevel:         logger.TRACE,
		Bind: []interface{}{
			app, &deploy.Deploy{}},
	})

	if err != nil {
		common.AppLogger.Error(err.Error())
	}
}

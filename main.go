package main

import (
	"embed"
	"flag"
	"fmt"
	"net"
	"os"
	"recovery-unit-deploy/service/common"
	"recovery-unit-deploy/service/deploy"

	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// lockPort 用于进程间通信的端口，应设为应用唯一的端口
const lockPort = 60629

// 检查应用是否已经运行
func isAlreadyRunning() bool {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", lockPort))
	if err != nil {
		// 端口已被占用，说明已有实例运行
		return true
	}
	defer conn.Close()
	return false
}

// 尝试获取实例锁
func acquireAppLock() (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", lockPort))
	if err != nil {
		return nil, fmt.Errorf("应用已在运行中: %v", err)
	}
	return listener, nil
}

func getWailsVersion() (string, error) {
	viper.SetConfigName("wails") // 文件名（不含扩展名）
	viper.AddConfigPath(".")     // 搜索当前目录
	viper.SetConfigType("json")  // 文件类型

	if err := viper.ReadInConfig(); err != nil {
		return "", err
	}

	info := viper.GetStringMapString("info")
	return info["productversion"], nil // 直接获取版本号
}

func main() {
	common.InitLogger("Deploy")

	defer common.CloseLogger()

	devMode := flag.Bool("dev", false, "Enable development mode")
	flag.Parse()
	common.CheckAdmin = !*devMode

	// 1. 防止双重启动
	if isAlreadyRunning() {
		common.AppLogger.Info("应用程序已在运行中")
		os.Exit(0)
	}

	// 获取应用锁
	listener, err := acquireAppLock()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln(err))
		os.Exit(1)
	}
	defer listener.Close()

	// Create an instance of the app structure
	app := NewApp()

	version, err := getWailsVersion() // 调用解析函数
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("读取版本失败: %v", err))
		version = "unknown"
	}
	common.AppLogger.Info(fmt.Sprintf("Current version is: %s", version))
	// Create application with options
	err = wails.Run(&options.App{
		Title:         "Deploy",
		Width:         1024,
		Height:        768,
		DisableResize: true, //禁用调整窗口尺寸
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:   &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:          app.startup,
		Logger:             common.AppLogger,
		LogLevel:           logger.TRACE,
		LogLevelProduction: logger.TRACE,
		Bind: []interface{}{
			app, &deploy.Deploy{}},
	})

	if err != nil {
		common.AppLogger.Error(err.Error())
	}
}

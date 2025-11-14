package main

import (
	"embed"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"recovery-unit-deploy/service/deploy"
	"runtime"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// lockPort 用于进程间通信的端口，应设为应用唯一的端口
const lockPort = 60629

var Version = "1.0.0" // 默认值
var BaseUrl = "https://deploy.ru.com/api-system"
var HasNewVersion = false

func hasNewVersion() bool {
	appInfo := api.GetAppVersionInfo("DEPLOY", common.GetOS())

	v1 := strings.Split(appInfo.Version, ".")
	v2 := strings.Split(Version, ".")

	maxLen := max(len(v1), len(v2))

	for i := 0; i < maxLen; i++ {
		num1, num2 := 0, 0
		if i < len(v1) {
			num1, _ = strconv.Atoi(v1[i]) // 忽略错误
		}
		if i < len(v2) {
			num2, _ = strconv.Atoi(v2[i])
		}

		if num1 > num2 {
			return true
		} else if num1 < num2 {
			return false
		}
	}
	return false
}

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

func isRunningAsRoot() bool {
	return os.Geteuid() == 0
}

func main() {
	if runtime.GOOS == "linux" {
		if !isRunningAsRoot() {
			// 获取当前可执行文件的绝对路径
			exePath, err := filepath.Abs(os.Args[0])
			params := strings.Join(os.Args[1:], " ")
			if err != nil {
				fmt.Printf("Error getting executable path: %v\n", err)
				return
			}

			// 构建pkexec命令，特别针对GUI应用传递环境变量
			cmd := exec.Command("pkexec",
				"env",
				"DISPLAY="+os.Getenv("DISPLAY"),
				"XAUTHORITY="+os.Getenv("XAUTHORITY"),
				"SUDO_USER="+os.Getenv("USER"),
				exePath,
				params)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to run with pkexec: %v\n", err)
			}
			// 父进程（非特权进程）退出
			os.Exit(0)
		}
	}

	common.InitLogger("Deploy")

	defer common.CloseLogger()

	common.AppLogger.Info(fmt.Sprintf("Current version is: %s", Version))

	common.AppLogger.Info(fmt.Sprintf("os.Args[0]: %s", os.Args[0]))

	args := os.Args[1:] // 忽略第一个参数（程序路径）
	isDebug := false
	isRestart := false
	for _, arg := range args {
		if arg == "-debug" {
			isDebug = true
		} else if arg == "-restart" {
			isRestart = true
			common.Restart = isRestart
			deploy.DeleteScheduledTask("Deploy")
		}
	}
	common.CheckAdmin = !isDebug

	startPage := ""
	if isRestart {
		startPage = "deploy"
	}

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
	app := NewApp(startPage)

	deploy := &deploy.Deploy{}
	deploy.InitClient(BaseUrl)

	if !isRestart && hasNewVersion() {
		deploy.HasNewVersion = true
	}

	if isRestart {
		deploy.LoadTemporaryInfo()
	}
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
			app, deploy},
	})

	if err != nil {
		common.AppLogger.Error(err.Error())
	}
}

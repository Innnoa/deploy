package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

var AppLogger logger.Logger
var LogFile *os.File

// 初始化日志系统
func InitLogger(appName string) {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("无法获取可执行路径:", err)
	}
	logDir := filepath.Dir(exePath)

	// 创建日志目录（如果不存在）
	logsDir := filepath.Join(logDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatal("创建日志目录失败:", err)
	}

	// 生成带毫秒的时间戳文件名
	timestamp := time.Now().Format("20060102150405")
	logFileName := filepath.Join(logsDir, fmt.Sprintf("%s_%s.log", appName, timestamp))

	// 创建日志文件
	LogFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("打开日志文件失败:", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, LogFile)
	log.SetOutput(multiWriter)

	// 设置自定义日志格式 (可选)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	AppLogger = logger.NewFileLogger(logFileName)

	// 写入初始化标记
	AppLogger.Info(fmt.Sprintf("===== 日志系统初始化完成 [%s] =====\n", logFileName))
}

func CloseLogger() {
	if LogFile != nil {
		AppLogger.Info("关闭日志文件")
		LogFile.Sync()
		LogFile.Close()
	}
}

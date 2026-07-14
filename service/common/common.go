package common

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

var CurrentComputerInfo ComputerInfo

var CurrentOA OAServer

var CurrentSeed SeedLabelInfo

var Server string
var Port string

var CheckAdmin bool

var DetailPCInfo DetailComputerInfo

var Restart bool

var CurrentUser string

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // 存在
	}
	if os.IsNotExist(err) {
		return false // 不存在
	}
	return true // 其他错误（如权限问题）
}

func IsUOS() bool {
	// 尝试读取 /etc/os-release 文件
	file, err := os.ReadFile("/etc/os-version")
	if err != nil {
		log.Println("Error reading /etc/os-version:", err)
		return false
	}

	// 检查文件内容
	content := string(file)
	return strings.Contains(content, "SystemName=UnionTech OS Desktop") || strings.Contains(content, "SystemName=UOS Desktop")
}

func GetOS() string {
	switch runtime.GOOS {
	case "windows":
		return "WIN"
	case "linux":
		if IsUOS() {
			return "UOS"
		}
		return "linux"
	}
	return "WIN"
}

func WriteFileWithSync(filePath string, data []byte) error {
	// 创建或打开文件。注意模式组合，特别是 O_SYNC 并非此处必需
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	// 使用defer确保文件句柄被关闭，尽管Sync后Close错误概率低，但仍需处理
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			AppLogger.Info(fmt.Sprintf("警告：关闭文件时出错: %v\n", closeErr))
		}
	}()

	// 写入数据到操作系统缓存
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// !!! 关键步骤：强制将数据同步到磁盘
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("同步文件到磁盘失败: %v", err)
	}

	AppLogger.Info(fmt.Sprintf("数据已确认写入磁盘: %s\n", filePath))
	return nil
}

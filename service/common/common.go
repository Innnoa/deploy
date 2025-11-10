package common

import (
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

func isUOS() bool {
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
	if runtime.GOOS == "windows" {
		return "WIN"
	} else if runtime.GOOS == "linux" {
		if isUOS() {
			return "UOS"
		}
		return "linux"
	}
	return "WIN"
}

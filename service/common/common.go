package common

import "os"

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

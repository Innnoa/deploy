//go:build !windows
// +build !windows

package deploy

import (
	"os"
	"recovery-unit-deploy/service/common"
)

func getUploadInfo() common.DetailComputerInfo {
	common.CurrentUser = ""
	return common.DetailPCInfo
}

func getLastKBCode() string {
	return os.Getenv("SEEDLABEL")
}

func checkSeedFile() bool {
	return true
}

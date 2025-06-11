//go:build !windows
// +build !windows

package deploy

import "recovery-unit-deploy/service/common"

func getUploadInfo() common.DetailComputerInfo {
	return common.DetailPCInfo
}

func getLastKBCode() string {
	return ""
}

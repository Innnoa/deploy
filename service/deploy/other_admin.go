//go:build !windows
// +build !windows

package deploy

import (
	"os"
	"recovery-unit-deploy/service/common"
)

func (p *Deploy) IsAdmin() bool {
	if !common.CheckAdmin {
		return true
	}

	// Unix-like 系统检查 root 权限
	return os.Geteuid() == 0
}

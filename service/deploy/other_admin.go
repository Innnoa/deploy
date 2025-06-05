//go:build !windows
// +build !windows

package deploy

import "os"

func (p *Deploy) IsAdmin() bool {
	// Unix-like 系统检查 root 权限
	return os.Geteuid() == 0
}

//go:build !windows
// +build !windows

package deploy

import "os/exec"

func setHideWindow(cmd *exec.Cmd) {
	// 空实现
}

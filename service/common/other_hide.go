//go:build !windows
// +build !windows

package common

import "os/exec"

func SetHideWindow(cmd *exec.Cmd) {
	// 空实现
}

//go:build !windows

package deploy

import "os/exec"

func reboot() {
	cmd := exec.Command("shutdown", "-r", "now")
	_ = cmd.Run()
}

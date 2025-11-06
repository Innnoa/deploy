//go:build !windows

package deploy

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/emersion/go-autostart"
)

func reboot() {
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART); err != nil {
		fmt.Println("reboot failed", err)
	} else {
		fmt.Println("reboot secuess")
	}
}

func createScheduledTask(taskName string, args []string) {
	exePath, _ := os.Executable()
	cmd := fmt.Sprintf(`"%s" %s`, exePath, strings.Join(args, " "))
	app := &autostart.App{
		Name:        "Deploy",
		DisplayName: "Deploy",
		Exec:        []string{cmd}, // 程序B的绝对路径
	}
	// 启用自启动
	app.Enable()
}

func DeleteScheduledTask(taskName string) {
	exePath, _ := os.Executable()
	app := &autostart.App{
		Name:        "Deploy",
		DisplayName: "Deploy",
		Exec:        []string{exePath},
	}
	// 禁用自启动
	app.Disable()
}

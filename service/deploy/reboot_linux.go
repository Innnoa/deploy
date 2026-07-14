//go:build linux
// +build linux

package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"recovery-unit-deploy/service/common"
	"strings"
)

func reboot() {
	output, err := runCommand("systemctl", "reboot")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("reboot failed:", err, output))
	} else {
		common.AppLogger.Info("reboot success")
	}
}

func createScheduledTask(taskName string, args []string) (string, error) {
	exePath, _ := os.Executable()
	cmd := fmt.Sprintf(`"%s" %s`, exePath, strings.Join(args, " "))
	autostartDir := filepath.Join("/etc", "xdg", "autostart")
	desktopFile := filepath.Join(autostartDir, taskName+".desktop")

	common.AppLogger.Info(fmt.Sprintf("desktopfile : %s", desktopFile))

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s
Icon=%s
Comment=Auto-start application
Terminal=false
Categories=Utility;
`, taskName, cmd, "")

	return desktopFile, common.WriteFileWithSync(desktopFile, []byte(content))
	// return desktopFile, os.WriteFile(desktopFile, []byte(content), 0644)
}

func DeleteScheduledTask(taskName string) error {
	// 假设条目创建在自动启动目录
	autostartDir := filepath.Join("/etc", "xdg", "autostart")
	desktopFile := filepath.Join(autostartDir, taskName+".desktop")

	// 2. 检查文件是否存在
	if _, err := os.Stat(desktopFile); os.IsNotExist(err) {
		return fmt.Errorf("启动条目文件不存在: %s", desktopFile)
	}

	// 3. 执行删除操作
	if err := os.Remove(desktopFile); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	fmt.Printf("成功删除自动启动条目: %s\n", desktopFile)
	return nil
}

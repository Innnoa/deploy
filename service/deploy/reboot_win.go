//go:build windows
// +build windows

package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"recovery-unit-deploy/service/common"
	"strings"
)

func createScheduledTask(taskName string, args []string) {
	exePath, _ := os.Executable()
	cmd := fmt.Sprintf(`"%s" %s`, exePath, strings.Join(args, " "))
	taskCmd := fmt.Sprintf(
		`schtasks /create /tn "%s" /tr "%s" /sc ONLOGON /RL HIGHEST /delay 0000:30 /f`,
		taskName, cmd,
	)

	execCmd := exec.Command("powershell", taskCmd)
	// 执行命令并捕获输出
	output, err := execCmd.CombinedOutput()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("创建scheduled task失败: %v\n输出: %s", err, common.DecodeByLocale(output)))
		return
	}
	common.AppLogger.Info(fmt.Sprintf("任务 '%s' 创建成功\n", taskName))
}

func DeleteScheduledTask(taskName string) {
	// 构造PowerShell命令：强制删除任务（/f跳过确认）
	cmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")

	// 执行命令并捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("删除失败: %v\n输出: %s", err, common.DecodeByLocale(output)))
		return
	}
	common.AppLogger.Info(fmt.Sprintf("任务 '%s' 已删除\n", taskName))
}

func reboot() {
	cmd := exec.Command("shutdown", "/r", "/t", "0")

	// 执行命令并等待其完成
	err := cmd.Run()

	// 错误处理
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("reboot failed: %v", err))
		return
	}

	common.AppLogger.Info("computer will reboot soon")
}

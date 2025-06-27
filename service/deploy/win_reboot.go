//go:build windows
// +build windows

package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"recovery-unit-deploy/service/common"
	"strings"

	"github.com/CodyGuo/win"
)

func getPrivileges() {
	var hToken win.HANDLE
	var tkp win.TOKEN_PRIVILEGES
	// 获取当前进程权限令牌
	win.OpenProcessToken(win.GetCurrentProcess(), win.TOKEN_ADJUST_PRIVILEGES|win.TOKEN_QUERY, &hToken)
	// 设置关机特权
	win.LookupPrivilegeValueA(nil, win.StringToBytePtr(win.SE_SHUTDOWN_NAME), &tkp.Privileges[0].Luid)
	tkp.PrivilegeCount = 1
	tkp.Privileges[0].Attributes = win.SE_PRIVILEGE_ENABLED
	win.AdjustTokenPrivileges(hToken, false, &tkp, 0, nil, nil)
}

func createScheduledTask(taskName string, args []string) {
	exePath, _ := os.Executable()
	cmd := fmt.Sprintf(`"%s" %s`, exePath, strings.Join(args, " "))
	taskCmd := fmt.Sprintf(
		`schtasks /create /tn "%s" /tr "%s" /sc ONSTART /RL HIGHEST /delay 0000:30 /f`,
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
	getPrivileges()
	win.ExitWindowsEx(win.EWX_REBOOT, 0) // 调用重启 API
}

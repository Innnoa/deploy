//go:build windows
// +build windows

package deploy

import (
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

func reboot() {
	getPrivileges()
	win.ExitWindowsEx(win.EWX_REBOOT, 0) // 调用重启 API
}

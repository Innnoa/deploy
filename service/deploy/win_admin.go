//go:build windows
// +build windows

package deploy

import "golang.org/x/sys/windows"

// 检查当前是否为管理员权限
func (p *Deploy) IsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// 使用推荐的API获取当前进程令牌
	token := windows.GetCurrentProcessToken()

	// 检查令牌是否属于管理员组
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}

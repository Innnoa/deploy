//go:build windows
// +build windows

package deploy

import (
	"fmt"
	"os/exec"
	"recovery-unit-deploy/service/common"
)

// 执行 w32tm 命令：配置域层级时间同步
func configureDomainTimeSync() error {
	cmd := exec.Command("w32tm", "/config", "/syncfromflags:domhier", "/update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("w32tm 命令执行失败: %v\n输出: %s", err, common.DecodeByLocale(output))
	}
	common.AppLogger.Info("w32tm config success")
	return nil
}

// 执行 net time 命令：强制同步目标服务器时间
func syncTimeWithTarget(target string) error {
	cmd := exec.Command("net", "time", "\\\\"+target, "/set", "/y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("net time 命令执行失败: %v\n输出: %s", err, string(output))
	}

	common.AppLogger.Info(fmt.Sprintf("The time has been synchronized to the target server %s\nOutput: %s", target, common.DecodeByLocale(output)))

	return nil
}

func syncTime() {
	// 执行 w32tm 配置
	if err := configureDomainTimeSync(); err != nil {
		common.AppLogger.Error(fmt.Sprintf("configureDomainTimeSync error: %v", err))
		return
	}

	// 设置目标服务器（示例：域控制器 DC01）
	targetServer := common.CurrentOA.ServerName // 实际使用时可从参数或配置读取
	if err := syncTimeWithTarget(targetServer); err != nil {
		common.AppLogger.Error(fmt.Sprintf("syncTimeWithTarget error: %v", err))
	}
}

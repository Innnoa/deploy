//go:build linux
// +build linux

package deploy

import (
	"fmt"
	"log"
	"os/exec"
	"recovery-unit-deploy/service/common"
	"runtime"
	"time"

	"github.com/beevik/ntp"
)

func syncTime() {
	// ntpTime, err := ntp.Time(common.CurrentOA.ServerName) // 可使用 time.apple.com, ntp.aliyun.com 等
	ntpTime, err := ntp.Time("ntp.ntsc.ac.cn")
	if err != nil {
		log.Fatalf("获取NTP时间失败: %v", err)
	}

	// 步骤2: 计算与本地时间的偏移量
	localTime := time.Now()
	offset := ntpTime.Sub(localTime)

	// 步骤3: 根据偏移量设置系统时间（需要root权限）
	// 注意：此操作会直接修改系统时间，请谨慎测试。
	if offset.Abs() > time.Second { // 例如，仅当偏差大于1秒时调整
		err = setSystemTime(ntpTime)
		if err != nil {
			common.AppLogger.Info(fmt.Sprintf("设置系统时间失败: %v", err))
		}
		common.AppLogger.Info("系统时间已同步！")
	} else {
		common.AppLogger.Info("时间偏差在允许范围内，无需调整。")
	}
}

// setSystemTime 根据操作系统调用命令设置时间
func setSystemTime(newTime time.Time) error {
	timeStr := newTime.Format("2006-01-02 15:04:05") // Go语言标准时间格式

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin": // Linux 或 macOS
		cmd = exec.Command("sudo", "date", "-s", timeStr)
	case "windows":
		// Windows 下分别设置日期和时间
		dateStr := newTime.Format("2006-01-02")
		timeStrWin := newTime.Format("15:04:05")
		cmd = exec.Command("cmd", "/C", "date", dateStr)
		err := cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command("cmd", "/C", "time", timeStrWin)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	return cmd.Run()
}

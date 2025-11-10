//go:build linux
// +build linux

package deploy

import (
	"fmt"
	"recovery-unit-deploy/service/common"
	"time"

	"github.com/beevik/ntp"
)

func syncTime() {
	// ntpTime, err := ntp.Time(common.CurrentOA.ServerName) // 可使用 time.apple.com, ntp.aliyun.com 等
	ntpTime, err := ntp.Time("ntp.ntsc.ac.cn")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("获取NTP时间失败: %v", err))
		return
	}

	// 步骤2: 计算与本地时间的偏移量
	localTime := time.Now()
	offset := ntpTime.Sub(localTime)

	// 步骤3: 根据偏移量设置系统时间（需要root权限）
	// 注意：此操作会直接修改系统时间，请谨慎测试。
	if offset.Abs() > time.Second { // 例如，仅当偏差大于1秒时调整
		err = setSystemTime(ntpTime)
		if err != nil {
			common.AppLogger.Error(fmt.Sprintf("设置系统时间失败: %v", err))
		}
		common.AppLogger.Info("系统时间已同步！")
	} else {
		common.AppLogger.Info("时间偏差在允许范围内，无需调整。")
	}
}

// setSystemTime 根据操作系统调用命令设置时间
func setSystemTime(newTime time.Time) error {
	timeStr := newTime.Format("2006-01-02 15:04:05") // Go语言标准时间格式

	_, err := runCommand("date", "-s", timeStr)
	return err
}

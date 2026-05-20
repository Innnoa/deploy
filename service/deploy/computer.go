package deploy

import (
	"net"
	"recovery-unit-deploy/service/common"
)

type DiskInfo struct {
	DeviceID  string // 盘符（如 "C"）
	FreeSpace string // 剩余空间（GB）
	Size      string // 总容量（GB）
}

// CPU 信息结构体
type CPUInfo struct {
	Name          string
	MaxClockSpeed float32
}

// 内存信息结构体
type MemoryInfo struct {
	TotalPhysical int64 // GB
}

// boot信息结构体
type SystemInfo struct {
	BootMode string
	Model    string // PC型号（如 "HP ProDesk 600 G3"）
}

func (c *Deploy) CheckSeedLabel() bool {
	return checkSeedFile()
}

func getIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// 排除未启用或回环接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil { // 优先返回 IPv4
					return ipNet.IP.String()
				}
			}
		}
	}
	return ""
}

func (c *Deploy) GetComputerInfo() common.ComputerInfo {
	if common.Restart {
		// getUploadInfo()
	} else {
		var info common.ComputerInfo

		name := getComputerName()

		ip := getIP()

		info.Name = name
		info.IP = ip

		common.CurrentComputerInfo = info

	}

	return common.CurrentComputerInfo
}

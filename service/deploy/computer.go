package deploy

import (
	"fmt"
	"net"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strings"
)

type DiskInfo struct {
	DeviceID  string // 盘符（如 "C"）
	FreeSpace string // 剩余空间（GB）
	Size      string // 总容量（GB）
}

func (c *Deploy) GetSeedLabel() common.SeedLabelInfo {
	kbcode := getLastKBCode()
	// kbcode = "KB5039334"
	seed := api.GetSeedLabel(kbcode)
	return seed
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

func setDiskInfo(disks []DiskInfo) {
	common.DetailPCInfo.NumOfDrive = fmt.Sprintf("%d", len(disks))
	if len(disks) > 0 {
		common.DetailPCInfo.SizeOfDrive1 = disks[0].Size

		if len(disks) > 1 {
			common.DetailPCInfo.SizeOfDrive2 = disks[1].Size
		}

		for _, d := range disks {
			if strings.EqualFold(d.DeviceID, "C") {
				common.DetailPCInfo.FreeSpaceC = d.FreeSpace
			} else if strings.EqualFold(d.DeviceID, "D") {
				common.DetailPCInfo.FreeSpaceD = d.FreeSpace
			}
		}

		common.DetailPCInfo.LastDrive = disks[len(disks)-1].DeviceID
		common.DetailPCInfo.SystemDrive = getOpSystemInfo()
	}
}

package deploy

import (
	"fmt"
	"log"
	"net"
	"os"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

func (c *Deploy) GetSeedLabel() common.SeedLabelInfo {
	kbcode := getLastKBCode()
	// kbcode = "KB5039334"
	seed := api.GetSeedLabel(kbcode)
	return seed
}

func (c *Deploy) CheckSeedLabel() bool {
	filename := fmt.Sprintf("C:\\%s.seedlabel.txt", common.CurrentSeed.SeedLabel)
	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	// 修改时间（跨平台通用）
	modTime := fileInfo.ModTime()

	// 创建时间（按系统处理）
	createTime := getCreationTime(fileInfo)

	strModTime := modTime.Format("2006-01-02 15:04:05")
	strCreateTime := createTime.Format("2006-01-02 15:04:05")
	fmt.Printf("修改时间: %s\n", strModTime)
	fmt.Printf("创建时间: %s\n", strCreateTime)

	return api.CheckSeedLabel(common.CurrentSeed.SeedLabel, strCreateTime, strModTime)
}

func getComputerName() string {
	name := os.Getenv("COMPUTERNAME")
	return name
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

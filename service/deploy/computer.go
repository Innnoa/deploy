package deploy

import (
	"net"
	"os"
	"recovery-unit-deploy/service/common"
)

func getComputerName() string {
	name := os.Getenv("COMPUTERNAME")
	return name
}

func getSeedLabel() string {
	seed := getSeed()
	return seed
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
	var info common.ComputerInfo

	name := getComputerName()
	seed := getSeedLabel()
	ip := getIP()

	info.Name = name
	info.Seed = seed
	info.IP = ip

	common.CurrentComputerInfo = info

	return common.CurrentComputerInfo
}

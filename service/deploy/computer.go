package deploy

import (
	"log"
	"net"
	"os"
	"recovery-unit-deploy/service/common"
	"runtime"

	"golang.org/x/sys/windows/registry"
)

func getComputerName() string {
	name := os.Getenv("COMPUTERNAME")
	return name
}

func getSeedLabel() string {

	var seed string
	if runtime.GOOS == "windows" {
		seed = getRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Longlabel")
	} else {
		seed = os.Getenv("SEEDLONGLABEL")
	}
	return seed
}

func getRegValue(key registry.Key, path string, name string) string {
	key, err := registry.OpenKey(key, path, registry.QUERY_VALUE)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer key.Close()

	value, _, err := key.GetStringValue(name)
	if err != nil {
		log.Println(err)
		return ""
	}
	return value
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

	return info
}

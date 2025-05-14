package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

func getComputerName() string {
	name := "C81369"
	return name
}

func getSeedLabel() string {
	seed := "CW10V24B"
	return seed
}

func getIP() string {
	ip := "192.168.14.110"
	return ip
}

func (c *Deploy) GetComputerInfo() common.ComputerInfo {
	var info common.ComputerInfo

	name := getComputerName()
	seed := getSeedLabel()
	ip := getIP()
	oa := api.GetOAServer(ip)

	info.Name = name
	info.Seed = seed
	info.OA = oa
	info.IP = ip

	return info
}

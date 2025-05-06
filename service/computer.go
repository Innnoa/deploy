package service

type ComputerInfo struct {
	Name     string `json:"name"`
	Seed     string `json:"seed"`
	IP       string `json:"ip"`
	OAServer string `json:"oaserver"`
}

func getComputerName() string {
	name := "C81369"
	return name
}

func getSeedLabel() string {
	seed := "CW10V24B"
	return seed
}

func getOAServer() string {
	server := "HPFS3OABAH3"
	return server
}

func getIP() string {
	ip := "192.168.14.110"
	return ip
}

func (c *ComputerInfo) GetComputerInfo() ComputerInfo {
	var info ComputerInfo

	name := getComputerName()
	seed := getSeedLabel()
	server := getOAServer()
	ip := getIP()

	info.Name = name
	info.Seed = seed
	info.OAServer = server
	info.IP = ip

	return info
}

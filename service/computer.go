package service

type ComputerInfo struct {
	Name string   `json:"name"`
	Seed string   `json:"seed"`
	IP   string   `json:"ip"`
	OA   OAServer `json:"oa"`
}

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

func (c *ComputerInfo) GetComputerInfo() ComputerInfo {
	var info ComputerInfo

	name := getComputerName()
	seed := getSeedLabel()
	ip := getIP()
	server := getOAServer(ip)

	info.Name = name
	info.Seed = seed
	info.OA = server
	info.IP = ip

	return info
}

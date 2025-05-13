package service

import "log"

type OAServer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	IP   string `json:"ip"`
}

var currentOA OAServer

func getOAServer(ip string) OAServer {
	log.Printf("get OAServer from ip: %s", ip)

	currentOA.Name = "HPFS3OABAH3"
	currentOA.IP = "192.168.14.3"
	return currentOA
}

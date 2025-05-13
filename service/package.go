package service

import "log"

type PackageInfo struct {
	ID        string `json:"id"`
	AppName   string `json:"app_name"`
	AppType   string `json:"app_type"`
	Path      string `json:"path"`
	WinFile   string `json:"win_file"`
	UOSFile   string `json:"uos_file"`
	KylinFile string `json:"kylin_file"`
	Status    string `json:"status"`
}

var localPrinterDriverList []PackageInfo
var allPackages []PackageInfo
var selectedLocalDriver PackageInfo
var selectedNetworkPrinters []Printer

func (p *PackageInfo) GetAllPackages(pol string) {
	//get all packages(application and driver) through pol number
	//call api
	var package0 PackageInfo
	package0.AppName = "FireFox08"
	package0.AppType = "ForceApp"
	package0.ID = "0B9111FE-F2F3-05AF-C7CE-35F536E15FDD"
	package0.Status = "Waiting"

	var package1 PackageInfo
	package1.AppName = "HP - LaserJet 4 Plus"
	package1.AppType = "Printer"
	package1.ID = "A2A50752-E653-5772-CBB4-5D89E4CB8E76"
	package1.Status = "Waiting"

	allPackages = append(allPackages, package0)
	allPackages = append(allPackages, package1)

	for _, value := range allPackages {
		if value.AppType == "Printer" {
			localPrinterDriverList = append(localPrinterDriverList, value)
		}
	}
}

func (p *PackageInfo) GetInstallPackages() []PackageInfo {
	var installPackages []PackageInfo
	for _, value := range allPackages {
		if value.AppType != "Printer" {
			installPackages = append(installPackages, value)
		}
	}

	installPackages = append(installPackages, selectedLocalDriver)
	//get network printer drivers through api
	log.Println("get network printer from list: ", selectedNetworkPrinters)

	var package0 PackageInfo
	package0.AppName = "HP - LaserJet 4 Plus"
	package0.AppType = "Printer"
	package0.ID = "A2A50752-E653-5772-CBB4-5D89E4CB8E76"
	package0.Status = "Waiting"
	installPackages = append(installPackages, package0)

	return installPackages
}

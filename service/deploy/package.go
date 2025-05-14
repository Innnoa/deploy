package deploy

import (
	"log"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

var localPrinterDriverList []common.PackageInfo
var allPackages []common.PackageInfo
var selectedLocalDriver common.PackageInfo
var selectedNetworkPrinters []common.Printer

func (p *Deploy) GetAllPackages(pol string) {
	//get all packages(application and driver) through pol number
	allPackages = api.GetAllPackages(pol)

	for _, value := range allPackages {
		if value.AppType == "Printer" {
			localPrinterDriverList = append(localPrinterDriverList, value)
		}
	}
}

func (p *Deploy) GetInstallPackages() []common.PackageInfo {
	var installPackages []common.PackageInfo
	for _, value := range allPackages {
		if value.AppType != "Printer" {
			installPackages = append(installPackages, value)
		}
	}

	installPackages = append(installPackages, selectedLocalDriver)
	//get network printer drivers through api
	log.Println("get network printer from list: ", selectedNetworkPrinters)

	networkPrinterDriver := api.GetPrinterDrivers(selectedNetworkPrinters)
	installPackages = append(installPackages, networkPrinterDriver...)

	return installPackages
}

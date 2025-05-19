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
var installedPackages []common.PackageInfo

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
	installedPackages = installedPackages[:0]
	for _, value := range allPackages {
		if value.AppType != "Printer" {
			installedPackages = append(installedPackages, value)
		}
	}

	installedPackages = append(installedPackages, selectedLocalDriver)
	//get network printer drivers through api
	log.Println("get network printer from list: ", selectedNetworkPrinters)

	networkPrinterDriver := api.GetNetworkPrinterDrivers(selectedNetworkPrinters)
	installedPackages = append(installedPackages, networkPrinterDriver...)

	return installedPackages
}

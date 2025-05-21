package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

var localPrinterDriverList []common.PackageInfo
var selectedLocalDriver common.PackageInfo
var selectedNetworkPrinters []common.Printer
var installedPackages []common.PackageInfo

func (p *Deploy) GetInstallPackages() []common.PackageInfo {
	installedPackages = installedPackages[:0]

	//get all packages(application only) through pol number and seed
	allPackages := api.GetAllPackages(common.CurrentComputerInfo.Name, common.CurrentComputerInfo.Seed)
	installedPackages = append(installedPackages, allPackages...)

	if selectedLocalDriver.ID != "" {
		installedPackages = append(installedPackages, selectedLocalDriver)
	}

	networkPrinterDriver := api.GetNetworkPrinterDrivers(selectedNetworkPrinters)
	installedPackages = append(installedPackages, networkPrinterDriver...)

	return installedPackages
}

package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

var localPrinterDriverList []common.PackageInfo
var selectedLocalDriver *common.PackageInfo
var networkPrinterDriverList []common.PrinterWithPackage
var selectedNetworkPrinters []common.Printer
var installedPackages []common.PackageInfo

func (p *Deploy) GetInstallPackages() []common.PackageInfo {
	installedPackages = installedPackages[:0]

	for _, p := range selectedNetworkPrinters {
		for _, value := range networkPrinterDriverList {
			if p.ID == value.ID {
				pi := common.PackageInfo{ID: value.AppId, AppName: value.AppName, AppType: "NETWORK", Path: value.Path, WinFile: value.WinFile, UOSFile: value.UOSFile, KylinFile: value.KylinFile, PolNo: p.PolNo, IP: p.IP}
				installedPackages = append(installedPackages, pi)
				break
			}
		}
	}

	if selectedLocalDriver != nil && selectedLocalDriver.ID != "" {
		installedPackages = append(installedPackages, *selectedLocalDriver)
	}

	//get all packages(application only) through pol number and seed
	allPackages := api.GetAllPackages(common.CurrentComputerInfo.Name, common.CurrentComputerInfo.Seed)
	installedPackages = append(installedPackages, allPackages...)

	return installedPackages
}

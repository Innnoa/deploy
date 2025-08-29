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
	if common.Restart {
		//do nothing
	} else {
		installedPackages = installedPackages[:0]

		for _, p := range selectedNetworkPrinters {
			for _, value := range networkPrinterDriverList {
				if p.ID == value.ID {
					pi := common.PackageInfo{ID: value.AppId, AppName: value.AppName, AppType: value.AppType, Path: value.Path, WinFile: value.WinFile, UOSFile: value.UOSFile, KylinFile: value.KylinFile, PolNo: p.PolNo, IP: p.IP}
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

		tasks := api.GetSeedTasks(common.CurrentSeed.SeedLabel)
		installedPackages = append(installedPackages, tasks...)

		sapps := api.GetCodesByGroup("SPECIAL_APP")
		for _, app := range sapps {
			if app.Code == "RU_SERVICE" {
				appid := app.Name
				ru := common.PackageInfo{AppName: "RU Service", ID: appid}
				installedPackages = append(installedPackages, ru)
				break
			}
		}
	}

	// uiShow := filter(installedPackages, func(p common.PackageInfo) bool { return strings.TrimSpace(p.AppName) != "Restart Machine" })

	return installedPackages
}

func getInstallPackages() []common.PackageInfo {
	// uiShow := filter(installedPackages, func(p common.PackageInfo) bool { return strings.TrimSpace(p.AppName) != "Restart Machine" })

	return installedPackages
}

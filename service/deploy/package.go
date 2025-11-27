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
					pi := common.PackageInfo{
						ID:            value.AppId,
						AppName:       value.AppName,
						AppType:       value.AppType,
						Path:          value.Path,
						WinFile:       value.WinFile,
						UOSFile:       value.UOSFile,
						KylinFile:     value.KylinFile,
						PolNo:         p.PolNo,
						IP:            p.IP,
						PrinterName:   value.PrinterName,
						PrinterDriver: value.PrinterDriver,
						Ppd:           value.Ppd}
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
		installedPackages = mergeAndDeduplicateByIssuetype(installedPackages, tasks)

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

func mergeAndDeduplicateByIssuetype(arr1, arr2 []common.PackageInfo) []common.PackageInfo {
	// 创建一个map，键为Issuetype，值为TotalIssue结构体
	mergedMap := make(map[string]common.PackageInfo)

	// 先遍历第一个数组，将元素存入map
	for _, item := range arr1 {
		mergedMap[item.AppName] = item
	}

	// 再遍历第二个数组
	for _, item := range arr2 {
		// 如果map中已存在相同的Issuetype，则用当前项覆盖（保留后出现的项）
		// 如果需要对数值型字段进行累加，可以在此处修改逻辑
		mergedMap[item.AppName] = item
	}

	// 将map中的值转换为切片
	result := make([]common.PackageInfo, 0, len(mergedMap))
	for _, value := range mergedMap {
		result = append(result, value)
	}

	return result
}

func getInstallPackages() []common.PackageInfo {
	// uiShow := filter(installedPackages, func(p common.PackageInfo) bool { return strings.TrimSpace(p.AppName) != "Restart Machine" })

	return installedPackages
}

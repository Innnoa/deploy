package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strings"
)

func (p *Deploy) GetNetworkPinterList(keyword string) []common.Printer {
	printers := api.GetNetworkPinterList(keyword)
	return printers
}

func (p *Deploy) GetPrinterModels() []string {
	models := api.GetPrinterModels()
	return models
}

func (p *Deploy) GetPrinterDrivers(model string) []common.PackageInfo {
	var drivers []common.PackageInfo

	for _, value := range localPrinterDriverList {
		if strings.HasPrefix(value.AppName, model) {
			drivers = append(drivers, value)
		}
	}
	return drivers
}

func (p *Deploy) SetSelectedPrinters(driverId string, printers []common.Printer) bool {
	for _, value := range allPackages {
		if value.ID == driverId {
			selectedLocalDriver = value
			break
		}
	}

	selectedNetworkPrinters = printers

	return true
}

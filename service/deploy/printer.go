package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

func (p *Deploy) GetNetworkPinterList(keyword string) []common.Printer {
	result := api.GetNetworkPinterList(keyword)
	var printers []common.Printer
	networkPrinterDriverList = networkPrinterDriverList[:0]

	for _, value := range result {
		p := common.Printer{ID: value.ID, PolNo: value.PolNo, IP: value.IP, AppId: value.AppId}
		printers = append(printers, p)

		networkPrinterDriverList = append(networkPrinterDriverList, value)
	}

	return printers
}

func (p *Deploy) GetPrinterModels() []common.PrinterModel {
	models := api.GetPrinterModels()
	return models
}

func (p *Deploy) GetSelectedLocalPrinterDrivers(id string) []common.PackageInfo {
	drivers := api.GetSelectedLocalPrinterDrivers(id)
	localPrinterDriverList = localPrinterDriverList[:0]
	localPrinterDriverList = drivers

	return drivers
}

func (p *Deploy) SetSelectedPrinters(driverId string, printers []common.Printer) bool {
	selectedLocalDriver = nil
	for _, value := range localPrinterDriverList {
		if value.ID == driverId {
			selectedLocalDriver = &value
			break
		}
	}

	selectedNetworkPrinters = printers

	return true
}

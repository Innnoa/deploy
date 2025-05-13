package service

import "strings"

type Printer struct {
	ID    string `json:"id"`
	PolNo string `json:"pol"`
	IP    string `json:"ip"`
}

func (p *Printer) GetNextworkPinterList(keyword string) []Printer {

	var printer0 Printer
	printer0.PolNo = "P87520"
	printer0.IP = "192.168.11.200"

	printers := []Printer{printer0}
	return printers
}

func (p *Printer) GetPrinterModels() []string {
	models := []string{"HP", "Epson", "Brother"}
	return models
}

func (p *Printer) GetPrinterDrivers(model string) []PackageInfo {
	var drivers []PackageInfo

	for _, value := range localPrinterDriverList {
		if strings.HasPrefix(value.AppName, model) {
			drivers = append(drivers, value)
		}
	}
	return drivers
}

func (p *Printer) SetSelectedPrinters(driverId string, printers []Printer) bool {
	for _, value := range allPackages {
		if value.ID == driverId {
			selectedLocalDriver = value
			break
		}
	}

	selectedNetworkPrinters = printers

	return true
}

package service

type Printer struct {
	PolNo string `json:"pol"`
	IP    string `json:"ip"`
}

func (c *Printer) GetNextworkPinterList(pol string, ip string) []Printer {

	var printer0 Printer
	printer0.PolNo = "P87520"
	printer0.IP = "192.168.11.200"

	printers := []Printer{printer0}
	return printers
}

func (c *Printer) GetPrinterModels() []string {
	models := []string{"HP", "Epson", "Brother"}
	return models
}

func (c *Printer) GetPrinterDrivers(model string) []string {
	drivers := []string{"HP - LaserJet 4 Plus", "HP - LaserJet 5"}
	return drivers
}

func (c *Printer) GetSelectedPrinters(driver string, printers []Printer) bool {
	//get driver installpath and file
	return true
}

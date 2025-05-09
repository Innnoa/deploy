package service

func (c *PackageInfo) GetPrinterModels() []string {
	models := []string{"HP", "Epson", "Brother"}
	return models
}

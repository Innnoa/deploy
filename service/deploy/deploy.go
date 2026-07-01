package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

var osDisplayNames = map[string]string{
	"WIN":   "Windows",
	"linux": "Linux",
	"UOS":   "UOS",
	"Kylin": "Kylin",
}

type Deploy struct {
	HasNewVersion bool
}

func (p *Deploy) InitClient(baseUrl string) {
	api.Client = api.NewAPIClient(baseUrl)
}

func (p *Deploy) CheckNewVersion() bool {
	return p.HasNewVersion
}

func (p *Deploy) GetOS() string {
	short := common.GetOS()
	if name, ok := osDisplayNames[short]; ok {
		return name
	}
	return "Windows"
}

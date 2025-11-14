package deploy

import (
	"recovery-unit-deploy/service/api"
)

type Deploy struct {
	HasNewVersion bool
	OS            string
}

func (p *Deploy) InitClient(baseUrl string) {
	api.Client = api.NewAPIClient(baseUrl)
}

func (p *Deploy) CheckNewVersion() bool {
	return p.HasNewVersion
}

func (p *Deploy) GetOS() string {
	return p.OS
}

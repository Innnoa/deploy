package deploy

import (
	"recovery-unit-deploy/service/api"
)

type Deploy struct {
	HasNewVersion bool
}

func (p *Deploy) InitClient() {
	api.Client = api.NewAPIClient("http://" + "deploy.ru.com:9900" + "/api-system")
}

func (p *Deploy) CheckNewVersion() bool {
	return p.HasNewVersion
}

package deploy

import (
	"recovery-unit-deploy/service/api"
)

type Deploy struct {
}

func (p *Deploy) InitClient() {
	api.Client = api.NewAPIClient("http://" + "deploy.ru.com:9900" + "/api-system")
}

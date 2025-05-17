package deploy

import (
	"recovery-unit-deploy/service/api"
)

type Deploy struct {
}

func (p *Deploy) InitClient(server string, port string) {
	api.Client = api.NewAPIClient("http://" + server + ":" + port + "/api-system")
}

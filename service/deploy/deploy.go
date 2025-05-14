package deploy

import (
	"net/http"
	"recovery-unit-deploy/service/api"
)

type Deploy struct {
}

func (p *Deploy) InitClient(server string, port string) {
	api.Client = api.HTTPClient{
		Client:  http.DefaultClient,
		BaseURL: "https://" + server + ":" + port,
	}
}

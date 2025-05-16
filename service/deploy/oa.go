package deploy

import (
	"recovery-unit-deploy/service/api"
)

func (c *Deploy) GetOAServer(ip string) string {
	oa := api.GetOAServer(ip)

	return oa
}

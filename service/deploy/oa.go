package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
)

func (c *Deploy) GetOAServer(computer common.ComputerInfo) common.ComputerInfo {
	oa := api.GetOAServer(computer.IP)

	computer.OA = oa

	return computer
}

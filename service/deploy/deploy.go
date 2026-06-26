package deploy

import (
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"runtime"
)

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
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "linux":
		if common.IsUOS() {
			return "UOS"
		}
		if common.IsKylin() {
			return "Kylin"
		}
		return "Linux"
	}
	return "Windows"
}

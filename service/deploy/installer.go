package deploy

import (
	"os"
	"path"
	"recovery-unit-deploy/service/common"
)

var installStatus []common.InstallStatus

func (p *Deploy) DoInstall() {
	for _, value := range allPackages {
		var status common.InstallStatus
		status.ID = value.ID
		status.Status = "Waiting"

		installStatus = append(installStatus, status)

		os.Setenv("SRC", path.Join(common.CurrentOA, value.Path))

	}
}

func (p *Deploy) GetInstallStatus() []common.InstallStatus {
	return installStatus
}

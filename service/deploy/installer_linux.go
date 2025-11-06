//go:build linux
// +build linux

package deploy

import (
	"fmt"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strings"
)

var tempFilePath = "/tmp/tool"

func installRU() error {
	return nil
}

func installRUService(src string) error {
	return nil
}

func createRUService(serviceName, binPath string) error {
	return nil
}

func createService(name, binPath string) error {
	return nil
}

func safeDeleteService(name string) error {
	return nil
}

func stopServiceIfRunning() error {
	return nil
}

func (p *Deploy) DeleteTempFiles() error {
	err := deleteTempFiles(tempFilePath)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("delete文件错误:", err))
	}
	return err
}

func (p *Deploy) DoInstall() {
	getUploadInfo()
	api.UploadPCInfo(common.DetailPCInfo)

	maintaskid, err := uploadInstallInfo()
	mainTask = maintaskid
	if err != nil {
		setAllStatusFail("upload task infomation failed")
		return
	}

	installPackages()
}

func (p *Deploy) InstallAfterReboot() {
	installPackages()
}

func rebootForInstall() {
	saveTemporaryInfo()
	createScheduledTask("Deploy", []string{"-restart"})
	reboot()
}

func installPackages() {
	for i := range installedPackages {
		if cancelling {
			break
		}
		if installedPackages[i].Status == common.Completed.String() ||
			installedPackages[i].Status == common.Failed.String() {
			continue
		}

		var app common.AppStatus
		app.ID = installedPackages[i].ID
		app.MainTask = mainTask
		api.StartInstall(app)
		installedPackages[i].Status = common.Running.String()

		if strings.TrimSpace(installedPackages[i].AppName) == "Restart Machine" {
			installedPackages[i].Status = common.Completed.String()
			api.InstallationSuccess(app)
			rebootForInstall()

			return
		} else if strings.TrimSpace(installedPackages[i].AppName) == "Time Sync" {
			syncTime()
			installedPackages[i].Status = common.Completed.String()
			api.InstallationSuccess(app)

			continue
		} else if strings.TrimSpace(installedPackages[i].AppName) == "RU Service" {
			err0 := installRU()
			if err0 != nil {
				common.AppLogger.Error(fmt.Sprintln("install ruservice failed:", err0))
				setPakcageStatusFailed(&installedPackages[i], err0, app)
			} else {
				installedPackages[i].Status = common.Completed.String()
				api.InstallationSuccess(app)
			}

			continue
		}

		_, err := runCommand("apt", "install", installedPackages[i].InstallPackageName, "-y")
		if err != nil {
			common.AppLogger.Error(fmt.Sprintln("failed to install the application:", err))
			setPakcageStatusFailed(&installedPackages[i], err, app)
			continue
		}

		installedPackages[i].Status = common.Completed.String()
		api.InstallationSuccess(app)
	}
}

//go:build linux
// +build linux

package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strings"
)

var tempFilePath = "/var/deploy"

func installRU() error {
	ru := api.GetAppVersionInfo("RU", common.GetOS())
	src := ru.InstallPath

	target := filepath.Join(tempFilePath, filepath.Base(src))
	var err error
	// 确保本地目录存在
	localDir := filepath.Dir(target)
	if err = os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %v", err)
	}

	switch common.CurrentOA.StorageType {
	case "SMB":
		downloadRUSmb(src, target)
	case "NGINX":
		downloadRUNginx(src, target)
	default:
		downloadRUSmb(src, target)
	}

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("download ruservice failed: %v", err))
		return err
	}

	err = installRUService(target)
	return err
}

func installRUService(target string) error {
	if err := checkDpkg(); err != nil {
		return err
	}
	if err := checkFile(target); err != nil {
		return err
	}

	if output, err := installDeb(target); err != nil {
		return fmt.Errorf("install ruservice failed: %v, output: %s", err, &output)
	}

	return nil
}

func checkDpkg() error {
	// 检查 dpkg 命令是否可用
	_, err := exec.LookPath("dpkg")
	if err != nil {
		return fmt.Errorf("'dpkg' command not found, ensure it is installed and in your PATH")
	}
	return nil
}

func checkFile(debPath string) error {
	// 检查 .deb 文件是否存在
	if _, err := os.Stat(debPath); os.IsNotExist(err) {
		return fmt.Errorf("deb file not found: %s", debPath)
	}
	return nil
}

func installDeb(debPath string) (string, error) {
	// 使用 dpkg 安装本地的 .deb 文件
	output, err := runCommand("sudo", "dpkg", "-i", debPath)

	return output, err
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

	err = scriptDownload()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("script download failed:", err))
		setAllStatusFail("script download failed")
		return
	}

	output, err := runCommand("sudo", "apt", "update")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("apt update failed:", err, output))
	}

	installPackages()
}

func (p *Deploy) InstallAfterReboot() {
	installPackages()
}

func rebootForInstall() {
	saveTemporaryInfo()
	_, err := createScheduledTask("Deploy", []string{"-restart"})
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("create autostart failed: %v", err))
		return
	}

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

		_, err := runCommand("apt", "install", "--reinstall", installedPackages[i].InstallPackageName, "-y")
		if err != nil {
			common.AppLogger.Error(fmt.Sprintln("failed to install the application:", err))
			setPakcageStatusFailed(&installedPackages[i], err, app)
			continue
		}

		if installedPackages[i].AppType == "Printer" {
			// 下载 sh 文件并执行
			if installedPackages[i].IP == "" {
				//本地打印机
				scriptPath := path.Join(tempFilePath, "uosPrinterLocal.sh")
				err = os.Chmod(scriptPath, 0755)
				if err != nil {
					setPakcageStatusFailed(&installedPackages[i], err, app)
					continue
				}
				common.AppLogger.Info("执行本地打印机脚本")
				output, err := common.RunScriptWithArgs(scriptPath, installedPackages[i].Ppd, installedPackages[i].PrinterName)
				if err != nil {
					common.AppLogger.Error(fmt.Sprintf("执行本地打印机脚本失败：%s, error: %v", output, err))
					setPakcageStatusFailed(&installedPackages[i], err, app)
					continue
				}
			} else {
				//网络打印机
				scriptPath := path.Join(tempFilePath, "uosPrinterNet.sh")
				err = os.Chmod(scriptPath, 0755)
				if err != nil {
					setPakcageStatusFailed(&installedPackages[i], err, app)
					continue
				}
				common.AppLogger.Info("执行网络打印机脚本")
				output, err := common.RunScriptWithArgs(scriptPath, installedPackages[i].Ppd, installedPackages[i].PrinterName, installedPackages[i].IP, installedPackages[i].PolNo)
				if err != nil {
					common.AppLogger.Error(fmt.Sprintf("执行网络打印机脚本失败：%s, %v", output, err))
					setPakcageStatusFailed(&installedPackages[i], err, app)
					continue
				}
			}
		}
		installedPackages[i].Status = common.Completed.String()
		api.InstallationSuccess(app)
	}
}

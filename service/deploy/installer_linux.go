//go:build linux
// +build linux

package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"recovery-unit-deploy/service/smb"
	"strconv"
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

func downloadRUSmb(src, target string) error {
	server := common.CurrentOA.ServerName
	username := common.CurrentOA.UserName //get from server
	encryptedPassword := common.CurrentOA.Password
	password := common.Decode(encryptedPassword) //get from server
	port := 0
	if common.CurrentOA.Port != "" {
		// 字符串转int
		_port, err := strconv.Atoi(common.CurrentOA.Port)
		if err != nil {
			common.AppLogger.Error(fmt.Sprintf("端口号转换失败: %v", err))
		}
		port = _port
	}
	client := smb.NewClient(server, port, common.CurrentOA.RootPath, username, password)
	err := client.Connect()
	if err != nil {
		common.AppLogger.Error(fmt.Sprint("连接smb服务器失败: %v", err))
		return err
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			common.AppLogger.Error(fmt.Sprintf("断开 SMB 连接失败: %v", err))
		}
	}()

	exists, _, err := client.FileExists(src)
	if !exists {
		errmsg := fmt.Sprintf("%s source is not exist: %v", src, err)
		common.AppLogger.Error(errmsg)
		return fmt.Errorf(errmsg)
	}

	err = client.DownloadFile(src, target)
	if err != nil {
		errmsg := fmt.Sprintf("下载 %s 失败: %v", src, err)
		return fmt.Errorf(errmsg)
	}

	return nil
}

func downloadRUNginx(src, target string) error {
	downloadUrl := fmt.Sprintf("http://%s:%s%s/%s", common.CurrentOA.ServerName, common.CurrentOA.Port, common.CurrentOA.BaseUrl, src)
	downloadUrl = strings.ReplaceAll(downloadUrl, "\\", "/")
	err := downloadFileWithBasicAuth(downloadUrl, common.CurrentOA.UserName, common.Decode(common.CurrentOA.Password), target)

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

//go:build windows
// +build windows

package deploy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var tempFilePath = "C:\\Temp\\tool"

func (p *Deploy) DeleteTempFiles() error {
	err := deleteTempFiles(tempFilePath)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("delete文件错误:", err))
	}
	exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
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

	target := tempFilePath
	_, exitErr := os.Stat(target)

	if os.IsNotExist(exitErr) {
		if err := os.MkdirAll(target, 0755); err != nil {
			common.AppLogger.Error(fmt.Sprintf("create local temp folder failed: %v", err))
			setAllStatusFail("create local temp folder failed")
			return
		}
	}

	appBats := api.GetCodesByGroup("APP_COMMON_FILES")
	if len(appBats) == 0 {
		setAllStatusFail("get APP_COMMON_FILES bat files infomation failed")
		return
	}

	printerBats := api.GetCodesByGroup("PRINTER_COMMON_FILES")
	if len(printerBats) == 0 {
		setAllStatusFail("get PRINTER_COMMON_FILES bat files infomation failed")
		return
	}

	beforeBats := append(appBats, printerBats...)

	switch common.CurrentOA.StorageType {
	case "SMB":
		smbInstall(target, beforeBats)
	case "NGINX":
		nginxInstall(target, beforeBats)
	default:
		smbInstall(target, beforeBats)
	}
}

func (p *Deploy) InstallAfterReboot() {
	var tempMount string
	if common.CurrentOA.StorageType != "NGINX" {
		_, m, ret := mount()
		if !ret {
			defer exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
			return
		}
		tempMount = m
	}

	target := tempFilePath
	installPackages(target, tempMount)
}

func rebootForInstall() {
	saveTemporaryInfo()
	createScheduledTask("Deploy", []string{"-restart"})
	reboot()
}

func installPackages(target, mount string) {
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
			err0 := installRU(target, mount)
			if err0 != nil {
				common.AppLogger.Error(fmt.Sprintln("install ruservice failed:", err0))
				setPakcageStatusFailed(&installedPackages[i], err0, app)
			} else {
				installedPackages[i].Status = common.Completed.String()
				api.InstallationSuccess(app)
			}

			continue
		}

		var err error
		err = downloadInstallFiles(target, mount, installedPackages[i])
		if err != nil {
			common.AppLogger.Error(fmt.Sprintln("download package files failed:", err))
			setPakcageStatusFailed(&installedPackages[i], err, app)
			continue
		}
		beforebat := ""
		shortSeed := common.CurrentComputerInfo.Seed[0:4]
		switch installedPackages[i].AppType {
		case "Printer":
			beforebat = "PrinterEntrance.bat"
			if installedPackages[i].IP == "" {
				_, err = common.RunScriptWithArgs(path.Join(target, beforebat), installedPackages[i].WinFile, installedPackages[i].PrinterName, installedPackages[i].PrinterDriver, "", "", shortSeed)
			} else {
				_, err = common.RunScriptWithArgs(path.Join(target, beforebat), installedPackages[i].WinFile, installedPackages[i].PrinterName, installedPackages[i].PrinterDriver, installedPackages[i].PolNo, installedPackages[i].IP, shortSeed)
			}
		default:
			beforebat = "AppEntrance.bat"
			_, err = common.RunScriptWithArgs(path.Join(target, beforebat), installedPackages[i].WinFile, shortSeed)
		}

		if err != nil {
			common.AppLogger.Error(fmt.Sprintln("failed to install the application:", err))
			setPakcageStatusFailed(&installedPackages[i], err, app)
			continue
		}

		installedPackages[i].Status = common.Completed.String()

		api.InstallationSuccess(app)
	}

	exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
}

func installRU(dir, mount string) error {
	ru := api.GetAppVersionInfo("RU", common.GetOS())
	url := ru.InstallPath

	src := filepath.Join(dir, filepath.Base(url))
	var err error
	switch common.CurrentOA.StorageType {
	case "SMB":
		err = smbCopyRUService(mount, url, src)
	case "NGINX":
		downloadUrl := fmt.Sprintf("http://%s:%s%s/%s", common.CurrentOA.ServerName, common.CurrentOA.Port, common.CurrentOA.BaseUrl, url)
		downloadUrl = strings.ReplaceAll(downloadUrl, "\\", "/")
		err = downloadFileWithBasicAuth(downloadUrl, common.CurrentOA.UserName, common.Decode(common.CurrentOA.Password), src)
	default:
		err = smbCopyRUService(mount, url, src)
	}

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("download ruservice failed: %v", err))
		return err
	}

	err = installRUService(src)
	return err
}

func installRUService(src string) error {
	target := "C:\\Program Files\\RU\\ruservice.exe"
	targetDir := filepath.Dir(target)
	_, err := os.Stat(targetDir)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			common.AppLogger.Error(fmt.Sprintf("create ru directory failed: %v", err))
			return err
		}
	}

	if err := safeDeleteService("ruservice"); err != nil {
		common.AppLogger.Error(fmt.Sprintf("failed to delete service: %v", err))
	}

	srcFile, err := os.Open(src)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("open ruservice src file failed: %v", err))
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(target)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("create ruservice targe file failed: %v", err))
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile) // 核心拷贝逻辑
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("copy ruservice targe file failed: %v", err))
		return err
	}

	err = createRUService("RUService", target)
	return err
}

func createRUService(serviceName, binPath string) error {
	scm, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer scm.Disconnect()

	// 创建新服务
	if err := createService(scm, serviceName, binPath); err != nil {
		common.AppLogger.Error(fmt.Sprintf("create service error: %v", err))
	}

	return nil
}

func createService(scm *mgr.Mgr, name, binPath string) error {
	// 配置服务参数
	config := mgr.Config{
		DisplayName: name,
		StartType:   mgr.StartAutomatic, // 自动启动
		Description: name,
	}

	// 创建服务
	service, err := scm.CreateService(name, binPath, config)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("create service %s error: %v", name, err))
		return err
	}
	service.Close()
	common.AppLogger.Info(fmt.Sprintf("service %s created", name))
	return nil
}

// 检查服务是否存在并返回服务句柄
func serviceExists(scm *mgr.Mgr, name string) (bool, *mgr.Service, error) {
	service, err := scm.OpenService(name)
	if err != nil {
		// 错误码1060表示服务不存在（Windows系统错误码）
		if err.Error() == "The specified service does not exist." {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, service, nil
}

// 停止服务（若正在运行）
func stopServiceIfRunning(service *mgr.Service) error {
	status, err := service.Query()
	if err != nil {
		return err
	}

	// 仅当服务运行时才停止
	if status.State == svc.Running {
		_, err = service.Control(svc.Stop)
		if err != nil {
			return err
		}
		common.AppLogger.Info("service already stopped")
	}
	return nil
}

// 安全删除服务（包括停止和删除）
func safeDeleteService(name string) error {
	scm, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer scm.Disconnect()

	exists, service, err := serviceExists(scm, name)
	if err != nil {
		return err
	}
	if !exists {
		common.AppLogger.Info("service does not exist")
		return nil
	}
	defer service.Close()

	// 停止运行中的服务
	if err := stopServiceIfRunning(service); err != nil {
		return err
	}

	// 删除服务
	if err := service.Delete(); err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		status, _ := service.Query()
		if status.State == svc.Stopped {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	common.AppLogger.Info("service has been deleted")
	return nil
}

func nginxInstall(target string, bats []common.GroupCode) {
	for _, bat := range bats {
		//Copy bat file that will run first before running app bat
		localPath := filepath.Join(target, filepath.Base(bat.Name))
		// nginx 下载路径拼接
		downloadUrl := fmt.Sprintf("http://%s:%s%s/%s", common.CurrentOA.ServerName, common.CurrentOA.Port, common.CurrentOA.BaseUrl, bat.Name)
		// 下载 downloadUrl 的文件
		downError := downloadFileWithBasicAuth(downloadUrl, common.CurrentOA.UserName, common.Decode(common.CurrentOA.Password), localPath)
		if downError != nil {
			common.AppLogger.Error(fmt.Sprintf("copy file %s failed: %v", downloadUrl, downError))
			continue
		}

		common.AppLogger.Info(fmt.Sprintf("copy file %s successful", bat.Name))
	}

	//2. 执行安装
	installPackages(target, "")
}

func smbInstall(target string, bats []common.GroupCode) {
	tempMount, _, ret := mount()
	if !ret {
		defer exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
		return
	}

	for _, bat := range bats {
		//Copy bat file that will run first before running app bat
		localPath := filepath.Join(target, filepath.Base(bat.Name))
		source := filepath.Join(tempMount, bat.Name)
		cmdCopy := fmt.Sprintf("copy %s %s", source, localPath)

		_, err := os.Stat(source)

		if os.IsNotExist(err) {
			common.AppLogger.Error(fmt.Sprintf("%s source is not exist.", source))
			continue
		}
		if output, err := exec.Command("cmd", "/C", cmdCopy).CombinedOutput(); err != nil {
			common.AppLogger.Error(fmt.Sprintf("%s copy common bat files failed: %v\n error: %s", cmdCopy, err, common.DecodeByLocale(output)))
			setAllStatusFail("copy common bat files failed")
			return
		}

		common.AppLogger.Info(fmt.Sprintf("common bat file %s copy successful", bat.Name))
	}

	installPackages(target, tempMount)
}

package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

var mainTask = ""

func CreateFileWithAutoDirs(filePath string) error {
	// 输入验证
	if len(filePath) == 0 {
		return errors.New("路径不能为空")
	}

	// 路径标准化处理
	normalizedPath := filepath.Clean(filePath)
	if !filepath.IsAbs(normalizedPath) {
		return errors.New("必须使用绝对路径")
	}

	// 提取父目录
	parentDir := filepath.Dir(normalizedPath)

	// 递归创建目录
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("目录创建失败: %v", err)
	}

	// 检查文件存在性
	if _, err := os.Stat(normalizedPath); os.IsNotExist(err) {
		// 创建并打开文件
		_, err := os.Create(normalizedPath)
		if err != nil {
			return fmt.Errorf("文件创建失败: %v", err)
		}
	}

	return nil
}

func uploadInstallInfo() (string, error) {
	var installInfo common.InstallInfo

	pols := []string{common.CurrentComputerInfo.Name}

	installInfo.Pols = pols

	appids := make([]string, 0, len(installedPackages))
	for _, p := range installedPackages {
		appids = append(appids, p.ID)
	}
	installInfo.AppIds = appids

	maintaskid, err := api.UploadInstallInfo(installInfo)
	return maintaskid, err
}

func (p *Deploy) DoInstall() {
	getUploadInfo()
	api.UploadPCInfo(common.DetailPCInfo)

	maintaskid, err := uploadInstallInfo()
	mainTask = maintaskid
	if err != nil {
		setAllStatusFail()
		return
	}

	// 配置参数
	server, tempMount, remotePath, ret := mount()
	if !ret {
		defer exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
		return
	}

	paths := api.GetCodesByGroup("COMMON_BAT_DEPLOY_PATH")

	if len(paths) == 0 {
		setAllStatusFail()
		return
	}

	src := paths[0].Name
	target := "C:/Temp/tool"
	_, err = os.Stat(target)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0755); err != nil {
			common.AppLogger.Error(fmt.Sprintf("创建本地文件失败: %v", err))
			setAllStatusFail()
			return
		}
	}

	beforeBats := api.GetCodesByGroup("COMMON_BAT")

	for _, bat := range beforeBats {
		//Copy bat file that will run first before running app bat
		localPath := filepath.Join(target, bat.Name)
		source := filepath.Join(tempMount, filepath.Base(remotePath), src, bat.Name)
		cmdCopy := fmt.Sprintf("copy %s %s", source, localPath)

		if output, err := exec.Command("cmd", "/C", cmdCopy).CombinedOutput(); err != nil {
			common.AppLogger.Error(fmt.Sprintf("%s 拷贝失败: %v\n输出: %s", cmdCopy, err, common.DecodeByLocale(output)))
			setAllStatusFail()
			return
		}

		common.AppLogger.Info(fmt.Sprintf("文件 %s 拷贝成功", path.Join(src, bat.Name)))
	}

	installPackages(target, server)
}

func mount() (string, string, string, bool) {
	server := ""
	if common.CurrentOA.IP != "" {
		server = common.CurrentOA.IP
	} else {
		server = common.CurrentOA.ServerName
	}
	username := common.CurrentOA.UserName //get from server
	encryptedPassword := common.CurrentOA.Password
	password := common.Decode(encryptedPassword) //get from server

	exec.Command("cmd", "/C", "net use Z: /delete /y").Run() // 确保卸载

	// 1️⃣ 挂载远程共享目录到本地临时路径（Windows）
	tempMount := "Z:"                                                // 临时驱动器盘符
	remotePath := "\\\\" + server + "\\" + common.CurrentOA.RootPath // 远程共享路径
	cmdMount := fmt.Sprintf(
		"net use %s %s /user:%s %s",
		tempMount, remotePath, username, password,
	)

	if output, err := exec.Command("cmd", "/C", cmdMount).CombinedOutput(); err != nil {
		common.AppLogger.Error(fmt.Sprintf("挂载失败: %v\n output: %s\n", err, common.DecodeByLocale(output)))
		setAllStatusFail()
		return "", "", "", false
	}

	common.AppLogger.Info("挂载成功")
	return server, tempMount, remotePath, true
}

func (p *Deploy) InstallAfterReboot() {
	server, _, _, ret := mount()
	if !ret {
		defer exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
		return
	}

	target := "C:/Temp/tool"

	installPackages(target, server)
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
		common.AppLogger.Error(fmt.Sprintf("CreateService %s error: %v", name, err))
		return err
	}
	service.Close()
	common.AppLogger.Info(fmt.Sprintf("Service %s created", name))
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
		common.AppLogger.Info("服务已停止")
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
		common.AppLogger.Info("服务不存在")
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

	common.AppLogger.Info("服务已删除")
	return nil
}

func createRUService(serviceName, binPath string) error {
	scm, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer scm.Disconnect()

	// 创建新服务
	if err := createService(scm, serviceName, binPath); err != nil {
		common.AppLogger.Error(fmt.Sprintf("Create error: %v", err))
	}

	return nil
}

func installRU() error {
	ru := api.GetAppVersionInfo("RU")
	url := ru.DownloadUrl
	resp, err := http.Get(url)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("download ruservice failed:", err))
		return err
	}
	defer resp.Body.Close() // 必须关闭响应体[1,3,6](@ref)

	if resp.StatusCode != http.StatusOK {
		common.AppLogger.Error(fmt.Sprintln("download ruservice failed: ", resp.Status))
		return fmt.Errorf("download ruservice failed: %s", resp.Status)
	}

	src := "C:\\Temp\\Tool\\ruservice.exe"
	outFile, err := os.Create(src)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("creat ruservice failed: ", err))
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body) // 流式复制避免内存溢出
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("save file failed: ", err))
		return err
	}

	target := "C:\\Program Files\\RU\\ruservice.exe"
	targetDir := filepath.Dir(target)
	_, err = os.Stat(targetDir)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			common.AppLogger.Error(fmt.Sprintf("create ru directory failed: %v", err))
			return err
		}
	}

	if err := safeDeleteService("ruservice"); err != nil {
		common.AppLogger.Error(fmt.Sprintf("操作失败: %v", err))
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

func installPackages(target string, server string) {
	for i := range installedPackages {
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
				common.AppLogger.Error(fmt.Sprintln("错误:", err0))
				installedPackages[i].Status = common.Failed.String()
				installedPackages[i].Error = err0.Error()
				var failedapp common.FailedAppStatus
				failedapp.ID = app.ID
				failedapp.MainTask = app.MainTask
				failedapp.Msg = installedPackages[i].Error
				api.InstallationFailed(failedapp)
			} else {
				installedPackages[i].Status = common.Completed.String()
				api.InstallationSuccess(app)
			}

			continue
		}

		beforebat := ""
		beforebatouput := ""
		var err error
		shortSeed := common.CurrentComputerInfo.Seed[0:4]
		longSeed := common.CurrentComputerInfo.Seed
		switch installedPackages[i].AppType {
		case "Force_App":
			beforebat = "CTALAN.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), longSeed, server, installedPackages[i].AppName, installedPackages[i].WinFile, shortSeed, installedPackages[i].Path)
		case "Security_Patch":
			beforebat = installedPackages[i].WinFile
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server)
		case "Others":
			beforebat = "OTHERS.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), longSeed, server, installedPackages[i].AppName, installedPackages[i].WinFile, shortSeed, installedPackages[i].Path)
		case "Seed_Tasks":
			beforebat = "OTHERS.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), longSeed, server, installedPackages[i].AppName, installedPackages[i].WinFile, shortSeed, installedPackages[i].Path)
		case "LOCAL":
			beforebat = "Printer.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), longSeed, server, installedPackages[i].AppName, installedPackages[i].WinFile, shortSeed, installedPackages[i].Path)
		case "NETWORK":
			beforebat = "PrintQ.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server, installedPackages[i].AppName, installedPackages[i].WinFile, installedPackages[i].PolNo, installedPackages[i].IP, installedPackages[i].Path)
		}

		if err != nil {
			common.AppLogger.Error(fmt.Sprintln("错误:", err))
			installedPackages[i].Status = common.Failed.String()
			installedPackages[i].Error = err.Error()
			var failedapp common.FailedAppStatus
			failedapp.ID = app.ID
			failedapp.MainTask = app.MainTask
			failedapp.Msg = installedPackages[i].Error
			api.InstallationFailed(failedapp)
			continue
		}
		common.AppLogger.Info(fmt.Sprintln("Bat输出:", beforebatouput))

		// 执行第二个cmd文件
		localCmd := path.Join("C:/Temp/tool", "JOB.CMD")
		if common.FileExists(localCmd) {
			cmdOutput, err := common.RunScript(localCmd)
			if err != nil {
				common.AppLogger.Error(fmt.Sprintln("JOB.CMD 执行错误:", err))
				installedPackages[i].Status = common.Failed.String()
				installedPackages[i].Error = err.Error()

				var failedapp common.FailedAppStatus
				failedapp.ID = app.ID
				failedapp.MainTask = app.MainTask
				failedapp.Msg = installedPackages[i].Error
				api.InstallationFailed(failedapp)
				deleteTempFile("C:\\Temp\\tool\\JOB.CMD")
				continue
			} else {
				common.AppLogger.Info(fmt.Sprintln("Cmd输出:", cmdOutput))
			}

			deleteTempFile("C:\\Temp\\tool\\JOB.CMD")
		}

		installedPackages[i].Status = common.Completed.String()

		api.InstallationSuccess(app)
	}

	err := deleteTempFiles("C:\\Temp\\tool")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("delete文件错误:", err))
	}
	exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
}

func setAllStatusFail() {
	for i := range installedPackages {
		installedPackages[i].Status = common.Failed.String()
		installedPackages[i].Error = "Can't connect to OA Server."
		var app common.FailedAppStatus
		app.ID = installedPackages[i].ID
		app.MainTask = mainTask
		app.Msg = installedPackages[i].Error
		api.InstallationFailed(app)
	}
}

func deleteTempFile(file string) error {
	if err := os.Remove(file); err != nil {
		common.AppLogger.Error(fmt.Sprintf("删除 %s 失败: %v", file, err))
	}

	return nil
}

func deleteTempFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	// 遍历并删除每个子项
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		// 递归删除子项（文件或目录）
		if err := os.RemoveAll(fullPath); err != nil {
			common.AppLogger.Error(fmt.Sprintf("删除 %s 失败: %v", fullPath, err))
		}
	}

	return nil
}

func (p *Deploy) DeleteTempFiles() error {
	err := deleteTempFiles("C:\\Temp\\tool")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("delete文件错误:", err))
	}
	exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
	return err
}

func (p *Deploy) GetInstallStatus() []common.PackageInfo {
	common.AppLogger.Info("GetInstallStatus")

	uiShow := getInstallPackages()
	return uiShow
}

func (p *Deploy) Reboot() {
	reboot()
}

func saveTemporaryInfo() {
	var tempInfo common.TempInfo
	tempInfo.Packages = append(tempInfo.Packages, installedPackages...)
	tempInfo.Server = common.CurrentOA
	tempInfo.Computer = common.CurrentComputerInfo

	// 序列化为JSON
	jsonData, err := json.Marshal(tempInfo)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("序列化安装包列表失败：%v", err))
	}

	// 写入文件（0644权限：用户读写，组和其他读）
	err = os.WriteFile("temp.json", jsonData, 0644)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("保存安装包列表失败：%v", err))
	}
}

func (p *Deploy) LoadTemporaryInfo(path string) {
	common.AppLogger.Info("start LoadTemporaryInfo")
	file, err := os.Open(path)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("文件 %s 打开失败: %v", path, err))
		return
	}
	defer file.Close() // 确保关闭文件

	var tempInfo common.TempInfo
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tempInfo); err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON解析失败: %v", err))
		return
	}

	installedPackages = append(installedPackages, tempInfo.Packages...)
	common.CurrentOA = tempInfo.Server
	common.CurrentComputerInfo = tempInfo.Computer
}

func rebootForInstall() {
	saveTemporaryInfo()
	createScheduledTask("Deploy", []string{"-restart"})
	reboot()
}

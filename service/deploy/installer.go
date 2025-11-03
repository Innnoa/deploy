package deploy

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
var cancelling = false

func CreateFileWithAutoDirs(filePath string) error {
	// 输入验证
	if len(filePath) == 0 {
		return errors.New("the file path cannot be empty")
	}

	// 路径标准化处理
	normalizedPath := filepath.Clean(filePath)
	if !filepath.IsAbs(normalizedPath) {
		return errors.New("absolute path must be used")
	}

	// 提取父目录
	parentDir := filepath.Dir(normalizedPath)

	// 递归创建目录
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("directory creation failed: %v", err)
	}

	// 检查文件存在性
	if _, err := os.Stat(normalizedPath); os.IsNotExist(err) {
		// 创建并打开文件
		_, err := os.Create(normalizedPath)
		if err != nil {
			return fmt.Errorf("file creation failed: %v", err)
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
		setAllStatusFail("upload task infomation failed")
		return
	}

	target := "C:/Temp/tool"
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

// 下载文件并使用Basic Auth认证
func downloadFileWithBasicAuth(url, username, password, outputFilePath string) error {
	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加Basic Auth认证头
	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 创建输出文件
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer outFile.Close()

	// 将响应内容写入文件
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("文件下载成功: %s\n", outputFilePath)
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
	installPackages(target, common.CurrentOA.ServerName, "")
}

func smbInstall(target string, bats []common.GroupCode) {
	server := common.CurrentOA.ServerName
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

	installPackages(target, server, tempMount)
}

func mount() (string, string, bool) {
	server := ""
	// if common.CurrentOA.IP != "" {
	// 	server = common.CurrentOA.IP
	// } else {
	server = common.CurrentOA.ServerName
	// }
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
		common.AppLogger.Error(fmt.Sprintf("mount OA Server failed: %v\n error: %s\n", err, common.DecodeByLocale(output)))
		setAllStatusFail("can't connect to OA Server")
		return "", "", false
	}

	common.AppLogger.Info("mount OA Server successful")
	return tempMount, remotePath, true
}

func (p *Deploy) InstallAfterReboot() {
	server := common.CurrentOA.ServerName

	var tempMount string
	if common.CurrentOA.StorageType != "NGINX" {
		_, m, ret := mount()
		if !ret {
			defer exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
			return
		}
		tempMount = m
	}

	target := "C:/Temp/tool"
	installPackages(target, server, tempMount)
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

func installRU(dir, mount string) error {
	ru := api.GetAppVersionInfo("RU")
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

func smbCopyRUService(mount string, url string, src string) error {
	source := filepath.Join(mount, url)
	cmdCopy := fmt.Sprintf("copy %s %s", source, src)

	if output, err := exec.Command("cmd", "/C", cmdCopy).CombinedOutput(); err != nil {
		common.AppLogger.Error(fmt.Sprintf("%s copy ruservice.exe failed: %v\n error: %s", cmdCopy, err, common.DecodeByLocale(output)))
		return fmt.Errorf("%s copy ruservice.exe failed: %v\n error: %s", cmdCopy, err, common.DecodeByLocale(output))
	}
	return nil
}

// 下载安装文件到本地
func downloadInstallFiles(target, mount string, pkg common.PackageInfo) error {
	var err error
	switch common.CurrentOA.StorageType {
	case "SMB":
		err = smbDownloadInstallFiles(target, mount, pkg)
	case "NGINX":
		err = nginxDownloadInstallFiles(target, pkg)
	default:
		err = smbDownloadInstallFiles(target, mount, pkg)
	}

	if err != nil {
		return err
	} else {
		return nil
	}
}

func smbDownloadInstallFiles(target, mount string, pkg common.PackageInfo) error {
	appType := pkg.AppType
	if appType == "Printer" {
		// 额外需要下载文件
		printerFiles := api.GetCodesByGroup("PRINTER_COMMON_FILES")
		for _, file := range printerFiles {
			// 提取文件名
			filename := filepath.Base(file.Name)
			localPath := filepath.Join(target, filename)
			source := filepath.Join(mount, file.Name)
			cmdCopy := fmt.Sprintf("copy %s %s", source, localPath)

			_, err := os.Stat(source)

			if os.IsNotExist(err) {
				common.AppLogger.Error(fmt.Sprintf("%s source is not exist.", source))
				continue
			}
			if output, err := exec.Command("cmd", "/C", cmdCopy).CombinedOutput(); err != nil {
				common.AppLogger.Error(fmt.Sprintf("%s copy printer common bat files failed: %v\n error: %s", cmdCopy, err, common.DecodeByLocale(output)))
				continue
			}

			common.AppLogger.Info(fmt.Sprintf("common printer bat file %s copy successful", file.Name))
		}

	}

	srcPath := filepath.Join(mount, pkg.Path)
	cmdCopy := fmt.Sprintf("robocopy %s %s  /e /z /mt:16", srcPath, target)
	_, err := os.Stat(srcPath)

	if os.IsNotExist(err) {
		common.AppLogger.Error(fmt.Sprintf("%s source is not exist.", srcPath))
		return err
	}

	// 执行命令
	cmd := exec.Command("cmd", "/C", cmdCopy)
	output, err := cmd.CombinedOutput()

	// 获取命令的退出状态码
	exitCode := cmd.ProcessState.ExitCode()

	common.AppLogger.Error(fmt.Sprintf("%s copy package files exitCode: %d, output: \n %s", cmdCopy, exitCode, common.DecodeByLocale(output)))
	// 判断：只有当退出码大于等于8时，才认为是需要处理的错误
	if err != nil && exitCode >= 8 {
		common.AppLogger.Error(fmt.Sprintf("%s copy package files failed: %v\n error: %s", cmdCopy, err, common.DecodeByLocale(output)))
		return err
	}

	common.AppLogger.Info(fmt.Sprintf("package files %s copy successful", pkg.Path))

	return nil
}

func nginxDownloadInstallFiles(target string, pkg common.PackageInfo) error {
	nginxDownloader := common.NewNginxDownloader(5, common.CurrentOA.UserName, common.Decode(common.CurrentOA.Password))
	appType := pkg.AppType
	if appType == "Printer" {
		// 额外需要下载文件
		printerFiles := api.GetCodesByGroup("PRINTER_COMMON_FILES")
		for _, file := range printerFiles {
			// 提取文件名
			filename := filepath.Base(file.Name)
			localPath := filepath.Join(target, filename)
			// nginx 下载路径拼接
			downloadUrl := fmt.Sprintf("http://%s:%s%s/%s", common.CurrentOA.ServerName, common.CurrentOA.Port, common.CurrentOA.BaseUrl, file.Name)
			downloadUrl = strings.ReplaceAll(downloadUrl, "\\", "/")
			downError := downloadFileWithBasicAuth(downloadUrl, common.CurrentOA.UserName, common.Decode(common.CurrentOA.Password), localPath)
			if downError != nil {
				common.AppLogger.Error(fmt.Sprintf("\"Printer 类型 文件 %s 拷贝 失败: %v", downloadUrl, downError))
				continue
			}

			common.AppLogger.Info(fmt.Sprintf("Printer 类型 文件 %s 拷贝成功", file.Name))
		}
	}

	url, err := url.Parse(pkg.Path)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("URL 解析失败: %v", err))
		return err
	}
	common.AppLogger.Info(fmt.Sprintf("URL 解析成功: %s", url.String()))
	downloadUrl := fmt.Sprintf("http://%s:%s%s/%s", common.CurrentOA.ServerName, common.CurrentOA.Port, common.CurrentOA.BaseUrl, url.String())
	downError := nginxDownloader.DownloadFromNginx(downloadUrl, target)
	if downError != nil {
		common.AppLogger.Error(fmt.Sprintf("文件 %s 拷贝 失败: %v", downloadUrl, downError))
		return downError
	}

	common.AppLogger.Info(fmt.Sprintf("文件 %s 拷贝成功", downloadUrl))
	return nil
}

func installPackages(target, server, mount string) {
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
				_, err = common.RunScriptWithArgs(path.Join(target, beforebat), installedPackages[i].PrinterName, installedPackages[i].PrinterDriver, shortSeed)
			} else {
				_, err = common.RunScriptWithArgs(path.Join(target, beforebat), installedPackages[i].PrinterName, installedPackages[i].PrinterDriver, installedPackages[i].PolNo, installedPackages[i].IP, shortSeed)
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

	// err := deleteTempFiles("C:\\Temp\\tool")
	// if err != nil {
	// 	common.AppLogger.Error(fmt.Sprintln("delete temp files failed:", err))
	// }
	exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
}

func setPakcageStatusFailed(pkg *common.PackageInfo, err error, app common.AppStatus) {
	pkg.Status = common.Failed.String()
	pkg.Error = err.Error()
	var failedapp common.FailedAppStatus
	failedapp.ID = app.ID
	failedapp.MainTask = app.MainTask
	failedapp.Msg = pkg.Error
	api.InstallationFailed(failedapp)
}

func setAllStatusFail(reason string) {
	for i := range installedPackages {
		installedPackages[i].Status = common.Failed.String()
		installedPackages[i].Error = reason
		var app common.FailedAppStatus
		app.ID = installedPackages[i].ID
		app.MainTask = mainTask
		app.Msg = installedPackages[i].Error
		api.InstallationFailed(app)
	}
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
	tempInfo.MaintaskId = mainTask

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
	mainTask = tempInfo.MaintaskId
}

func rebootForInstall() {
	saveTemporaryInfo()
	createScheduledTask("Deploy", []string{"-restart"})
	reboot()
}

func (p *Deploy) CancelInatallation() {
	cancelling = true
	for _, value := range installedPackages {
		common.AppLogger.Info(fmt.Sprintf("CancelInatallation package.status : %s", value.Status))
		if value.Status == "" || value.Status == common.Waiting.String() || value.Status == common.Running.String() {
			value.Status = common.Canceled.String()

			var app common.AppStatus
			app.ID = value.ID
			app.MainTask = mainTask
			api.CancelInstallation(app)
		}
	}

	p.DeleteTempFiles()

	os.Exit(0)
}

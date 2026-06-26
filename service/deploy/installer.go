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
)

var mainTask = ""
var cancelling = false
var kylinSubmitted = false
var lastKylinPoll time.Time

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

func (p *Deploy) GetInstallStatus() []common.PackageInfo {
	common.AppLogger.Info("GetInstallStatus")

	if common.IsKylin() && kylinSubmitted && time.Since(lastKylinPoll) > 10*time.Second {
		lastKylinPoll = time.Now()
		pollKylinInstallStatus()
	}

	uiShow := getInstallPackages()
	return uiShow
}

func pollKylinInstallStatus() {
	for i := range installedPackages {
		if installedPackages[i].Status != common.Running.String() {
			continue
		}

		resp, err := api.GetKylinAppStatus(installedPackages[i].ID, mainTask)
		if err != nil {
			common.AppLogger.Error(fmt.Sprintf("poll Kylin status failed for %s: %v", installedPackages[i].AppName, err))
			continue
		}

		if resp.Data.Status != "SUCCESS" {
			continue
		}

		var app common.AppStatus
		app.ID = installedPackages[i].ID
		app.MainTask = mainTask

		for _, row := range resp.Data.Rows {
			if row.IStatusInstallOK == 1 {
				installedPackages[i].Status = common.Completed.String()
				api.InstallationSuccess(app)
				break
			}
			if row.IStatusInstallFail == 1 || row.IStatusDownloadFail == 1 {
				installedPackages[i].Status = common.Failed.String()
				installedPackages[i].Error = row.StrResult
				api.InstallationFailed(common.FailedAppStatus{AppStatus: app, Msg: row.StrResult})
				break
			}
		}
	}
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

	common.AppLogger.Error(fmt.Sprintf("temp json：%s", string(jsonData)))

	os.MkdirAll(tempFilePath, 0755)
	// 写入文件（0644权限：用户读写，组和其他读）
	err = common.WriteFileWithSync(path.Join(tempFilePath, "temp.json"), jsonData)
	// err = os.WriteFile(path.Join(tempFilePath, "temp.json"), jsonData, 0644)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("保存安装包列表失败：%v", err))
	} else {
		common.AppLogger.Info("saveTemporaryInfo success")
	}
}

func (p *Deploy) LoadTemporaryInfo() {
	common.AppLogger.Info("start LoadTemporaryInfo")

	path := filepath.Join(tempFilePath, "temp.json")
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

	if err := os.Remove(path); err != nil {
		common.AppLogger.Error(fmt.Sprintf("删除 %s 失败: %v", path, err))
	}
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

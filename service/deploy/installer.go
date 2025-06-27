package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strings"
)

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

func uploadInstallInfo() error {
	var installInfo common.InstallInfo

	pols := []string{common.CurrentComputerInfo.Name}

	installInfo.Pols = pols

	appids := make([]string, 0, len(installedPackages))
	for _, p := range installedPackages {
		appids = append(appids, p.ID)
	}
	installInfo.AppIds = appids

	err := api.UploadInstallInfo(installInfo)
	return err
}

func (p *Deploy) DoInstall() {
	getUploadInfo()
	api.UploadPCInfo(common.DetailPCInfo)

	err := uploadInstallInfo()
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

	root := "pcms"
	deploy := "deploy"
	target := "C:/Temp/tool"

	_, err = os.Stat(target)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0755); err != nil {
			common.AppLogger.Error(fmt.Sprintf("创建本地文件失败: %v", err))
			setAllStatusFail()
			return
		}

	}

	beforeBats := []string{"CTALAN.bat", "OTHERS.bat", "Printer.bat", "PrintQ.bat"}

	for _, bat := range beforeBats {
		//Copy bat file that will run first before running app bat
		localPath := filepath.Join(target, bat)
		source := filepath.Join(tempMount, filepath.Base(remotePath), root, deploy, bat)
		cmdCopy := fmt.Sprintf("copy %s %s", source, localPath)

		if output, err := exec.Command("cmd", "/C", cmdCopy).CombinedOutput(); err != nil {
			common.AppLogger.Error(fmt.Sprintf("%s 拷贝失败: %v\n输出: %s", cmdCopy, err, common.DecodeByLocale(output)))
			setAllStatusFail()
			return
		}

		common.AppLogger.Info(fmt.Sprintf("文件 %s 拷贝成功", path.Join(root, deploy, bat)))
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
	tempMount := "Z:"                             // 临时驱动器盘符
	remotePath := "\\\\" + server + "\\" + "seed" // 远程共享路径
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

	err := deleteTempFiles("C:\\Temp\\tool")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("delete文件错误:", err))
	}

}

func installPackages(target string, server string) {
	for i := range installedPackages {
		if installedPackages[i].Status == common.Completed.String() ||
			installedPackages[i].Status == common.Failed.String() {
			continue
		}

		var app common.AppId
		app.ID = installedPackages[i].ID
		api.StartInstall(app)
		installedPackages[i].Status = common.Running.String()

		if strings.TrimSpace(installedPackages[i].AppName) == "Restart Machine" {
			installedPackages[i].Status = common.Completed.String()
			api.InstallationSuccess(app)
			rebootForInstall()

			return
		}

		beforebat := ""
		beforebatouput := ""
		var err error
		shortSeed := common.CurrentComputerInfo.Seed[0:4]
		longSeed := common.CurrentComputerInfo.Seed
		switch strings.ToUpper(installedPackages[i].AppType) {
		case "APP":
			beforebat = "CTALAN.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, longSeed, installedPackages[i].AppName)
		case "SECURITYPATCH":
			beforebat = installedPackages[i].WinFile
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server)
		case "OTHERS":
			beforebat = "OTHERS.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, longSeed, installedPackages[i].AppName)
		case "Task":
			beforebat = "OTHERS.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, longSeed, installedPackages[i].AppName)
		case "LOCAL":
			beforebat = "Printer.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server, installedPackages[i].AppName, installedPackages[i].WinFile, shortSeed)
		case "NETWORK":
			beforebat = "PrintQ.bat"
			beforebatouput, err = common.RunScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, installedPackages[i].PolNo, installedPackages[i].IP)
		}

		if err != nil {
			common.AppLogger.Error(fmt.Sprintln("错误:", err))
			installedPackages[i].Status = common.Failed.String()
			installedPackages[i].Error = err.Error()
			api.InstallationFailed(app)
			continue
		}
		common.AppLogger.Info(fmt.Sprintln("Bat输出:", beforebatouput))

		// 执行第二个cmd文件
		localCmd := path.Join("C:/Temp/tool", "JOB.CMD")
		cmdOutput, err := common.RunScript(localCmd)
		if err != nil {
			common.AppLogger.Error(fmt.Sprintln("JOB.CMD 执行错误:", err))
			installedPackages[i].Status = common.Failed.String()
			installedPackages[i].Error = err.Error()
			api.InstallationFailed(app)
			deleteTempFile("C:\\Temp\\tool\\JOB.CMD")
			continue
		} else {
			common.AppLogger.Info(fmt.Sprintln("Cmd输出:", cmdOutput))
		}

		deleteTempFile("C:\\Temp\\tool\\JOB.CMD")
		installedPackages[i].Status = common.Completed.String()

		api.InstallationSuccess(app)

		defer exec.Command("cmd", "/C", "net use Z: /delete /y").Run()
	}
}

func setAllStatusFail() {
	for i := range installedPackages {
		installedPackages[i].Status = common.Failed.String()
		installedPackages[i].Error = "Can't connect to OA Server."
		var app common.AppId
		app.ID = installedPackages[i].ID
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

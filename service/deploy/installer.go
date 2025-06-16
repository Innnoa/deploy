package deploy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

type Charset string

const (
	UTF8    Charset = "UTF-8"
	GB18030 Charset = "GB18030" // Windows 中文编码
)

func ConvertByte2String(byteData []byte, charset Charset) string {
	switch charset {
	case GB18030:
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(byteData)
		return string(decodeBytes)
	default:
		return string(byteData)
	}
}

func runScriptWithArgs(scriptPath string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	if len(args) == 6 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1], args[2], args[3], args[4], args[5])
	} else if len(args) == 7 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1], args[2], args[3], args[4], args[5], args[6])
	} else if len(args) == 4 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1], args[2], args[3])
	} else if len(args) == 2 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1])
	}

	setHideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := ConvertByte2String(stdout.Bytes(), GB18030)
	errMsg := ConvertByte2String(stderr.Bytes(), GB18030)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s失败，退出码: %d\n错误输出: %s", scriptPath, exitErr.ExitCode(), errMsg)
		}
		return "", fmt.Errorf("启动%s失败: %v", scriptPath, err)
	}
	return output, nil
}

func runScript(scriptPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cmd", "/C", scriptPath)
	setHideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := ConvertByte2String(stdout.Bytes(), GB18030)
	errMsg := ConvertByte2String(stderr.Bytes(), GB18030)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s失败，退出码: %d\n错误输出: %s", scriptPath, exitErr.ExitCode(), errMsg)
		}
		return "", fmt.Errorf("启动%s失败: %v", scriptPath, err)
	}
	return output, nil
}

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
	err := uploadInstallInfo()
	if err != nil {
		setAllStatusFail()
		return
	}

	// 配置参数
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
		common.AppLogger.Error(fmt.Sprintf("挂载失败: %v\n output: %s\n", err, ConvertByte2String(output, GB18030)))
		setAllStatusFail()
		return
	}

	common.AppLogger.Info("挂载成功")

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
			common.AppLogger.Error(fmt.Sprintf("%s 拷贝失败: %v\n输出: %s", cmdCopy, err, ConvertByte2String(output, GB18030)))
			setAllStatusFail()
			return
		}

		common.AppLogger.Info(fmt.Sprintf("文件 %s 拷贝成功", path.Join(root, deploy, bat)))
	}

	for i := range installedPackages {
		var app common.AppId
		app.ID = installedPackages[i].ID
		api.StartInstall(app)

		installedPackages[i].Status = common.Running.String()

		// remoteFile := path.Join(installedPackages[i].Path, installedPackages[i].WinFile)
		// localPath := path.Join(target, installedPackages[i].WinFile)

		// // 执行拷贝
		// if err := copyFromSMB(rootShare, remoteFile, localPath); err != nil {
		// 	common.AppLogger.Error(fmt.Sprintln("操作失败:", err))
		// 	installedPackages[i].Status = common.Failed.String()
		// 	installedPackages[i].Error = "Copy file from OA Server failed."
		// 	api.InstallationFailed(app)
		// 	continue
		// } else {
		// 	common.AppLogger.Info(fmt.Sprintln("文件成功拷贝至:", localPath))
		// }

		beforebat := ""
		beforebatouput := ""
		shortSeed := common.CurrentComputerInfo.Seed[0:4]
		longSeed := common.CurrentComputerInfo.Seed
		switch strings.ToUpper(installedPackages[i].AppType) {
		case "APP":
			beforebat = "CTALAN.bat"
			beforebatouput, err = runScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, longSeed, installedPackages[i].AppName)
		case "SECURITYPATCH":
			beforebat = installedPackages[i].WinFile
			beforebatouput, err = runScriptWithArgs(path.Join(target, beforebat), shortSeed, server)
		case "OTHERS":
			beforebat = "OTHERS.bat"
			beforebatouput, err = runScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, longSeed, installedPackages[i].AppName)
		case "LOCAL":
			beforebat = "Printer.bat"
			beforebatouput, err = runScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile)
		case "NETWORK":
			beforebat = "PrintQ.bat"
			beforebatouput, err = runScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, installedPackages[i].PolNo, installedPackages[i].IP)
		}

		// beforebatouput, err := runScriptWithArgs(path.Join(target, beforebat), shortSeed, server, "", installedPackages[i].WinFile, longSeed, installedPackages[i].AppName)
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
		cmdOutput, err := runScript(localCmd)
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
	}

	err = deleteTempFiles("C:\\Temp\\tool")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintln("delete文件错误:", err))
	}

	getUploadInfo()
	api.UploadPCInfo(common.DetailPCInfo)

	defer exec.Command("cmd", "/C", "net use Z: /delete /y").Run() // 确保卸载
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
	return installedPackages
}

func (p *Deploy) Reboot() {
	// reboot()
}

func (p *Deploy) SaveTemporaryInfo() {

}

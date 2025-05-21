package deploy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"time"

	"github.com/hirochachacha/go-smb2"
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

func runScript(scriptPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cmd", "/C", scriptPath)
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

// 创建SMB客户端连接
func connectSMB(server, username, password, shareName string) (*smb2.Share, func(), error) {
	// 1. 建立TCP连接
	conn, err := net.Dial("tcp", server+":445")
	if err != nil {
		return nil, nil, fmt.Errorf("连接服务器失败: %v", err)
	}

	// 2. 配置NTLM认证
	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
		},
	}

	// 3. 创建会话
	session, err := dialer.Dial(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("认证失败: %v", err)
	}

	// 4. 挂载共享文件夹
	share, err := session.Mount(shareName)
	if err != nil {
		session.Logoff()
		conn.Close()
		return nil, nil, fmt.Errorf("挂载共享失败: %v", err)
	}

	// 返回共享对象及清理函数
	cleanup := func() {
		share.Umount()
		session.Logoff()
		conn.Close()
	}
	return share, cleanup, nil
}

// 从SMB复制文件到本地
func copyFromSMB(share *smb2.Share, remotePath, localPath string) error {
	// 1. 打开远程文件
	remoteFile, err := share.Open(remotePath)
	if err != nil {
		return fmt.Errorf("无法打开远程文件: %v", err)
	}
	defer remoteFile.Close()

	// 2. 创建本地文件
	err = CreateFileWithAutoDirs(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %v", err)
	}

	localFile, err := os.OpenFile(localPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer localFile.Close()

	// 3. 复制文件内容
	if _, err := io.Copy(localFile, remoteFile); err != nil {
		return fmt.Errorf("文件复制失败: %v", err)
	}

	// 4. 同步写入（可选）
	if err := localFile.Sync(); err != nil {
		return fmt.Errorf("文件同步失败: %v", err)
	}

	return nil
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
	username := "Administrator" //get from server
	password := "Deepit123"     //get from server
	shareName := "MasterOAServer"

	// 连接SMB
	share, cleanup, err := connectSMB(server, username, password, shareName)
	defer cleanup()
	if err != nil {
		fmt.Println("连接失败:", err)
		setAllStatusFail()

		return
	}

	for i := range installedPackages {
		var app common.AppId
		app.ID = installedPackages[i].ID
		api.StartInstall(app)

		installedPackages[i].Status = common.Running.String()

		remoteFile := path.Join("jobs", installedPackages[i].WinFile)
		localPath := path.Join("C:/Temp/tool", installedPackages[i].WinFile)

		// 执行拷贝
		if err := copyFromSMB(share, remoteFile, localPath); err != nil {
			fmt.Println("操作失败:", err)
			installedPackages[i].Status = common.Failed.String()
			installedPackages[i].Error = "Copy file from OA Server failed."
			api.InstallationFailed(app)
			continue
		} else {
			fmt.Println("文件成功拷贝至:", localPath)
		}

		os.Setenv("SRC", "\\\\"+server+"\\"+shareName+"\\"+installedPackages[i].Path)
		// 执行第一个bat文件
		batOutput, err := runScript(localPath)
		if err != nil {
			fmt.Println("错误:", err)
			installedPackages[i].Status = common.Failed.String()
			installedPackages[i].Error = err.Error()
			api.InstallationFailed(app)
			continue
		}
		fmt.Println("Bat输出:", batOutput)

		// 执行第二个cmd文件
		localCmd := path.Join("C:/Temp/tool", "JOB.CMD")
		cmdOutput, err := runScript(localCmd)
		if err != nil {
			fmt.Println("JOB.CMD 执行错误:", err)
			installedPackages[i].Status = common.Failed.String()
			installedPackages[i].Error = err.Error()
			api.InstallationFailed(app)
		} else {
			fmt.Println("Cmd输出:", cmdOutput)
		}

		err = deleteByOSCommand("C:\\Temp\\tool")
		if err != nil {
			fmt.Println("delete文件错误:", err)
			installedPackages[i].Status = common.Failed.String()
			installedPackages[i].Error = err.Error()
			api.InstallationFailed(app)
			continue
		}

		installedPackages[i].Status = common.Completed.String()

		api.InstallationSuccess(app)
	}
}

func setAllStatusFail() {
	for _, value := range installedPackages {
		value.Status = common.Failed.String()
		value.Error = "Can't connect to OA Server."
		var app common.AppId
		app.ID = value.ID
		api.InstallationFailed(app)
	}
}

func deleteByOSCommand(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	// 遍历并删除每个子项
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		// 递归删除子项（文件或目录）
		if err := os.RemoveAll(fullPath); err != nil {
			return fmt.Errorf("删除 %s 失败: %w", fullPath, err)
		}
	}

	return nil
}

func (p *Deploy) GetInstallStatus() []common.PackageInfo {
	fmt.Printf("GetInstallStatus")
	return installedPackages
}

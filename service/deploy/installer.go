package deploy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"recovery-unit-deploy/service/common"
	"time"

	"github.com/hirochachacha/go-smb2"
)

var installStatus []common.InstallStatus

func runScript(scriptPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cmd", "/C", scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s失败，退出码: %d\n错误输出: %s", scriptPath, exitErr.ExitCode(), stderr.String())
		}
		return "", fmt.Errorf("启动%s失败: %v", scriptPath, err)
	}
	return stdout.String(), nil
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
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %v", err)
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

func (p *Deploy) DoInstall() {
	// 配置参数
	server := common.CurrentOA
	username := "Administrator" //get from server
	password := "Deepit123"     //get from server
	shareName := "Master OA Server"

	for _, value := range allPackages {
		var status common.InstallStatus
		status.ID = value.ID
		status.Status = "Waiting"

		installStatus = append(installStatus, status)

		remoteFile := value.WinFile
		localPath := path.Join("C:\\Temp\\Tool", value.WinFile)

		// 连接SMB
		share, cleanup, err := connectSMB(server, username, password, shareName)
		defer cleanup()
		if err != nil {
			fmt.Println("连接失败:", err)
		}

		// 执行拷贝
		if err := copyFromSMB(share, remoteFile, localPath); err != nil {
			fmt.Println("操作失败:", err)
		} else {
			fmt.Println("文件成功拷贝至:", localPath)
		}
	}

	for _, value := range allPackages {
		os.Setenv("SRC", path.Join(common.CurrentOA, value.Path))
		// 执行第一个bat文件
		batOutput, err := runScript("script.bat")
		if err != nil {
			fmt.Println("错误:", err)
			continue
		}
		fmt.Println("Bat输出:", batOutput)

		// 执行第二个cmd文件
		cmdOutput, err := runScript("script.cmd")
		if err != nil {
			fmt.Println("错误:", err)
			continue
		}
		fmt.Println("Cmd输出:", cmdOutput)
	}

}

func (p *Deploy) GetInstallStatus() []common.InstallStatus {
	return installStatus
}

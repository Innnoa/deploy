package smb

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/hirochachacha/go-smb2"
)

// Client SMB客户端结构体
type Client struct {
	Hostname string
	Port     string
	RootPath string
	Username string
	Password string
	conn     net.Conn
	session  *smb2.Session
	share    *smb2.Share
}

// NewClient 创建新的SMB客户端
func NewClient(hostname string, port int, rootPath, username, password string) *Client {
	//port 可能为 0，默认使用 445
	if port == 0 {
		port = 445
	}
	return &Client{
		Hostname: hostname,
		Port:     fmt.Sprintf("%d", port),
		RootPath: rootPath,
		Username: username,
		Password: password,
	}
}

// Connect 连接到SMB服务器
func (c *Client) Connect() error {
	// 连接到SMB服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", c.Hostname, c.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to SMB server: %v", err)
	}
	c.conn = conn

	// 创建SMB会话
	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     c.Username,
			Password: c.Password,
		},
	}

	session, err := dialer.DialContext(context.Background(), conn)
	if err != nil {
		err := c.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close SMB session: %v", err)
		}
		return fmt.Errorf("failed to create SMB session: %v", err)
	}
	c.session = session

	// 连接到共享目录
	share, err := session.Mount(c.RootPath)
	if err != nil {
		err := c.session.Logoff()
		if err != nil {
			return err
		}
		err = c.conn.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to mount SMB share: %v", err)
	}
	c.share = share

	return nil
}

// Disconnect 断开SMB连接
func (c *Client) Disconnect() error {
	var lastErr error

	// 按顺序关闭资源
	if c.share != nil {
		if err := c.share.Umount(); err != nil {
			lastErr = err
		}
	}

	if c.session != nil {
		if err := c.session.Logoff(); err != nil {
			lastErr = err
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			lastErr = err
		}
	}

	// 重置状态
	c.share = nil
	c.session = nil
	c.conn = nil

	return lastErr
}

// ensureConnected 确保连接已建立
func (c *Client) ensureConnected() error {
	if c.conn == nil || c.session == nil || c.share == nil {
		return c.Connect()
	}
	return nil
}

// normalizePath 规范化SMB路径
func normalizePath(path string) string {
	// SMB路径使用正斜杠
	path = filepath.ToSlash(path)

	return path
}

// UploadFile 上传文件到SMB服务器
func (c *Client) UploadFile(localPath, remotePath string) error {
	// 确保连接已建立
	if err := c.ensureConnected(); err != nil {
		return fmt.Errorf("connection error: %v", err)
	}

	// 打开本地文件
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer srcFile.Close()

	// 获取文件信息
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 规范化远程路径
	normRemotePath := normalizePath(remotePath)

	// 创建远程文件
	dstFile, err := c.share.Create(normRemotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer dstFile.Close()

	// 设置文件大小
	if err := dstFile.Truncate(fileInfo.Size()); err != nil {
		return fmt.Errorf("failed to set file size: %v", err)
	}

	// 复制文件内容
	bytesWritten, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	if bytesWritten != fileInfo.Size() {
		return fmt.Errorf("file size mismatch: expected %d bytes, wrote %d bytes", fileInfo.Size(), bytesWritten)
	}

	return nil
}

// DownloadFile 从SMB服务器下载文件
func (c *Client) DownloadFile(remotePath, localPath string) error {
	// 确保连接已建立
	if err := c.ensureConnected(); err != nil {
		return fmt.Errorf("connection error: %v", err)
	}

	// 规范化远程路径
	normRemotePath := normalizePath(remotePath)

	// 打开远程文件
	srcFile, err := c.share.Open(normRemotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %v", err)
	}
	defer srcFile.Close()

	// 获取文件信息
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 确保本地目录存在
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %v", err)
	}

	// 创建本地文件
	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer dstFile.Close()

	// 复制文件内容
	bytesWritten, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	if bytesWritten != fileInfo.Size() {
		return fmt.Errorf("file size mismatch: expected %d bytes, wrote %d bytes", fileInfo.Size(), bytesWritten)
	}

	return nil
}

// ListFiles 列出SMB服务器目录中的文件
func (c *Client) ListFiles(remoteDir string) ([]string, error) {
	// 确保连接已建立
	if err := c.ensureConnected(); err != nil {
		return nil, fmt.Errorf("connection error: %v", err)
	}

	// 规范化远程路径
	normRemoteDir := normalizePath(remoteDir)

	// 打开目录
	entries, err := c.share.ReadDir(normRemoteDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	// 收集文件名
	var fileNames []string
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
	}

	return fileNames, nil
}

// FileExists 检查SMB服务器上的文件是否存在
func (c *Client) FileExists(remotePath string) (bool, bool, error) {
	// 确保连接已建立
	if err := c.ensureConnected(); err != nil {
		return false, false, fmt.Errorf("connection error: %v", err)
	}

	// 规范化远程路径
	normRemotePath := normalizePath(remotePath)

	// 尝试获取文件信息
	f, err := c.share.Stat(normRemotePath)
	if err != nil {
		// 如果是文件不存在的错误，返回false
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to check file existence: %v", err)
	}
	dir := f.IsDir()

	return true, dir, nil
}

// 目录下载，递归下载目录下的所有文件
func (c *Client) DownloadDir(remoteDir, localDir string) error {
	// 确保连接已建立
	if err := c.ensureConnected(); err != nil {
		return fmt.Errorf("connection error: %v", err)
	}

	// 规范化远程路径
	normRemoteDir := normalizePath(remoteDir)

	// 列出目录下的所有文件
	fileNames, err := c.ListFiles(normRemoteDir)
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	// 递归下载每个文件
	for _, fileName := range fileNames {
		remotePath := filepath.Join(normRemoteDir, fileName)
		remotePath = normalizePath(remotePath)
		localPath := filepath.Join(localDir, fileName)

		// 如果是目录，递归调用
		if _, isDir, err := c.FileExists(remotePath); err == nil && isDir {
			if err := c.DownloadDir(remotePath, localPath); err != nil {
				return fmt.Errorf("failed to download directory %s: %v", remotePath, err)
			}
		} else if err == nil {
			// 如果是文件，直接下载
			if err := c.DownloadFile(remotePath, localPath); err != nil {
				return fmt.Errorf("failed to download file %s: %v", remotePath, err)
			}
		} else {
			return fmt.Errorf("failed to check file existence %s: %v", remotePath, err)
		}
	}

	return nil
}

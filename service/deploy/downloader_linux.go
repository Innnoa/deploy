//go:build linux
// +build linux

package deploy

import (
	"fmt"
	"path/filepath"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"recovery-unit-deploy/service/smb"
	"strconv"
	"strings"
)

func scriptDownload() error {
	bats := api.GetCodesByGroup("UOS_PRINTER_COMMON_FILES")
	if len(bats) == 0 {
		return fmt.Errorf("get UOS_PRINTER_COMMON_FILES sh files infomation failed")
	}
	switch common.CurrentOA.StorageType {
	case "SMB":
		return smbDownload(bats)
	case "NGINX":
		return nginxDownload(bats)
	default:
		return smbDownload(bats)
	}
}

func smbDownload(beforeBats []common.GroupCode) error {
	server := common.CurrentOA.ServerName
	username := common.CurrentOA.UserName //get from server
	encryptedPassword := common.CurrentOA.Password
	password := common.Decode(encryptedPassword) //get from server
	port := 0
	if common.CurrentOA.Port != "" {
		// 字符串转int
		_port, err := strconv.Atoi(common.CurrentOA.Port)
		if err != nil {
			return fmt.Errorf("端口号转换失败: %v", err)
		}
		port = _port
	}
	client := smb.NewClient(server, port, common.CurrentOA.RootPath, username, password)
	err := client.Connect()
	if err != nil {
		return fmt.Errorf("连接smb服务器失败: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			common.AppLogger.Warning(fmt.Sprintf("断开 SMB 连接失败: %v", err))
		}
	}()

	for _, bat := range beforeBats {
		//Copy bat file that will run first before running app bat
		fileName := filepath.Base(bat.Name)
		localPath := filepath.Join(tempFilePath, fileName)
		source := bat.Name
		exists, _, err := client.FileExists(source)
		if !exists {
			common.AppLogger.Error(fmt.Sprintf("%s source is not exist: %v", source, err))
			continue
		}
		err = client.DownloadFile(source, localPath)
		if err != nil {
			return fmt.Errorf("下载 %s 失败: %v", source, err)
		}
		common.AppLogger.Info(fmt.Sprintf("文件 %s 拷贝成功", bat.Name))
	}
	return nil
}

func nginxDownload(beforeBats []common.GroupCode) error {
	for _, bat := range beforeBats {
		fileName := filepath.Base(bat.Name)
		localPath := filepath.Join(tempFilePath, fileName)
		// nginx 下载路径拼接
		downloadUrl := fmt.Sprintf("http://%s:%s%s/%s", common.CurrentOA.ServerName, common.CurrentOA.Port, common.CurrentOA.BaseUrl, bat.Name)
		// 下载 downloadUrl 的文件
		downError := downloadFileWithBasicAuth(downloadUrl, common.CurrentOA.UserName, common.Decode(common.CurrentOA.Password), localPath)
		if downError != nil {
			return fmt.Errorf("文件 %s 拷贝 失败: %v", downloadUrl, downError)
		}

		common.AppLogger.Info(fmt.Sprintf("文件 %s 拷贝成功", bat.Name))
	}
	return nil
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

//go:build windows
// +build windows

package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"recovery-unit-deploy/service/common"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

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

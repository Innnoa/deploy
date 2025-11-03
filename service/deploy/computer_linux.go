//go:build linux
// +build linux

package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type OEMInfo struct {
	Basic       BasicInfo  `json:"basic"`
	Custom      CustomInfo `json:"custom_info"`
	ToolVersion string     `json:"tools_version"`
	Version     string     `json:"version"`
}

type BasicInfo struct {
	ISOId     string `json:"iso_id"`
	OEMCode   string `json:"oem_code"`
	TimeStamp string `json:"timestamp"`
	Type      int    `json:"type"`
}

type CustomInfo struct {
	CustomizedKernel bool `json:"customized_kernel"`
	DevelopMode      bool `json:"develop_mode"`
}

type BlockDevice struct {
	Name       string        `json:"name"`               // 设备名称，如sda、sda1
	MajMin     string        `json:"maj:min"`            // 主次设备号
	Rm         bool          `json:"rm"`                 // 是否为可移动设备
	Size       string        `json:"size"`               // 设备大小
	Ro         bool          `json:"ro"`                 // 是否为只读设备
	Type       string        `json:"type"`               // 设备类型：disk、part、rom等
	Mountpoint *string       `json:"mountpoint"`         // 挂载点（使用指针允许null值）
	Children   []BlockDevice `json:"children,omitempty"` // 子设备（分区）
}

// BlockDeviceList 表示最外层的JSON结构
type BlockDeviceList struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

type DiskSpaceInfo struct {
	Filesystem string
	Size       string
	Used       string
	Available  string
	UsePercent string
	MountPoint string
}

func runCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // 标准输出
	cmd.Stderr = &stderr // 标准错误

	// 执行命令
	err := cmd.Run()

	// 获取输出内容
	output := stdout.String()
	errMsg := stderr.String()

	// 构建返回的error
	var resultErr error
	if err != nil || errMsg != "" {
		// 合并命令执行错误和stderr内容
		errorDetails := fmt.Sprintf("命令执行错误: %v, 标准错误: %s", err, errMsg)
		resultErr = fmt.Errorf(errorDetails)
	}

	return output, resultErr
}

func parseDfOutput(output string) (string, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 2 {
		return "", fmt.Errorf("输出行数不足")
	}

	// 解析数据行
	dataLine := strings.TrimSpace(lines[1])
	fields := strings.Fields(dataLine)

	if len(fields) < 6 {
		return "", fmt.Errorf("输出格式不符合预期，期望至少6个字段，得到%d个", len(fields))
	}

	// 返回可用空间（第4个字段）
	return fields[3], nil
}

func getDiskInfo() []DiskInfo {
	var disks []DiskInfo

	output, err := runCommand("lsblk", "-J")

	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("命令执行失败: %s\n", err))
		return disks
	}

	var deviceList BlockDeviceList

	// 反序列化JSON字符串
	err = json.Unmarshal([]byte(output), &deviceList)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("JSON反序列化失败: %v", err))
		return disks
	}

	for _, device := range deviceList.BlockDevices {
		common.AppLogger.Info(fmt.Sprintf("设备: %s, 类型: %s, 大小: %s\n", device.Name, device.Type, device.Size))
		if strings.EqualFold(device.Type, "disk") {
			var disk DiskInfo
			disk.DeviceID = device.Name
			disk.Size = device.Size
			var fs int = 0

			if len(device.Children) > 0 {
				for _, child := range device.Children {
					output, err := runCommand("df", "-B1", child.Name)
					if output != "" && err == nil {
						freeSpace, err := parseDfOutput(output)
						if err == nil {
							num, err := strconv.Atoi(freeSpace)
							if err != nil {
								common.AppLogger.Error(fmt.Sprintf("%s 转换失败：%v\n", freeSpace, err))
							} else {
								fs += num
							}
						} else {
							common.AppLogger.Error(fmt.Sprintf("解析df输出失败: %s\n", err))
						}
					}
				}
			}

			gb := float64(fs) / (1024 * 1024 * 1024)
			disk.FreeSpace = fmt.Sprintf("%.2fGB", gb)

			disks = append(disks, disk)
		}
	}

	return disks
}

func getComputerName() string {
	hostname := os.Getenv("HOSTNAME")

	parts := strings.Split(hostname, ".")

	var computerName string
	if len(parts) > 0 {
		computerName = parts[0]
	}

	return computerName
}

func getUploadInfo() common.DetailComputerInfo {
	common.DetailPCInfo.PolNo = common.CurrentComputerInfo.Name
	common.DetailPCInfo.IP = common.CurrentComputerInfo.IP
	common.DetailPCInfo.Seedlabel = common.CurrentComputerInfo.Seed
	common.DetailPCInfo.SP = common.CurrentComputerInfo.Seed[len(common.CurrentComputerInfo.Seed)-3:]

	return common.DetailPCInfo
}

func getLastKBCode() string {
	data, err := os.ReadFile("/etc/oem-info")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("读取文件出错:", err))
		return ""
	}

	var info OEMInfo

	err = json.Unmarshal(data, &info)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("解析JSON出错:", err))
		return ""
	}

	common.DetailPCInfo.KBCode = info.Basic.OEMCode

	common.AppLogger.Info(fmt.Sprintf("oem-code is: %s\n", common.DetailPCInfo.KBCode))
	return common.DetailPCInfo.KBCode
}

func getCreationTime(fileInfo os.FileInfo) time.Time {
	stat := fileInfo.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
}

func checkSeedFile() bool {
	filename := fmt.Sprintf("/etc/seedinfo/%s.seedlabel.txt", common.CurrentSeed.SeedLabel)
	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	// 修改时间（跨平台通用）
	modTime := fileInfo.ModTime()

	// 创建时间（按系统处理）
	createTime := getCreationTime(fileInfo)

	strModTime := modTime.Format("2006-01-02 15:04:05")
	strCreateTime := createTime.Format("2006-01-02 15:04:05")

	common.AppLogger.Info(fmt.Sprintf("seedlabel file create time: %s\n", createTime))
	common.AppLogger.Info(fmt.Sprintf("seedlabel file modify time: %s\n", strModTime))

	return api.CheckSeedLabel(common.CurrentSeed.SeedLabel, strModTime, strCreateTime)
}

func getOpSystemInfo() string {

	return "Linux"
}

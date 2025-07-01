//go:build windows
// +build windows

package deploy

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"recovery-unit-deploy/service/common"
	"regexp"
	"strings"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows/registry"
)

type Win32_ComputerSystem struct {
	BootupState string // 启动状态（如 "Normal boot", "Failover boot"）
	Model       string // PC型号（如 "HP ProDesk 600 G3"）
}

type Win32_PhysicalMemory struct {
	Capacity uint64 // 内存容量 (字节)
}

// CPU 信息结构体
type CPUInfo struct {
	Name          string
	MaxClockSpeed float32
}

// 内存信息结构体
type MemoryInfo struct {
	TotalPhysical int64 // GB
}

// boot信息结构体
type SystemInfo struct {
	BootMode string
	Model    string // PC型号（如 "HP ProDesk 600 G3"）
}

type Win32_Processor struct {
	MaxClockSpeed uint32
	Name          string
}

type Win32_OperatingSystem struct {
	SystemDirectory string // 系统目录路径（如 "C:\Windows\System32"）
}

type Win32_LogicalDisk struct {
	DeviceID  string // 盘符（如 "C:"）
	FreeSpace uint64 // 剩余空间（字节）
	Size      uint64 // 总容量（字节）
}

type DiskInfo struct {
	DeviceID  string // 盘符（如 "C"）
	FreeSpace string // 剩余空间（GB）
	Size      string // 总容量（GB）
}

func getUploadInfo() common.DetailComputerInfo {
	common.DetailPCInfo.PolNo = common.CurrentComputerInfo.Name
	common.DetailPCInfo.IP = common.CurrentComputerInfo.IP
	common.DetailPCInfo.Seedlabel = common.CurrentComputerInfo.Seed
	common.DetailPCInfo.SP = common.CurrentComputerInfo.Seed[len(common.CurrentComputerInfo.Seed)-3:]

	disks := getDiskInfo()
	common.DetailPCInfo.NumOfDrive = fmt.Sprintf("%d", len(disks))
	if len(disks) > 0 {
		common.DetailPCInfo.SizeOfDrive1 = disks[0].Size

		if len(disks) > 1 {
			common.DetailPCInfo.SizeOfDrive2 = disks[1].Size
		}

		for _, d := range disks {
			if strings.EqualFold(d.DeviceID, "C") {
				common.DetailPCInfo.FreeSpaceC = d.FreeSpace
			} else if strings.EqualFold(d.DeviceID, "D") {
				common.DetailPCInfo.FreeSpaceD = d.FreeSpace
			}
		}

		common.DetailPCInfo.LastDrive = disks[len(disks)-1].DeviceID
		common.DetailPCInfo.SystemDrive = getOpSystemInfo()
	}

	cpu := getCPUInfo()
	common.DetailPCInfo.CpuType = fmt.Sprintf("%s @ %.2f GHz", cpu.Name, cpu.MaxClockSpeed)
	common.DetailPCInfo.CpuSpeed = fmt.Sprintf("%.2f GHz", cpu.MaxClockSpeed)

	ram := getMemoryInfo()
	common.DetailPCInfo.Ram = fmt.Sprintf("%d MB", ram.TotalPhysical)

	sysInfo := getSystemInfo()
	common.DetailPCInfo.BootEnv = sysInfo.BootMode
	common.DetailPCInfo.PCModel = sysInfo.Model

	common.DetailPCInfo.OS = getOS()

	return common.DetailPCInfo
}

func getSystemInfo() SystemInfo {
	var systems []Win32_ComputerSystem
	query := wmi.CreateQuery(&systems, "")
	if err := wmi.QueryNamespace(query, &systems, "root\\cimv2"); err != nil {
		log.Fatalf("无法查询系统信息: %v", err)
	}
	if len(systems) == 0 {
		log.Fatal("WMI查询失败或未获取数据")
	}

	bootState := systems[0].BootupState
	bootMode := ""
	if bootState == "Normal boot" { // 通常UEFI启动为Normal boot
		bootMode = "UEFI"
	} else {
		bootMode = "Legacy"
	}
	model := systems[0].Model

	return SystemInfo{
		BootMode: bootMode,
		Model:    model,
	}
}

// 获取内存信息
func getMemoryInfo() MemoryInfo {

	var osList []Win32_PhysicalMemory
	query := wmi.CreateQuery(&osList, "")
	if err := wmi.QueryNamespace(query, &osList, "root\\cimv2"); err != nil {
		log.Fatalf("无法查询内存信息: %v", err)
	}

	if len(osList) == 0 {
		log.Fatal("未找到操作系统信息")
	}

	os := osList[0]

	// 转换 byte 到 GB
	const byteToMB = 1024.0 * 1024.0
	totalPhysical := int64(float64(os.Capacity) / byteToMB)

	return MemoryInfo{
		TotalPhysical: totalPhysical,
	}
}

// 获取CPU信息
func getCPUInfo() CPUInfo {
	// 正确的 CPU 信息结构

	var cpus []Win32_Processor
	query := wmi.CreateQuery(&cpus, "")
	if err := wmi.QueryNamespace(query, &cpus, "root\\cimv2"); err != nil {
		log.Fatalf("无法查询CPU信息: %v", err)
	}

	if len(cpus) == 0 {
		log.Fatal("未找到CPU信息")
	}

	cpu := cpus[0] // 如果有多个CPU，通常第一个是系统使用的

	return CPUInfo{
		Name:          cpu.Name,
		MaxClockSpeed: float32(cpu.MaxClockSpeed) / 1000, // MHz to GHz
	}
}

func getOpSystemInfo() string {
	var opSys []Win32_OperatingSystem
	query := wmi.CreateQuery(&opSys, "")

	if err := wmi.QueryNamespace(query, &opSys, "root\\cimv2"); err != nil {
		log.Fatalf("无法查询操作系统信息: %v", err)
	}

	if len(opSys) == 0 {
		log.Fatal("WMI查询失败或未获取数据")
	}

	return opSys[0].SystemDirectory[0:1]
}

func getDiskInfo() []DiskInfo {
	var wdisks []Win32_LogicalDisk

	query := wmi.CreateQuery(&wdisks, "")
	if err := wmi.QueryNamespace(query, &wdisks, "root\\cimv2"); err != nil {
		log.Fatalf("无法查询磁盘信息: %v", err)
	}
	if len(wdisks) == 0 {
		log.Fatal("WMI查询失败或未获取数据")
	}

	var disks []DiskInfo

	for _, wd := range wdisks {
		freeGB := fmt.Sprintf("%.2f GB", float64(wd.FreeSpace)/(1024*1024*1024))
		totalGB := fmt.Sprintf("%.2f GB", float64(wd.Size)/(1024*1024*1024))

		var disk = DiskInfo{DeviceID: wd.DeviceID[0:1], FreeSpace: freeGB, Size: totalGB}
		disks = append(disks, disk)
	}

	return disks
}

func getLastKBCode() string {
	var kbList []string
	cmd := exec.Command("systeminfo")
	common.SetHideWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("获取systeminfo失败: %v\n", err))
		return ""
	}

	// 解析补丁信息
	kbRegex := regexp.MustCompile(`KB\d+`)
	lines := strings.Split(out.String(), "\n")

	for _, line := range lines {
		matches := kbRegex.FindAllString(line, -1)
		kbList = append(kbList, matches...)
	}

	// 去重并输出
	kbList = removeDuplicates(kbList)

	if len(kbList) > 0 {
		common.DetailPCInfo.KBCode = kbList[len(kbList)-1]
	}

	common.AppLogger.Info(fmt.Sprintf("Last KBCode is: %s\n", common.DetailPCInfo.KBCode))
	return common.DetailPCInfo.KBCode
}

func removeDuplicates(list []string) []string {
	keys := make(map[string]bool)
	var unique []string
	for _, item := range list {
		if _, exists := keys[item]; !exists {
			keys[item] = true
			unique = append(unique, item)
		}
	}
	return unique
}

func getOS() string {
	return common.GetRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "ProductName")
}

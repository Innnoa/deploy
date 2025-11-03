//go:build windows
// +build windows

package deploy

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
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

	logonId, lastsignon := getLoginInfo()
	common.DetailPCInfo.LogonId = logonId
	common.DetailPCInfo.LastSignon = lastsignon

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

func checkSeedFile() bool {
	filename := fmt.Sprintf("C:\\%s.seedlabel.txt", common.CurrentSeed.SeedLabel)
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

// 定义Windows API函数及常量
var (
	modAdvapi32      = syscall.NewLazyDLL("Advapi32.dll")
	procOpen         = modAdvapi32.NewProc("OpenEventLogW")
	procRead         = modAdvapi32.NewProc("ReadEventLogW")
	procClose        = modAdvapi32.NewProc("CloseEventLog")
	procConvertSid   = modAdvapi32.NewProc("ConvertStringSidToSidW")
	lookupAccountSid = modAdvapi32.NewProc("LookupAccountSidW")
)

const (
	EVENTLOG_BACKWARDS_READ  = 0x0008
	EVENTLOG_SEQUENTIAL_READ = 0x0001
	EVENTLOG_SUCCESS         = 0x0000
)

type EVENTLOGRECORD struct {
	Length              uint32 // 记录总长度（含填充字节）[1,2](@ref)
	Reserved            uint32 // 固定值 0x654c664c ("eLfL")，用于标识结构[1,2](@ref)
	RecordNumber        uint32 // 事件记录号（日志内唯一序号）[1,3](@ref)
	TimeGenerated       uint32 // 事件生成时间（UTC 秒数，自 1970-01-01）[1,2](@ref)
	TimeWritten         uint32 // 事件写入日志时间（UTC 秒数）[1,2](@ref)
	EventID             uint32 // 事件标识符（与事件源关联）[1,7](@ref)
	EventType           uint16 // 事件类型（见下方常量定义）[1,2](@ref)
	NumStrings          uint16 // 描述字符串数量（0-256）[1,2](@ref)
	EventCategory       uint16 // 事件分类（由事件源定义）[1,2](@ref)
	ReservedFlags       uint16 // 保留标志（指示是否含 XML）[1](@ref)
	ClosingRecordNumber uint32 // 保留字段（固定为 0）[1,2](@ref)
	StringOffset        uint32 // 描述字符串区域的偏移地址[1,2](@ref)
	UserSidLength       uint32 // 用户 SID 长度（字节数，0 表示无 SID）[1,2](@ref)
	UserSidOffset       uint32 // 用户 SID 的偏移地址[1,2](@ref)
	DataLength          uint32 // 事件二进制数据长度[1,2](@ref)
	DataOffset          uint32 // 事件二进制数据的偏移地址[1,2](@ref)
}

var logonTypeArray = []string{
	"2", "7", "11",
}

var excludeSIDArray = []string{
	"S-1-5-90", "S-1-5-96", "S-1-5-18", "S-1-5-19", "S-1-5-20", "S-1-5-6", "S-1-5-80",
}

func getLoginInfo() (string, string) {
	// 1. 打开安全日志
	serverName, _ := windows.UTF16PtrFromString("")
	logName, _ := windows.UTF16PtrFromString("Security")
	handle, _, err := procOpen.Call(uintptr(unsafe.Pointer(serverName)), uintptr(unsafe.Pointer(logName)))
	if handle == 0 {
		common.AppLogger.Error(fmt.Sprintf("procOpen error: %v", err))
		procClose.Call(handle)

		return "", ""
	}
	// 2. 反向读取日志
	buf := make([]byte, 1024) // 初始缓冲区
	var bytesRead uint32
	var minBytesNeeded uint32
	for {
		ret, _, err := procRead.Call(
			handle,
			EVENTLOG_BACKWARDS_READ|EVENTLOG_SEQUENTIAL_READ,
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			uintptr(unsafe.Pointer(&bytesRead)),
			uintptr(unsafe.Pointer(&minBytesNeeded)),
		)
		if ret == 0 { // 读取失败
			common.AppLogger.Error(fmt.Sprintf("procRead error: %v", err))
			break
		}

		// 3. 解析每条记录
		ptr := unsafe.Pointer(&buf[0])
		for bytesRead > 0 {
			record := (*EVENTLOGRECORD)(ptr)
			common.AppLogger.Info(fmt.Sprintf("record: %v", record))
			if record.EventID == 4624 { // 登录成功事件
				strs := parseStrings(record)

				if !slices.Contains(excludeSIDArray, strs[4]) && slices.Contains(logonTypeArray, strs[8]) {
					_, name, _ := sidToUsername(strs[4])
					common.CurrentUser = name
					common.AppLogger.Info(fmt.Sprintf("username: %s", name))
					timestamp := int64(record.TimeGenerated)
					// 转换为 time.Time 对象
					t := time.Unix(timestamp, 0).UTC() // 明确指定 UTC 时区
					location, err := time.LoadLocation("Asia/Shanghai")
					if err != nil {
						common.AppLogger.Error(fmt.Sprintf("LoadLocation error: %v", err))
						location = time.FixedZone("CST", 8*3600) // UTC+8 for Shanghai
					}
					localTime := t.In(location).Format("2006-01-02 15:04:05")
					common.AppLogger.Info(fmt.Sprintf("localTime: %s", localTime))

					return name, localTime
				}
			}
			// 移动到下一条记录
			ptr = unsafe.Pointer(uintptr(ptr) + uintptr(record.Length))
			bytesRead -= record.Length
		}
	}
	defer procClose.Call(handle)

	return "", ""
}

func sidToUsername(sid string) (domain, user string, err error) {
	// 将字符串 SID 转为二进制 SID
	var psid *syscall.SID
	sidPtr, _ := syscall.UTF16PtrFromString(sid)
	ret, _, _ := procConvertSid.Call(uintptr(unsafe.Pointer(sidPtr)), uintptr(unsafe.Pointer(&psid)))
	if ret == 0 {
		return "", "", syscall.GetLastError()
	}
	defer syscall.LocalFree((syscall.Handle)(unsafe.Pointer(psid)))

	// 查询账户名和域名
	var (
		name, domainBuf [128]uint16
		nameLen         = uint32(len(name))
		domainLen       = uint32(len(domainBuf))
		sidType         uint32
	)
	ret, _, _ = lookupAccountSid.Call(
		0, // 本地系统
		uintptr(unsafe.Pointer(psid)),
		uintptr(unsafe.Pointer(&name[0])),
		uintptr(unsafe.Pointer(&nameLen)),
		uintptr(unsafe.Pointer(&domainBuf[0])),
		uintptr(unsafe.Pointer(&domainLen)),
		uintptr(unsafe.Pointer(&sidType)),
	)
	if ret == 0 {
		return "", "", syscall.GetLastError()
	}

	return syscall.UTF16ToString(domainBuf[:]), syscall.UTF16ToString(name[:]), nil
}

func parseStrings(record *EVENTLOGRECORD) []string {
	base := uintptr(unsafe.Pointer(record)) + uintptr(record.StringOffset)
	ptr := unsafe.Pointer(base)
	strings := make([]string, 0, record.NumStrings)
	for i := 0; i < int(record.NumStrings); i++ {
		s := windows.UTF16PtrToString((*uint16)(ptr))
		strings = append(strings, s)
		// 移动到下一个字符串（+2字节乘字符串长度）
		ptr = unsafe.Pointer(uintptr(ptr) + uintptr(2*(len(s)+1)))
	}
	return strings
}

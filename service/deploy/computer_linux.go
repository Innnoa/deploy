//go:build linux
// +build linux

package deploy

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"recovery-unit-deploy/service/api"
	"recovery-unit-deploy/service/common"
	"strconv"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/shirou/gopsutil/v3/cpu" // 注意：推荐使用v3版本
	"github.com/shirou/gopsutil/v3/mem"
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
	TimeStamp int    `json:"timestamp"`
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

// 定义与Linux系统utmp结构体对应的Go结构体
type Utmp struct {
	Type    int16     // 记录类型
	Pad1    [2]byte   // 对齐填充
	Pid     int32     // 登录进程ID
	Line    [32]byte  // 终端设备名（例如tty1, pts/0）
	Id      [4]byte   // 终端名称缩写或inittab ID
	User    [32]byte  // 用户名
	Host    [256]byte // 远程主机名（如果是远程登录）
	Exit    [2]int16  // 进程退出状态（由init设置）
	Session int32     // 会话ID
	Tv      Timeval   // 时间戳
	AddrV6  [4]int32  // IPv6地址（网络字节序）
	Pad2    [20]byte  // 保留字段
}

type Timeval struct {
	Sec  int32 // 秒
	Usec int32 // 微秒
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
	if err != nil {
		// 合并命令执行错误和stderr内容
		resultErr = fmt.Errorf("error is: %v, errMsg is: %s", err, errMsg)
	}

	common.AppLogger.Info(fmt.Sprintf("command output: %s", output))
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

			if device.Mountpoint != nil && *device.Mountpoint == "/" {
				common.DetailPCInfo.SystemDrive = device.Name
			}

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

func setDiskInfo(disks []DiskInfo) {
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
	}
}

func getComputerName() string {
	hostname, err := os.Hostname()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("获取主机名失败: %v", err))
	}
	common.AppLogger.Info(fmt.Sprintf("os.Hostname: %s\n", hostname))

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

	disks := getDiskInfo()
	setDiskInfo(disks)

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

func getLoginInfo() (string, string) {
	var username string = ""
	var lastsignon string = ""

	// 打开utmp文件
	file, err := os.Open("/var/run/utmp")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("无法打开utmp文件: %v", err))
		return "", ""
	}
	defer file.Close()

	var utmp Utmp
	binary.Size(utmp)

	for {
		// 读取一条utmp记录
		err := binary.Read(file, binary.LittleEndian, &utmp)
		if err != nil {
			break // 可能已读到文件末尾
		}

		// 检查是否为用户登录记录（类型为USER_PROCESS，通常值为7）
		if utmp.Type == 7 {
			// 将字节数组转换为字符串，并去除末尾的空字符
			username = string(utmp.User[:])
			for i, c := range username {
				if c == 0 {
					username = username[:i]
					break
				}
			}

			// 转换时间戳
			loginTime := time.Unix(int64(utmp.Tv.Sec), int64(utmp.Tv.Usec)*1000)
			lastsignon = loginTime.Format("2006-01-02 15:04:05")
		}
	}

	return username, lastsignon
}

func getOS() string {
	var os string = ""

	output, err := runCommand("lsb_release", "-a")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("命令执行失败: %s\n", err))
		return os
	}

	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和没有冒号的行
		if line == "" || !strings.Contains(line, ":") {
			continue
		}

		// 按冒号分割键值对
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		result[key] = value
	}

	if description, exists := result["Description"]; exists {
		os = description
	}

	return os
}

func getSystemInfo() SystemInfo {
	systemInfo := SystemInfo{}

	exist := common.PathExists("/sys/firmware/efi")
	if exist {
		systemInfo.BootMode = "UEFI"
	} else {
		systemInfo.BootMode = "Legacy"
	}

	const dmiProductNameFile = "/sys/class/dmi/id/product_name"

	// 读取文件内容
	data, err := os.ReadFile(dmiProductNameFile)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("读取DMI文件失败: %v", err))
	}

	// 去除内容中的换行符和首尾空白字符
	model := strings.TrimSpace(string(data))
	systemInfo.Model = model

	return systemInfo
}

func getMemoryInfo() MemoryInfo {
	var mi MemoryInfo
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("   获取内存信息失败: %v\n", err))
	} else {
		// 将字节转换为MB，1 MB = 1024 * 1024 字节
		totalMB := float64(memInfo.Total) / (1024 * 1024)
		mi = MemoryInfo{
			TotalPhysical: int64(totalMB),
		}
	}

	return mi
}

func getCPUInfo() CPUInfo {
	var cpuInfo CPUInfo
	cpuInfos, err := cpu.Info()
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("   获取CPU信息失败: %v\n", err))
	} else {
		// 通常所有逻辑CPU的核心信息是一致的，取第一个即可
		if len(cpuInfos) > 0 {
			cpuModel := cpuInfos[0].ModelName
			cpuSpeed := cpuInfos[0].Mhz
			cpuInfo = CPUInfo{
				Name:          cpuModel,
				MaxClockSpeed: float32(cpuSpeed) / 1000, // MHz to GHz
			}
		}
	}

	return cpuInfo
}

func (c *Deploy) GetSeedLabel() common.SeedInfo {
	kbcode := getLastKBCode()
	common.AppLogger.Info(fmt.Sprintf("oem info is: %s\n", kbcode))
	// kbcode = "KB5039334"
	seedlable := api.GetSeedLabel(kbcode)
	common.AppLogger.Info(fmt.Sprintf("seedlabel is: %s\n", seedlable))
	if len(strings.TrimSpace(seedlable)) == 0 {
		return common.SeedInfo{}
	}
	return api.GetSeedLabelBySeed(seedlable)
}

func getLastKBCode() string {
	data, err := os.ReadFile("/etc/oem-info")
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("读取文件出错: %v", err))
		return ""
	}

	var info OEMInfo

	err = json.Unmarshal(data, &info)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("解析JSON出错: %v", err))
		return ""
	}

	common.DetailPCInfo.KBCode = info.Basic.OEMCode

	common.AppLogger.Info(fmt.Sprintf("oem-code is: %s\n", common.DetailPCInfo.KBCode))
	return common.DetailPCInfo.KBCode
}

func getCreationTime(filePath string) time.Time {
	t, err := times.Stat(filePath)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("Error getting file times: %v\n", err))
		return time.Now()
	}

	if t.HasBirthTime() {
		// 使用 BirthTime() 方法获取文件的创建时间
		return t.BirthTime()
	} else {
		common.AppLogger.Error("无法获取文件的BirthTime（当前文件系统可能不支持）")
		if t.HasChangeTime() {
			return t.ChangeTime()
		} else {
			common.AppLogger.Error("无法获取文件的ChangeTime（当前文件系统可能不支持）")
		}
	}

	return time.Now()
}

func checkSeedFile() bool {
	filename := fmt.Sprintf("/etc/seedinfo/%s.seedlabel.txt", common.CurrentSeed.SeedLabel)
	fileInfo, err := os.Stat(filename)
	if err != nil {
		common.AppLogger.Error(fmt.Sprintf("seedlabel文件检查失败: %v", err))
		return false
	}

	// 修改时间（跨平台通用）
	modTime := fileInfo.ModTime()

	// 创建时间（按系统处理）
	createTime := getCreationTime(filename)

	strModTime := modTime.Format("2006-01-02 15:04:05")
	strCreateTime := createTime.Format("2006-01-02 15:04:05")

	common.AppLogger.Info(fmt.Sprintf("seedlabel file create time: %s\n", createTime))
	common.AppLogger.Info(fmt.Sprintf("seedlabel file modify time: %s\n", strModTime))

	return api.CheckSeedLabel(common.CurrentSeed.SeedLabel, strModTime, strCreateTime)
}

func getOpSystemInfo() string {

	return "Linux"
}

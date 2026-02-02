//go:build windows
// +build windows

package common

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/registry"
)

func GetRegValue(key registry.Key, path string, name string) string {
	return getRegValue(key, path, name)
}

func getRegValue(key registry.Key, path string, name string) string {
	key, err := registry.OpenKey(key, path, registry.QUERY_VALUE)
	if err != nil {
		AppLogger.Error(err.Error())
		return ""
	}
	defer key.Close()

	value, _, err := key.GetStringValue(name)
	if err != nil {
		AppLogger.Error(err.Error())
		return ""
	}
	return value
}

func setRegValue(key registry.Key, path string, name string, value interface{}) {
	key, err := registry.OpenKey(key, path, registry.QUERY_VALUE)
	if err != nil {
		AppLogger.Error(err.Error())
	}
	defer key.Close()

	switch v := value.(type) {
	case string:
		err = key.SetStringValue(name, v)
	case int:
		err = key.SetDWordValue(name, uint32(v))
	default:
		err = fmt.Errorf("该类型尚未处理：%v", v)
	}
	if err != nil {
		AppLogger.Error(err.Error())
	}
}

func GetSeed() string {
	label := getRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Label")
	version := getRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Version")
	return fmt.Sprintf("%sV%s", label, version)
}

func UpdateLocalReg() {
	setRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Longlabel", CurrentSeed.SeedLabel)
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	setRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Deployed", formattedTime)
	setRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "OAServer", CurrentOA.ServerName)
}

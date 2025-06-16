//go:build windows
// +build windows

package common

import (
	"fmt"

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
	return getRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Longlabel")
}

func UpdateLocalSeed(seed string) {
	setRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Longlabel", seed)
	setRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Label", seed[0:4])
}

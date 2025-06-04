//go:build windows
// +build windows

package deploy

import (
	"recovery-unit-deploy/service/common"

	"golang.org/x/sys/windows/registry"
)

func getRegValue(key registry.Key, path string, name string) string {
	key, err := registry.OpenKey(key, path, registry.QUERY_VALUE)
	if err != nil {
		common.AppLogger.Error(err.Error())
		return ""
	}
	defer key.Close()

	value, _, err := key.GetStringValue(name)
	if err != nil {
		common.AppLogger.Error(err.Error())
		return ""
	}
	return value
}

func getSeed() string {
	return getRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Longlabel")
}

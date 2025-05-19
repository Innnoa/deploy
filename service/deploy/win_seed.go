//go:build windows
// +build windows

package deploy

import (
	"log"

	"golang.org/x/sys/windows/registry"
)

func getRegValue(key registry.Key, path string, name string) string {
	key, err := registry.OpenKey(key, path, registry.QUERY_VALUE)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer key.Close()

	value, _, err := key.GetStringValue(name)
	if err != nil {
		log.Println(err)
		return ""
	}
	return value
}

func getSeed() string {
	return getRegValue(registry.LOCAL_MACHINE, "SOFTWARE\\HKPF\\Seed", "Longlabel")
}

//go:build linux
// +build linux

package deploy

import (
	"os"
	"syscall"
	"time"
)

func getCreationTime(fileInfo os.FileInfo) time.Time {
	stat := fileInfo.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
}

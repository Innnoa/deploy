//go:build windows
// +build windows

package deploy

import (
	"os"
	"syscall"
	"time"
)

func getCreationTime(fileInfo os.FileInfo) time.Time {
	winAttr := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, winAttr.CreationTime.Nanoseconds())
}

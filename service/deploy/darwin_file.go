//go:build darwin
// +build darwin

package deploy

import (
	"os"
	"time"
)

func getCreationTime(fileInfo os.FileInfo) time.Time {
	var time time.Time
	return time
}

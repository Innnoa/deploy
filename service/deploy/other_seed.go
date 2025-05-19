//go:build !windows

package deploy

import "os"

func getSeed() string {
	return os.Getenv("SEEDLONGLABEL")
}

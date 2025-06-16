//go:build !windows

package common

import "os"

func GetSeed() string {
	return os.Getenv("Longlabel")
}

func UpdateLocalSeed(seed string) {
	os.Setenv("Longlabel", seed)
	os.Setenv("Label", seed[0:4])
}

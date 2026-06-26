package common

import (
	"os"
	"testing"
)

func TestGetOSOnCurrentSystem(t *testing.T) {
	osName := GetOS()
	if osName == "" {
		t.Fatal("GetOS() returned empty string")
	}
	t.Logf("current OS: %s", osName)
}

func TestIsKylinPositive(t *testing.T) {
	content := `ID="kylin"
NAME="Kylin Linux Advanced Server"
VERSION="V10 (Lance)"
`

	tmpFile, err := os.CreateTemp("", "os-release-kylin-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	old := readOsRelease
	readOsRelease = func() ([]byte, error) {
		return os.ReadFile(tmpFile.Name())
	}
	defer func() { readOsRelease = old }()

	if !IsKylin() {
		t.Error("IsKylin() should detect Kylin from ID=kylin")
	}
}

func TestIsKylinPositiveUbuntuKylin(t *testing.T) {
	content := `NAME="Ubuntu Kylin"
ID=ubuntu
`

	tmpFile, err := os.CreateTemp("", "os-release-ubuntukylin-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	old := readOsRelease
	readOsRelease = func() ([]byte, error) {
		return os.ReadFile(tmpFile.Name())
	}
	defer func() { readOsRelease = old }()

	if !IsKylin() {
		t.Error("IsKylin() should detect Ubuntu Kylin via case-insensitive kylin match")
	}
}

func TestIsKylinNegative(t *testing.T) {
	content := `ID=arch
NAME="Arch Linux"
`

	tmpFile, err := os.CreateTemp("", "os-release-arch-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	old := readOsRelease
	readOsRelease = func() ([]byte, error) {
		return os.ReadFile(tmpFile.Name())
	}
	defer func() { readOsRelease = old }()

	if IsKylin() {
		t.Error("IsKylin() should NOT detect Arch as Kylin")
	}
}

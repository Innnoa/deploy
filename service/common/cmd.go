package common

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

func DecodeByLocale(output []byte) string {
	var decoder *encoding.Decoder
	lang := getSystemLocale()
	switch lang {
	case "zh-CN": // 简体中文
		decoder = simplifiedchinese.GB18030.NewDecoder()
	case "zh-TW", "zh-HK": // 繁体中文
		decoder = traditionalchinese.Big5.NewDecoder()
	default: // 英语及其他
		return string(output) // 默认UTF-8或Latin1
	}

	decoded, _, _ := transform.Bytes(decoder, output)
	return string(decoded)
}

func getWindowsLocale() (string, error) {
	cmd := exec.Command("powershell", "Get-Culture | Select-Object -ExpandProperty Name")
	out, err := cmd.Output()
	return string(out), err
}

func getMacLocale() (string, error) {
	cmd := exec.Command("defaults", "read", "-g", "AppleLocale")
	out, err := cmd.Output()
	return string(out), err
}

func getLinuxLocale() (string, error) {
	// 方法2：执行命令
	cmd := exec.Command("locale", "charmap")
	out, err := cmd.Output()
	return string(out), err // 如 UTF-8
}

// 获取系统语言（Windows示例）
func getSystemLocale() string {
	var language string
	var err error

	if runtime.GOOS == "windows" {
		language, err = getWindowsLocale()
	} else if runtime.GOOS == "linux" {
		language, err = getLinuxLocale()
	} else if runtime.GOOS == "darwin" {
		language, err = getMacLocale()
	}

	if err != nil {
		language = "zh-HK"
	}

	return strings.TrimSpace(language)
}

func RunScriptWithArgs(scriptPath string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	if len(args) == 6 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1], args[2], args[3], args[4], args[5])
	} else if len(args) == 7 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1], args[2], args[3], args[4], args[5], args[6])
	} else if len(args) == 4 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1], args[2], args[3])
	} else if len(args) == 2 {
		cmd = exec.CommandContext(ctx, "cmd", "/C", scriptPath, args[0], args[1])
	}

	SetHideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := DecodeByLocale(stdout.Bytes())
	errMsg := DecodeByLocale(stderr.Bytes())
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s失败，退出码: %d\n错误输出: %s", scriptPath, exitErr.ExitCode(), errMsg)
		}
		return "", fmt.Errorf("启动%s失败: %v", scriptPath, err)
	}
	return output, nil
}

func RunScript(scriptPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cmd", "/C", scriptPath)
	SetHideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := DecodeByLocale(stdout.Bytes())
	errMsg := DecodeByLocale(stderr.Bytes())
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s失败，退出码: %d\n错误输出: %s", scriptPath, exitErr.ExitCode(), errMsg)
		}
		return "", fmt.Errorf("启动%s失败: %v", scriptPath, err)
	}
	return output, nil
}

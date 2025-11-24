//go:build windows
// +build windows

package common

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

func RunScriptWithArgs(scriptPath string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	args = append([]string{"/C", scriptPath}, args...)

	cmd = exec.CommandContext(ctx, "cmd", args...)

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

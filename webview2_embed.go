//go:build windows

package main

import (
	"embed"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"recovery-unit-deploy/service/common"
)

//go:embed webview2/*
var webview2FS embed.FS

func ensureWebView2Runtime() string {
	minVersion := "94.0.992.31"

	if hasSystemWebView2(minVersion) {
		common.AppLogger.Info("检测到系统已安装 WebView2 (>= " + minVersion + ")，直接使用")
		return ""
	}
	common.AppLogger.Info("未检测到系统 WebView2，将使用内嵌运行时")

	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	cacheDir := filepath.Join(filepath.Dir(exePath), "WebView2Runtime")

	if cached := findWebView2Dir(cacheDir); cached != "" {
		common.AppLogger.Info("使用已缓存的 WebView2 运行时: " + cached)
		return cached
	}

	actualDir := extractEmbeddedCab(cacheDir)
	if actualDir == "" {
		actualDir = extractExternalCab(cacheDir)
	}

	return actualDir
}

func extractEmbeddedCab(cacheDir string) string {
	cabEntry := findCabInEmbed()
	if cabEntry == nil {
		return ""
	}

	common.AppLogger.Info("正在解压内嵌的 WebView2 运行时: " + cabEntry.Name())
	cabData, err := fs.ReadFile(webview2FS, "webview2/"+cabEntry.Name())
	if err != nil {
		common.AppLogger.Error("读取内嵌 CAB 失败: " + err.Error())
		return ""
	}

	return extractCabData(cabData, cacheDir)
}

func extractExternalCab(cacheDir string) string {
	cabPath := findCabNextToExe()
	if cabPath == "" {
		common.AppLogger.Info("未找到 WebView2 CAB 文件，将使用系统检测")
		return ""
	}

	common.AppLogger.Info("正在解压外部 WebView2 运行时: " + cabPath)
	cabData, err := os.ReadFile(cabPath)
	if err != nil {
		common.AppLogger.Error("读取外部 CAB 失败: " + err.Error())
		return ""
	}

	return extractCabData(cabData, cacheDir)
}

func extractCabData(cabData []byte, cacheDir string) string {
	os.MkdirAll(cacheDir, 0755)

	tmpCab := filepath.Join(cacheDir, "runtime.cab")
	if err := os.WriteFile(tmpCab, cabData, 0644); err != nil {
		common.AppLogger.Error("写入临时 CAB 失败: " + err.Error())
		return ""
	}
	defer os.Remove(tmpCab)

	cmd := exec.Command("expand", tmpCab, "-F:*", cacheDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		common.AppLogger.Error("CAB 解压失败: " + string(output))
		return ""
	}

	actualDir := findWebView2Dir(cacheDir)
	if actualDir == "" {
		common.AppLogger.Error("解压后未找到 msedgewebview2.exe")
		return ""
	}

	exec.Command("icacls", actualDir, "/grant", "*S-1-15-2-2:(OI)(CI)(RX)").Run()
	exec.Command("icacls", actualDir, "/grant", "*S-1-15-2-1:(OI)(CI)(RX)").Run()

	common.AppLogger.Info("WebView2 运行时已解压到: " + actualDir)
	return actualDir
}

func findWebView2Dir(root string) string {
	if _, err := os.Stat(filepath.Join(root, "msedgewebview2.exe")); err == nil {
		return root
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(root, entry.Name())
			if _, err := os.Stat(filepath.Join(subDir, "msedgewebview2.exe")); err == nil {
				return subDir
			}
		}
	}
	return ""
}

func findCabInEmbed() fs.DirEntry {
	entries, err := webview2FS.ReadDir("webview2")
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".cab" {
			return entry
		}
	}
	return nil
}

func findCabNextToExe() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exePath)

	candidates := []string{"webview2_runtime.cab"}
	for _, p := range candidates {
		if _, err := os.Stat(filepath.Join(exeDir, p)); err == nil {
			return filepath.Join(exeDir, p)
		}
	}

	entries, _ := os.ReadDir(exeDir)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".cab" {
			return filepath.Join(exeDir, entry.Name())
		}
	}
	return ""
}

func hasSystemWebView2(minVersion string) bool {
	keys := []string{
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
		`HKCU\Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
	}
	for _, key := range keys {
		out, err := exec.Command("reg", "query", key, "/v", "pv").Output()
		if err != nil {
			continue
		}
		version := parseRegVersion(string(out))
		if version != "" && compareVersion(version, minVersion) >= 0 {
			return true
		}
	}
	return false
}

func parseRegVersion(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		for i, p := range parts {
			if p == "REG_SZ" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return ""
}

func compareVersion(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	for i := 0; i < len(ap) || i < len(bp); i++ {
		va, vb := 0, 0
		if i < len(ap) {
			va, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			vb, _ = strconv.Atoi(bp[i])
		}
		if va > vb {
			return 1
		}
		if va < vb {
			return -1
		}
	}
	return 0
}

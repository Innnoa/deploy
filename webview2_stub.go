//go:build !windows

package main

func ensureWebView2Runtime() string {
	return ""
}

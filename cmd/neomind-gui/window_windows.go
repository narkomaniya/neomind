//go:build windows

package main

import (
	"syscall"

	"github.com/webview/webview_go"
)

var user32 = syscall.NewLazyDLL("user32.dll")

// minimizeWindow minimizes the native WebView window (SW_MINIMIZE).
func minimizeWindow(w webview.WebView) {
	hwnd := uintptr(w.Window())
	if hwnd == 0 {
		return
	}
	user32.NewProc("ShowWindow").Call(hwnd, 6)
}

//go:build windows

package main

import (
	"os/exec"
	"syscall"

	"github.com/webview/webview_go"
)

var user32 = syscall.NewLazyDLL("user32.dll")

// hideProcess keeps helper processes (like the OpenCode CLI) from flashing a
// console window on top of the editor.
func hideProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}

// minimizeWindow minimizes the native WebView window (SW_MINIMIZE).
func minimizeWindow(w webview.WebView) {
	hwnd := uintptr(w.Window())
	if hwnd == 0 {
		return
	}
	user32.NewProc("ShowWindow").Call(hwnd, 6)
}

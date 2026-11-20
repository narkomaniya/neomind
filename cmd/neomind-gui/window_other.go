//go:build !windows

package main

import (
	"os/exec"

	"github.com/webview/webview_go"
)

// hideProcess is a no-op outside Windows.
func hideProcess(cmd *exec.Cmd) {}

// minimizeWindow is a no-op outside Windows; the native window manager handles it.
func minimizeWindow(w webview.WebView) {}

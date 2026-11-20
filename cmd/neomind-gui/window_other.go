//go:build !windows

package main

import "github.com/webview/webview_go"

// minimizeWindow is a no-op outside Windows; the native window manager handles it.
func minimizeWindow(w webview.WebView) {}

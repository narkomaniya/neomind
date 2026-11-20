//go:build !windows

package main

import "github.com/webview/webview_go"

// setWindowIcon is a no-op outside Windows: the platform window manager uses
// the application bundle/icon set by the build script.
func setWindowIcon(w webview.WebView, debug bool) {}

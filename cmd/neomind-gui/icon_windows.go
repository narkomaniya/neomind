//go:build windows

package main

import (
	"log"
	"syscall"

	"github.com/webview/webview_go"
)

var (
	modUser32    = syscall.NewLazyDLL("user32.dll")
	modKernel32  = syscall.NewLazyDLL("kernel32.dll")
	procLoadIcon = modUser32.NewProc("LoadIconW")
	procSendMsg  = modUser32.NewProc("SendMessageW")
	procGetMod   = modKernel32.NewProc("GetModuleHandleW")
)

const (
	wmSetIcon = 0x0080
	iconSmall = 0
	iconBig   = 1
	appIconID = 1 // resource ID emitted by windres from assets/neomind.ico
)

// setWindowIcon overrides the generic IDI_APPLICATION icon that WebView assigns
// to its window class with the NeoMind icon embedded in the executable.
func setWindowIcon(w webview.WebView, debug bool) {
	hwnd := uintptr(w.Window())
	if hwnd == 0 {
		if debug {
			log.Println("icon: native window handle is not ready")
		}
		return
	}
	hInst, _, _ := procGetMod.Call(0)
	hIcon, _, _ := procLoadIcon.Call(hInst, uintptr(appIconID))
	if hIcon == 0 {
		if debug {
			log.Println("icon: embedded icon resource not found")
		}
		return
	}
	procSendMsg.Call(hwnd, wmSetIcon, iconSmall, hIcon)
	procSendMsg.Call(hwnd, wmSetIcon, iconBig, hIcon)
	if debug {
		log.Printf("icon: applied hwnd=%x hicon=%x", hwnd, hIcon)
	}
	// Do not DestroyIcon: this handle belongs to the executable resource.
}

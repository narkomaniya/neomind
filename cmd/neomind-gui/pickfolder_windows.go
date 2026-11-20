//go:build windows

package main

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	modShell32       = syscall.NewLazyDLL("shell32.dll")
	modOle32         = syscall.NewLazyDLL("ole32.dll")
	procBrowseFolder = modShell32.NewProc("SHBrowseForFolderW")
	procPathFromPIDL = modShell32.NewProc("SHGetPathFromIDListW")
	procCoTaskFree   = modOle32.NewProc("CoTaskMemFree")
)

type browseInfoW struct {
	HwndOwner      uintptr
	PidlRoot       uintptr
	PszDisplayName *uint16
	LpszTitle      *uint16
	UlFlags        uint32
	Lpfn           uintptr
	LParam         uintptr
	IImage         int32
}

// pickFolder shows the native Windows folder picker. An empty result means
// the user cancelled the dialog.
func pickFolder(dialogTitle string) (string, error) {
	if strings.TrimSpace(dialogTitle) == "" {
		dialogTitle = "Select folder"
	}
	title, err := syscall.UTF16PtrFromString(dialogTitle)
	if err != nil {
		return "", err
	}
	bi := browseInfoW{
		LpszTitle: title,
		UlFlags:   0x51, // RETURNONLYFSDIRS | NEWDIALOGSTYLE | EDITBOX
	}
	pidl, _, _ := procBrowseFolder.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", nil
	}
	defer procCoTaskFree.Call(pidl)

	buf := make([]uint16, 260)
	ok, _, _ := procPathFromPIDL.Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
	if ok == 0 {
		return "", nil
	}
	path := syscall.UTF16ToString(buf)
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	return path, nil
}

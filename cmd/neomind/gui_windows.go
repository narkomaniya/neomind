//go:build windows

package main

import "syscall"

// guiSysProcAttr detaches the windowed GUI from the calling console so opening
// NeoMind never creates or keeps a terminal window around.
func guiSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x00000008} // DETACHED_PROCESS
}

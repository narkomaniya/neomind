//go:build !windows

package main

import "syscall"

// guiSysProcAttr is a no-op on platforms without a separate console window.
func guiSysProcAttr() *syscall.SysProcAttr { return nil }

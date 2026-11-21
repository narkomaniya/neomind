package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Windows autorun module. Uses the per-user Run key, so no admin rights are
// needed and nothing is written outside the user profile.

const startupRunKey = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
const startupValue = "NeoMind"

func startupPath() (string, error) {
	return os.Executable()
}

func startupStatus() (string, error) {
	path, _ := startupPath()
	enabled := false
	if runtime.GOOS == "windows" {
		out, err := exec.Command("reg", "query", startupRunKey, "/v", startupValue).CombinedOutput()
		enabled = err == nil && strings.Contains(strings.ToLower(string(out)), "neomind-gui.exe")
	}
	b, _ := json.Marshal(map[string]interface{}{"enabled": enabled, "path": path, "supported": runtime.GOOS == "windows"})
	return string(b), nil
}

func startupEnable() (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("autostart is only available on Windows")
	}
	path, err := startupPath()
	if err != nil {
		return "", err
	}
	out, err := exec.Command("reg", "add", startupRunKey, "/v", startupValue, "/t", "REG_SZ", "/d", `"`+path+`"`, "/f").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return startupStatus()
}

func startupDisable() (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("autostart is only available on Windows")
	}
	out, err := exec.Command("reg", "delete", startupRunKey, "/v", startupValue, "/f").CombinedOutput()
	if err != nil && !strings.Contains(string(out), "ERROR: The system was unable to find") {
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return startupStatus()
}

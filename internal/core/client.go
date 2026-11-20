// Package core запускает Rust-ядро и разбирает его JSON-ответы.
package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Note — одна запись из хранилища ядра.
type Note struct {
	TS   int64  `json:"ts"`
	Text string `json:"text"`
}

// Response — общий вид ответа neomind-core.
type Response struct {
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	Operation string `json:"operation,omitempty"`
	Version   string `json:"version,omitempty"`
	Core      string `json:"core,omitempty"`
	CLI       string `json:"cli,omitempty"`
	Notes     []Note `json:"notes,omitempty"`
	Count     int    `json:"count"`
	DataDir   string `json:"data_dir"`
}

func coreName() string {
	if runtime.GOOS == "windows" {
		return "neomind-core.exe"
	}
	return "neomind-core"
}

// Path ищет ядро: сначала NEOMIND_CORE, потом рядом с CLI, потом в PATH.
func Path() (string, error) {
	if p := strings.TrimSpace(os.Getenv("NEOMIND_CORE")); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		return "", fmt.Errorf("NEOMIND_CORE points to a missing file: %s", p)
	}
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), coreName())
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath(coreName()); err == nil {
		return p, nil
	}
	return "", errors.New("neomind-core not found (run scripts/build.* or set NEOMIND_CORE)")
}

// Call выполняет одну команду ядра и возвращает разобранный ответ.
func Call(args ...string) (Response, error) {
	var out Response
	core, err := Path()
	if err != nil {
		return out, err
	}
	cmd := exec.Command(core, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return out, fmt.Errorf("core failed: %s", msg)
		}
		return out, fmt.Errorf("core failed: %w", err)
	}
	line := strings.TrimSpace(stdout.String())
	if line == "" {
		return out, errors.New("core returned no answer")
	}
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		return out, fmt.Errorf("bad core answer: %w", err)
	}
	if !out.OK && out.Error != "" {
		return out, errors.New(out.Error)
	}
	return out, nil
}

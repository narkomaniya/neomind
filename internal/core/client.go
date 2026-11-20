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
	"strconv"
	"strings"
)

// Note — одна запись из хранилища ядра.
type Note struct {
	ID      int64  `json:"id"`
	TS      int64  `json:"ts"`
	Created int64  `json:"created"`
	Text    string `json:"text"`
}

// CreatedMillis возвращает время создания в миллисекундах.
// Старые записи ядра могли хранить секунды в поле ts.
func (n Note) CreatedMillis() int64 {
	if n.Created > 0 {
		return n.Created
	}
	if n.ID > 0 {
		if n.ID < 1_000_000_000_000 {
			return n.ID * 1000
		}
		return n.ID
	}
	if n.TS < 1_000_000_000_000 {
		return n.TS * 1000
	}
	return n.TS
}

// Key — стабильный идентификатор заметки.
func (n Note) Key() int64 {
	if n.ID != 0 {
		return n.ID
	}
	return n.TS
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
	Count     int    `json:"count,omitempty"`
	DataDir   string `json:"data_dir,omitempty"`
	ID        int64  `json:"id,omitempty"`
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

// DataDir возвращает папку хранилища (совпадает с логикой Rust-ядра).
func DataDir() string {
	if p := strings.TrimSpace(os.Getenv("NEOMIND_DATA_DIR")); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		if u := os.Getenv("USERPROFILE"); u != "" {
			home = u
		}
	}
	return filepath.Join(home, ".neomind")
}

func Status() (Response, error) { return Call("status") }

func List() ([]Note, error) {
	r, err := Call("note_list")
	return r.Notes, err
}

func Add(text string) (int64, error) {
	r, err := Call("note_add", text)
	return r.ID, err
}

func Update(id int64, text string) error {
	_, err := Call("note_update", strconv.FormatInt(id, 10), text)
	return err
}

func Delete(id int64) error {
	_, err := Call("note_delete", strconv.FormatInt(id, 10))
	return err
}

func Clear() error {
	_, err := Call("note_clear")
	return err
}

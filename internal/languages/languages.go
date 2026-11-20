// Package languages holds the catalogue of languages/toolchains that can be
// installed in one batch via winget, plus a background install manager.
package languages

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Language is one catalogue entry.
type Language struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Size      string `json:"size"`
	Desc      string `json:"desc"`
	Winget    string `json:"winget"`
	Check     string `json:"check"`
	Installed bool   `json:"installed"`
}

// Catalog lists languages with verified winget identifiers.
var Catalog = []Language{
	{ID: "go", Name: "Go", Version: "1.27", Size: "~75 MB", Desc: "CLI, services, tooling", Winget: "GoLang.Go", Check: "go"},
	{ID: "rust", Name: "Rust", Version: "stable", Size: "~150 MB", Desc: "cores, systems code", Winget: "Rustlang.Rustup", Check: "cargo"},
	{ID: "python", Name: "Python", Version: "3.13", Size: "~30 MB", Desc: "scripting, data, AI", Winget: "Python.Python.3.13", Check: "py"},
	{ID: "node", Name: "Node.js", Version: "LTS", Size: "~30 MB", Desc: "JS, frontend, tooling", Winget: "OpenJS.NodeJS", Check: "node"},
	{ID: "java", Name: "Java (JDK)", Version: "21", Size: "~200 MB", Desc: "JVM, Android, enterprise", Winget: "Microsoft.OpenJDK.21", Check: "java"},
	{ID: "cpp", Name: "C / C++ (MinGW)", Version: "WinLibs", Size: "~120 MB", Desc: "compiler and standard library", Winget: "BrechtSanders.WinLibs.POSIX.UCRT", Check: "gcc"},
	{ID: "csharp", Name: ".NET SDK", Version: "9", Size: "~200 MB", Desc: "C#, F#, web and desktop", Winget: "Microsoft.DotNet.SDK.9", Check: "dotnet"},
	{ID: "ruby", Name: "Ruby", Version: "3.x", Size: "~40 MB", Desc: "Ruby, Rails, scripting", Winget: "RubyInstallerTeam.RubyWithDevKit", Check: "ruby"},
	{ID: "php", Name: "PHP", Version: "8.3", Size: "~25 MB", Desc: "web, CMS, backend", Winget: "PHP.PHP.8.3", Check: "php"},
	{ID: "perl", Name: "Perl", Version: "5.x", Size: "~50 MB", Desc: "text processing, automation", Winget: "StrawberryPerl.StrawberryPerl", Check: "perl"},
}

// present checks both PATH and the usual Windows install locations: the GUI is
// often launched from a shortcut before a fresh PATH reaches the process.
func present(cmd string) bool {
	if _, err := exec.LookPath(cmd); err == nil {
		return true
	}
	if runtime.GOOS != "windows" {
		return false
	}
	home := os.Getenv("USERPROFILE")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	local := os.Getenv("LOCALAPPDATA")
	programs := os.Getenv("ProgramFiles")

	var candidates []string
	switch strings.ToLower(cmd) {
	case "go":
		candidates = []string{filepath.Join(programs, "Go", "bin", "go.exe")}
	case "cargo":
		candidates = []string{filepath.Join(home, ".cargo", "bin", "cargo.exe")}
	case "py", "python":
		if local != "" {
			candidates, _ = filepath.Glob(filepath.Join(local, "Programs", "Python", "Python*", "python.exe"))
		}
		candidates = append(candidates, filepath.Join(programs, "Python", "python.exe"))
	case "node":
		candidates = []string{filepath.Join(programs, "nodejs", "node.exe")}
	case "java":
		if programs != "" {
			found, _ := filepath.Glob(filepath.Join(programs, "Eclipse Adoptium", "jdk*", "bin", "java.exe"))
			candidates = append(candidates, found...)
		}
	case "gcc":
		if local != "" {
			found, _ := filepath.Glob(filepath.Join(local, "Microsoft", "WinGet", "Packages", "*mingw64*", "mingw64", "bin", "gcc.exe"))
			candidates = append(candidates, found...)
		}
	case "dotnet":
		candidates = []string{filepath.Join(programs, "dotnet", "dotnet.exe")}
	case "ruby":
		candidates = []string{filepath.Join(programs, "Ruby*", "bin", "ruby.exe")}
	case "php":
		candidates = []string{filepath.Join(programs, "PHP", "php.exe")}
	case "perl":
		candidates = []string{filepath.Join(programs, "Strawberry", "perl", "bin", "perl.exe")}
	}
	for _, p := range candidates {
		if p != "" {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	}
	return false
}

// Snapshot returns the catalogue with up-to-date install flags.
func Snapshot() []Language {
	out := make([]Language, len(Catalog))
	copy(out, Catalog)
	for i := range out {
		out[i].Installed = present(out[i].Check)
		if out[i].ID == "python" && !out[i].Installed {
			out[i].Installed = present("python")
		}
	}
	return out
}

// JSON returns the catalogue as a JSON string for the GUI.
func JSON() string {
	b, _ := json.Marshal(Snapshot())
	return string(b)
}

// Job is the state of a background installation.
type Job struct {
	ID       string   `json:"id"`
	Running  bool     `json:"running"`
	Current  string   `json:"current"`
	Total    int      `json:"total"`
	Done     []string `json:"done"`
	Failed   []string `json:"failed"`
	Messages []string `json:"messages"`
}

// Manager keeps the active install jobs.
type Manager struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

func NewManager() *Manager { return &Manager{jobs: make(map[string]*Job)} }

func (m *Manager) Start(ids []string) string {
	id := fmt.Sprintf("job-%d", time.Now().UnixNano())
	job := &Job{ID: id, Running: true, Total: len(ids), Done: []string{}, Failed: []string{}, Messages: []string{}}
	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			job.Running = false
			job.Current = ""
			m.mu.Unlock()
		}()

		if runtime.GOOS != "windows" {
			for _, langID := range ids {
				name := langID
				for i := range Catalog {
					if Catalog[i].ID == langID {
						name = Catalog[i].Name
						break
					}
				}
				m.addMessage(job, false, name, "winget is only available on Windows")
			}
			return
		}
		for _, langID := range ids {
			var lang *Language
			for i := range Catalog {
				if Catalog[i].ID == langID {
					lang = &Catalog[i]
					break
				}
			}
			if lang == nil {
				continue
			}
			m.mu.Lock()
			job.Current = lang.Name
			m.mu.Unlock()

			cmd := exec.Command("winget", "install", "-e", "--id", lang.Winget,
				"--silent", "--disable-interactivity",
				"--accept-package-agreements", "--accept-source-agreements")
			out, err := cmd.CombinedOutput()
			detail := strings.TrimSpace(string(out))
			if len(detail) > 400 {
				detail = detail[len(detail)-400:]
			}
			if err != nil {
				msg := err.Error()
				if detail != "" {
					msg = detail
				}
				m.addMessage(job, false, lang.Name, msg)
			} else {
				m.addMessage(job, true, lang.Name, "installed")
			}
		}
	}()
	return id
}

func (m *Manager) addMessage(job *Job, ok bool, name, text string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ok {
		job.Done = append(job.Done, name)
	} else {
		job.Failed = append(job.Failed, name)
	}
	job.Messages = append(job.Messages, text)
	if len(job.Messages) > 60 {
		job.Messages = job.Messages[len(job.Messages)-60:]
	}
}

// Progress returns a snapshot of a job.
func (m *Manager) Progress(id string) (Job, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok {
		return Job{}, false
	}
	copyJob := *j
	copyJob.Done = append([]string{}, j.Done...)
	copyJob.Failed = append([]string{}, j.Failed...)
	copyJob.Messages = append([]string{}, j.Messages...)
	return copyJob, true
}

func (j Job) JSON() string {
	b, _ := json.Marshal(j)
	return string(b)
}

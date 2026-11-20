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

// Group is a visual section in the languages catalog.
type Group string

const (
	GroupGeneral       Group = "general"
	GroupWeb           Group = "web"
	GroupGames         Group = "games"
	GroupCybersecurity Group = "cybersecurity"
	GroupData          Group = "data"
	GroupEditor        Group = "editor"
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
	Group     Group  `json:"group"`
	Order     int    `json:"order,omitempty"`
	Installed bool   `json:"installed"`
}

// Catalog lists languages and toolchains with winget identifiers. The flat
// slice is kept so install logic does not change; Group/Order are UI metadata.
var Catalog = []Language{
	// General
	{ID: "go", Name: "Go", Version: "1.27", Size: "~75 MB", Desc: "CLI, services, tooling", Winget: "GoLang.Go", Check: "go", Group: GroupGeneral, Order: 1},
	{ID: "rust", Name: "Rust", Version: "stable", Size: "~150 MB", Desc: "cores, systems code", Winget: "Rustlang.Rustup", Check: "cargo", Group: GroupGeneral, Order: 2},
	{ID: "python", Name: "Python", Version: "3.13", Size: "~30 MB", Desc: "scripting, data, AI", Winget: "Python.Python.3.13", Check: "py", Group: GroupGeneral, Order: 3},
	{ID: "node", Name: "Node.js", Version: "LTS", Size: "~30 MB", Desc: "JS, frontend, tooling", Winget: "OpenJS.NodeJS", Check: "node", Group: GroupGeneral, Order: 4},
	{ID: "java", Name: "Java (JDK)", Version: "21", Size: "~200 MB", Desc: "JVM, Android, enterprise", Winget: "Microsoft.OpenJDK.21", Check: "java", Group: GroupGeneral, Order: 5},
	{ID: "cpp", Name: "C / C++ (MinGW)", Version: "WinLibs", Size: "~120 MB", Desc: "compiler and standard library", Winget: "BrechtSanders.WinLibs.POSIX.UCRT", Check: "gcc", Group: GroupGeneral, Order: 6},
	{ID: "csharp", Name: ".NET SDK", Version: "9", Size: "~200 MB", Desc: "C#, F#, web and desktop", Winget: "Microsoft.DotNet.SDK.9", Check: "dotnet", Group: GroupGeneral, Order: 7},
	{ID: "ruby", Name: "Ruby", Version: "3.x", Size: "~40 MB", Desc: "Ruby, Rails, scripting", Winget: "RubyInstallerTeam.RubyWithDevKit", Check: "ruby", Group: GroupGeneral, Order: 8},
	{ID: "php", Name: "PHP", Version: "8.3", Size: "~25 MB", Desc: "web, CMS, backend", Winget: "PHP.PHP.8.3", Check: "php", Group: GroupGeneral, Order: 9},
	{ID: "perl", Name: "Perl", Version: "5.x", Size: "~50 MB", Desc: "text processing, automation", Winget: "StrawberryPerl.StrawberryPerl", Check: "perl", Group: GroupGeneral, Order: 10},
	{ID: "swift", Name: "Swift", Version: "stable", Size: "~150 MB", Desc: "iOS, macOS, server-side", Winget: "Swift.Swift", Check: "swift", Group: GroupGeneral, Order: 11},
	{ID: "kotlin", Name: "Kotlin", Version: "stable", Size: "~80 MB", Desc: "JVM, Android, multiplatform", Winget: "JetBrains.Kotlin", Check: "kotlinc", Group: GroupGeneral, Order: 12},
	{ID: "dart", Name: "Dart", Version: "stable", Size: "~60 MB", Desc: "Flutter, web, CLI", Winget: "Google.Dart", Check: "dart", Group: GroupGeneral, Order: 13},
	{ID: "r", Name: "R", Version: "4.x", Size: "~100 MB", Desc: "statistics, data analysis", Winget: "RProject.R", Check: "Rscript", Group: GroupGeneral, Order: 14},
	{ID: "julia", Name: "Julia", Version: "1.x", Size: "~120 MB", Desc: "scientific computing", Winget: "JuliaLang.Julia", Check: "julia", Group: GroupGeneral, Order: 15},
	{ID: "zig", Name: "Zig", Version: "stable", Size: "~50 MB", Desc: "systems programming", Winget: "ZigSoftwareFoundation.Zig", Check: "zig", Group: GroupGeneral, Order: 16},
	{ID: "lua", Name: "Lua", Version: "5.x", Size: "~10 MB", Desc: "embedding, scripting, games", Winget: "Lua.Lua", Check: "lua", Group: GroupGeneral, Order: 17},
	{ID: "haskell", Name: "Haskell (GHC)", Version: "9.x", Size: "~250 MB", Desc: "functional programming", Winget: "GHC.GHC", Check: "ghc", Group: GroupGeneral, Order: 18},
	{ID: "scala", Name: "Scala", Version: "3.x", Size: "~100 MB", Desc: "JVM, functional/OOP", Winget: "Scala.Scala", Check: "scala", Group: GroupGeneral, Order: 19},
	{ID: "elixir", Name: "Elixir", Version: "1.x", Size: "~40 MB", Desc: "BEAM, concurrent systems", Winget: "Elixir-lang.Elixir", Check: "elixir", Group: GroupGeneral, Order: 20},
	{ID: "erlang", Name: "Erlang/OTP", Version: "26", Size: "~120 MB", Desc: "distributed systems", Winget: "Erlang.OTP", Check: "erl", Group: GroupGeneral, Order: 21},
	{ID: "clojure", Name: "Clojure", Version: "1.x", Size: "~60 MB", Desc: "functional Lisp on the JVM", Winget: "Clojure.Clojure", Check: "clojure", Group: GroupGeneral, Order: 22},
	{ID: "fortran", Name: "Fortran", Version: "GNU", Size: "~180 MB", Desc: "scientific and numeric code", Winget: "GNU.MinGW", Check: "gfortran", Group: GroupGeneral, Order: 23},
	{ID: "ocaml", Name: "OCaml", Version: "5.x", Size: "~90 MB", Desc: "functional systems language", Winget: "OCaml.OCaml", Check: "ocaml", Group: GroupGeneral, Order: 24},
	{ID: "nim", Name: "Nim", Version: "2.x", Size: "~40 MB", Desc: "compiled, Python-like syntax", Winget: "Nim-lang.Nim", Check: "nim", Group: GroupGeneral, Order: 25},
	{ID: "groovy", Name: "Groovy", Version: "4.x", Size: "~70 MB", Desc: "JVM scripting and build", Winget: "Apache.Groovy", Check: "groovy", Group: GroupGeneral, Order: 26},
	// Web
	{ID: "typescript", Name: "TypeScript", Version: "5.x", Size: "~30 MB", Desc: "typed JavaScript compiler", Winget: "Microsoft.TypeScript.MSBuild", Check: "tsc", Group: GroupWeb, Order: 1},
	{ID: "deno", Name: "Deno", Version: "stable", Size: "~40 MB", Desc: "secure JS/TS runtime", Winget: "DenoLand.Deno", Check: "deno", Group: GroupWeb, Order: 2},
	{ID: "bun", Name: "Bun", Version: "stable", Size: "~40 MB", Desc: "fast JS runtime and bundler", Winget: "Oven-sh.Bun", Check: "bun", Group: GroupWeb, Order: 3},
	{ID: "pnpm", Name: "pnpm", Version: "stable", Size: "~20 MB", Desc: "fast package manager", Winget: "pnpm.pnpm", Check: "pnpm", Group: GroupWeb, Order: 4},
	{ID: "yarn", Name: "Yarn", Version: "stable", Size: "~20 MB", Desc: "JavaScript package manager", Winget: "Yarnpkg.Yarn", Check: "yarn", Group: GroupWeb, Order: 5},
	{ID: "composer", Name: "Composer", Version: "2.x", Size: "~15 MB", Desc: "PHP dependency manager", Winget: "GetComposer.Composer", Check: "composer", Group: GroupWeb, Order: 6},
	{ID: "angular", Name: "Angular CLI", Version: "latest", Size: "~40 MB", Desc: "Angular project tooling", Winget: "Angular.AngularCLI", Check: "ng", Group: GroupWeb, Order: 7},
	{ID: "vuecli", Name: "Vue CLI", Version: "5.x", Size: "~35 MB", Desc: "Vue project scaffolding", Winget: "VueJS.VueCLI", Check: "vue", Group: GroupWeb, Order: 8},
	// Games
	{ID: "godot", Name: "Godot", Version: "4.x", Size: "~60 MB", Desc: "open 2D/3D game engine", Winget: "GodotEngine.GodotEngine", Check: "godot", Group: GroupGames, Order: 1},
	{ID: "unity", Name: "Unity Hub", Version: "latest", Size: "~120 MB", Desc: "Unity editor manager", Winget: "UnityTechnologies.UnityHub", Check: "unity-hub", Group: GroupGames, Order: 2},
	{ID: "unreal", Name: "Epic Games Launcher", Version: "latest", Size: "~200 MB", Desc: "Unreal Engine launcher", Winget: "EpicGames.EpicGamesLauncher", Check: "UnrealEditor", Group: GroupGames, Order: 3},
	{ID: "blender", Name: "Blender", Version: "4.x", Size: "~300 MB", Desc: "3D modeling and rendering", Winget: "BlenderFoundation.Blender", Check: "blender", Group: GroupGames, Order: 4},
	{ID: "gamemaker", Name: "GameMaker", Version: "latest", Size: "~250 MB", Desc: "2D game creation", Winget: "YoYoGames.GameMaker", Check: "gamemaker", Group: GroupGames, Order: 5},
	{ID: "love2d", Name: "LÖVE", Version: "11", Size: "~20 MB", Desc: "Lua game framework", Winget: "LOVE2D.LOVE2D", Check: "love", Group: GroupGames, Order: 6},
	{ID: "renpy", Name: "Ren'Py", Version: "8.x", Size: "~130 MB", Desc: "visual novel engine", Winget: "RenPy.RenPy", Check: "renpy", Group: GroupGames, Order: 7},
	// Cybersecurity
	{ID: "nmap", Name: "Nmap", Version: "7.x", Size: "~40 MB", Desc: "network discovery and security auditing", Winget: "Nmap.Nmap", Check: "nmap", Group: GroupCybersecurity, Order: 1},
	{ID: "wireshark", Name: "Wireshark", Version: "4.x", Size: "~200 MB", Desc: "network protocol analyzer", Winget: "WiresharkFoundation.Wireshark", Check: "tshark", Group: GroupCybersecurity, Order: 2},
	{ID: "burp", Name: "Burp Suite Community", Version: "latest", Size: "~150 MB", Desc: "web application security testing", Winget: "PortSwigger.BurpSuiteCommunity", Check: "burpsuite", Group: GroupCybersecurity, Order: 3},
	{ID: "ghidra", Name: "Ghidra", Version: "11.x", Size: "~300 MB", Desc: "software reverse engineering", Winget: "NationalSecurityAgency.Ghidra", Check: "ghidra", Group: GroupCybersecurity, Order: 4},
	{ID: "metasploit", Name: "Metasploit Framework", Version: "6.x", Size: "~400 MB", Desc: "penetration testing framework", Winget: "Rapid7.Metasploit", Check: "msfconsole", Group: GroupCybersecurity, Order: 5},
	{ID: "radare2", Name: "radare2", Version: "5.x", Size: "~60 MB", Desc: "reverse engineering framework", Winget: "radareorg.Radare2", Check: "r2", Group: GroupCybersecurity, Order: 6},
	{ID: "openvpn", Name: "OpenVPN Connect", Version: "3.x", Size: "~60 MB", Desc: "VPN client", Winget: "OpenVPNTechnologies.OpenVPNConnect", Check: "openvpn", Group: GroupCybersecurity, Order: 7},
	{ID: "hashcat", Name: "hashcat", Version: "6.x", Size: "~30 MB", Desc: "password recovery tool", Winget: "hashcat.hashcat", Check: "hashcat", Group: GroupCybersecurity, Order: 8},
	// Data
	{ID: "sqlite", Name: "SQLite", Version: "3.x", Size: "~10 MB", Desc: "embedded SQL database", Winget: "sqlite.org.SQLite", Check: "sqlite3", Group: GroupData, Order: 1},
	{ID: "postgresql", Name: "PostgreSQL", Version: "16", Size: "~350 MB", Desc: "relational database server", Winget: "PostgreSQL.PostgreSQL", Check: "psql", Group: GroupData, Order: 2},
	{ID: "mysql", Name: "MySQL", Version: "8.x", Size: "~300 MB", Desc: "relational database server", Winget: "Oracle.MySQL", Check: "mysql", Group: GroupData, Order: 3},
	{ID: "mongodb", Name: "MongoDB", Version: "7.x", Size: "~300 MB", Desc: "document database", Winget: "MongoDB.Server", Check: "mongosh", Group: GroupData, Order: 4},
	{ID: "duckdb", Name: "DuckDB", Version: "1.x", Size: "~60 MB", Desc: "in-process analytical SQL", Winget: "DuckDB.DuckDB", Check: "duckdb", Group: GroupData, Order: 5},
	// Editor / tooling
	{ID: "git", Name: "Git", Version: "latest", Size: "~100 MB", Desc: "version control", Winget: "Git.Git", Check: "git", Group: GroupEditor, Order: 1},
	{ID: "githubcli", Name: "GitHub CLI", Version: "latest", Size: "~30 MB", Desc: "GitHub from the command line", Winget: "GitHub.cli", Check: "gh", Group: GroupEditor, Order: 2},
	{ID: "vscode", Name: "Visual Studio Code", Version: "latest", Size: "~120 MB", Desc: "code editor", Winget: "Microsoft.VisualStudioCode", Check: "code", Group: GroupEditor, Order: 3},
	{ID: "neovim", Name: "Neovim", Version: "stable", Size: "~30 MB", Desc: "terminal code editor", Winget: "Neovim.Neovim", Check: "nvim", Group: GroupEditor, Order: 4},
	{ID: "jetbrains", Name: "JetBrains Toolbox", Version: "latest", Size: "~100 MB", Desc: "IDE manager", Winget: "JetBrains.Toolbox", Check: "jetbrains-toolbox", Group: GroupEditor, Order: 5},
	{ID: "cmake", Name: "CMake", Version: "latest", Size: "~60 MB", Desc: "cross-platform build system", Winget: "Kitware.CMake", Check: "cmake", Group: GroupEditor, Order: 6},
	{ID: "ninja", Name: "Ninja", Version: "latest", Size: "~5 MB", Desc: "fast build tool", Winget: "Ninja-build.Ninja", Check: "ninja", Group: GroupEditor, Order: 7},
	{ID: "powershell", Name: "PowerShell 7", Version: "7.x", Size: "~80 MB", Desc: "cross-platform shell", Winget: "Microsoft.PowerShell", Check: "pwsh", Group: GroupEditor, Order: 8},
	{ID: "docker", Name: "Docker Desktop", Version: "latest", Size: "~500 MB", Desc: "containers and images", Winget: "Docker.DockerDesktop", Check: "docker", Group: GroupEditor, Order: 9},
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

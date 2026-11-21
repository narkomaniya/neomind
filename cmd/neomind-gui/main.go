// neomind-gui — the native NeoMind desktop window (WebView2 / WebKit / GTK).
// All logic stays in the Rust core; the GUI only calls it through bindings.
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/narkomaniya/neomind/internal/core"
	"github.com/narkomaniya/neomind/internal/languages"
	"github.com/narkomaniya/neomind/internal/mods"
	"github.com/webview/webview_go"
)

//go:embed ui/index.html
var uiFS embed.FS

var langManager = languages.NewManager()

func bind(w webview.WebView, name string, f interface{}) {
	if err := w.Bind(name, f); err != nil {
		log.Printf("bind %s: %v", name, err)
	}
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// writeDiskFile mirrors a file into a real folder chosen with the native dialog.
// Internal folders are plain names, so they are intentionally skipped here.
func writeDiskFile(folder, name, text string) error {
	if folder == "" || !filepath.IsAbs(folder) {
		return nil
	}
	clean := filepath.Base(name)
	if clean != name || clean == "." || clean == ".." {
		return fmt.Errorf("invalid file name")
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(folder, clean), []byte(text), 0o644)
}

type bootFolder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type bootConfig struct {
	Locale  string       `json:"locale"`
	Folders []bootFolder `json:"folders"`
	Light   bool         `json:"light"`
}

// readBootConfig reads one-shot installer choices: the preset UI language and
// the starter project folders selected in the setup wizard. The locale file is
// consumed so a later in-app language change is never overwritten.
func readBootConfig() string {
	cfg := bootConfig{Folders: []bootFolder{}, Light: os.Getenv("NEOMIND_LIGHT") == "1"}
	dir := core.DataDir()
	if dir != "" {
		localeFile := filepath.Join(dir, "locale")
		if b, err := os.ReadFile(localeFile); err == nil {
			cfg.Locale = strings.TrimSpace(string(b))
			_ = os.Remove(localeFile)
		}
		presetDir := filepath.Join(dir, "presets")
		entries, _ := os.ReadDir(presetDir)
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(presetDir, e.Name()))
			if err != nil {
				continue
			}
			var f bootFolder
			if err := json.Unmarshal(b, &f); err == nil && f.Name != "" && f.Path != "" {
				cfg.Folders = append(cfg.Folders, f)
			}
		}
	}
	out, _ := json.Marshal(cfg)
	return string(out)
}

func bindAll(w webview.WebView) {
	bind(w, "nm_list", func() (string, error) {
		notes, err := core.List()
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(notes)
		return string(b), err
	})

	bind(w, "nm_add", func(text string) (string, error) {
		id, err := core.Add(text)
		return strconv.FormatInt(id, 10), err
	})

	bind(w, "nm_add_file", func(name, folder, text string) (string, error) {
		if err := writeDiskFile(folder, name, text); err != nil {
			return "", err
		}
		id, err := core.AddFile(name, folder, text)
		return strconv.FormatInt(id, 10), err
	})

	bind(w, "nm_update_file", func(id, name, folder, text string) (string, error) {
		if err := writeDiskFile(folder, name, text); err != nil {
			return "", err
		}
		n, err := parseID(id)
		if err != nil {
			return "", err
		}
		return "ok", core.UpdateFile(n, name, folder, text)
	})

	bind(w, "nm_pick_folder", func(title string) (string, error) {
		return pickFolder(title)
	})

	bind(w, "nm_boot", func() (string, error) {
		return readBootConfig(), nil
	})

	bind(w, "nm_minimize", func() (string, error) {
		minimizeWindow(w)
		return "ok", nil
	})

	bind(w, "nm_close", func() (string, error) {
		go func() {
			time.Sleep(60 * time.Millisecond)
			w.Terminate()
		}()
		return "ok", nil
	})

	bind(w, "nm_ai_models", func() (string, error) { return nmAIModels() })
	bind(w, "nm_ai_new", func() (string, error) { return nmAINew() })
	bind(w, "nm_ai_ask", func(text, model string) (string, error) { return nmAIAsk(text, model) })

	bind(w, "nm_discord_setup", func(clientID, asset string) (string, error) {
		if err := discordSave(clientID, asset); err != nil {
			return "", err
		}
		return discordStatus(), nil
	})
	bind(w, "nm_discord_status", func() (string, error) { return discordStatus(), nil })
	bind(w, "nm_discord_activity", func(project, language, file string, started int64) (string, error) {
		discordSetActivity(project, language, file, started)
		return "ok", nil
	})

	bind(w, "nm_startup_status", func() (string, error) { return startupStatus() })
	bind(w, "nm_startup_enable", func() (string, error) { return startupEnable() })
	bind(w, "nm_startup_disable", func() (string, error) { return startupDisable() })

	bind(w, "nm_app_update_check", func() (string, error) { return appUpdateCheck() })
	bind(w, "nm_app_update_apply", func(downloadURL, shaURL, version string) (string, error) {
		return appUpdateApply(downloadURL, shaURL, version)
	})

	bind(w, "nm_github_status", func() (string, error) { return githubStatus(), nil })
	bind(w, "nm_github_device_start", func() (string, error) { return githubDeviceStart() })
	bind(w, "nm_github_device_poll", func() (string, error) { return githubDevicePoll() })
	bind(w, "nm_github_commits", func() (string, error) { return githubCommits() })
	bind(w, "nm_github_signout", func() (string, error) {
		if err := githubSignOut(); err != nil {
			return "", err
		}
		return "ok", nil
	})

	bind(w, "nm_lightning", func(enable bool) (string, error) {
		exe, err := os.Executable()
		if err != nil {
			return "", err
		}
		cmd := exec.Command(exe)
		cmd.Dir = filepath.Dir(exe)
		env := make([]string, 0, len(os.Environ())+1)
		for _, item := range os.Environ() {
			if !strings.HasPrefix(item, "NEOMIND_LIGHT=") {
				env = append(env, item)
			}
		}
		if enable {
			env = append(env, "NEOMIND_LIGHT=1")
		}
		cmd.Env = env
		if err := cmd.Start(); err != nil {
			return "", err
		}
		go func() {
			time.Sleep(120 * time.Millisecond)
			w.Terminate()
		}()
		return "ok", nil
	})

	bind(w, "nm_update", func(id, text string) (string, error) {
		n, err := parseID(id)
		if err != nil {
			return "", err
		}
		return "ok", core.Update(n, text)
	})

	bind(w, "nm_delete", func(id string) (string, error) {
		n, err := parseID(id)
		if err != nil {
			return "", err
		}
		return "ok", core.Delete(n)
	})

	bind(w, "nm_clear", func() (string, error) {
		return "ok", core.Clear()
	})

	bind(w, "nm_status", func() (string, error) {
		r, err := core.Status()
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(r)
		return string(b), err
	})

	bind(w, "nm_open_folder", func() (string, error) {
		dir := core.DataDir()
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("explorer", dir)
		case "darwin":
			cmd = exec.Command("open", dir)
		default:
			cmd = exec.Command("xdg-open", dir)
		}
		if err := cmd.Start(); err != nil {
			return "", err
		}
		return "ok", nil
	})

	bind(w, "nm_languages", func() (string, error) {
		return languages.JSON(), nil
	})

	bind(w, "nm_install_languages", func(idsJSON string) (string, error) {
		var ids []string
		if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil {
			return "", err
		}
		if len(ids) == 0 {
			return "", fmt.Errorf("no languages selected")
		}
		return langManager.Start(ids), nil
	})

	bind(w, "nm_install_progress", func(jobID string) (string, error) {
		job, ok := langManager.Progress(jobID)
		if !ok {
			return "", fmt.Errorf("unknown install job: %s", jobID)
		}
		return job.JSON(), nil
	})

	bind(w, "nm_mods", func() (string, error) {
		return mods.JSON(), nil
	})

	bind(w, "nm_mods_top", func() (string, error) {
		return mods.TopJSON(), nil
	})

	bind(w, "nm_mods_search", func(query string) (string, error) {
		return mods.Search(query)
	})

	bind(w, "nm_mods_open", func(rawURL string) (string, error) {
		if err := mods.Open(rawURL); err != nil {
			return "", err
		}
		return "ok", nil
	})

	bind(w, "nm_set_title", func(title string) (string, error) {
		w.SetTitle(title)
		return "ok", nil
	})
}

func main() {
	if len(os.Args) >= 4 && os.Args[1] == "--neomind-update" {
		runAppUpdater(os.Args[2], os.Args[3], "")
		return
	}
	html, err := uiFS.ReadFile("ui/index.html")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := core.Path(); err != nil {
		fmt.Fprintln(os.Stderr, "NeoMind:", err)
		os.Exit(1)
	}

	// Serve from localhost instead of a data: URL: webview bindings are
	// guaranteed to be available on a normal http origin.
	debug := os.Getenv("NEOMIND_DEBUG") != ""
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, _ *http.Request) {
		if debug {
			log.Println("GUI: page requested")
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = rw.Write(html)
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()

	// Keep the embedded browser lean: no GPU process, extensions or background
	// network work. Existing flags (used for debugging) are left untouched.
	if os.Getenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS") == "" {
		_ = os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS",
			"--disable-gpu --disable-extensions --disable-background-networking --disable-features=Translate,BackForwardCache,MediaRouter --renderer-process-limit=1")
	}

	w := webview.New(false)
	defer w.Destroy()
	setWindowIcon(w, debug)
	w.SetTitle("NeoMind — Notes")
	w.SetSize(1180, 760, webview.HintMin)
	bindAll(w)
	discordLoad()
	githubLoad()
	if os.Getenv("NEOMIND_LIGHT") == "1" {
		discordDisable()
	}
	go discord.run()
	if debug {
		bind(w, "nm_report", func(msg string) (string, error) {
			log.Println("GUI JS:", msg)
			return "ok", nil
		})
	}
	w.Navigate(fmt.Sprintf("http://%s/", listener.Addr().String()))
	w.Run()
}

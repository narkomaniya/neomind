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
	"runtime"
	"strconv"

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
}

func main() {
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

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("NeoMind — Notes")
	w.SetSize(1180, 760, webview.HintMin)
	bindAll(w)
	if debug {
		bind(w, "nm_report", func(msg string) (string, error) {
			log.Println("GUI JS:", msg)
			return "ok", nil
		})
	}
	w.Navigate(fmt.Sprintf("http://%s/", listener.Addr().String()))
	w.Run()
}

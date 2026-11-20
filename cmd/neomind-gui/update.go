package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/narkomaniya/neomind/internal/core"
)

// Silent self-update from the public GitHub release. The updater runs in a
// detached copy of the app so it can survive the editor closing.

const githubReleasesURL = "https://api.github.com/repos/narkomaniya/neomind/releases/latest"

type appUpdateInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	URL       string `json:"url,omitempty"`
	SHA       string `json:"sha,omitempty"`
	Notes     string `json:"notes,omitempty"`
	Current   string `json:"current"`
}

type ghReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type ghRelease struct {
	TagName string           `json:"tag_name"`
	Assets  []ghReleaseAsset `json:"assets"`
	Body    string           `json:"body"`
}

func versionParts(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, _ := strconv.Atoi(p)
		out = append(out, n)
	}
	return out
}

func versionNewer(a, b string) bool {
	pa, pb := versionParts(a), versionParts(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x > y {
			return true
		}
		if x < y {
			return false
		}
	}
	return false
}

func appVersion() string {
	r, err := core.Status()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(r.Version)
}

func appUpdateCheck() (string, error) {
	current := appVersion()
	req, _ := http.NewRequest(http.MethodGet, githubReleasesURL, nil)
	req.Header.Set("User-Agent", "NeoMind")
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("GitHub returned %s", resp.Status)
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	ver := strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")
	info := appUpdateInfo{Version: ver, Current: current}
	if ver == "" || !versionNewer(ver, current) {
		return jsonMarshal(info)
	}
	for _, a := range rel.Assets {
		if a.Name == "NeoMind-Setup-"+ver+".exe" {
			info.URL = a.URL
		}
		if a.Name == "NeoMind-Setup-"+ver+".exe.sha256" {
			info.SHA = a.URL
		}
	}
	if info.URL == "" || info.SHA == "" {
		return jsonMarshal(info)
	}
	info.Available = true
	if len(rel.Body) > 600 {
		rel.Body = rel.Body[:600]
	}
	info.Notes = rel.Body
	return jsonMarshal(info)
}

func jsonMarshal(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func downloadFile(rawURL, path string) error {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("download returned %s", resp.Status)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func appUpdateApply(downloadURL, shaURL, version string) (string, error) {
	if !strings.HasPrefix(downloadURL, "https://github.com/") || !strings.HasPrefix(shaURL, "https://github.com/") {
		return "", fmt.Errorf("invalid update URL")
	}
	dir, err := os.MkdirTemp("", "neomind-update-")
	if err != nil {
		return "", err
	}
	installer := filepath.Join(dir, "NeoMind-Setup-"+version+".exe")
	sumFile := filepath.Join(dir, "NeoMind-Setup-"+version+".exe.sha256")
	if err := downloadFile(downloadURL, installer); err != nil {
		return "", err
	}
	if err := downloadFile(shaURL, sumFile); err != nil {
		return "", err
	}
	expected, err := os.ReadFile(sumFile)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(expected))
	if len(fields) == 0 {
		return "", fmt.Errorf("checksum file is empty")
	}
	f, err := os.Open(installer)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	_ = f.Close()
	if err != nil {
		return "", err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, fields[0]) {
		return "", fmt.Errorf("checksum mismatch")
	}
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(self, "--neomind-update", installer, self, strconv.Itoa(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return "", err
	}
	_ = cmd.Process.Release()
	return `{"ok":true}`, nil
}

func runAppUpdater(installer, app, pid string) {
	// Give the editor a moment to close before replacing its files.
	_ = pid
	time.Sleep(1500 * time.Millisecond)
	cmd := exec.Command(installer, "/VERYSILENT", "/NORESTART", "/CLOSEAPPLICATIONS", "/FORCECLOSEAPPLICATIONS")
	_ = cmd.Run()
	_ = os.Remove(installer)
	next := exec.Command(app)
	_ = next.Start()
	_ = next.Process.Release()
}

func parsePID(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

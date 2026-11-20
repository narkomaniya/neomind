package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/narkomaniya/neomind/internal/core"
)

// GitHub auth uses the device flow: no callback server and no token is typed
// into the editor. The OAuth client below is the public GitHub CLI client;
// GitHub explicitly allows embedding its credentials in version control.

const (
	githubClientID     = "178c6fc778ccc68e1d6a"
	githubClientSecret = "34ddeff2b558a23d38fba8a6de74f086ede1cc0b"
	githubDeviceURL    = "https://github.com/login/device/code"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubAPIURL       = "https://api.github.com"
)

type githubToken struct {
	Token string `json:"token"`
	Login string `json:"login"`
}

type githubClient struct {
	mu         sync.Mutex
	token      string
	login      string
	deviceCode string
	expires    time.Time
	interval   int
}

var github = &githubClient{interval: 5}

type githubDevice struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type githubError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func githubPath() string {
	dir := core.DataDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "github.json")
}

func githubLoad() {
	path := githubPath()
	if path == "" {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var stored githubToken
	if err := json.Unmarshal(b, &stored); err != nil {
		return
	}
	github.mu.Lock()
	github.token = strings.TrimSpace(stored.Token)
	github.login = strings.TrimSpace(stored.Login)
	github.mu.Unlock()
}

func githubSave(token, login string) error {
	path := githubPath()
	if path == "" {
		return fmt.Errorf("no data folder")
	}
	b, err := json.Marshal(githubToken{Token: token, Login: login})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	github.mu.Lock()
	github.token = token
	github.login = login
	github.mu.Unlock()
	return nil
}

func githubStatus() string {
	github.mu.Lock()
	defer github.mu.Unlock()
	return fmt.Sprintf(`{"authorized":%t,"login":%q}`, github.token != "", github.login)
}

func githubSignOut() error {
	github.mu.Lock()
	github.token = ""
	github.login = ""
	github.mu.Unlock()
	path := githubPath()
	if path == "" {
		return nil
	}
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func githubPostForm(rawURL string, values url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var ge githubError
		_ = json.NewDecoder(resp.Body).Decode(&ge)
		if ge.ErrorDescription != "" {
			return nil, fmt.Errorf("%s", ge.ErrorDescription)
		}
		if ge.Error != "" {
			return nil, fmt.Errorf("%s", ge.Error)
		}
		return nil, fmt.Errorf("GitHub returned %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func githubDeviceStart() (string, error) {
	data, err := githubPostForm(githubDeviceURL, url.Values{
		"client_id": {githubClientID},
		"scope":     {"repo,read:user"},
	})
	if err != nil {
		return "", err
	}
	var dev githubDevice
	if err := json.Unmarshal(data, &dev); err != nil {
		return "", err
	}
	if dev.DeviceCode == "" {
		return "", fmt.Errorf("GitHub did not return a device code")
	}
	if dev.Interval <= 0 {
		dev.Interval = 5
	}
	github.mu.Lock()
	github.deviceCode = dev.DeviceCode
	github.interval = dev.Interval
	github.expires = time.Now().Add(time.Duration(dev.ExpiresIn) * time.Second)
	github.mu.Unlock()
	out := map[string]interface{}{
		"user_code":        dev.UserCode,
		"verification_uri": dev.VerificationURI,
		"interval":         dev.Interval,
		"expires_in":       dev.ExpiresIn,
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func githubLogin(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, githubAPIURL+"/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "NeoMind")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("GitHub returned %s", resp.Status)
	}
	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}
	if user.Login == "" {
		return "", fmt.Errorf("GitHub did not return a login")
	}
	return user.Login, nil
}

func githubDevicePoll() (string, error) {
	github.mu.Lock()
	code := github.deviceCode
	expired := time.Now().After(github.expires)
	github.mu.Unlock()
	if code == "" || expired {
		out := map[string]string{"status": "expired"}
		b, _ := json.Marshal(out)
		return string(b), nil
	}
	data, err := githubPostForm(githubTokenURL, url.Values{
		"client_id":     {githubClientID},
		"device_code":   {code},
		"client_secret": {githubClientSecret},
		"grant_type":    {"urn:ietf:params:oauth:grant-type:device_code"},
	})
	if err != nil {
		return "", err
	}
	var reply struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(data, &reply); err != nil {
		return "", err
	}
	if reply.AccessToken != "" {
		login, err := githubLogin(reply.AccessToken)
		if err != nil {
			return "", err
		}
		if err := githubSave(reply.AccessToken, login); err != nil {
			return "", err
		}
		github.mu.Lock()
		github.deviceCode = ""
		github.mu.Unlock()
		out := map[string]string{"status": "ok", "login": login}
		b, _ := json.Marshal(out)
		return string(b), nil
	}
	status := "error"
	switch reply.Error {
	case "authorization_pending", "slow_down":
		status = "pending"
	case "expired_token":
		status = "expired"
	}
	out := map[string]string{"status": status}
	if reply.Error != "" && status == "error" {
		out["message"] = reply.Error
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func githubAPI(token, path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, githubAPIURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "NeoMind")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("GitHub authorization expired or rate limited")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub returned %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func githubCommits() (string, error) {
	github.mu.Lock()
	token, login := github.token, github.login
	github.mu.Unlock()
	if token == "" || login == "" {
		return "", fmt.Errorf("not signed in")
	}
	q := url.QueryEscape("author:" + login)
	data, err := githubAPI(token, "/search/commits?q="+q+"&sort=committer-date&order=desc&per_page=30")
	if err != nil {
		return "", err
	}
	var result struct {
		Items []struct {
			SHA        string `json:"sha"`
			HTMLURL    string `json:"html_url"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Date string `json:"date"`
				} `json:"author"`
			} `json:"commit"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	type commit struct {
		Repo    string `json:"repo"`
		Message string `json:"message"`
		Date    string `json:"date"`
		URL     string `json:"url"`
		SHA     string `json:"sha"`
	}
	items := make([]commit, 0, len(result.Items))
	for _, item := range result.Items {
		msg := strings.SplitN(item.Commit.Message, "\n", 2)[0]
		sha := item.SHA
		if len(sha) > 7 {
			sha = sha[:7]
		}
		items = append(items, commit{
			Repo:    item.Repository.FullName,
			Message: msg,
			Date:    item.Commit.Author.Date,
			URL:     item.HTMLURL,
			SHA:     sha,
		})
	}
	b, err := json.Marshal(items)
	return string(b), err
}

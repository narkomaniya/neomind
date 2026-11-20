package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// The AI panel talks to the local OpenCode service through its own CLI. Using
// `opencode api` keeps the user's config, free models and sign-in intact — the
// GUI never needs to know a port or an auth token.

var (
	aiOnce    sync.Once
	aiExe     string
	aiErr     error
	aiMu      sync.Mutex
	aiSession string
	aiModel   string
)

func opencodeExe() (string, error) {
	aiOnce.Do(func() {
		aiExe, aiErr = exec.LookPath("opencode")
		if aiErr != nil {
			aiErr = fmt.Errorf("opencode is not installed or not in PATH")
		}
	})
	return aiExe, aiErr
}

func opencodeAPI(timeout time.Duration, args ...string) ([]byte, error) {
	exe, err := opencodeExe()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, args...)
	hideProcess(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return out, nil
}

type aiEnvelope struct {
	Data json.RawMessage `json:"data"`
}

type aiSessionData struct {
	ID string `json:"id"`
}

type aiModelInfo struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerID"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Enabled    bool   `json:"enabled"`
	Cost       []struct {
		Input  float64 `json:"input"`
		Output float64 `json:"output"`
	} `json:"cost"`
}

type aiMessage struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Time struct {
		Created   int64 `json:"created"`
		Completed int64 `json:"completed"`
	} `json:"time"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func aiCreateSession() (string, error) {
	out, err := opencodeAPI(60*time.Second, "api", "post", "/api/session", "--data", `{}`)
	if err != nil {
		return "", err
	}
	var env aiEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		return "", fmt.Errorf("could not read OpenCode answer")
	}
	var s aiSessionData
	if err := json.Unmarshal(env.Data, &s); err != nil || s.ID == "" {
		return "", fmt.Errorf("could not create OpenCode session")
	}
	return s.ID, nil
}

func aiEnsureSession() (string, error) {
	if aiSession != "" {
		return aiSession, nil
	}
	id, err := aiCreateSession()
	if err == nil {
		aiSession = id
	}
	return id, err
}

// nm_ai_models lists the free chat models shipped with OpenCode.
func nmAIModels() (string, error) {
	out, err := opencodeAPI(60*time.Second, "api", "get", "/api/model")
	if err != nil {
		return "", err
	}
	var env struct {
		Data []aiModelInfo `json:"data"`
	}
	if err := json.Unmarshal(out, &env); err != nil {
		return "", fmt.Errorf("could not read OpenCode models")
	}
	type pick struct {
		ID         string `json:"id"`
		ProviderID string `json:"providerID"`
		Name       string `json:"name"`
	}
	list := make([]pick, 0, 12)
	for _, m := range env.Data {
		if m.Type != "chat" || m.Enabled == false || m.ID == "" {
			continue
		}
		free := len(m.Cost) == 0
		for _, c := range m.Cost {
			if c.Input == 0 && c.Output == 0 {
				free = true
			}
		}
		if !free {
			continue
		}
		name := m.Name
		if name == "" {
			name = m.ID
		}
		list = append(list, pick{ID: m.ID, ProviderID: m.ProviderID, Name: name})
		if len(list) >= 12 {
			break
		}
	}
	b, err := json.Marshal(list)
	return string(b), err
}

// nm_ai_new starts a clean conversation; the previous one stays in OpenCode.
func nmAINew() (string, error) {
	aiMu.Lock()
	defer aiMu.Unlock()
	aiSession = ""
	aiModel = ""
	if _, err := aiEnsureSession(); err != nil {
		return "", err
	}
	return "ok", nil
}

// nm_ai_ask sends a prompt and waits for the finished assistant answer.
func nmAIAsk(text, model string) (string, error) {
	aiMu.Lock()
	defer aiMu.Unlock()

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("empty prompt")
	}
	id, err := aiEnsureSession()
	if err != nil {
		return "", err
	}

	if model != "" && model != aiModel {
		parts := strings.SplitN(model, ":", 2)
		modelID, provider := parts[0], "opencode"
		if len(parts) == 2 {
			provider = parts[1]
		}
		body, _ := json.Marshal(map[string]interface{}{
			"model": map[string]string{"id": modelID, "providerID": provider},
		})
		if _, err := opencodeAPI(60*time.Second, "api", "post", "/api/session/"+id+"/model", "--data", string(body)); err != nil {
			return "", err
		}
		aiModel = model
	}

	body, _ := json.Marshal(map[string]string{"text": text})
	reply, err := opencodeAPI(180*time.Second, "api", "post", "/api/session/"+id+"/prompt", "--data", string(body))
	if err != nil {
		return "", err
	}
	var sent aiEnvelope
	if err := json.Unmarshal(reply, &sent); err != nil {
		return "", fmt.Errorf("could not read OpenCode answer")
	}
	var user aiMessage
	if err := json.Unmarshal(sent.Data, &user); err != nil {
		return "", fmt.Errorf("could not read OpenCode answer")
	}

	deadline := time.Now().Add(180 * time.Second)
	for time.Now().Before(deadline) {
		out, err := opencodeAPI(30*time.Second, "api", "get", "/api/session/"+id+"/message")
		if err != nil {
			return "", err
		}
		var env struct {
			Data []aiMessage `json:"data"`
		}
		if err := json.Unmarshal(out, &env); err != nil {
			return "", fmt.Errorf("could not read OpenCode messages")
		}
		for i := len(env.Data) - 1; i >= 0; i-- {
			m := env.Data[i]
			if m.Type != "assistant" || m.Time.Created < user.Time.Created {
				continue
			}
			if m.Time.Completed == 0 {
				continue
			}
			var answer strings.Builder
			for _, p := range m.Content {
				if p.Type == "text" {
					answer.WriteString(p.Text)
				}
			}
			if s := strings.TrimSpace(answer.String()); s != "" {
				return s, nil
			}
		}
		time.Sleep(900 * time.Millisecond)
	}
	return "", fmt.Errorf("OpenCode did not answer in time")
}

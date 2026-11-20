// Package mods contains a curated, cosmetic-only catalogue inspired by VS Code
// themes and UI tweaks. Nothing here is installed into NeoMind: entries only
// point to open source repositories and marketplace listings, so third-party
// code can never break the app.
package mods

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Mod is one curated entry.
type Mod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Source      string `json:"source"`
	Marketplace string `json:"marketplace"`
	License     string `json:"license"`
	Safe        bool   `json:"safe"`
}

func avatar(owner string) string {
	return "https://github.com/" + owner + ".png?size=64"
}

func marketplace(name string) string {
	return "https://marketplace.visualstudio.com/search?term=" + url.QueryEscape(name) + "&target=VSCode"
}

// Catalog holds only cosmetic themes/icon themes: no arbitrary extensions,
// no code execution, no changes to NeoMind storage or the Rust core.
var Catalog = []Mod{
	{ID: "dracula", Name: "Dracula Official", Author: "Dracula", Kind: "Theme", Description: "Dark theme with a distinctive purple accent.", Icon: avatar("dracula"), Source: "https://github.com/dracula/vscode", Marketplace: marketplace("Dracula Official"), License: "MIT", Safe: true},
	{ID: "one-dark-pro", Name: "One Dark Pro", Author: "Binaryify", Kind: "Theme", Description: "Atom's iconic One Dark palette for VS Code.", Icon: avatar("Binaryify"), Source: "https://github.com/Binaryify/OneDark-Pro", Marketplace: marketplace("One Dark Pro"), License: "MIT", Safe: true},
	{ID: "github-theme", Name: "GitHub Theme", Author: "GitHub", Kind: "Theme", Description: "Light and dark themes matching github.com.", Icon: avatar("github"), Source: "https://github.com/github/vscode-github-theme", Marketplace: marketplace("GitHub Theme"), License: "see source", Safe: true},
	{ID: "material-theme", Name: "Material Theme", Author: "Equinusocio", Kind: "Theme + icons", Description: "Material Design colours and file icons.", Icon: avatar("material-theme"), Source: "https://github.com/material-theme/vsc-material-theme", Marketplace: marketplace("Material Theme"), License: "see source", Safe: true},
	{ID: "nord", Name: "Nord", Author: "Nord", Kind: "Theme", Description: "Arctic, blue-toned colour palette.", Icon: avatar("nordtheme"), Source: "https://github.com/nordtheme/vscode-nord", Marketplace: marketplace("Nord"), License: "MIT", Safe: true},
	{ID: "tokyo-night", Name: "Tokyo Night", Author: "enkia", Kind: "Theme", Description: "Calm dark theme inspired by Tokyo at night.", Icon: avatar("enkia"), Source: "https://github.com/enkia/tokyo-night-vscode-theme", Marketplace: marketplace("Tokyo Night"), License: "MIT", Safe: true},
	{ID: "night-owl", Name: "Night Owl", Author: "Sarah Drasner", Kind: "Theme", Description: "High-contrast dark theme for long sessions.", Icon: avatar("sdras"), Source: "https://github.com/sdras/night-owl-vscode-theme", Marketplace: marketplace("Night Owl"), License: "MIT", Safe: true},
	{ID: "everforest", Name: "Everforest", Author: "sainnhe", Kind: "Theme", Description: "Forest-inspired, low-fatigue green palette.", Icon: avatar("sainnhe"), Source: "https://github.com/sainnhe/everforest-vscode", Marketplace: marketplace("Everforest"), License: "MIT", Safe: true},
	{ID: "catppuccin", Name: "Catppuccin", Author: "Catppuccin", Kind: "Theme", Description: "Soothing pastel palette with many variants.", Icon: avatar("catppuccin"), Source: "https://github.com/catppuccin/vscode", Marketplace: marketplace("Catppuccin"), License: "MIT", Safe: true},
	{ID: "gruvbox", Name: "Gruvbox", Author: "lewis6991", Kind: "Theme", Description: "Retro warm contrast, popular in terminals.", Icon: avatar("lewis6991"), Source: "https://github.com/lewis6991/vscode-gruvbox", Marketplace: marketplace("Gruvbox"), License: "MIT", Safe: true},
	{ID: "kanagawa", Name: "Kanagawa", Author: "rebin", Kind: "Theme", Description: "Wave and ink inspired dark palette.", Icon: avatar("rebin"), Source: "https://github.com/rebin/VSCodeKanagawaTheme", Marketplace: marketplace("Kanagawa"), License: "MIT", Safe: true},
	{ID: "cobalt2", Name: "Cobalt2", Author: "Wes Bos", Kind: "Theme", Description: "Classic high-contrast cobalt dark theme.", Icon: avatar("wesbos"), Source: "https://github.com/wesbos/cobalt2-vscode", Marketplace: marketplace("Cobalt2"), License: "MIT", Safe: true},
	{ID: "synthwave84", Name: "SynthWave '84", Author: "Robb Owen", Kind: "Theme", Description: "Neon retro theme inspired by the 1980s.", Icon: avatar("robb0wen"), Source: "https://github.com/robb0wen/synthwave-vscode", Marketplace: marketplace("SynthWave '84"), License: "MIT", Safe: true},
	{ID: "quiet-light", Name: "Quiet Light", Author: "Microsoft", Kind: "Built-in theme", Description: "Clean, bright default theme from VS Code.", Icon: avatar("microsoft"), Source: "https://github.com/microsoft/vscode/tree/main/extensions/theme-quietlight", Marketplace: marketplace("Quiet Light"), License: "MIT", Safe: true},
	{ID: "solarized-dark", Name: "Solarized Dark", Author: "Microsoft", Kind: "Built-in theme", Description: "The original Solarized palette, built in.", Icon: avatar("microsoft"), Source: "https://github.com/microsoft/vscode/tree/main/extensions/theme-solarized-dark", Marketplace: marketplace("Solarized Dark"), License: "MIT", Safe: true},
	{ID: "monokai", Name: "Monokai", Author: "Microsoft", Kind: "Built-in theme", Description: "Legendary Monokai syntax colours, built in.", Icon: avatar("microsoft"), Source: "https://github.com/microsoft/vscode/tree/main/extensions/theme-monokai", Marketplace: marketplace("Monokai"), License: "MIT", Safe: true},
	{ID: "red", Name: "Red", Author: "Microsoft", Kind: "Built-in theme", Description: "Minimal built-in theme with a red accent.", Icon: avatar("microsoft"), Source: "https://github.com/microsoft/vscode/tree/main/extensions/theme-red", Marketplace: marketplace("Red"), License: "MIT", Safe: true},
}

// JSON returns the safe catalogue for the GUI.
func JSON() string {
	b, _ := json.Marshal(Catalog)
	return string(b)
}

// Open launches an allow-listed HTTPS URL in the default browser. Only source
// repositories, marketplace pages and GitHub avatars are accepted.
func Open(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return fmt.Errorf("only https links are allowed")
	}
	host := strings.ToLower(u.Hostname())
	allowed := map[string]bool{
		"github.com":                   true,
		"www.github.com":               true,
		"marketplace.visualstudio.com": true,
		"raw.githubusercontent.com":    true,
	}
	if !allowed[host] {
		return fmt.Errorf("link host is not allowed: %s", host)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u.String())
	case "darwin":
		cmd = exec.Command("open", u.String())
	default:
		cmd = exec.Command("xdg-open", u.String())
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = os.Stdout.Sync()
	return nil
}

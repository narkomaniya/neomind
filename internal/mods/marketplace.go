package mods

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const marketplaceAPI = "https://marketplace.visualstudio.com/_apis/public/gallery/extensionquery"

// apiExtension mirrors the part of the Marketplace response we display.
type apiExtension struct {
	Publisher struct {
		PublisherName string `json:"publisherName"`
		DisplayName   string `json:"displayName"`
	} `json:"publisher"`
	ExtensionName    string `json:"extensionName"`
	DisplayName      string `json:"displayName"`
	ShortDescription string `json:"shortDescription"`
	Versions         []struct {
		Files []struct {
			AssetType string `json:"assetType"`
			Source    string `json:"source"`
		} `json:"files"`
		Properties []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"properties"`
	} `json:"versions"`
	Statistics []struct {
		StatisticName string  `json:"statisticName"`
		Value         float64 `json:"value"`
	} `json:"statistics"`
}

func (e apiExtension) toMod() MarketplaceMod {
	id := e.Publisher.PublisherName + "." + e.ExtensionName
	name := e.DisplayName
	if strings.TrimSpace(name) == "" {
		name = e.ExtensionName
	}
	publisher := e.Publisher.DisplayName
	if strings.TrimSpace(publisher) == "" {
		publisher = e.Publisher.PublisherName
	}

	description := e.ShortDescription
	if len(description) > 180 {
		description = description[:177] + "..."
	}

	var icon, source string
	executes := false
	if len(e.Versions) > 0 {
		v := e.Versions[0]
		for _, f := range v.Files {
			if f.AssetType == "Microsoft.VisualStudio.Services.Icons.Default" {
				icon = f.Source
				break
			}
		}
		for _, p := range v.Properties {
			switch p.Key {
			case "Microsoft.VisualStudio.Services.Links.Source",
				"Microsoft.VisualStudio.Services.Links.GitHub",
				"Microsoft.VisualStudio.Services.Links.Learn":
				if source == "" && strings.HasPrefix(p.Value, "https://") {
					source = p.Value
				}
			case "Microsoft.VisualStudio.Code.ExecutesCode":
				executes = strings.EqualFold(p.Value, "true")
			}
		}
	}

	var installs, ratingCount int64
	var rating float64
	for _, s := range e.Statistics {
		switch s.StatisticName {
		case "install":
			installs = int64(s.Value)
		case "weightedRating":
			rating = s.Value
		case "ratingcount":
			ratingCount = int64(s.Value)
		}
	}

	return MarketplaceMod{
		ID:           id,
		Name:         name,
		Publisher:    publisher,
		Description:  description,
		Icon:         icon,
		Source:       source,
		Marketplace:  "https://marketplace.visualstudio.com/items?itemName=" + id,
		Installs:     installs,
		Rating:       rating,
		RatingCount:  ratingCount,
		ExecutesCode: executes,
	}
}

// TopJSON returns the bundled top-100 catalog for the GUI.
func TopJSON() string {
	b, _ := json.Marshal(TopCatalog)
	return string(b)
}

// Search queries the live VS Code Marketplace. An empty query returns the
// bundled top-100. Search is read-only: no VSIX is downloaded or installed.
func Search(query string) (string, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return TopJSON(), nil
	}
	if len(q) > 100 {
		return "", fmt.Errorf("search query is too long")
	}

	criteria := []map[string]string{
		{"filterType": "8", "value": "Microsoft.VisualStudio.Code"},
		{"filterType": "10", "value": q},
	}
	body, err := json.Marshal(map[string]any{
		"filters": []map[string]any{{
			"criteria":   criteria,
			"pageNumber": 1,
			"pageSize":   50,
			"sortBy":     4,
			"sortOrder":  0,
		}},
		"flags": 918,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, marketplaceAPI, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json;api-version=7.1-preview.1")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NeoMind/0.4.0")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("marketplace returned %s", resp.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", err
	}

	var payload struct {
		Results []struct {
			Extensions []apiExtension `json:"extensions"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}

	found := make([]MarketplaceMod, 0, 50)
	for _, group := range payload.Results {
		for _, ext := range group.Extensions {
			found = append(found, ext.toMod())
		}
	}
	out, err := json.Marshal(found)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

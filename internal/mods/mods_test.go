package mods

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTopCatalogIsCompleteAndSafe(t *testing.T) {
	if len(TopCatalog) != 100 {
		t.Fatalf("expected 100 extensions, got %d", len(TopCatalog))
	}
	ids := make(map[string]bool)
	for _, m := range TopCatalog {
		if ids[m.ID] {
			t.Fatalf("duplicate id: %s", m.ID)
		}
		ids[m.ID] = true
		if m.Name == "" || m.Publisher == "" {
			t.Fatalf("%s has an empty name or publisher", m.ID)
		}
		if !strings.HasPrefix(m.Marketplace, "https://marketplace.visualstudio.com/items?itemName=") {
			t.Fatalf("%s has an invalid marketplace link", m.ID)
		}
		if m.Source != "" && !strings.HasPrefix(m.Source, "https://github.com/") {
			t.Fatalf("%s has a source link outside the allowlist: %s", m.ID, m.Source)
		}
		if m.Icon != "" && !strings.HasPrefix(m.Icon, "https://") {
			t.Fatalf("%s has an invalid icon link", m.ID)
		}
	}
}

func TestTopJSON(t *testing.T) {
	if !strings.Contains(TopJSON(), `"ms-python.python"`) {
		t.Fatal("top JSON does not contain the most installed extension")
	}
}

func TestSearchEmptyReturnsTop(t *testing.T) {
	got, err := Search("")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"ms-python.python"`) {
		t.Fatal("empty search did not return the top catalog")
	}
}

func TestSearchLive(t *testing.T) {
	got, err := Search("rust")
	if err != nil {
		t.Skipf("marketplace is unavailable: %v", err)
	}
	var found []MarketplaceMod
	if err := json.Unmarshal([]byte(got), &found); err != nil {
		t.Fatal(err)
	}
	if len(found) == 0 {
		t.Fatal("live search returned no results")
	}
	for _, m := range found {
		if !strings.HasPrefix(m.Marketplace, "https://marketplace.visualstudio.com/items?itemName=") {
			t.Fatalf("%s has an invalid marketplace link: %s", m.ID, m.Marketplace)
		}
	}
}

func TestCatalogIsSafeAndWellFormed(t *testing.T) {
	if len(Catalog) == 0 {
		t.Fatal("empty catalogue")
	}
	ids := make(map[string]bool)
	for _, m := range Catalog {
		if !m.Safe {
			t.Fatalf("%s is not marked safe", m.ID)
		}
		if ids[m.ID] {
			t.Fatalf("duplicate id: %s", m.ID)
		}
		ids[m.ID] = true
		if m.Source == "" || m.Marketplace == "" || m.Icon == "" {
			t.Fatalf("%s has an empty link/icon", m.ID)
		}
	}
}

func TestOpenRejectsUnknownHosts(t *testing.T) {
	if err := Open("https://example.com/x"); err == nil {
		t.Fatal("expected unknown host to be rejected")
	}
	if err := Open("http://github.com/x"); err == nil {
		t.Fatal("expected non-https link to be rejected")
	}
}

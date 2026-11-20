package mods

import "testing"

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

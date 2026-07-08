package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"llm-router/internal/model"
)

func writeProfileTo(path string, p model.Profile) error {
	b, _ := json.MarshalIndent(p, "", "  ")
	return os.WriteFile(path, b, 0o600)
}

func readProfileFrom(path string) (model.Profile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return model.Profile{}, err
	}
	var p model.Profile
	err = json.Unmarshal(b, &p)
	return p, err
}

func TestProfileRoundTripAndExport(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "p1.json")
	prof := model.Profile{Name: "p1", Targets: []model.Target{{
		CLI: "claude", ProviderID: "anthropic", BaseURL: "https://x", APIKey: "secret",
	}}}
	if err := writeProfileTo(p, prof); err != nil {
		t.Fatal(err)
	}
	got, err := readProfileFrom(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "p1" || got.Targets[0].BaseURL != "https://x" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	// Export はシークレットを除外
	exp := filepath.Join(dir, "export.json")
	if err := Export(exp, []model.Profile{prof}); err != nil {
		t.Fatal(err)
	}
	imported, err := Import(exp)
	if err != nil {
		t.Fatal(err)
	}
	if imported[0].Targets[0].APIKey != "" {
		t.Fatalf("secret not stripped in export: %q", imported[0].Targets[0].APIKey)
	}
	if imported[0].Targets[0].BaseURL != "https://x" {
		t.Fatalf("base url lost in export: %q", imported[0].Targets[0].BaseURL)
	}
}

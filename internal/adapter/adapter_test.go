package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCodexReadWriteRoundTrip(t *testing.T) {
	p := writeTemp(t, "config.toml", `
model = "gpt-5.5"
openai_base_url = "https://example.com/v1"

[model_providers.deepseek]
base_url = "https://ds.example.com"
env_key = "DEEPSEEK_KEY"
`)
	// ConfigDir を一時ディレクトリに差し替えるため、filepath を直接扱うアダプタではなく
	// ヘルパ経由で検証する。ここでは load/save TOML のラウンドトリップを確認。
	m, err := loadTOML(p)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := deepGet(m, "openai_base_url"); !ok || tomlString(v) != "https://example.com/v1" {
		t.Fatalf("openai_base_url mismatch: %v", v)
	}
	if v, ok := deepGet(m, "model_providers.deepseek.base_url"); !ok || tomlString(v) != "https://ds.example.com" {
		t.Fatalf("deepseek base_url mismatch: %v", v)
	}
	deepSet(m, "openai_base_url", "https://changed.example.com")
	if err := saveTOML(p, m); err != nil {
		t.Fatal(err)
	}
	// 他キーが保持されていること
	m2, _ := loadTOML(p)
	if v, ok := deepGet(m2, "model_providers.deepseek.base_url"); !ok || tomlString(v) != "https://ds.example.com" {
		t.Fatalf("deepseek lost after round-trip: %v", v)
	}
}

func TestJSONCStripper(t *testing.T) {
	in := []byte(`{
  // comment
  "model": "anthropic/claude-sonnet-4-5",
  "provider": {
    "anthropic": { "options": { "baseURL": "https://x", "apiKey": "sk", } }, // tail comma
  },
}`)
	out := stripJSONC(in)
	m := map[string]interface{}{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("parse after strip failed: %v", err)
	}
	if m["model"] != "anthropic/claude-sonnet-4-5" {
		t.Fatalf("model lost: %v", m["model"])
	}
	prov := m["provider"].(map[string]interface{})
	anth := prov["anthropic"].(map[string]interface{})
	opts := anth["options"].(map[string]interface{})
	if opts["baseURL"] != "https://x" || opts["apiKey"] != "sk" {
		t.Fatalf("provider options lost: %v", opts)
	}
}

func TestOpenCodeRead(t *testing.T) {
	p := writeTemp(t, "opencode.jsonc", `{
  "model": "anthropic/claude-sonnet-4-5",
  "provider": {
    "openai": { "options": { "baseURL": "https://o.example.com" } },
    "anthropic": { "options": { "baseURL": "https://a.example.com", "apiKey": "sk-ant" } }
  }
}`)
	m, err := loadJSONC(p)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := deepGet(m, "provider.openai.options.baseURL"); !ok || tomlString(v) != "https://o.example.com" {
		t.Fatalf("openai baseURL mismatch: %v", v)
	}
}

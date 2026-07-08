package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"

	"llm-router/internal/model"
	"llm-router/internal/platform"
)

// ClaudeAdapter は Claude Code を扱う。base_url / api_key は環境変数参照。
type ClaudeAdapter struct{}

func NewClaude() *ClaudeAdapter { return &ClaudeAdapter{} }

func (a *ClaudeAdapter) ID() model.CLIID        { return model.CLIClaude }
func (a *ClaudeAdapter) Name() string           { return "Claude Code" }
func (a *ClaudeAdapter) ConfigDir() string      { return platform.CLIDirs()["claude"] }
func (a *ClaudeAdapter) ConfigPath() string     { return filepath.Join(a.ConfigDir(), "settings.json") }

func (a *ClaudeAdapter) Read() ([]model.Target, error) {
	t := model.Target{
		CLI:        model.CLIClaude,
		ProviderID: "anthropic",
		Provider:   "anthropic",
		Enabled:    true,
	}
	if m, ok := readJSONField(filepath.Join(a.ConfigDir(), "settings.json"), "model"); ok {
		t.Model = m
		t.ReadSource = model.SourceRef{Kind: model.SourceConfig, Path: "settings.json"}
	}
	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" {
		t.BaseURL = v
		t.ReadSource = model.SourceRef{Kind: model.SourceEnv, Path: "ANTHROPIC_BASE_URL"}
	}
	if v := firstEnv("ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "CLAUDE_API_KEY"); v != "" {
		t.APIKey = v
		t.ReadSource = model.SourceRef{Kind: model.SourceEnv, Path: "ANTHROPIC_API_KEY"}
	}
	return []model.Target{t}, nil
}

// WriteConfig: Claude Code は base_url/api_key を設定ファイルに持たないため、
// model のみを settings.json へ書き込む（BaseURL/APIKey は env 系反映先へ）。
func (a *ClaudeAdapter) WriteConfig(t model.Target) error {
	if t.Model == "" {
		return nil
	}
	return writeJSONField(filepath.Join(a.ConfigDir(), "settings.json"), "model", t.Model)
}

func (a *ClaudeAdapter) EnvNames(t model.Target) (string, string) {
	return "ANTHROPIC_BASE_URL", "ANTHROPIC_API_KEY"
}

// ---- JSON ヘルパ（設定ファイルの特定フィールドのみ更新・他キー保持） ----

func readJSONField(path, key string) (string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", false
	}
	if v, ok := m[key].(string); ok {
		return v, true
	}
	return "", false
}

func writeJSONField(path, key, val string) error {
	m := map[string]interface{}{}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &m)
	}
	m[key] = val
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}

package adapter

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"llm-router/internal/model"
	"llm-router/internal/platform"
)

// CodexAdapter は Codex (OpenAI) を扱う。
// 組込 openai プロバイダは openai_base_url / OPENAI_API_KEY。
// カスタムは model_providers.<id>.{base_url, env_key}。
type CodexAdapter struct{}

func NewCodex() *CodexAdapter { return &CodexAdapter{} }

func (a *CodexAdapter) ID() model.CLIID     { return model.CLICodex }
func (a *CodexAdapter) Name() string        { return "Codex" }
func (a *CodexAdapter) ConfigDir() string   { return platform.CLIDirs()["codex"] }
func (a *CodexAdapter) ConfigPath() string  { return filepath.Join(a.ConfigDir(), "config.toml") }

func (a *CodexAdapter) Read() ([]model.Target, error) {
	cfgPath := filepath.Join(a.ConfigDir(), "config.toml")
	m, _ := loadTOML(cfgPath)

	modelName := ""
	if v, ok := deepGet(m, "model"); ok {
		modelName = tomlString(v)
	}
	targets := []model.Target{}

	// 組込 openai
	openai := model.Target{
		CLI:        model.CLICodex,
		ProviderID: "openai",
		Provider:   "openai",
		Enabled:    true,
		Model:      modelName,
	}
	openai.BaseURL = ""
	if v, ok := deepGet(m, "openai_base_url"); ok {
		openai.BaseURL = tomlString(v)
	}
	openai.APIKey = os.Getenv("OPENAI_API_KEY")
	if openai.BaseURL != "" {
		openai.ReadSource = model.SourceRef{Kind: model.SourceConfig, Path: cfgPath}
	} else if openai.APIKey != "" {
		openai.ReadSource = model.SourceRef{Kind: model.SourceEnv, Path: "OPENAI_API_KEY"}
	}
	targets = append(targets, openai)

	// カスタム model_providers
	if v, ok := deepGet(m, "model_providers"); ok {
		pm, ok := v.(map[string]interface{})
		if !ok {
			pm = map[string]interface{}{}
		}
		for id, raw := range pm {
			pm, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			ct := model.Target{
				CLI:        model.CLICodex,
				ProviderID: id,
				Provider:   id,
				Enabled:    true,
				Model:      modelName,
			}
			ct.BaseURL = tomlString(pm["base_url"])
			envKey := tomlString(pm["env_key"])
			if envKey == "" {
				envKey = "OPENAI_API_KEY"
			}
			ct.APIKey = os.Getenv(envKey)
			ct.ReadSource = model.SourceRef{Kind: model.SourceConfig, Path: cfgPath}
			targets = append(targets, ct)
		}
	}
	return targets, nil
}

func (a *CodexAdapter) WriteConfig(t model.Target) error {
	cfgPath := filepath.Join(a.ConfigDir(), "config.toml")
	m, _ := loadTOML(cfgPath)
	if t.ProviderID == "openai" {
		if t.BaseURL != "" {
			deepSet(m, "openai_base_url", t.BaseURL)
		}
	} else {
		if t.BaseURL != "" {
			deepSet(m, "model_providers."+t.ProviderID+".base_url", t.BaseURL)
		}
		deepSet(m, "model_providers."+t.ProviderID+".env_key", "OPENAI_API_KEY")
	}
	if t.Model != "" {
		deepSet(m, "model", t.Model)
	}
	return saveTOML(cfgPath, m)
}

func (a *CodexAdapter) EnvNames(t model.Target) (string, string) {
	// base_url は設定ファイルベースのため env 名は空。
	return "", "OPENAI_API_KEY"
}

// ---- TOML ヘルパ（他キー保持のため map 経由で読み書き） ----

func loadTOML(path string) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := toml.Unmarshal(b, &m); err != nil {
		return m, err
	}
	return m, nil
}

func saveTOML(path string, m map[string]interface{}) error {
	out, err := toml.Marshal(m)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}

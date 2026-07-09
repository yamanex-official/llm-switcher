package adapter

import (
	"path/filepath"

	"llm-switcher/internal/model"
	"llm-switcher/internal/platform"
)

// OpenCodeAdapter は OpenCode を扱う。設定は opencode.jsonc。
// provider.<id>.options.baseURL / apiKey を使用。
type OpenCodeAdapter struct{}

func NewOpenCode() *OpenCodeAdapter { return &OpenCodeAdapter{} }

func (a *OpenCodeAdapter) ID() model.CLIID    { return model.CLIOpenCode }
func (a *OpenCodeAdapter) Name() string       { return "OpenCode" }
func (a *OpenCodeAdapter) ConfigDir() string  { return platform.CLIDirs()["opencode"] }
func (a *OpenCodeAdapter) ConfigPath() string { return filepath.Join(a.ConfigDir(), "opencode.jsonc") }

func (a *OpenCodeAdapter) Read() ([]model.Target, error) {
	cfgPath := filepath.Join(a.ConfigDir(), "opencode.jsonc")
	m, _ := loadJSONC(cfgPath)
	modelName := ""
	if v, ok := deepGet(m, "model"); ok {
		if s, ok := v.(string); ok {
			modelName = s
		}
	}

	targets := []model.Target{}
	if v, ok := deepGet(m, "provider"); ok {
		pm, _ := v.(map[string]interface{})
		for id, raw := range pm {
		prov, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		opts, _ := prov["options"].(map[string]interface{})
		t := model.Target{
			CLI:        model.CLIOpenCode,
			ProviderID: id,
			Provider:   id,
			Enabled:    true,
			Model:      modelName,
		}
		if opts != nil {
			if v, ok := opts["baseURL"].(string); ok {
				t.BaseURL = v
			}
			if v, ok := opts["apiKey"].(string); ok {
				t.APIKey = v
			}
		}
		t.ReadSource = model.SourceRef{Kind: model.SourceConfig, Path: cfgPath}
		targets = append(targets, t)
	}
	}
	// provider が一つもない場合は openai を仮表示
	if len(targets) == 0 {
		targets = append(targets, model.Target{
			CLI: model.CLIOpenCode, ProviderID: "openai", Provider: "openai",
			Enabled: true, Model: modelName,
			ReadSource: model.SourceRef{Kind: model.SourceDefault, Path: cfgPath},
		})
	}
	return targets, nil
}

func (a *OpenCodeAdapter) WriteConfig(t model.Target) error {
	cfgPath := filepath.Join(a.ConfigDir(), "opencode.jsonc")
	m, _ := loadJSONC(cfgPath)
	deepSet(m, "model", t.Model)
	deepSet(m, "provider."+t.ProviderID+".options.baseURL", t.BaseURL)
	deepSet(m, "provider."+t.ProviderID+".options.apiKey", t.APIKey)
	return saveJSONC(cfgPath, m)
}

func (a *OpenCodeAdapter) EnvNames(t model.Target) (string, string) {
	switch t.ProviderID {
	case "openai":
		return "OPENAI_BASE_URL", "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_BASE_URL", "ANTHROPIC_API_KEY"
	case "gemini":
		return "", "GEMINI_API_KEY"
	default:
		return "", "OPENAI_API_KEY"
	}
}

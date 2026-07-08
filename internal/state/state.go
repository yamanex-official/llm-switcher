package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"llm-router/internal/model"
	"llm-router/internal/platform"
)

func dir() (string, error) {
	base, err := platform.AppConfigDir()
	if err != nil {
		return "", err
	}
	return base, os.MkdirAll(base, 0o700)
}

// AppConfigDirSafe はアプリ設定ディレクトリを返す（呼び出し側用）。
func AppConfigDirSafe() (string, error) {
	return dir()
}

func profilePath(name string) (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	safe := strings.ReplaceAll(name, string(filepath.Separator), "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, "..", "__")
	cleanPath := filepath.Clean(filepath.Join(d, "profile-"+safe+".json"))
	if !strings.HasPrefix(cleanPath, filepath.Clean(d)+string(filepath.Separator)) {
		return "", fmt.Errorf("profile name escapes config dir")
	}
	return cleanPath, nil
}

func SaveProfile(p model.Profile) error {
	p.UpdatedAt = time.Now()
	path, err := profilePath(p.Name)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func LoadProfile(name string) (model.Profile, error) {
	var p model.Profile
	path, err := profilePath(name)
	if err != nil {
		return p, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(b, &p)
	return p, err
}

func ListProfiles() ([]string, error) {
	d, err := dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "profile-") && strings.HasSuffix(n, ".json") {
			names = append(names, strings.TrimSuffix(strings.TrimPrefix(n, "profile-"), ".json"))
		}
	}
	return names, nil
}

// Export は全プロファイルを1ファイルに出力（シークレットは除外）。
func Export(path string, profiles []model.Profile) error {
	safe := make([]model.Profile, 0, len(profiles))
	for _, p := range profiles {
		cp := p
		for i := range cp.Targets {
			cp.Targets[i].APIKey = ""
		}
		safe = append(safe, cp)
	}
	b, err := json.MarshalIndent(struct {
		Profiles []model.Profile `json:"profiles"`
	}{Profiles: safe}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func Import(path string) ([]model.Profile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var in struct {
		Profiles []model.Profile `json:"profiles"`
	}
	if err := json.Unmarshal(b, &in); err != nil {
		return nil, err
	}
	return in.Profiles, nil
}

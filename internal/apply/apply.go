package apply

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result は 1 反映先の適用結果。
type Result struct {
	Kind    string
	Target  string
	Changed bool
	Error   error
}

// Options は適用対象の選択。
type Options struct {
	ConfigFile  bool
	EnvFile     bool
	Profile     bool
	OSEnv       bool
	ConfigPaths []string // 設定ファイルのパス（トランザクション用スナップショット対象）
}

const (
	markerStart = "# >>> llm-switcher >>>"
	markerEnd   = "# <<< llm-switcher <<<"
)

// EnvVar は書き込む環境変数。
type EnvVar struct {
	Name  string
	Value string
}

// Apply は選択された反映先へ一括適用する（トランザクション）。
func Apply(opts Options, configWrite func() error, envVars []EnvVar, homeDir string) []Result {
	results := []Result{}

	type backup struct {
		path string
		data []byte
	}
	backups := []backup{}

	snapshot := func(path string) error {
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				backups = append(backups, backup{path: path, data: nil})
				return nil
			}
			return err
		}
		backups = append(backups, backup{path: path, data: b})
		return nil
	}

	rollback := func() {
		for _, bk := range backups {
			if bk.data == nil {
				_ = os.Remove(bk.path)
			} else {
				_ = os.WriteFile(bk.path, bk.data, 0o600)
			}
		}
	}

	// 1. 設定ファイルのスナップショット（書込前に取得）
	if opts.ConfigFile {
		for _, p := range opts.ConfigPaths {
			if err := snapshot(p); err != nil {
				results = append(results, Result{Kind: "config", Error: fmt.Errorf("snapshot: %w", err)})
				return results
			}
		}
	}

	if opts.ConfigFile && configWrite != nil {
		if err := configWrite(); err != nil {
			results = append(results, Result{Kind: "config", Target: "CLI 設定ファイル", Changed: false, Error: err})
			rollback()
			return results
		}
		results = append(results, Result{Kind: "config", Target: "CLI 設定ファイル", Changed: true})
	}

	if opts.EnvFile {
		p := filepath.Join(homeDir, ".env")
		if err := snapshot(p); err != nil {
			results = append(results, Result{Kind: "envfile", Error: err})
			rollback()
			return results
		}
		if err := writeEnvFile(p, envVars); err != nil {
			results = append(results, Result{Kind: "envfile", Error: err})
			rollback()
			return results
		}
		results = append(results, Result{Kind: "envfile", Target: p, Changed: true})
	}

	if opts.Profile {
		info := detectShell(homeDir)
		if info.ProfilePath == "" {
			results = append(results, Result{
				Kind:    "profile",
				Target:  string(info.Type),
				Error:   fmt.Errorf("このシェル (%s) はプロファイルに対応していません。代わりに OS 環境変数の反映を使用してください", info.Type),
			})
			rollback()
			return results
		}
		p := info.ProfilePath
		if err := snapshot(p); err != nil {
			results = append(results, Result{Kind: "profile", Error: err})
			rollback()
			return results
		}
		if err := writeProfileForShell(p, envVars, info.Type); err != nil {
			results = append(results, Result{Kind: "profile", Error: err})
			rollback()
			return results
		}
		results = append(results, Result{Kind: "profile", Target: p, Changed: true})
	}

	if opts.OSEnv {
		if err := writeOSEnv(envVars); err != nil {
			results = append(results, Result{Kind: "osenv", Error: err})
			rollback()
			return results
		}
		results = append(results, Result{Kind: "osenv", Target: "OS 環境変数", Changed: true})
	}

	return results
}

func validateNoNewline(val string) error {
	if strings.ContainsAny(val, "\r\n") {
		return fmt.Errorf("value contains newline characters")
	}
	return nil
}

func writeEnvFile(path string, vars []EnvVar) error {
	for _, v := range vars {
		if err := validateNoNewline(v.Value); err != nil {
			return fmt.Errorf("env var %s: %w", v.Name, err)
		}
	}
	existing := map[string]string{}
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if idx := strings.Index(line, "="); idx > 0 {
				existing[line[:idx]] = line[idx+1:]
			}
		}
	}
	merged := map[string]string{}
	for k, v := range existing {
		merged[k] = v
	}
	for _, v := range vars {
		merged[v.Name] = v.Value
	}
	var sb strings.Builder
	for k, v := range merged {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(sb.String()), 0o600)
}

func writeProfileLines(path string, vars []EnvVar, formatLine func(name, value string) string) error {
	for _, v := range vars {
		if err := validateNoNewline(v.Value); err != nil {
			return fmt.Errorf("profile var %s: %w", v.Name, err)
		}
	}
	lines := []string{}
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		inBlock := false
		for sc.Scan() {
			line := sc.Text()
			if strings.TrimSpace(line) == markerStart {
				inBlock = true
				continue
			}
			if strings.TrimSpace(line) == markerEnd {
				inBlock = false
				continue
			}
			if !inBlock {
				lines = append(lines, line)
			}
		}
	}
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l + "\n")
	}
	sb.WriteString(markerStart + "\n")
	for _, v := range vars {
		sb.WriteString(formatLine(v.Name, v.Value) + "\n")
	}
	sb.WriteString(markerEnd + "\n")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(sb.String()), 0o600)
}

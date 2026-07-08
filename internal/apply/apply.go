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
	Kind    string // "config" | "envfile" | "profile" | "osenv"
	Target  string // 説明
	Changed bool
	Error   error
}

// Options は適用対象の選択。
type Options struct {
	ConfigFile bool
	EnvFile    bool
	Profile    bool
	OSEnv      bool
}

// Marker はシェルプロファイル管理ブロックの開始/終了マーカー。
const (
	markerStart = "# >>> llm-router >>>"
	markerEnd   = "# <<< llm-router <<<"
)

// EnvVar は書き込む環境変数。
type EnvVar struct {
	Name  string
	Value string
}

// Apply は選択された反映先へ一括適用する（トランザクション）。
// configWrite: 設定ファイル書き込み（adapter が行う）。nil ならスキップ。
// envVars: .env / シェルプロファイル / OS環境変数へ書き出す変数群。
// homeDir: .env 等のベースディレクトリ。
func Apply(opts Options, configWrite func() error, envVars []EnvVar, homeDir string) []Result {
	results := []Result{}

	type backup struct {
		path string
		data []byte
	}
	backups := []backup{}

	snapshot := func(path string) error {
		if _, err := os.Stat(path); err != nil {
			backups = append(backups, backup{path: path, data: nil})
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
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
				_ = os.WriteFile(bk.path, bk.data, 0o644)
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
		p, err := detectProfilePath(homeDir)
		if err != nil {
			results = append(results, Result{Kind: "profile", Error: err})
			rollback()
			return results
		}
		if err := snapshot(p); err != nil {
			results = append(results, Result{Kind: "profile", Error: err})
			rollback()
			return results
		}
		if err := writeProfile(p, envVars); err != nil {
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

// writeEnvFile は .env の該当行を更新（他行は保持）。
func writeEnvFile(path string, vars []EnvVar) error {
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
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

// writeProfile はマーカーブロックのみを管理（ブロック外は非破壊）。
func writeProfile(path string, vars []EnvVar) error {
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
		sb.WriteString(fmt.Sprintf("export %s=%s\n", v.Name, v.Value))
	}
	sb.WriteString(markerEnd + "\n")
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

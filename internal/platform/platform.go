package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// AppConfigDir は llm-switcher 自身の設定ディレクトリを返す。
// 各 OS で os.UserConfigDir() が適切な場所を返す:
//   - Windows: %APPDATA%\llm-switcher
//   - macOS:   ~/Library/Application Support/llm-switcher
//   - Linux:   ~/.config/llm-switcher
func AppConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "llm-switcher"), nil
}

// CLIDirs は対象 CLI のユーザー単位設定ディレクトリを返す。
func CLIDirs() map[string]string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		prof := os.Getenv("USERPROFILE")
		if prof == "" {
			prof = home
		}
		return map[string]string{
			"claude":   filepath.Join(prof, ".claude"),
			"codex":    filepath.Join(prof, ".codex"),
			"opencode": filepath.Join(prof, ".config", "opencode"),
		}
	}
	return map[string]string{
		"claude":   filepath.Join(home, ".claude"),
		"codex":    filepath.Join(home, ".codex"),
		"opencode": filepath.Join(home, ".config", "opencode"),
	}
}

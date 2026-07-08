package apply

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func detectProfilePath(homeDir string) (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
	}
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return filepath.Join(homeDir, ".zshrc"), nil
	}
	return filepath.Join(homeDir, ".bashrc"), nil
}

// writeOSEnv は OS 環境変数を設定する。
// Windows: setx（ユーザースコープ）。Unix: 真の OS スコープがないため ~/.profile へマーカー書き込み。
func writeOSEnv(vars []EnvVar) error {
	if runtime.GOOS == "windows" {
		for _, v := range vars {
			cmd := exec.Command("setx", v.Name, v.Value)
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("setx %s failed: %w (%s)", v.Name, err, string(out))
			}
		}
		return nil
	}
	p := filepath.Join(os.Getenv("HOME"), ".profile")
	return writeProfile(p, vars)
}

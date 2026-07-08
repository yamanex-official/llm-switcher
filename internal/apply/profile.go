package apply

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type ShellType string

const (
	ShellBash       ShellType = "bash"
	ShellZsh        ShellType = "zsh"
	ShellFish       ShellType = "fish"
	ShellCsh        ShellType = "csh"
	ShellTcsh       ShellType = "tcsh"
	ShellPowerShell ShellType = "powershell"
	ShellCmd        ShellType = "cmd"
)

type ShellInfo struct {
	Type        ShellType
	ProfilePath string // 空文字列の場合はプロファイル非対応（cmd.exe 等）
	OSEnvPath   string // Unix で OS 環境変数書き込み先ファイル
}

func detectShell(homeDir string) ShellInfo {
	if runtime.GOOS == "windows" {
		return detectWindowsShell(homeDir)
	}
	return detectUnixShell(homeDir)
}

func detectWindowsShell(homeDir string) ShellInfo {
	if isWSL() {
		return detectUnixShell(homeDir)
	}
	comspec := os.Getenv("COMSPEC")
	shell := os.Getenv("SHELL")
	if strings.Contains(strings.ToLower(comspec), "cmd") &&
		!strings.Contains(strings.ToLower(shell), "powershell") &&
		!strings.Contains(strings.ToLower(shell), "pwsh") {
		return ShellInfo{Type: ShellCmd, ProfilePath: "", OSEnvPath: ""}
	}
	return ShellInfo{
		Type:        ShellPowerShell,
		ProfilePath: filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
	}
}

func isWSL() bool {
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		return true
	}
	if v := os.Getenv("WSL_DISTRO_NAME"); v != "" {
		return true
	}
	return false
}

func detectUnixShell(homeDir string) ShellInfo {
	shell := os.Getenv("SHELL")
	name := filepath.Base(shell)
	switch {
	case strings.Contains(name, "fish"):
		return ShellInfo{
			Type:        ShellFish,
			ProfilePath: filepath.Join(homeDir, ".config", "fish", "config.fish"),
			OSEnvPath:   filepath.Join(homeDir, ".config", "fish", "config.fish"),
		}
	case strings.Contains(name, "tcsh"):
		return ShellInfo{
			Type:        ShellTcsh,
			ProfilePath: filepath.Join(homeDir, ".tcshrc"),
			OSEnvPath:   filepath.Join(homeDir, ".login"),
		}
	case strings.Contains(name, "csh"):
		return ShellInfo{
			Type:        ShellCsh,
			ProfilePath: filepath.Join(homeDir, ".cshrc"),
			OSEnvPath:   filepath.Join(homeDir, ".login"),
		}
	case strings.Contains(name, "zsh"):
		return ShellInfo{
			Type:        ShellZsh,
			ProfilePath: filepath.Join(homeDir, ".zshrc"),
			OSEnvPath:   filepath.Join(homeDir, ".profile"),
		}
	default:
		return ShellInfo{
			Type:        ShellBash,
			ProfilePath: filepath.Join(homeDir, ".bashrc"),
			OSEnvPath:   filepath.Join(homeDir, ".profile"),
		}
	}
}

func detectProfilePath(homeDir string) (string, error) {
	info := detectShell(homeDir)
	if info.ProfilePath == "" {
		return "", fmt.Errorf("プロファイル非対応のシェルです (shell=%s)。代わりに OS 環境変数の反映を使用してください", info.Type)
	}
	return info.ProfilePath, nil
}

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
	home := os.Getenv("HOME")
	info := detectUnixShell(home)
	p := info.OSEnvPath
	if p == "" {
		p = filepath.Join(home, ".profile")
	}
	return writeProfileForShell(p, vars, info.Type)
}

func writeProfileForShell(path string, vars []EnvVar, shell ShellType) error {
	var prefix string
	switch shell {
	case ShellFish:
		return writeFishExport(path, vars)
	case ShellCsh, ShellTcsh:
		return writeCshExport(path, vars)
	case ShellPowerShell:
		return writePwshExport(path, vars)
	default:
		prefix = "export "
	}
	return writeProfileLines(path, vars, func(name, value string) string {
		return fmt.Sprintf("%s%s=%s", prefix, name, value)
	})
}

func writeFishExport(path string, vars []EnvVar) error {
	return writeProfileLines(path, vars, func(name, value string) string {
		return fmt.Sprintf("set -gx %s %s", name, value)
	})
}

func writeCshExport(path string, vars []EnvVar) error {
	return writeProfileLines(path, vars, func(name, value string) string {
		return fmt.Sprintf("setenv %s %s", name, value)
	})
}

func writePwshExport(path string, vars []EnvVar) error {
	return writeProfileLines(path, vars, func(name, value string) string {
		return fmt.Sprintf("$env:%s = \"%s\"", name, value)
	})
}

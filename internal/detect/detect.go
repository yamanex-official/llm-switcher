package detect

import (
	"os/exec"

	"llm-router/internal/model"
	"llm-router/internal/platform"
)

// targets は検出対象 CLI とその実行ファイル名。
var targets = []struct {
	ID  model.CLIID
	Name string
	Exe  string
}{
	{model.CLIClaude, "Claude Code", "claude"},
	{model.CLICodex, "Codex", "codex"},
	{model.CLIOpenCode, "OpenCode", "opencode"},
}

// Detect は各 CLI のインストール有無と設定ディレクトリを検出する。
func Detect() []model.DetectedCLI {
	dirs := platform.CLIDirs()
	out := make([]model.DetectedCLI, 0, len(targets))
	for _, t := range targets {
		d := model.DetectedCLI{ID: t.ID, Name: t.Name}
		if p, err := exec.LookPath(t.Exe); err == nil {
			d.Installed = true
			d.Path = p
			d.Note = "設定: " + dirs[string(t.ID)]
		} else {
			d.Note = "未検出"
		}
		out = append(out, d)
	}
	return out
}

package main

import (
	"fmt"

	"llm-switcher/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := tui.New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		// 非 TTY 環境（CI/パイプ）では静的描画して終了（起動確認用）。
		fmt.Println(tui.RenderStatic())
	}
}

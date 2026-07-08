package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"llm-router/internal/adapter"
	"llm-router/internal/apply"
	"llm-router/internal/detect"
	"llm-router/internal/model"
	"llm-router/internal/state"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	scrDashboard screen = iota
	scrEdit
	scrApply
	scrResult
	scrProfiles
)

type appModel struct {
	adapters   []adapter.Adapter
	targets    []model.Target
	installed  map[model.CLIID]bool
	state      screen
	cursor     int
	selIdx     int
	editing    model.Target
	editField  int
	inputs     []textinput.Model
	opts       apply.Options
	results    []apply.Result
	msg        string
	homeDir    string
	// profiles
	profiles    []string
	profCursor  int
	nameInput   textinput.Model
	nameFocused bool
}

func New() appModel {
	home, _ := os.UserHomeDir()
	m := appModel{homeDir: home}
	m.load()
	m.refreshProfiles()
	m.nameInput = textinput.New()
	m.nameInput.Placeholder = "新規プロファイル名"
	m.nameInput.Width = 30
	return m
}

func (m *appModel) load() {
	m.adapters = adapter.All()
	m.targets = []model.Target{}
	for _, a := range m.adapters {
		ts, err := a.Read()
		if err == nil {
			m.targets = append(m.targets, ts...)
		}
	}
	m.installed = map[model.CLIID]bool{}
	for _, d := range detect.Detect() {
		m.installed[d.ID] = d.Installed
	}
}

func (m *appModel) refreshProfiles() {
	names, _ := state.ListProfiles()
	m.profiles = names
}

func (m appModel) Init() tea.Cmd { return nil }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case scrDashboard:
		return m.updateDashboard(msg)
	case scrEdit:
		return m.updateEdit(msg)
	case scrApply:
		return m.updateApply(msg)
	case scrResult:
		if key, ok := msg.(tea.KeyMsg); ok {
			if key.String() == "q" || key.String() == "esc" || key.String() == "enter" {
				m.state = scrDashboard
				m.msg = ""
				return m, nil
			}
		}
	case scrProfiles:
		return m.updateProfiles(msg)
	}
	return m, nil
}

func (m appModel) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.targets)-1 {
			m.cursor++
		}
	case "p":
		m.state = scrProfiles
		m.profCursor = 0
	case "enter", "e":
		if len(m.targets) == 0 {
			return m, nil
		}
		m.selIdx = m.cursor
		m.editing = m.targets[m.cursor]
		m.startEdit()
		m.state = scrEdit
	}
	return m, nil
}

func (m *appModel) startEdit() {
	fields := []string{m.editing.BaseURL, m.editing.APIKey, m.editing.Model}
	m.inputs = make([]textinput.Model, 0, 3)
	for _, v := range fields {
		ti := textinput.New()
		ti.SetValue(v)
		ti.Width = 50
		m.inputs = append(m.inputs, ti)
	}
	m.inputs[0].Focus()
	m.editField = 0
}

func (m appModel) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = scrDashboard
		return m, nil
	case "tab", "down":
		m.inputs[m.editField].Blur()
		m.editField = (m.editField + 1) % len(m.inputs)
		m.inputs[m.editField].Focus()
		return m, nil
	case "shift+tab", "up":
		m.inputs[m.editField].Blur()
		m.editField = (m.editField + len(m.inputs) - 1) % len(m.inputs)
		m.inputs[m.editField].Focus()
		return m, nil
	case "enter":
		m.commitEdit()
		m.opts = apply.Options{ConfigFile: true}
		m.state = scrApply
		return m, nil
	}
	var cmd tea.Cmd
	m.inputs[m.editField], cmd = m.inputs[m.editField].Update(msg)
	return m, cmd
}

func (m *appModel) commitEdit() {
	m.editing.BaseURL = m.inputs[0].Value()
	m.editing.APIKey = m.inputs[1].Value()
	m.editing.Model = m.inputs[2].Value()
}

func (m appModel) updateApply(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = scrEdit
		return m, nil
	case "up", "k":
		m.cursor = (m.cursor + 3) % 4
	case "down", "j":
		m.cursor = (m.cursor + 1) % 4
	case " ":
		switch m.cursor {
		case 0:
			m.opts.ConfigFile = !m.opts.ConfigFile
		case 1:
			m.opts.EnvFile = !m.opts.EnvFile
		case 2:
			m.opts.Profile = !m.opts.Profile
		case 3:
			m.opts.OSEnv = !m.opts.OSEnv
		}
	case "enter", "a":
		m.doApply()
		m.state = scrResult
	}
	return m, nil
}

func (m *appModel) doApply() {
	ad := findAdapter(m.adapters, m.editing.CLI)
	var configWrite func() error
	if ad != nil {
		t := m.editing
		configWrite = func() error { return ad.WriteConfig(t) }
	}
	baseEnv, keyEnv := "", ""
	if ad != nil {
		baseEnv, keyEnv = ad.EnvNames(m.editing)
	}
	vars := []apply.EnvVar{}
	if baseEnv != "" && m.editing.BaseURL != "" {
		vars = append(vars, apply.EnvVar{Name: baseEnv, Value: m.editing.BaseURL})
	}
	if keyEnv != "" && m.editing.APIKey != "" {
		vars = append(vars, apply.EnvVar{Name: keyEnv, Value: m.editing.APIKey})
	}
	m.results = apply.Apply(m.opts, configWrite, vars, m.homeDir)
	m.load()
}

func (m *appModel) updateProfiles(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = scrDashboard
		m.nameFocused = false
		m.nameInput.Blur()
		return m, nil
	case "tab":
		m.nameFocused = !m.nameFocused
		if m.nameFocused {
			m.nameInput.Focus()
		} else {
			m.nameInput.Blur()
		}
		return m, nil
	case "s":
		name := strings.TrimSpace(m.nameInput.Value())
		if name != "" {
			_ = state.SaveProfile(model.Profile{Name: name, Targets: m.targets})
			m.refreshProfiles()
			m.nameInput.SetValue("")
		}
		return m, nil
	case "e":
		m.exportAll()
		return m, nil
	case "i":
		m.importAll()
		m.refreshProfiles()
		return m, nil
	case "enter":
		if m.profCursor < len(m.profiles) {
			m.loadProfileApply(m.profiles[m.profCursor])
			m.state = scrDashboard
		}
		return m, nil
	case "up", "k":
		if m.profCursor > 0 {
			m.profCursor--
		}
	case "down", "j":
		if m.profCursor < len(m.profiles)-1 {
			m.profCursor++
		}
	}
	if m.nameFocused {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *appModel) loadProfileApply(name string) {
	p, err := state.LoadProfile(name)
	if err != nil {
		m.msg = "load failed: " + err.Error()
		return
	}
	for _, t := range p.Targets {
		ad := findAdapter(m.adapters, t.CLI)
		if ad == nil {
			continue
		}
		_ = ad.WriteConfig(t)
		baseEnv, keyEnv := ad.EnvNames(t)
		vars := []apply.EnvVar{}
		if baseEnv != "" && t.BaseURL != "" {
			vars = append(vars, apply.EnvVar{Name: baseEnv, Value: t.BaseURL})
		}
		if keyEnv != "" && t.APIKey != "" {
			vars = append(vars, apply.EnvVar{Name: keyEnv, Value: t.APIKey})
		}
		_ = apply.Apply(apply.Options{EnvFile: true}, nil, vars, m.homeDir)
	}
	m.load()
	m.msg = "profile applied: " + name
}

func (m *appModel) exportAll() {
	names, _ := state.ListProfiles()
	var ps []model.Profile
	for _, n := range names {
		if p, err := state.LoadProfile(n); err == nil {
			ps = append(ps, p)
		}
	}
	dir, _ := state.AppConfigDirSafe()
	_ = state.Export(filepath.Join(dir, "export.json"), ps)
}

func (m *appModel) importAll() {
	dir, _ := state.AppConfigDirSafe()
	ps, err := state.Import(filepath.Join(dir, "export.json"))
	if err != nil {
		m.msg = "import failed: " + err.Error()
		return
	}
	for _, p := range ps {
		_ = state.SaveProfile(p)
	}
}

func findAdapter(as []adapter.Adapter, id model.CLIID) adapter.Adapter {
	for _, a := range as {
		if a.ID() == id {
			return a
		}
	}
	return nil
}

// ---- View ----

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	noStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	curStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
)

func maskKey(s string) string {
	if s == "" {
		return noStyle.Render("(空)")
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

func (m appModel) View() string {
	switch m.state {
	case scrDashboard:
		return m.viewDashboard()
	case scrEdit:
		return m.viewEdit()
	case scrApply:
		return m.viewApply()
	case scrResult:
		return m.viewResult()
	case scrProfiles:
		return m.viewProfiles()
	}
	return ""
}

func (m appModel) viewDashboard() string {
	s := titleStyle.Render("LLMルーター — ダッシュボード") + "\n\n"
	if len(m.targets) == 0 {
		s += noStyle.Render("接続先が見つかりません\n")
	}
	for i, t := range m.targets {
		cursor := "  "
		if i == m.cursor {
			cursor = curStyle.Render("▸ ")
		}
		status := noStyle.Render("未設定")
		if t.BaseURL != "" || t.APIKey != "" {
			status = okStyle.Render("設定済")
		}
		install := ""
		if !m.installed[t.CLI] {
			install = warnStyle.Render(" (未インストール)")
		}
		line := fmt.Sprintf("%s%s/%s  [%s]  %s%s\n", cursor, t.CLI, t.ProviderID, status, t.BaseURL, install)
		s += line
	}
	s += "\n" + noStyle.Render("(↑/↓ 選択  Enter 編集  p プロファイル  q 終了)")
	if m.msg != "" {
		s += "\n" + warnStyle.Render(m.msg)
	}
	return s
}

func (m appModel) viewEdit() string {
	labels := []string{"Base URL", "API Key", "Model"}
	s := titleStyle.Render(fmt.Sprintf("編集: %s/%s", m.editing.CLI, m.editing.ProviderID)) + "\n\n"
	for i, ti := range m.inputs {
		mark := "  "
		if i == m.editField {
			mark = curStyle.Render("▶ ")
		}
		s += fmt.Sprintf("%s%s:\n%s\n\n", mark, labels[i], ti.View())
	}
	s += noStyle.Render("(Tab 移動  Enter 次へ  Esc 戻る)")
	return s
}

func (m appModel) viewApply() string {
	opts := []struct {
		on    bool
		label string
	}{
		{m.opts.ConfigFile, "CLI 設定ファイル"},
		{m.opts.EnvFile, ".env ファイル"},
		{m.opts.Profile, "シェルプロファイル"},
		{m.opts.OSEnv, "OS 環境変数"},
	}
	s := titleStyle.Render("反映先を選択") + "\n\n"
	for i, o := range opts {
		cursor := "  "
		if i == m.cursor {
			cursor = curStyle.Render("▸ ")
		}
		box := "[ ]"
		if o.on {
			box = "[x]"
		}
		s += fmt.Sprintf("%s%s %s\n", cursor, box, o.label)
	}
	s += "\n" + noStyle.Render("(Space 切替  Enter 適用  Esc 戻る)")
	return s
}

func (m appModel) viewResult() string {
	s := titleStyle.Render("反映結果") + "\n\n"
	if len(m.results) == 0 {
		s += noStyle.Render("反映先が選択されていません\n")
	}
	for _, r := range m.results {
		stateS := okStyle.Render("OK")
		if r.Error != nil {
			stateS = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("ERR")
		}
		line := fmt.Sprintf("[%s] %s", stateS, r.Kind)
		if r.Target != "" {
			line += " (" + r.Target + ")"
		}
		if r.Error != nil {
			line += ": " + r.Error.Error()
		}
		s += line + "\n"
	}
	s += "\n" + noStyle.Render("(Enter/Esc ダッシュボードへ)")
	return s
}

func (m appModel) viewProfiles() string {
	s := titleStyle.Render("プロファイル") + "\n\n"
	s += fmt.Sprintf("%s 新規名: %s\n\n", curStyle.Render("▶"), m.nameInput.View())
	s += noStyle.Render("(Tab 名前入力  s 保存)\n\n")
	if len(m.profiles) == 0 {
		s += noStyle.Render("保存済プロファイルなし\n")
	}
	for i, n := range m.profiles {
		cursor := "  "
		if i == m.profCursor {
			cursor = curStyle.Render("▸ ")
		}
		s += fmt.Sprintf("%s%s\n", cursor, n)
	}
	s += "\n" + noStyle.Render("(↑/↓ 選択  Enter 適用  e エクスポート  i インポート  Esc 戻る)")
	return s
}

// RenderStatic は非 TTY 環境向けの静的描画。
func RenderStatic() string {
	m := New()
	var sb strings.Builder
	sb.WriteString(m.viewDashboard())
	return sb.String()
}

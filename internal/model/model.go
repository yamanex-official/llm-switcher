package model

import "time"

// CLIID は管理対象 CLI の識別子。
type CLIID string

const (
	CLIClaude   CLIID = "claude"
	CLICodex    CLIID = "codex"
	CLIOpenCode CLIID = "opencode"
)

// DetectedCLI は検出結果の 1 行分（CLI × インストール状態）。
type DetectedCLI struct {
	ID        CLIID
	Name      string
	Installed bool
	Path      string // 実行ファイルパス（検出時のみ）
	Note      string // 設定ディレクトリ等の補足
}

// SourceKind は値の読み取り元の種別。
type SourceKind string

const (
	SourceConfig  SourceKind = "config"
	SourceEnv     SourceKind = "env"
	SourceDefault SourceKind = "default"
)

// SourceRef は読み取った値の由来（表示用）。
type SourceRef struct {
	Kind SourceKind
	Path string // 設定ファイルパス または 環境変数名
}

// Target は管理対象の 1 接続 = (CLI, provider_id) の組。
type Target struct {
	CLI        CLIID
	ProviderID string // CLI 内での一意 ID: "anthropic" | "openai" | "deepseek" | ...
	Provider   string // "openai" | "anthropic" | "gemini" | "カスタム"
	BaseURL    string
	APIKey     string
	Model      string
	Enabled    bool
	ReadSource SourceRef // 最後に読み取った値の由来
}

// Profile は名前付き接続先セット。
type Profile struct {
	Name      string
	Targets   []Target
	UpdatedAt time.Time
}

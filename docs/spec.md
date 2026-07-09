# LLMルーター 仕様書

> 最終更新: 2026-07-08（初版）→ レビュー指摘反映済み
> 関連: `docs/review.md`（批判的レビュー指摘リスト）

## 1. 概要・目的

各種 AI コーディング CLI（Codex / Claude Code / OpenCode）が API 接続に使用する
**環境変数・接続設定（Base URL / API Key / Model 等）を一元管理する TUI アプリケーション**。

手動で `~/.claude/settings.json` や `~/.codex/config.toml`、`opencode.jsonc`、各種環境変数を
編集・同期する手間をなくし、以下を TUI 上で完結させる。

- 各 CLI の **現在の接続設定の確認**（検出・読み取り・優先順位解決）
- 環境変数 / Base URL / モデル等の **編集**
- 変更の **反映（apply）**：設定ファイル / `.env` / シェルプロファイル / OS 環境変数

### 背景（なぜ必要か）

LLM ルーター（LiteLLM / Cloudflare AI Gateway 等）を経由して複数プロバイダを切り替える運用では、
CLI ごとに異なるファイル形式（JSON / TOML / JSONC）と環境変数名（`ANTHROPIC_BASE_URL` /
`OPENAI_BASE_URL` / `GEMINI_API_KEY` 等）を覚え、都度書き換える必要がある。これを統合する。

### 提供形態・OS

- 対象 OS：Windows 11 / macOS / Ubuntu（Linux）の 3 環境で同じ操作フロー。
- 配布：単一実行ファイル（Go クロスコンパイル）。エンドユーザーはダウンロードして実行するだけ（第 9 章）。

### フェーズ分け（スコープ管理）

本プロジェクトは個人開発のため、機能をフェーズに分割する（第 13 章マイルストーンも参照）。

- **v1.0（MVP）**: CLI 検出・閲覧・編集・反映（反映先は **CLI 設定ファイルのみ**）。3 OS 単一バイナリ。
- **v1.1+**: `.env` / シェルプロファイル / OS 環境変数への反映、プロファイル管理、構文検証・疎通チェック、エクスポート/インポート。

---

## 2. 対象 CLI と管理対象

管理の最小単位は **「CLI × プロバイダ」**（レビュー A1）。1 つの CLI に複数プロバイダが
共存しうる（Codex の `model_providers.*`、OpenCode の `provider.*`）ため、1 行 = 1 CLI ではなく
1 行 = 1 (CLI, provider_id) とする。

| CLI | プロバイダ指定 | キー環境変数 | Base URL 指定 | 設定ファイル（ユーザー単位） |
|---|---|---|---|---|
| **Claude Code** | 単一（anthropic） | `ANTHROPIC_API_KEY`<br>`ANTHROPIC_AUTH_TOKEN`<br>`CLAUDE_API_KEY` | `ANTHROPIC_BASE_URL`（env） | `~/.claude/settings.json`<br>`~/.claude.json`（認証） |
| **Codex (OpenAI)** | `model_providers.<id>`（複数可） | 組込 `openai` → `OPENAI_API_KEY`<br>カスタムは `env_key` で任意 | `openai_base_url`（user config）<br>カスタムは `model_providers.<id>.base_url` | `~/.codex/config.toml`<br>プロジェクト `.codex/config.toml` |
| **OpenCode** | `provider.<id>`（複数可） | `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` 等（プロバイダ依存） | `provider.<id>.options.baseURL` | `~/.config/opencode/opencode.jsonc`（global）<br>プロジェクト `opencode.jsonc` |
| **Antigravity CLI (Google)** | 未確定 | `GEMINI_API_KEY`/`GOOGLE_API_KEY`（※） | （※） | （※） |

> **Antigravity は v2 以降の対象**（レビュー C1）。公式ドキュメントが未特定のため v1.0 スコープ外。
> 情報が揃い次第ベストエフォートで追加。v1.0 は上記 3 CLI を対象とする。
>
> 出典：Claude Code `docs.claude.com` / Codex `developers.openai.com/codex`（config-reference, environment-variables）/
> OpenCode `opencode.ai/docs`（config, providers）。

共通フィールド（各 Target が持つ）:

- `provider`（openai / anthropic / gemini / カスタム）
- `base_url`（API エンドポイント）
- `api_key`（シークレット）
- `model`（既定モデル、CLI が対応する場合）
- `enabled`（有効/無効）

---

## 3. 機能要件

### v1.0（MVP）

#### F1. CLI 検出・ステータス表示
- インストール済み CLI を自動検出（実行ファイルの有無を OS ごとの探索パスから判定）。
- 各 (CLI, provider) の現在値を読み取り一覧表示（レビュー A2 の優先順位で解決）。
- CLI 未検出の行は「未インストール」としてグレーアウト表示。
- 設定ファイルが存在しない場合は「未設定」と表示し、テンプレート作成を促す。

#### F2. 閲覧
- 設定ファイル / 環境変数から `base_url` / `api_key` / `model` 等をパースして表示。
- `api_key` はデフォルトでマスク表示（末尾数文字のみ、または `****`）。表示切替可。
- 各値の**読み取り元**（設定ファイルパス / 環境変数名）を表示。

#### F3. 編集
- TUI フォームで各フィールドを編集（テキスト入力、選択、トグル）。
- メモリ上の `Target` を更新。

#### F4. 反映（apply）— v1.0 は (a) のみ
- **(a) CLI 設定ファイル**：編集値を該当キーだけ更新（他キーは保持）。上書き前 `.bak` 作成。
  - v1.1 以降で (b)(c)(d) を追加（下記「将来機能」）。

### 将来機能（v1.1+）

#### F5. 追加の反映先
- (b) `.env` ファイル（第 8 章）
- (c) シェルプロファイル（第 8 章、マーカー方式）
- (d) OS 環境変数（第 8 章）

#### F6. プロファイル管理
- 「自宅 LiteLLM」「Cloudflare GW」「直接接続」等、接続先セットを名前付きプロファイルとして保存。
- プロファイル切替で一括反映。

#### F7. 検証
- 反映前に JSON/TOML/JSONC の構文検証。
- （任意）`base_url` への軽量な疎通チェック（HEAD / models エンドポイント）。

#### F8. エクスポート/インポート
- 設定を YAML/JSON でエクスポート（シークレットは除外または暗号化）。
- 別マシンへインポート。

---

## 4. 非機能要件

- **3 OS 同一動作**：Windows 11 / macOS / Ubuntu で同じ TUI・同じ操作フロー。OS 差は `internal/platform` に隠蔽。
- **エンドユーザーは何もインストールしなくてよい**：単一実行ファイルのみ。実行時ランタイム不要。
  Go で `windows/amd64`・`darwin/arm64`・`linux/amd64` をクロスコンパイルし GitHub Releases で提供。
  （開発者側は Go ツールチェインを 1 回導入するが、公開ツールの維持コストと捉える。）
- **OS 差の吸収**：パス・シェル・環境変数反映は `runtime.GOOS` で分岐する共通レイヤに集約。
- **設定はローカルのみ**：クラウド同期なし。
- **破壊的変更の防御**：既存設定ファイルは上書き前に `.bak` を作成し、トランザクション適用（第 8 章）。
- **監査性**：反映履歴をローカルに記録（オプション）。

---

## 5. 技術構成

- **言語: Go**（単一静的バイナリ、実行時依存ゼロ、3 OS 対応）。
- **TUI: Bubble Tea + Lipgloss**（小規模・成熟）。代替: Go 標準ライブラリのみ（要相談）。
- **パーサ**:
  - TOML: `BurntSushi/toml`
  - JSON/JSONC: **JSONC パーサの実現性リスクあり（レビュー C4）**。Go 標準には JSONC がなく、
    サードパーティも限定的。方針: `jsonc-parser`（`github.com/PaesslerAG/jsonc`）を評価し、
    不十分なら「コメント・末尾カンマを除去してから `encoding/json` でパース」する自前プリプロセッサを実装。
    いずれにせよラウンドトリップ（読み書きで他キーを保持）の検証が必要。
- **OS 抽象化**: `runtime.GOOS` で分岐する `internal/platform` が
  ホーム・設定ディレクトリ（`os.UserConfigDir()`）・シェルプロファイル・環境変数反映を集約。
- **CLI アダプタ**: `internal/adapter` に各 CLI ごとの読み書きロジックを配置
  （ClaudeAdapter / CodexAdapter / OpenCodeAdapter）。1 (CLI, provider_id) に対する
  読み取り・書き込みキー写像をカプセル化（レビュー A1/D6）。
- **CI ビルド**: GitHub Actions で 3 OS をクロスコンパイルし Releases へアップロード。

### ディレクトリ構成（予定）
```
llm-switcher/
├── cmd/llm-switcher/main.go        # エントリポイント
├── internal/
│   ├── detect/                   # CLI 検出
│   ├── adapter/                  # claude/codex/opencode 読み書きアダプタ
│   ├── model/                    # データモデル (Target, Profile)
│   ├── apply/                    # 反映ロジック (file/.env/profile/osenv) + トランザクション
│   ├── platform/                 # OS差吸収 (GOOS 分岐)
│   └── tui/                      # Bubble Tea 画面
├── configs/                      # テンプレート
├── docs/spec.md
├── docs/review.md
├── go.mod / go.sum
├── .github/workflows/release.yml
└── README.md
```

---

## 6. データモデル

管理単位は **CLI × プロバイダ**（レビュー A1）。`Source` は複数ありうるため参照型に分割（D6）。

```go
// 管理対象の 1 接続 = (CLI, provider_id) の組
type Target struct {
    CLI        string // "claude" | "codex" | "opencode"
    ProviderID string // CLI 内での一意 ID: "anthropic" | "openai" | "deepseek" | ...
    Provider   string // "openai" | "anthropic" | "gemini" | "カスタム"
    BaseURL    string
    APIKey     string // マスク表示・保存時は平文（ローカルのみ）
    Model      string
    Enabled    bool
    ReadSource SourceRef // 最後に読み取った値の由来（表示用）
}

// 読み取り元（複数の可能性があるため構造化）
type SourceRef struct {
    Kind string // "config" | "env" | "default" | "profile"
    Path string // 設定ファイルパス または 環境変数名
}

// 名前付き接続先セット（v1.1 のプロファイル機能で使用）
type Profile struct {
    Name        string
    Targets     []Target
    UpdatedAt   time.Time
}
```

- アプリ自身の状態は `os.UserConfigDir()` 配下の `llm-switcher/`（第 11 章「OS 固有」参照）に保存。
  ※第 6 章旧版の `~/.config/llm-switcher/` は Linux のみで正しく、Windows/macOS では一致しないため
  `os.UserConfigDir()` に一本化（レビュー B4）。
- **シークレットは平文でローカル保存**。OS のファイル権限で保護（Ubuntu は `0700`）。
  必要なら v1.1 で OS キーチェーン連携を検討。

---

## 7. 読み取り優先順位（レビュー A2）

「現在値」の表示は、各 CLI が実際に解決する優先順位を完全には再現できないため、
**本ツールの解決ルール**を以下の通り固定し、かつ読み取り元を表示する。

1. **CLI 設定ファイル**（ユーザー単位 > プロジェクト単位）での明示値を優先。
2. 設定ファイルに値がない場合は、**環境変数**（OS 環境変数 / シェルプロファイル / `.env` で export 済み）を使用。
3. いずれもない場合は「未設定（default なし）」とする。

- 各値の `Source` を表示し、どの層から来たかをユーザーに提示。
- CLI ごとの詳細優先順位（例: Codex は `config.toml` の `model_providers.<id>.env_key` が env を参照等）
  はアダプタ内で吸収し、ユーザーには「設定ファイル値」または「環境変数値」のいずれかとして表示。

---

## 8. 反映（apply）先と挙動

### 8.1 トランザクション適用（レビュー B1）

複数ファイルへ同時反映する場合の整合性を保証する。

1. 反映対象の全ファイルを `.bak` にスナップショット（元の内容を退避）。
2. 各ファイルを **テンポラリ書き出し → アトミック rename** で更新。
3. いずれかの書き込みに失敗した場合、成功済みの分も含め **全ファイルを `.bak` から復元**しロールバック。
4. 全成功時のみ `.bak` を破棄（または N 世代保持）。
5. 反映前に dry-run プレビュー（変更 diff）を表示し確認プロンプト。

### 8.2 v1.0 対象: (a) CLI 設定ファイル

- アダプタが対象 (CLI, provider_id) に対応するファイルの該当キーだけを更新（他キーは保持 = ラウンドトリップ）。
- 上書き前 `.bak` 作成。パスは各 OS の規定位置（第 11 章）。
- **Codex の注意**: プロジェクト `.codex/config.toml` では `openai_base_url` 等は無視される仕様のため、
  ユーザー単位 `~/.codex/config.toml` へ書き込む。
- **OpenCode の注意（レビュー B3）**: v1.0 は BaseURL を **直接値** として `opencode.jsonc` に書き込む
  （決定論的で単純）。`{env:VAR}` / `{file:path}` による間接参照は v1.1 でオプション化する。

### 8.3 将来対象: (b)(c)(d)（v1.1+）

| 反映先 | 挙動 | OS ごとの実装（`internal/platform` が吸収） |
|---|---|---|
| (b) `.env` | プロジェクトまたはグローバルの `.env` へ `KEY=VALUE` を書き出し/更新 | 対象: **カレントディレクトリの `.env`**（Git ルートまで遡る）（レビュー C2）。既存 `.env` をパースして該当行を置換（3 OS 共通） |
| (c) シェルプロファイル | 管理ブロックのみを追加/更新 | **マーカー方式**（下記 8.4）でブロック外には一切触らない |
| (d) OS 環境変数 | 永続環境変数を設定 | **Windows**: `setx` / Windows API（`Machine` は UAC 昇格）。**macOS/Ubuntu**: 真の OS スコープが標準でないため (c) プロファイルへの export と等価化して動作統一 |

### 8.4 シェルプロファイル編集の安全方式（レビュー A3）

プロファイルを直接書き換えず、以下の管理ブロックのみを管理する。

```
# >>> llm-switcher >>>
export OPENAI_BASE_URL=https://...
export OPENAI_API_KEY=sk-...
# <<< llm-switcher <<<
```

- ブロック外の行（ユーザーの既存設定）には一切触らない。
- 更新時は該当ブロック全体を置換。ブロックがなければ末尾に追加。
- マーカー検出ロジックは「開始マーカー〜終了マーカー」の正規表現で行い、ネストしない。
- **並行利用の注意（レビュー C6）**: 対象 CLI が実行中の場合、反映前に警告を出す（CLI 側の再読み込み挙動は保証外）。

---

## 9. 配布・インストール（エンドユーザー視点）

- **前提条件: なし**。ユーザーは何もインストールしなくてよい。
- GitHub Releases から自身の OS 向けバイナリをダウンロードし実行するだけ。
- **整合性検証（レビュー D2）**: Releases に `checksums.txt`（SHA256）を同梱し、ダウンロードの
  改ざんチェックを可能にする。
- **macOS Gatekeeper（レビュー C3）**: v1.0 は未署名。README に「右クリック→開く」または
  `xattr -dr com.apple.quarantine llm-switcher` の手動許可手順を記載。codesign + notarize は v1.1 で検討。
- オプション（利便性のみ・必須ではない）: Windows `scoop` / macOS `brew` / Ubuntu `deb` 提供（v1.1）。

---

## 10. TUI 画面設計

```
[起動]
  │
  ▼
(1) ダッシュボード（表形式）
    ┌─ CLI │ Provider │ BaseURL │ Key │ 状態 ─┐
    │ claude   anthropic   (env)     ****  設定済み │
    │ codex    openai      https://… ****  設定済み │
    │ opencode anthropic   -         -     未設定   │
    │ opencode openai      https://… ****  設定済み │
    └─────────────────────────────────────┘
    CLI 未検出は行をグレーアウト「未インストール」
  │ 行選択 → Enter
  ▼
(2) 接続編集画面
    ├─ ProviderID (表示) / BaseURL(編集) / APIKey(マスク・表示切替) / Model / Enabled
    └─ 保存 → Target 更新
  │
  ▼
(3) 反映先選択 + プレビュー（v1.0 は [x] CLI設定ファイル のみ）
    └─ Apply（dry-run diff 表示 → 確認 → 書き込み → 結果表示）
```

- ダッシュボードの「状態」列: 設定ファイル or 環境変数で値が解決されていれば「設定済み」、
  未検出 CLI は「未インストール」、値なしは「未設定」。
- キーバインド: vim 風または矢印 + Enter（一貫させる）。ヘルプ画面 `?`。
- 日本語 UI のマルチバイト文字幅計算は、Windows Terminal / pwsh / macOS Terminal で検証が必要
  （レビュー D3: Bubble Tea / Lipgloss の幅計算リスク）。

---

## 11. OS 固有の考慮（3 環境共通動作のための抽象化）

OS 差は `internal/platform` に集約し、TUI は OS を意識しない。

### 共通
- アプリ設定ディレクトリ: **`os.UserConfigDir()` 配下の `llm-switcher/`**（レビュー B4 統一）。
  - Windows: `%APPDATA%\llm-switcher\`
  - macOS: `~/Library/Application Support/llm-switcher/`
  - Ubuntu: `~/.config/llm-switcher/`
- ホーム: `os.UserHomeDir()`。改行コード: 設定ファイル書き出しは常に **LF**。
  パス区切りは `filepath` で自動選択。

### シェル対応マトリクス

| シェル | OS | プロファイル反映 | OS 環境変数反映 | プロファイル書式 |
|---|---|---|---|---|
| PowerShell (pwsh 7+) | Windows | ✓ (`$PROFILE`) | ✓ (`setx`) | `$env:KEY = "VALUE"` |
| コマンドプロンプト (cmd.exe) | Windows | × (非対応) | ✓ (`setx`) | — |
| Bash | WSL / Linux | ✓ (`~/.bashrc`) | ✓ (`~/.profile`) | `export KEY=VALUE` |
| Zsh | macOS / Linux | ✓ (`~/.zshrc`) | ✓ (`~/.profile`) | `export KEY=VALUE` |
| Fish | Linux | ✓ (`~/.config/fish/config.fish`) | ✓ (`~/.config/fish/config.fish`) | `set -gx KEY VALUE` |
| Csh | Linux | ✓ (`~/.cshrc`) | ✓ (`~/.login`) | `setenv KEY VALUE` |
| Tcsh | Linux | ✓ (`~/.tcshrc`) | ✓ (`~/.login`) | `setenv KEY VALUE` |

### Windows 11
- CLI 設定ディレクトリ: `%USERPROFILE%\.claude` / `%USERPROFILE%\.codex` / `%USERPROFILE%\.config\opencode`
- シェル自動判定: WSL 実行中は Linux 側のシェル検出にフォールバック。
  WSL 以外では `COMSPEC` / `SHELL` 環境変数で PowerShell と cmd.exe を判定。
- プロファイル: PowerShell `$PROFILE`（pwsh 7+）。cmd.exe はシェルプロファイルの標準機構がないため、
  「シェルプロファイル」反映先は非対応。代わりに「OS 環境変数」（`setx`）を使用すること。
- OS 環境変数: `setx` コマンド（ユーザースコープ）。PowerShell・cmd.exe 両対応。
  `Machine` スコープは要 UAC 昇格のため未サポート。

### macOS
- CLI 設定ディレクトリ: `~/.claude` / `~/.codex` / `~/.config/opencode`
- シェルプロファイル: 既定 zsh `~/.zshrc`（bash は `~/.bashrc`）。`$SHELL` から判定。
- 永続環境変数: 標準手段がないためプロファイルへの export で統一。

### Ubuntu（Linux）
- CLI 設定ディレクトリ: `~/.claude` / `~/.codex` / `~/.config/opencode`
- シェルプロファイル: 既定 bash `~/.bashrc`。`$SHELL` から zsh / fish / csh / tcsh を自動判定。
- 永続環境変数: プロファイルへの export で統一（シェルに応じて書式を自動切替）。
- ファイル権限: シークレット保存ディレクトリは `0700`。

### WSL（Windows Subsystem for Linux）
- Windows バイナリを WSL 内で実行した場合、`/proc/sys/fs/binfmt_misc/WSLInterop` または
  `WSL_DISTRO_NAME` 環境変数により WSL 環境を自動検出し、Linux のシェル検出ロジックにフォールバックする。
- WSL 内の Linux シェル（bash/zsh/fish 等）のプロファイルに書き込み可能。
- `setx` は Windows 側の環境変数を設定する。WSL 内で実行する場合は Linux バイナリの使用を推奨。

---

## 12. セキュリティ・シークレット取扱

- API Key は画面でデフォルトマスク。
- エクスポート時はシークレットを除外（または暗号化）。
- ハードコード禁止（ソース内にキーを書かない）。
- コミット対象外: `llm-switcher` の状態ディレクトリを `.gitignore` に含める。
- `gitleaks` を導入し、キー混入を CI/フックで検知（dev ルート方針に準拠）。

---

## 13. テスト戦略（レビュー C5）

- **単体テスト**: 各アダプタの設定ファイル読み取り・書き込み（ラウンドトリップで他キー保持を確認）。
- **platform テスト**: `GOOS` 分岐（パス解決・プロファイル判定）の表ベーステスト。
- **apply トランザクションテスト**: 途中失敗時のロールバック挙動。
- **JSONC パーサテスト**: コメント・末尾カンマ・文字列内 `//` のエッジケース。
- **3 OS E2E**: 最低限、手動確認項目リスト（検出→閲覧→編集→反映→再起動後の永続）を CI やリリース前チェックリストに含める。

---

## 14. マイルストーン（開発手順）

1. **M1 スケルトン**: Go モジュール初期化、Bubble Tea ダッシュボード、CLI 検出。**3 OS ビルド・起動確認**。
2. **M2 読み取り**: Claude/Codex/OpenCode アダプタの読み取り + 優先順位解決（第 7 章） + 閲覧画面。
3. **M3 編集**: 接続編集フォーム + メモリモデル更新。
4. **M4 反映(a)**: CLI 設定ファイルへのトランザクション適用 + バックアップ/ロールバック + dry-run プレビュー。
5. **M5 v1.0 仕上げ**: 構文検証、gitleaks/README、Windows/macOS/Ubuntu 動作確認、**Releases 公開（チェックサム付）**。
6. **M6 v1.1 反映(b)(c)(d)**: `.env` / シェルプロファイル（マーカー方式）/ OS 環境変数へ反映。
7. **M7 v1.1 プロファイル + 検証 + export/import**: F5 〜 F8。
8. **M8 v1.1 配布整備**: macOS 署名/notarize、scoop/brew/deb 提供、Antigravity 追加（情報揃い次第）。

---

## 15. 未確定・要調査事項

- **JSONC パーサの選定（C4）**: `jsonc-parser` の評価、または自前プリプロセッサの採否。
- **macOS 署名の要否（C3）**: v1.0 は未署名・手動許可、v1.1 で notarize 検討。
- **OpenCode `{env:VAR}` 間接参照（B3）**: v1.0 は直接値書き込み、v1.1 でオプション化。
- **OS 環境変数反映の権限モデル（Windows）**: ユーザー/システムの既定と UAC 挙動。
- **Antigravity CLI 公式仕様（C1）**: 設定ファイル・Base URL 指定・環境変数名。v2 対象。

---

## 16. 免責事項

- **環境変数反映の即時性**: シェルプロファイルおよび OS 環境変数への変更は、**反映後に新規シェルセッションを開始するまで有効になりません**。
  既存の開いているターミナルウィンドウでは変更が反映されないため、再起動または新しいターミナルを開いてください。
- **シェル自動判定の限界**: `$SHELL` 環境変数に基づいてシェルを判定します。`$SHELL` が設定されていない、
  またはシンボリックリンクで実シェル名と異なる場合、正しく検出できない可能性があります。
  その場合は手動で適切な反映先を選択してください。
- **cmd.exe の制約**: コマンドプロンプトには標準的なシェルプロファイル機構がありません。
  「シェルプロファイル」反映先は cmd.exe では使用できません。代わりに「OS 環境変数」（`setx`）を使用してください。
- **WSL の制約**: Windows バイナリを WSL 内で実行した場合、Linux シェルとして認識されます。
  `setx` による OS 環境変数設定は WSL 側ではなく Windows 側の環境変数に作用します。
  WSL 内で Linux の挙動を期待する場合は Linux 版バイナリの使用を推奨します。
- **ファイル破損**: 本ツールは既存設定ファイルのバックアップ（`.bak`）を自動作成しますが、
  重要な設定ファイルは必ずご自身でもバックアップを取ってください。
- **API キーの漏洩**: 本ツールは API キーをローカルファイルに平文で保存します。
  ファイル権限 (`0600`) による保護に依存しており、暗号化は行いません。
  共有マシンでの使用には十分注意してください。
- **設定の競合**: 本ツールの管理対象外の方法（手動編集・他ツール）で設定ファイルや環境変数を変更した場合、
  本ツールの表示と実際の状態が一致しなくなる可能性があります。

## 17. 参照

- Claude Code CLI reference — https://docs.claude.com/en/docs/claude-code/cli-reference
- Codex Configuration Reference — https://developers.openai.com/codex/config-reference
- Codex Environment Variables — https://developers.openai.com/codex/environment-variables
- OpenCode Config — https://opencode.ai/docs/config/
- OpenCode Providers — https://opencode.ai/docs/providers/
- 批判的レビュー指摘 — `docs/review.md`

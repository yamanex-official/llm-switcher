# llm-switcher

対話型 AI コマンドラインツールの API 接続設定を一元管理する TUI アプリケーション。

<img width="1262" height="563" alt="2026-07-08_17h17_23" src="https://github.com/user-attachments/assets/e89fb912-5cec-4a0e-9a14-0d3216cd5e90" />

---

**対応 CLI**: Claude Code / Codex (OpenAI) / OpenCode  
**対応 OS**: Windows / macOS / Linux（Ubuntu/Zorin OS想定）

---

## インストール

### バイナリ（推奨）

[Releases](https://github.com/yamanex-official/llm-switcher/releases) からご自身の OS 向けバイナリをダウンロードし、実行するだけ。

| OS | ファイル |
|---|---|
| Windows | `llm-switcher-windows-amd64.exe` |
| macOS (Apple Silicon) | `llm-switcher-darwin-arm64` |
| Ubuntu / Linux | `llm-switcher-linux-amd64` |

macOS で初回実行時にブロックされた場合:

```bash
# 検疫属性を解除（ダウンロードしたバイナリ用）
xattr -dr com.apple.quarantine llm-switcher-darwin-arm64
# または right-click → Open で実行許可
```

### ソースからビルド

```bash
# Go 1.26+ が必要
git clone https://github.com/yamanex-official/llm-switcher.git
cd llm-switcher
make build-all    # 3 OS 分をビルド
# または
go build -o bin/llm-switcher ./cmd/llm-switcher
```

---

## 使い方

```bash
./bin/llm-switcher
```

TUI が起動し、検出された CLI と現在の接続設定を表示します。

<img width="1262" height="563" alt="2026-07-08_17h17_23" src="https://github.com/user-attachments/assets/bf8431e9-0df5-43b6-af2e-99ee3164bd9f" />

### キー操作

| 画面 | キー | 操作 |
|---|---|---|
| ダッシュボード | `↑` `↓` / `j` `k` | 行選択 |
|  | `Enter` / `e` | 選択行の編集画面へ |
|  | `r` | 設定を再読み込み |
|  | `p` | プロファイル管理画面へ |
|  | `q` | 終了 |
| 編集 | `Tab` / `↓` | 次のフィールドへ |
|  | `Shift+Tab` / `↑` | 前のフィールドへ |
|  | `Enter` | 反映先選択画面へ |
|  | `Esc` | ダッシュボードへ戻る |
| 反映先選択 | `Space` | チェック切替 |
|  | `Enter` | 反映実行 |
|  | `Esc` | 編集画面へ戻る |
| プロファイル | `s` | 入力名で現在の設定をプロファイル保存 |
|  | `Enter` | 選択プロファイルを適用 |
|  | `e` | 全プロファイルエクスポート |
|  | `i` | エクスポート済ファイルをインポート |
|  | `Esc` | ダッシュボードへ戻る |

### 反映先

編集後の設定は以下のいずれか（または複数）に反映できます：

- [x] **CLI 設定ファイル** — 各 CLI の `config.toml` / `opencode.jsonc` / `settings.json` を直接更新
- [ ] **.env ファイル** — `$HOME/.env` に環境変数を出力
- [ ] **シェルプロファイル** — シェルに応じたプロファイルに管理ブロックを追記
- [ ] **OS 環境変数** — Windows は `setx`、macOS/Linux はシェルプロファイルで対応

#### シェル対応

| シェル | プロファイル反映 | OS 環境変数反映 |
|---|---|---|
| PowerShell (pwsh 7+) | ✓ `$PROFILE` | ✓ (`setx`) |
| コマンドプロンプト | × | ✓ (`setx`) |
| Bash / Zsh | ✓ | ✓ |
| Fish | ✓ | ✓ |
| Csh / Tcsh | ✓ | ✓ |

> **注意**: コマンドプロンプト (cmd.exe) は標準的なシェルプロファイル機構がないため、シェルプロファイル反映先は使用できません。代わりに「OS 環境変数」を使用してください。

シェルプロファイルは `# >>> llm-switcher >>>` 〜 `# <<< llm-switcher <<<` のブロック内のみを管理し、ブロック外の既存設定は一切破壊しません。シェルに応じて適切な書式（bash/zsh: `export`、fish: `set -gx`、csh/tcsh: `setenv`、PowerShell: `$env:`）が自動選択されます。

---

## 管理対象 CLI と設定ファイル

<img width="1262" height="563" alt="2026-07-08_17h17_36" src="https://github.com/user-attachments/assets/fc32623e-379d-4fa3-99c6-99802f7345ec" />


| CLI | 設定ファイル | 主な管理環境変数 |
|---|---|---|
| **Claude Code** | `~/.claude/settings.json` | `ANTHROPIC_BASE_URL`, `ANTHROPIC_API_KEY` |
| **Codex** | `~/.codex/config.toml` | `OPENAI_API_KEY`, `openai_base_url` |
| **OpenCode** | `~/.config/opencode/opencode.jsonc` | プロバイダ依存 |

---

## プロファイル

「自宅 LiteLLM」「Cloudflare GW」「直結」など、接続先セットを名前付きで保存し、ワンキーで切替可能です。

プロファイル画面（ダッシュボードで `p`）で名前を入力し `s` で保存、`Enter` で適用します。

---

## ライセンス

MIT

## 免責事項

- **環境変数の反映は新規シェルセッション開始後に有効になります**。既存のターミナルでは変更が反映されません。
- **cmd.exe はシェルプロファイル機能をサポートしません**。代わりに「OS 環境変数」の反映を使用してください。
- **WSL 内で Windows バイナリを実行した場合**、`setx` は Windows 側の環境変数に作用します。WSL 内では Linux バイナリの使用を推奨します。
- API キーはローカルファイルに平文で保存されます。共有マシンでの使用に注意してください。
- 設定ファイルのバックアップ (`.bak`) は自動作成されますが、重要なファイルは必ずご自身でもバックアップしてください。

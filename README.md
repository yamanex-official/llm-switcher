# LLM Router (llm-router)

対話型 AI コマンドラインツールの API 接続設定を一元管理する TUI アプリケーション。

**対応 CLI**: Claude Code / Codex (OpenAI) / OpenCode  
**対応 OS**: Windows / macOS / Ubuntu

---

## インストール

### バイナリ（推奨）

[Releases](https://github.com/yamanex-official/llm-router/releases) からご自身の OS 向けバイナリをダウンロードし、実行するだけ。

| OS | ファイル |
|---|---|
| Windows | `llm-router-windows-amd64.exe` |
| macOS (Apple Silicon) | `llm-router-darwin-arm64` |
| Ubuntu / Linux | `llm-router-linux-amd64` |

macOS で初回実行時にブロックされた場合: `xattr -dr com.apple.quarantine llm-router-darwin-arm64`

### ソースからビルド

```bash
# Go 1.26+ が必要
git clone https://github.com/yamanex-official/llm-router.git
cd llm-router
make build-all    # 3 OS 分をビルド
# または
go build -o bin/llm-router ./cmd/llm-router
```

---

## 使い方

```bash
./bin/llm-router
```

TUI が起動し、検出された CLI と現在の接続設定を表示します。

### キー操作

| 画面 | キー | 操作 |
|---|---|---|
| ダッシュボード | `↑` `↓` / `j` `k` | 行選択 |
|  | `Enter` / `e` | 選択行の編集画面へ |
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
- [ ] **シェルプロファイル** — `~/.zshrc` / `~/.bashrc` にマーカー管理ブロックを追記
- [ ] **OS 環境変数** — Windows は `setx`、macOS/Ubuntu はシェルプロファイルで対応

シェルプロファイルは `# >>> llm-router >>>` 〜 `# <<< llm-router <<<` のブロック内のみを管理し、ブロック外の既存設定は一切破壊しません。

---

## 管理対象 CLI と設定ファイル

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

---

## 開発

```bash
go vet ./...      # 静的解析
go test ./...     # テスト
make build-all    # 3 OS クロスビルド
make tidy         # 依存整理
```

# 仕様書レビュー指摘

## 致命的な問題（実装前に直すべき）

### A1. データモデルの設計不足
`Connection` は「1 CLI = 1 Provider」だが、実際は 1 CLI に複数プロバイダが存在する。
- Codex: `model_providers.openai` / `model_providers.deepseek` / `model_providers.ollama` 等、複数定義可
- OpenCode: `provider.anthropic` / `provider.openai` / `provider.gemini` 等、複数定義可
- Claude Code: `ANTHROPIC_BASE_URL` をプロキシに向けることで実質複数プロバイダを扱える

接続の管理単位を「CLI単位」にするのか「CLI × プロバイダ」にするのか、ここを間違えると全レイヤが破綻する。
→ `CLI × Provider(provider_id)` のマトリックスにする必要あり。

### A2. 読み取り元の優先順位が未定義
同じ環境変数が複数箇所（シェルプロファイル、.env、OS環境変数、設定ファイル）に存在した場合、
どれを「現在値」として表示するのか定義されていない。
各 CLI は独自の優先順位を持つ（Codex: config.toml > env、OpenCode: project > global > {env:VAR} 展開）。
このツールの読み取り優先順位ルールを明示しないと「現在値」が一意に決まらない。

### A3. シェルプロファイル編集の実現性・危険性
`export KEY=VALUE` を bashrc/zshrc/$PROFILE に「追加/更新」する挙動が曖昧:
- 既存行の検出・置換ロジック（同名キーの値違い、コメント行、eval 内での定義）未定義
- プロファイル破壊時の影響（シェルが起動しなくなる）が仕様に考慮されていない
- 追記による重複定義の副作用も検討されていない

対策案: マーカーコメント `# >>> llm-router >>>` 〜 `# <<< llm-router <<<` で
管理セクションを限定し、セクション外の行には一切触らない方針を採る。

### A4. M1 〜 M7 のスコープが広すぎる
F1（検出）→ F2（閲覧）→ F3（編集）→ F4（4つの反映先）→ F5（プロファイル）→ F6（検証）→ F7（export/import）
+ M7（3OS CI/release）。個人開発で数ヶ月かかる規模。

当初ユーザー要求は「各LLMが使用する環境変数を**確認・編集・反映**ができる TUI」。
→ v1.0 MVP: F1（検出）+ F2（閲覧）+ F3/F4（編集・反映、反映先は CLI 設定ファイルのみ）。
プロファイル・export/import・疎通チェックは v1.1+ に延期。

---

## 重大な不足（実装中に詰まる可能性が高い）

### B1. 反映のトランザクション性
複数ファイルを同時更新（CLI設定ファイル + .env + シェルプロファイルを一括apply）した場合、
一部だけ成功・一部だけ失敗した状態の整合性が保証されない。
「.bak から復元」では不十分: 2ファイル成功・1ファイル失敗時にどの .bak で戻すか判断がつかない。

→ Atomic write（tmpファイル書き→rename）＋ 全ファイル成功後にのみ .bak 破棄、
失敗時は成功した分も含めて全ロールバック のトランザクションパターンが必要。

### B2. TUIの画面設計が具体性不足
「ダッシュボード → CLI一覧 → 編集 → apply」では実装できない。最低限:
- ダッシュボード: 各行に表示するカラム（CLI名 / プロバイダID / BaseURL / 設定状態）
- 接続ステータスの定義（設定ファイル有無？疎通確認？環境変数のみ？）
- CLI 非検出時の UX（グレーアウト・「未インストール」表示）
- 編集画面のフィールド順序とバリデーションルール

### B3. OpenCode の `{env:VAR}` / `{file:path}` 変数展開
OpenCode は設定ファイル内で `{env:OPENAI_API_KEY}` のように環境変数を間接参照する。
このツールが BaseURL を書き込む場合:
- 直接値を opencode.jsonc に書くのか
- `{env:LLM_ROUTER_BASE_URL}` のような間接参照を書くのか
どちらを選択するかで挙動が全く異なる。仕様に明示すべき。

### B4. アプリ自身の設定ディレクトリが仕様内で矛盾
- 第6章: `~/.config/llm-router/`
- 第10章: `os.UserConfigDir()` 配下の `llm-router/`

`os.UserConfigDir()` は:
- Windows: `%APPDATA%`（例: `C:\Users\takum\AppData\Roaming`）
- macOS: `~/Library/Application Support/`
- Linux: `~/.config/`

`~/.config/llm-router/` は Linux では正しいが Windows/macOS では実際に作成されるパスと一致しない。
`os.UserConfigDir()` に一本化し、各OSでの実際のパス例を明記する。

---

## 中程度の問題

### C1. Antigravity CLI を対象CLIに含める是非
未確定のままスコープに入れると、進捗のブロッカーになりうる。
→ 「v2 以降対応」「情報が揃い次第ベストエフォート」と切り分ける。

### C2. `.env` の対象ディレクトリが不明
「プロジェクトまたはグローバルの `.env`」とあるが:
- プロジェクト = カレントディレクトリ？ Git ルート？
- グローバル = `$HOME/.env`？
仕様に明示が必要。

### C3. macOS バイナリ署名 (Gatekeeper)
GitHub Releases からダウンロードした未署名バイナリは、初回実行時に
「開発元を確認できないため開けません」と macOS Gatekeeper にブロックされる。
codesign + notarize の要否を仕様に明記する（v1.0 ではスキップし README で手動許可手順を案内するなど）。

### C4. JSONC パーサの実現性リスク
Go の標準ライブラリには JSONC（末尾カンマ・コメント対応）が存在しない。
サードパーティも数が少なく品質にばらつきあり。
自前実装する場合はバックスラッシュエスケープ・文字列内 `//` 等のエッジケースに注意。
リスクとして仕様に追記すべき。

### C5. テスト戦略不在
- パーサー単位テスト（各 CLI の設定ファイル読み取り）
- `internal/platform` の OS 分岐テスト
- 反映の backup/rollback 挙動テスト
- 3 OS での E2E テスト（最低限、手動確認項目リスト）
いずれも仕様に書かれていない。

### C6. 並行利用の注意
Claude Code / Codex が実行中に設定ファイルを書き換えた場合の挙動は CLI 側の実装依存。
v1 では「実行中のCLIがある場合は反映前に警告」程度のガードを推奨。

---

## 軽微な問題

### D1. セクション番号「7.5」
「配布」が第7章と第8章の間に不自然に挿入されている。独立した章（8→9にずらす）にすべき。

### D2. 「ダウンロードのみ」はセキュリティ上の注意が必要
GitHub Releases からの直ダウンロードは、改ざんチェック（チェックサム/SHA256）がない場合
サプライチェーンリスクがある。最低限 checksums.txt の提供を仕様に入れる。

### D3. Windows ターミナルでのマルチバイト文字
日本語 UI の幅計算が Windows Terminal / cmd.exe / pwsh で一致するとは限らない。
Bubble Tea / Lipgloss が正しく扱えるか検証が必要。リスクとして追記。

### D4. M7（配布）は単一マイルストーンとして軽すぎる
CI 設定 + クロスコンパイル + macOS 署名 + checksums + brew/scoop formula は
それなりの作業量。M6 の仕上げに統合するか、M7 のタスクを具体化する。

### D5. 参照URLの劣化リスク
第13章の参照URLはプロバイダ側のリニューアルで変わりうる。Wayback Machine の
アーカイブURLも併記するか、適宜更新前提とする旨を付記。

### D6. `internal/model/Connection` に `Source` フィールドがあるが
CLI × Provider のマトリックス化（A1）に伴い、`Source` が単一のパス/変数名では
足りなくなる（config file と env var 両方から読むケース）。再設計時に見直し。

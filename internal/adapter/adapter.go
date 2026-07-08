package adapter

import (
	"os"
	"strings"

	"llm-router/internal/model"
)

// Adapter は各 CLI の設定読み書きをカプセル化する。
type Adapter interface {
	ID() model.CLIID
	Name() string
	ConfigDir() string
	ConfigPath() string // 設定ファイルのフルパス（スナップショット用）
	// Read は現在の接続設定を (CLI, provider) ごとに読み取る。
	Read() ([]model.Target, error)
	// WriteConfig は base_url / model / api_key(設定ファイルが対応する場合) を設定ファイルへ書き込む。
	WriteConfig(t model.Target) error
	// EnvNames は env 系反映先(.env/シェル/OS環境変数)で使用する変数名を返す。
	// base_url が設定ファイルベースの場合は空文字。
	EnvNames(t model.Target) (baseURLEnv, apiKeyEnv string)
}

// All は全アダプタを返す。
func All() []Adapter {
	return []Adapter{NewClaude(), NewCodex(), NewOpenCode()}
}

func firstEnv(names ...string) string {
	for _, n := range names {
		if v := os.Getenv(n); v != "" {
			return v
		}
	}
	return ""
}

// deepGet はドット区切りパスで map から値を取得する。
func deepGet(m map[string]interface{}, path string) (interface{}, bool) {
	cur := m
	keys := strings.Split(path, ".")
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return nil, false
		}
		if i == len(keys)-1 {
			return v, true
		}
		nm, ok := v.(map[string]interface{})
		if !ok {
			return nil, false
		}
		cur = nm
	}
	return nil, false
}

// deepSet はドット区切りパスで map に値を設定する（途中の map は生成）。
func deepSet(m map[string]interface{}, path string, val interface{}) {
	keys := strings.Split(path, ".")
	cur := m
	for i, k := range keys {
		if i == len(keys)-1 {
			cur[k] = val
			return
		}
		nm, ok := cur[k].(map[string]interface{})
		if !ok {
			nm = map[string]interface{}{}
			cur[k] = nm
		}
		cur = nm
	}
}

func tomlString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

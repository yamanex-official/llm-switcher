package adapter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
)

func loadJSONC(path string) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	stripped := stripJSONC(b)
	if err := json.Unmarshal(stripped, &m); err != nil {
		return m, err
	}
	return m, nil
}

func saveJSONC(path string, m map[string]interface{}) error {
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	return os.WriteFile(path, out, 0o600)
}

// stripJSONC は JSONC のコメントと末尾カンマを除去する（文字列内は考慮）。
func stripJSONC(b []byte) []byte {
	var out bytes.Buffer
	inString := false
	inLineComment := false
	inBlockComment := false
	i := 0
	for i < len(b) {
		c := b[i]
		switch {
		case inLineComment:
			if c == '\n' {
				inLineComment = false
				out.WriteByte('\n')
			}
			i++
			continue
		case inBlockComment:
			if c == '*' && i+1 < len(b) && b[i+1] == '/' {
				inBlockComment = false
				i += 2
				continue
			}
			if c == '\n' {
				out.WriteByte('\n')
			}
			i++
			continue
		case inString:
			out.WriteByte(c)
			if c == '\\' && i+1 < len(b) {
				out.WriteByte(b[i+1])
				i += 2
				continue
			}
			if c == '"' {
				inString = false
			}
			i++
			continue
		}
		// 文字列外
		if c == '"' {
			inString = true
			out.WriteByte(c)
			i++
			continue
		}
		if c == '/' && i+1 < len(b) && b[i+1] == '/' {
			inLineComment = true
			i += 2
			continue
		}
		if c == '/' && i+1 < len(b) && b[i+1] == '*' {
			inBlockComment = true
			i += 2
			continue
		}
		out.WriteByte(c)
		i++
	}
	return removeTrailingCommas(out.Bytes())
}

func removeTrailingCommas(b []byte) []byte {
	var out bytes.Buffer
	n := len(b)
	for i := 0; i < n; i++ {
		if b[i] == ',' {
			j := i + 1
			for j < n && (b[j] == ' ' || b[j] == '\t' || b[j] == '\n' || b[j] == '\r') {
				j++
			}
			if j < n && (b[j] == '}' || b[j] == ']') {
				continue // カンマをスキップ
			}
		}
		out.WriteByte(b[i])
	}
	return out.Bytes()
}

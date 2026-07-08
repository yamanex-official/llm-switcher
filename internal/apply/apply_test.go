package apply

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyEnvFileMerge(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	os.WriteFile(p, []byte("EXISTING=1\n# comment\nFOO=bar\n"), 0o644)

	results := Apply(Options{EnvFile: true}, nil,
		[]EnvVar{{Name: "OPENAI_BASE_URL", Value: "https://x"}, {Name: "FOO", Value: "baz"}},
		dir)
	if len(results) != 1 || results[0].Error != nil {
		t.Fatalf("apply failed: %+v", results)
	}
	b, _ := os.ReadFile(p)
	got := string(b)
	if !strings.Contains(got, "OPENAI_BASE_URL=https://x") {
		t.Fatalf(".env missing new var:\n%s", got)
	}
	if !strings.Contains(got, "FOO=baz") {
		t.Fatalf(".env FOO not updated:\n%s", got)
	}
	if !strings.Contains(got, "EXISTING=1") {
		t.Fatalf(".env existing var lost:\n%s", got)
	}
}

func TestApplyProfileMarkerNonDestructive(t *testing.T) {
	dir := t.TempDir()
	p, err := detectProfilePath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(p, []byte("echo hello\nMY=own\n"), 0o644)

	Apply(Options{Profile: true}, nil,
		[]EnvVar{{Name: "OPENAI_BASE_URL", Value: "https://x"}}, dir)

	b, _ := os.ReadFile(p)
	got := string(b)
	if !strings.Contains(got, "echo hello") || !strings.Contains(got, "MY=own") {
		t.Fatalf("profile outer content destroyed:\n%s", got)
	}
	if !strings.Contains(got, markerStart) || !strings.Contains(got, markerEnd) {
		t.Fatalf("marker block missing:\n%s", got)
	}
	if !strings.Contains(got, "export OPENAI_BASE_URL=https://x") {
		t.Fatalf("managed var missing:\n%s", got)
	}

	// 再適用でブロックが更新されること（外側は維持）
	Apply(Options{Profile: true}, nil,
		[]EnvVar{{Name: "OPENAI_BASE_URL", Value: "https://y"}}, dir)
	b, _ = os.ReadFile(p)
	got = string(b)
	if strings.Count(got, markerStart) != 1 {
		t.Fatalf("marker duplicated on re-apply:\n%s", got)
	}
	if !strings.Contains(got, "export OPENAI_BASE_URL=https://y") {
		t.Fatalf("managed var not updated:\n%s", got)
	}
}

func TestApplyRollbackOnConfigFailure(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	os.WriteFile(envPath, []byte("A=1\n"), 0o644)

	fail := false
	configWrite := func() error {
		fail = true
		if fail {
			return os.ErrPermission
		}
		return nil
	}
	results := Apply(Options{ConfigFile: true, EnvFile: true}, configWrite,
		[]EnvVar{{Name: "OPENAI_BASE_URL", Value: "https://x"}}, dir)

	// config 失敗 → rollback で .env も元に戻る
	b, _ := os.ReadFile(envPath)
	if string(b) != "A=1\n" {
		t.Fatalf("env file not rolled back: %q", string(b))
	}
	_ = results
}

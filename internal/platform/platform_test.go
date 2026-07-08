package platform

import (
	"runtime"
	"strings"
	"testing"
)

func TestCLIDirsCrossPlatform(t *testing.T) {
	dirs := CLIDirs()
	for _, k := range []string{"claude", "codex", "opencode"} {
		if dirs[k] == "" {
			t.Fatalf("missing dir for %s", k)
		}
	}
	if runtime.GOOS != "windows" {
		if dirs["opencode"] == "" || !strings.HasSuffix(dirs["opencode"], ".config/opencode") {
			t.Fatalf("opencode dir unexpected: %s", dirs["opencode"])
		}
	}
}

func TestAppConfigDir(t *testing.T) {
	d, err := AppConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if d == "" {
		t.Fatal("empty config dir")
	}
	if runtime.GOOS == "windows" {
		// %APPDATA%\llm-router
		if !strings.Contains(d, "llm-router") {
			t.Fatalf("windows config dir unexpected: %s", d)
		}
	}
}

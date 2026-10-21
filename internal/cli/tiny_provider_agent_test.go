package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/testenv"
)

func TestTinyProviderEditsKeepAgentModelUsable(t *testing.T) {
	home := t.TempDir()
	testenv.SetHome(t, home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))
	path := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"model":"sonnet","theme":"dark"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"fixture", "backup"} {
		if err := provider.Save(provider.Provider{ID: id, Name: id, Key: "fixture", Chat: "http://127.0.0.1:1/v1", Models: []string{"m"}}); err != nil {
			t.Fatal(err)
		}
	}
	a, err := agent.Find("claude")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Pick("model", "fixture/m"); err != nil {
		t.Fatal(err)
	}
	moved, err := renameProvider("fixture", "renamed")
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Values()["model"]; got != "renamed/m" || len(moved) == 0 {
		t.Fatalf("rename left Agent on %q; moved=%v", got, moved)
	}
	changes, err := providerChange(func() error { return provider.SetOff("renamed", true) })
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Values()["model"]; got != "backup/m" || len(changes) == 0 {
		t.Fatalf("disabled provider left Agent on %q; moved=%v", got, changes)
	}
	if err := a.Disconnect(); err != nil {
		t.Fatal(err)
	}
	if got := a.Values()["model"]; got != "sonnet" {
		t.Fatalf("disconnect restored %q, want sonnet", got)
	}
}

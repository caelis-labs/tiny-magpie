package gui

import (
	"path/filepath"
	"testing"
"time"

	"github.com/yetone/magpie/internal/agentenv"
)

// These fixtures are shared by retained provider, key and GUI tests.
type folderOnly struct{ Windows }

func (folderOnly) OpenFolder(string) error { return nil }

// sandboxHome gives the test a machine of its own: no agents, no magpie files.
func sandboxHome(t *testing.T) string {
	t.Helper()
	h := t.TempDir()
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(h, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(h, ".cache"))
	t.Setenv("PATH", "")
	for _, v := range agentenv.Vars {
		t.Setenv(v, "")
	}
	t.Setenv("APPDATA", "")
	t.Setenv("LOCALAPPDATA", "")
	return h
}

func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		if ok() {
			return
		}
	}
	t.Fatal(what)
}

package gui

// Shared sandbox support for Tiny boundary tests.
import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/agentenv"
)

func builtinHome(t *testing.T) http.Handler {
	h := t.TempDir()
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(h, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(h, ".cache"))
	return Handler(nil, nil)
}

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

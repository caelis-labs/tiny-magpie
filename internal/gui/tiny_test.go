package gui

import (
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/provider"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTinyManagementSurface(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	h := Handler(nil, nil)
	for _, path := range []string{"/api/agents/cli/codex", "/api/agents/install", "/api/sessions/resume", "/api/profile/save", "/api/backup/export", "/api/library", "/api/gateway/take-over", "/api/codex/daemon/restart", "/api/usage/workbuddy-checkin"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{}`)))
		if w.Code != 404 {
			t.Errorf("%s: %d %s", path, w.Code, w.Body)
		}
	}
	for _, native := range []bool{false, true} {
		if on, why := gatewayMode(native); !on || why != "tiny-magpie" {
			t.Fatalf("GUI can leave gateway mode: %v %q", on, why)
		}
	}
}

func TestTinyProviderEditLeavesAgentConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, ".config"))
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	path := filepath.Join(root, ".codex", "config.toml")
	os.MkdirAll(filepath.Dir(path), 0700)
	before := []byte("model_provider = \"magpie\"\nmodel = \"fixture/m\"\n[model_providers.magpie]\nbase_url = \"http://127.0.0.1:3425/v1\"\nwire_api = \"responses\"\n")
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	if err := provider.Save(provider.Provider{ID: "fixture", Name: "fixture", Chat: "http://127.0.0.1:1/v1", Models: []string{"m"}}); err != nil {
		t.Fatal(err)
	}
	h := Handler(nil, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/provider/delete", strings.NewReader(`{"id":"fixture"}`)))
	if w.Code != 200 {
		t.Fatalf("remove: %d %s", w.Code, w.Body)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatalf("provider edit changed Agent config: %v %q", err, after)
	}
}

func TestTinyDoesNotAttachToAnotherGateway(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"name":"magpie","version":"99.0.0","models":42,"window":true}`))
	}))
	defer other.Close()
	t.Setenv("MAGPIE_ADDR", strings.TrimPrefix(other.URL, "http://"))
	before := served.Swap(nil)
	defer served.Store(before)
	state := providersState()
	if state.Gateway.Running || state.Gateway.Mine || state.Gateway.Window || !state.Gateway.Taken {
		t.Fatalf("attached to other gateway: %+v", state.Gateway)
	}
	// Trace must not attach to the other process either.
	mux := http.NewServeMux()
	traceRoutes(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/gateway/trace", nil))
	if w.Code != 200 || strings.Contains(w.Body.String(), `"mine":true`) {
		t.Fatalf("attached trace: %d %s", w.Code, w.Body)
	}
	if !gateway.Running() {
		t.Fatal("other gateway was stopped")
	}
}

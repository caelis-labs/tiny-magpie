package gateway

import (
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTinyRemovedSurface(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	h := New().Handler()
	for _, path := range []string{"/v1/images/generations", "/images/edits", "/v1/videos", "/v1/audio/transcriptions", "/mcp/any", "/mcp/magpie/tools", "/alpha/search"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{}`)))
		if w.Code != http.StatusNotFound {
			t.Errorf("%s: %d %s", path, w.Code, w.Body)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	for _, removed := range []string{"images/", "audio/", "/mcp/"} {
		if strings.Contains(w.Body.String(), removed) {
			t.Errorf("advertised removed API %s", removed)
		}
	}
}

func TestTinyCopiedSettingsCannotRecordConversations(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := settings.SetGatewayConversations(true); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/v1/responses", nil)
	r.Header.Set(SessionHeader, "fixture-session")
	w := httptest.NewRecorder()
	wrapped, done := recordConversation(w, r, provider.Responses, []byte(`{"input":"private fixture"}`))
	done()
	if wrapped != w {
		t.Fatal("copied upstream recording consent enabled body capture")
	}
}

func TestTinyImportedNestedIntentGroup(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	dir := filepath.Join(root, "tiny-magpie")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	config := `{"noAutoGroups":true,"providers":[{"id":"fixture","name":"Fixture","key":"fixture","chat":"http://127.0.0.1:1/v1","models":["m"]}],"groups":[{"id":"inner","name":"Inner","members":["fixture/m"],"classifier":"fixture/m"},{"id":"outer","name":"Outer","members":["group/inner"]}]}`
	if err := os.WriteFile(filepath.Join(dir, "providers.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"inner", "outer"} {
		w := httptest.NewRecorder()
		New().Handler().ServeHTTP(w, httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(`{"model":"group/`+id+`","messages":[{"role":"user","content":"hello"}]}`)))
		if w.Code != 400 || !strings.Contains(w.Body.String(), "intent routing is not available") {
			t.Fatalf("%s: %d %s", id, w.Code, w.Body)
		}
	}
}

package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/provider"
)

// Exercise the real upstream writer through Tiny's public management API,
// including preservation of unrelated settings and restoration on disconnect.
func TestTinyAgentModelAndRestore(t *testing.T) {
	home := sandboxHome(t)
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	path := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"theme":"dark","model":"sonnet","env":{"MY_SETTING":"keep"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := provider.Save(provider.Provider{ID: "fixture", Name: "Fixture", Key: "fixture", Chat: "http://127.0.0.1:1/v1", Models: []string{"m", "other"}}); err != nil {
		t.Fatal(err)
	}
	h := Handler(nil, nil)
	call := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(body)))
		return w
	}
	for _, m := range []string{"fixture/m", "fixture/other"} {
		w := call("/api/set", `{"agent":"claude","field":"model","value":"`+m+`"}`)
		if w.Code != 200 {
			t.Fatalf("pick: %d %s", w.Code, w.Body)
		}
		if v, _ := edit.GetJSON(path, "model"); v != m {
			t.Fatalf("model %q", v)
		}
	}
	a, err := agent.Find("claude")
	if err != nil {
		t.Fatal(err)
	}
	fields := agentFields(a, a.Values())
	b, _ := json.Marshal(fields)
	for _, f := range fields {
		if f.Label == "sign-in" || f.Label == "memories" || f.Label == "auto mode" || f.Label == "subagents" {
			t.Fatalf("complex field exposed: %s", b)
		}
	}
	for _, f := range []string{"login", "auto_mode", "subagent"} {
		if w := call("/api/set", `{"agent":"claude","field":"`+f+`","value":"fixture/m"}`); w.Code != http.StatusBadRequest {
			t.Fatalf("%s: %d %s", f, w.Code, w.Body)
		}
	}
	w := call("/api/agents/disconnect/claude", `{}`)
	if w.Code != 200 {
		t.Fatalf("disconnect: %d %s", w.Code, w.Body)
	}
	for key, want := range map[string]string{"theme": "dark", "model": "sonnet", "env.MY_SETTING": "keep", "env.ANTHROPIC_BASE_URL": ""} {
		if got, _ := edit.GetJSON(path, key); got != want {
			t.Errorf("%s: %q, want %q", key, got, want)
		}
	}
}

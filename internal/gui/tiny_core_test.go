package gui

import (
	"net/http/httptest"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
)

func TestTinyClassifiersAndRetrievalInventory(t *testing.T) {
	builtinHome(t)
	s := providersState()
	found := false
	for _, p := range s.Presets {
		if p.ID == "typesafe" {
			found = true
		}
	}
	if !found {
		t.Fatal("TypeSafe missing from Add provider")
	}
	p := provider.Provider{ID: "fixture", Name: "Fixture", Chat: "http://127.0.0.1:1/v1", Key: "fixture"}
	if err := catalog.SaveLive(p.ID, p.Chat, []catalog.Model{{ID: "chat", Name: "Chat"}, {ID: "vector", Name: "Vector", Retrieval: "embedding"}, {ID: "rerank-1", Name: "Rank", Retrieval: "rerank"}}); err != nil {
		t.Fatal(err)
	}
	info := providerInfo(p, nil)
	if len(info.Retrievals) != 2 || len(info.Models) != 1 || info.Models[0].ID != "chat" {
		t.Fatalf("mixed model capabilities: %+v %+v", info.Models, info.Retrievals)
	}
}

func TestTinyRetiredSettingsHaveNoAPI(t *testing.T) {
	h := builtinHome(t)
	for _, path := range []string{"/api/privacy-model", "/api/masking", "/api/settings/redact-rules", "/api/settings/redact-rule", "/api/settings/codex-credits", "/api/plugins/builtin"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 404 {
			t.Errorf("%s: %d %s", path, w.Code, w.Body)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/fonts", nil))
	if w.Code != 404 {
		t.Fatalf("fonts: %d", w.Code)
	}
}

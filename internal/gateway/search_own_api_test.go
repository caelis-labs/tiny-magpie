package gateway

import (
	"strings"

	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// A request that offers web search, at a provider that doesn't search by
// itself but serves the client's own API, is translated on that API, not
// on Chat (#997, wlj521: Codex on Xiaomi MiMo's pay as you go, whose
// /v1/responses is served, showed "Responses → Chat" — Codex offers its
// web_search on every turn). The provider is the preset's, the three URLs
// of its pay as you go region; the search tool, which the provider would
// turn away or ignore, isn't sent to it.
func TestSearchOfferedStaysOnTheClientsAPI(t *testing.T) {
	srv, seen := relayOfEvery(t)
	fresh(t)
	if err := provider.Save(provider.Provider{ID: "xiaomi", Name: "Xiaomi MiMo", Key: "k", Preset: "xiaomi",
		Chat: srv.URL + "/v1", Responses: srv.URL + "/v1", Anthropic: srv.URL,
		Models: []string{"mimo-v2.6-flash"}}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []struct{ name, path, body, want string }{
		{"Codex", "/v1/responses", `{"model":"xiaomi/mimo-v2.6-flash","stream":true,"store":false,"reasoning":{"effort":"medium"},"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]}],
			"tools":[{"type":"function","name":"shell","description":"run","parameters":{"type":"object","properties":{}}},{"type":"web_search","external_web_access":false}]}`, "/v1/responses"},
		{"Claude Code", "/v1/messages", `{"model":"xiaomi/mimo-v2.6-flash","max_tokens":100,"stream":true,"messages":[{"role":"user","content":"hi"}],
			"tools":[{"name":"Read","description":"read","input_schema":{"type":"object"}},{"type":"web_search_20250305","name":"web_search","max_uses":8}]}`, "/v1/messages"},
		{"Chat", "/v1/chat/completions", `{"model":"xiaomi/mimo-v2.6-flash","stream":true,"messages":[{"role":"user","content":"hi"}],"web_search_options":{}}`, "/v1/chat/completions"},
	} {
		seen()
		if code, b := post(t, r.path, r.body); code != 200 {
			t.Fatalf("%s: %d %s", r.name, code, b)
		}
		paths, bodies := seen()
		if len(paths) != 1 || paths[0] != r.want {
			t.Errorf("%s with web search offered went to %v, want %s", r.name, paths, r.want)
			continue
		}
		if strings.Contains(bodies[0], "web_search") {
			t.Errorf("%s: the search tool reached a provider that doesn't search: %s", r.name, bodies[0])
		}
	}
}

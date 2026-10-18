package gateway

import (
	"net/http"

	"testing"
)

// A call Codex makes on a hidden thread of its own — the title of a new
// chat — comes without x-openai-subagent; its turn metadata names the
// thread's source (#314). magpie's own web search says so too.
func TestCallKindThreadSource(t *testing.T) {
	for want, h := range map[string]map[string]string{
		"thread_title":    {"x-codex-turn-metadata": `{"session_id":"s","thread_source":"thread_title","turn_id":"t"}`},
		"guardian_review": {"x-codex-turn-metadata": `{"thread_source":"guardian_review"}`},
		"guardian":        {"x-openai-subagent": "guardian", "x-codex-turn-metadata": `{"thread_source":"guardian_review"}`},
		"web_search":      {"User-Agent": SearchAgent},
		"":                {"x-codex-turn-metadata": `{"thread_source":"user","turn_id":"t"}`, "User-Agent": "codex_cli_rs/0.159.2"},
	} {
		hh := http.Header{}
		for k, v := range h {
			hh.Set(k, v)
		}
		if got := callKind(hh); got != want {
			t.Errorf("%v: %q, want %q", h, got, want)
		}
	}
	for _, meta := range []string{"", "{", `{"thread_source":"subagent"}`, `{"turn_id":"t"}`} {
		if got := threadSource(meta); got != "" {
			t.Errorf("threadSource(%q) = %q", meta, got)
		}
	}
}

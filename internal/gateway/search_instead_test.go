package gateway

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// pickedSearchSetup is a search for glm/glm-flash, which can't search, with
// ds/deepseek-flash, on DeepSeek's Responses API, the first searcher: one
// that answers without searching (blind), or one whose every search is
// refused (failing). next is another searcher that would search.
func pickedSearchSetup(t *testing.T, failing bool, nextOn provider.Protocol) (ds, next *fakeDeepSeek, result func() string) {
	t.Helper()
	searchSandbox(t)
	forget := func() {
		lastNamedFail.Lock()
		lastNamedFail.f = SearchFailure{}
		lastNamedFail.Unlock()
	}
	forget()
	t.Cleanup(forget)
	ds = newFakeDeepSeek(t)
	ds.blind = true
	first := ds.URL
	if failing {
		broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/models") {
				w.Write([]byte(`{"data":[{"id":"deepseek-flash"}]}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":{"message":"search quota used up"}}`))
		}))
		t.Cleanup(broken.Close)
		first = broken.URL
	}
	glm, result := newBlindModel(t)
	next = newFakeDeepSeek(t)
	searchesOn(t, provider.Responses, first)
	searchesOn(t, nextOn, next.URL)
	ds2 := provider.Provider{ID: "ds2", Name: "DeepSeek 2", Key: "k", Models: []string{"deepseek-flash"}}
	if nextOn == provider.Anthropic {
		ds2.Anthropic = next.URL + "/anthropic"
	} else {
		ds2.Responses = next.URL + "/v1"
	}
	save(t, provider.Provider{ID: "ds", Name: "DeepSeek", Key: "k", Responses: first + "/v1", Models: []string{"deepseek-flash"}},
		provider.Provider{ID: "glm", Name: "GLM", Key: "k", Chat: glm.URL, Models: []string{"glm-flash"}}, ds2)
	return ds, next, result
}

func searchForGLM(t *testing.T, s *Server) {
	t.Helper()
	body := `{"model":"glm/glm-flash","max_tokens":3000,"stream":true,"messages":[{"role":"user","content":"用 web_search 工具搜索：今天上海天气"}],"tools":[{"type":"web_search_20250305","name":"web_search"}]}`
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body)))
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

func TestTinySystemOnePassesThrough(t *testing.T) {
	fresh(t)
	var sent string
	vendor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/systemone" || r.Header.Get("Authorization") != "Bearer fixture-key" {
			t.Errorf("vendor route: %s %s", r.URL.Path, r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		sent = string(body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"model":"jev-latest","answers":{"kind":{"type":"choice","choice":"code","confidence":0.9}}}`)
	}))
	defer vendor.Close()
	if err := provider.Save(provider.Provider{ID: "typesafe", Name: "TypeSafe", Decide: vendor.URL + "/v1", Key: "fixture-key", Models: []string{"jev-latest"}}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/v1/systemone", "/systemone"} {
		code, raw := postRetrieval(t, New(), path, `{"model":"typesafe/jev-latest","state":{"text":"hello"},"questions":{"kind":{"type":"choice","criteria":{"code":null,"other":null}}}}`)
		if code != 200 || !strings.Contains(raw, `"choice":"code"`) || !strings.Contains(sent, `"model":"jev-latest"`) || !strings.Contains(sent, `"text":"hello"`) {
			t.Fatalf("%s: %d %s sent %s", path, code, raw, sent)
		}
	}
}

func TestTinyOldRedactionDoesNotTransformRequests(t *testing.T) {
	fresh(t)
	s := settings.Load()
	s.Redact = true
	s.RedactPersonal = true
	s.RedactWords = []string{"fixture-value"}
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	body := []byte(`{"input":"fixture-value person@example.test"}`)
	got, masked, done, err := redacted(w, body)
	done()
	if err != nil || got != w || string(masked) != string(body) {
		t.Fatalf("request changed: %s %v", masked, err)
	}
}

func TestTinyDoesNotKeepRequestBodies(t *testing.T) {
	s := New()
	s.trace.keepBody(1, provider.Chat, []byte(`{"messages":[{"role":"user","content":"private fixture"}]}`))
	if _, ok := s.trace.keptBodyOf(1); ok {
		t.Fatal("retained removed context-text body")
	}
}

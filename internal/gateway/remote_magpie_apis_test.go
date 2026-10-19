package gateway

import (
	"context"
	"encoding/json"

	"net/http/httptest"
	"slices"
	"strings"

	"testing"
	"time"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/usage"
)

// A remote's exposed retrieval models survive discovery, and their requests
// keep the input and options through both gateways.
func TestRemoteMagpieRetrieval(t *testing.T) {
	s, lib := shelved(t)
	remote := httptest.NewServer(s.Handler())
	t.Cleanup(remote.Close)
	if err := provider.Save(provider.Provider{ID: "office", Preset: provider.RemoteMagpiePreset, Chat: remote.URL, Key: "remote-key"}); err != nil {
		t.Fatal(err)
	}
	p, _ := provider.Find("office")
	if _, err := p.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	live, _ := catalog.LiveAll("office")
	if !slices.ContainsFunc(live, func(m catalog.Model) bool { return m.ID == "lib/embed-1" }) {
		t.Fatalf("remote embedding model lost: %v", live)
	}
	// The shared list has no separate retrieval kind, so these stay in
	// the local catalog just as they do in the remote's agent list.
	entries := provider.Catalog()
	for _, id := range []string{"office/lib/embed-1", "office/lib/rerank-1"} {
		if slices.ContainsFunc(entries, func(e provider.Entry) bool { return e.ID == id }) {
			t.Errorf("retrieval model leaked into chat catalog: %s", id)
		}
	}
	for _, tc := range []struct{ path, body, result, model string }{
		{"/v1/embeddings", `{"model":"office/lib/embed-1","input":["hello","world"],"dimensions":2,"encoding_format":"float"}`, `"embedding":[0.1,-0.2]`, "embed-1"},
		{"/v1/rerank", `{"model":"office/lib/rerank-1","query":"magpie","documents":["a crow","a magpie"],"top_n":2,"return_documents":true}`, `"relevance_score":0.9`, "rerank-1"},
	} {
		req := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
		req.Header.Set("User-Agent", "pi/1.0")
		rec := httptest.NewRecorder()
		New().Handler().ServeHTTP(rec, req)
		code, raw := rec.Code, rec.Body.String()
		if code != 200 || !strings.Contains(raw, tc.result) {
			t.Fatalf("%s: %d %s", tc.path, code, raw)
		}
		got := lib.got(tc.path)
		if len(got) != 1 || gjsonModel([]byte(got[0])) != tc.model {
			t.Fatalf("vendor request: %v", got)
		}
		var asked, sent map[string]any
		json.Unmarshal([]byte(tc.body), &asked)
		json.Unmarshal([]byte(got[0]), &sent)
		asked["model"] = tc.model
		a, _ := json.Marshal(asked)
		b, _ := json.Marshal(sent)
		if string(a) != string(b) {
			t.Errorf("retrieval options changed: %s; want %s", b, a)
		}
	}
	var recorded int
	for _, r := range usage.Load(time.Time{}) {
		if r.Provider == "lib" {
			recorded++
			if r.Agent != "pi" || r.Via != hostName() {
				t.Errorf("remote retrieval usage lost caller: %+v", r)
			}
		}
	}
	if recorded != 2 {
		t.Errorf("remote retrieval usage: %d records; want 2", recorded)
	}
}

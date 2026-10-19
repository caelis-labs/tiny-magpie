package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

func TestTinyAutomationToolCallsStayWithClient(t *testing.T) {
	fresh(t)
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	// No real OS tool may run even if this boundary regresses.
	bin, marker := t.TempDir(), filepath.Join(t.TempDir(), "unexpected-tool")
	for _, name := range []string{"screencapture", "osascript"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nprintf unexpected > '"+marker+"'\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	vendor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"fixture","object":"chat.completion","model":"fixture","choices":[{"index":0,"message":{"role":"assistant","content":null,"tool_calls":[{"id":"capture-1","type":"function","function":{"name":"screencapture","arguments":"{\"action\":\"capture\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
	}))
	defer vendor.Close()
	if err := provider.Save(provider.Provider{ID: "fixture", Name: "Fixture", Chat: vendor.URL + "/v1", Key: "fixture", Models: []string{"fixture"}}); err != nil {
		t.Fatal(err)
	}
	gw := httptest.NewServer(New().Handler())
	defer gw.Close()
	res, err := http.Post(gw.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"fixture/fixture","messages":[{"role":"user","content":"capture fixture"}],"tools":[{"type":"function","function":{"name":"screencapture","parameters":{"type":"object"}}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || !strings.Contains(string(body), `"name":"screencapture"`) || !strings.Contains(string(body), `"finish_reason":"tool_calls"`) {
		t.Fatalf("client did not receive its tool call: %d %s", res.StatusCode, body)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("gateway ran the client's OS tool: %v", err)
	}
	// The retained subscription bridge may only run vendor WebSearch itself.
	for _, web := range []bool{false, true} {
		args := claudeCLIArgs("fixture", `{}`, "", web)
		want := ""
		if web {
			want = "WebSearch"
		}
		i := slices.Index(args, "--tools")
		if i < 0 || i+1 >= len(args) || args[i+1] != want {
			t.Fatalf("subscription bridge enabled local tools: %v", args)
		}
	}
}

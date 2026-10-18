package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTinyRejectsPureToolExtension(t *testing.T) {
	sandbox(t)
	file := filepath.Join(t.TempDir(), "tools.js")
	if err := os.WriteFile(file, []byte(`export default async () => ({tool:{ping:{args:{},execute:async()=>"pong"}}})`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Add(t.Context(), file); err != nil {
		t.Fatal(err)
	}
	loaded, err := Plugins(t.Context())
	if err != nil || len(loaded) != 1 || !strings.Contains(loaded[0].Error, "tool-only plugins are not supported") {
		t.Fatalf("pure tool extension: %+v %v", loaded, err)
	}
	tools, err := Tools(t.Context())
	if err != nil || len(tools) != 0 {
		t.Fatalf("rejected tools exposed: %+v %v", tools, err)
	}
}

func TestTinyMarketKeepsGatewayExtensions(t *testing.T) {
	rows, err := parseMarket([]byte(`{"plugins":[
        {"package":"@fixture/provider","community":true},
        {"package":"@fixture/middleware","community":true,"kind":"middleware"},
        {"package":"@fixture/tools","community":true,"kind":"tools"},
        {"package":"@fixture/page","community":true,"kind":"page"},
        {"package":"@fixture/agent","community":true,"kind":"agent"}
    ]}`))
	if err != nil || len(rows) != 2 || rows[0].Package != "@fixture/provider" || rows[1].Package != "@fixture/middleware" {
		t.Fatalf("gateway market: %+v %v", rows, err)
	}
}

func TestTinyRejectsPureAgentExtension(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "fixture.agent.js")
	if err := os.WriteFile(file, []byte("export default {}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := notPlugin(file); err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("pure agent file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"main":"index.js","magpie":{"agent":"fixture.agent.js"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := notPlugin(dir); err != nil {
		t.Fatalf("mixed provider rejected: %v", err)
	}
}

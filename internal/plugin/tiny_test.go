package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

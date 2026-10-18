package plugin

import (
	"path/filepath"
	"testing"
)

// A mixed package retains its Provider facet without advertising tools on
// Tiny's removed public MCP surface. Pure tool rejection is tested separately.
func TestMixedProviderToolsStayPrivate(t *testing.T) {
	sandbox(t)
	abs, err := filepath.Abs("testdata/tools")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Add(t.Context(), abs); err != nil {
		t.Fatal(err)
	}
	providers, err := Providers(t.Context())
	if err != nil || len(providers) != 1 || providers[0].ID != "tools-fixture" {
		t.Fatalf("mixed package lost its Provider: %+v %v", providers, err)
	}
	tools, err := Tools(t.Context())
	if err != nil || len(tools) != 0 {
		t.Fatalf("mixed package advertised removed tools: %+v %v", tools, err)
	}
}

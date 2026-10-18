package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/plugin"
	"github.com/yetone/magpie/internal/testenv"
)

// Standalone pages fail at installation; standalone tools fail when the
// host discovers their hooks. Neither advertises a usable extension.
func TestPluginListRejectsStandaloneToolsAndPage(t *testing.T) {
	bun := testenv.Bun(t)
	groupsHome(t)
	t.Setenv("MAGPIE_BUN", bun)
	t.Cleanup(plugin.Settle)
	dir := t.TempDir()
	tools, page := filepath.Join(dir, "tools.js"), filepath.Join(dir, "page")
	if err := os.WriteFile(tools, []byte(`export default async () => ({tool:{ping:{args:{},execute:async()=>"pong"}}})`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(page, os.DirFS("../gui/testdata/page-plugin")); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.Add(context.Background(), tools); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.Add(context.Background(), page); err == nil || !strings.Contains(err.Error(), "page plugins are not supported") {
		t.Fatalf("standalone page: %v", err)
	}
	out, err := stdoutOf(t, func() error { return listPlugins(context.Background(), false) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "tool-only plugins are not supported") {
		t.Errorf("the rejected extension's reason is missing:\n%s", out)
	}
	if strings.Contains(out, "_ping") || strings.Contains(out, "Usage glance") || strings.Contains(out, page) {
		t.Errorf("a removed extension is advertised:\n%s", out)
	}
	if err := pluginCmd([]string{"plugin", "off", tools}); err != nil {
		t.Fatal(err)
	}
	out, err = stdoutOf(t, func() error { return listPlugins(context.Background(), false) })
	if err != nil || strings.Contains(out, "tool-only plugins are not supported") {
		t.Errorf("disabled extension still loads: %s, %v", out, err)
	}
}

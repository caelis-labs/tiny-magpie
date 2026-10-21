package agentplug

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/plugin"
)

func TestTinyMixedPluginDoesNotLoadAgentFacet(t *testing.T) {
	home := t.TempDir()
	for _, key := range []string{"HOME", "USERPROFILE"} {
		t.Setenv(key, home)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	dir := t.TempDir()
	for name, body := range map[string]string{
		"package.json":     `{"main":"index.js","magpie":{"agent":"fixture.agent.js","hooks":"fixture.hooks.js"}}`,
		"fixture.agent.js": `throw new Error("Agent facet must not run");`,
		"fixture.hooks.js": `export async function beforeRequest() {}`,
		"index.js":         `export default async () => ({})`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	path := plugin.ListPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(plugin.List{Plugins: []plugin.Entry{{Spec: dir}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	Reload()
	t.Cleanup(Reload)
	list, errs := List()
	if len(list) != 0 || len(errs) != 0 {
		t.Fatalf("unsupported Agent facet loaded: agents=%v errors=%v", list, errs)
	}
	if got := plugin.Gateway(dir); got != filepath.Join(dir, "fixture.hooks.js") {
		t.Fatalf("retained gateway hook missing: %q", got)
	}
}

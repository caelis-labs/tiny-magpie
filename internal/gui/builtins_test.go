package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/pluginrt"

	_ "github.com/yetone/magpie/plugins/all"
)

func builtinHome(t *testing.T) http.Handler {
	h := t.TempDir()
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(h, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(h, ".cache"))
	return Handler(nil, nil)
}

func postBuiltin(h http.Handler, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/plugins/builtin", strings.NewReader(body)))
	return rec
}

// A plugins.yml that doesn't parse leaves every plugin on, says so, and
// no switch writes over it.
func TestBuiltinBrokenFile(t *testing.T) {
	h := builtinHome(t)
	if err := os.MkdirAll(filepath.Dir(pluginrt.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pluginrt.Path(), []byte("plugins: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, b := range builtinsState() {
		if b.Broken == "" {
			t.Fatalf("not told broken: %+v", b)
		}
	}
	if rec := postBuiltin(h, `{"id":"library","on":false}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("switched over a broken file: %d", rec.Code)
	}
	if b, _ := os.ReadFile(pluginrt.Path()); string(b) != "plugins: [\n" {
		t.Fatalf("written over: %q", b)
	}
}

// A plugin's settings are set only as their schema allows, kept in
// plugins.yml, and handed to the page in boot.js; a request says on or
// config, not both.
func TestBuiltinConfig(t *testing.T) {
	h := builtinHome(t)
	if _, err := pluginrt.Settle(); err != nil {
		t.Fatal(err)
	}
	gw := func() builtinJSON {
		for _, b := range builtinsState() {
			if b.ID == "gateway" {
				return b
			}
		}
		t.Fatal("no gateway plugin")
		return builtinJSON{}
	}
	if b := gw(); b.Schema == nil || b.Schema.Properties["callsPage"] == nil || b.Config["callsPage"] != int64(20) {
		t.Fatalf("before: %+v", b)
	}
	if rec := postBuiltin(h, `{"id":"gateway","config":{"callsPage":50}}`); rec.Code != http.StatusOK {
		t.Fatalf("50: %d %s", rec.Code, rec.Body)
	}
	if b := gw(); b.Config["callsPage"] != int64(50) || !b.On {
		t.Fatalf("after 50: %+v", b)
	}
	for _, bad := range []string{
		`{"id":"gateway","config":{"callsPage":30}}`,
		`{"id":"gateway","config":{"other":1}}`,
		`{"id":"gateway","on":true,"config":{"callsPage":10}}`,
		`{"id":"gateway"}`,
		`{"id":"library","config":{"callsPage":10}}`,
	} {
		if rec := postBuiltin(h, bad); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", bad, rec.Code, rec.Body)
		}
	}
	if b := gw(); b.Config["callsPage"] != int64(50) {
		t.Fatalf("a refused one written: %+v", b)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/boot.js", nil))
	if !strings.Contains(rec.Body.String(), `"pluginConfig":{"gateway":{"callsPage":50}}`) {
		t.Fatalf("boot.js: %s", rec.Body)
	}
}

// Every file a built-in plugin says it has is on disk, loaded by the page
// and named with its hash so a new one isn't read from the cache: a script
// after the kernel and before app.js, which looks for its pages, a
// stylesheet in the head with the others.
func TestBuiltinClientsLoaded(t *testing.T) {
	page, err := os.ReadFile("assets/index.html")
	if err != nil {
		t.Fatal(err)
	}
	src := string(page)
	at := func(name string) int {
		tag := `<script src="` + name + `"></script>`
		if strings.HasSuffix(name, ".css") {
			tag = `<link rel="stylesheet" href="` + name + `">`
		}
		i := strings.Index(src, tag)
		if i < 0 {
			t.Fatalf("index.html doesn't load %s", name)
		}
		return i
	}
	kernel, app, start := at("kernel.js"), at("app.js"), at("start.js")
	if kernel > app || app > start {
		t.Fatal("kernel.js, app.js and start.js out of order")
	}
	n := 0
	for _, m := range pluginrt.Registered() {
		for _, c := range m.Client {
			name := "plugins/" + m.ID + "/" + c
			if _, err := os.Stat(filepath.Join("assets", name)); err != nil {
				t.Errorf("%s: %v", name, err)
			}
			attr := "src"
			if strings.HasSuffix(name, ".css") {
				attr = "href"
				if i := at(name); i > strings.Index(src, "</head>") {
					t.Errorf("%s isn't in the head", name)
				}
			} else if c == "page.js" {
				// a page's code that runs on app.js's helpers: after app.js,
				// before start.js shows the first page
				if i := at(name); i < app || i > start {
					t.Errorf("%s isn't between app.js and start.js", name)
				}
			} else if i := at(name); i < kernel || i > app {
				t.Errorf("%s isn't between kernel.js and app.js", name)
			}
			if !strings.Contains(string(versionedPage(page)), attr+`="`+name+`?v=`) {
				t.Errorf("%s isn't named with its hash", name)
			}
			n++
		}
	}
	if n == 0 {
		t.Fatal("no plugin has a script")
	}
}

// Every string of every built-in plugin is said in every language.
func TestBuiltinsTranslated(t *testing.T) {
	src, err := os.ReadFile("assets/i18n.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range []string{"zh: {", `"zh-TW": {`, "ja: {", "de: {"} {
		_, block, ok := strings.Cut(string(src), "\n  "+lang+"\n")
		if !ok {
			t.Fatalf("no %s block", lang)
		}
		block, _, _ = strings.Cut(block, "\n  },")
		for _, m := range pluginrt.Registered() {
			says := []string{m.Name, m.Sub, m.One, m.Many}
			if sc := m.Schema(); sc != nil {
				for _, p := range sc.Properties {
					says = append(says, p.Title, p.Description)
				}
			}
			for _, s := range says {
				if s == "" {
					continue
				}
				k, _ := json.Marshal(s)
				if !strings.Contains(block, "\n    "+string(k)+": ") {
					t.Errorf("%s doesn't say %s", lang, k)
				}
			}
		}
	}
}

func TestBuiltinSwitch(t *testing.T) {
	h := builtinHome(t)
	if _, err := pluginrt.Settle(); err != nil {
		t.Fatal(err)
	}
	if len(builtinsState()) != 8 {
		t.Fatalf("assembly: %+v", builtinsState())
	}
	if err := pluginrt.SetPreset(pluginrt.PresetEssentials); err != nil {
		t.Fatal(err)
	}
	if w := postBuiltin(h, `{"id":"groups","on":true}`); w.Code != 200 || !pluginrt.Running().Active("groups") {
		t.Fatalf("groups: %d %s", w.Code, w.Body)
	}
	for _, id := range []string{"library", "sessions", "search", "tools", "warmups", "otel", "nope"} {
		if w := postBuiltin(h, `{"id":"`+id+`","on":true}`); w.Code != 400 {
			t.Fatalf("resurrected %s", id)
		}
	}
}

func TestBootSaysBuiltinsOff(t *testing.T) {
	h := builtinHome(t)
	for _, preset := range []string{pluginrt.PresetEssentials, pluginrt.PresetAll} {
		if err := pluginrt.SetPreset(preset); err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/boot.js", nil))
		for _, id := range []string{"library", "sessions", "search", "tools", "warmups", "otel", "tiny-removed"} {
			if !strings.Contains(w.Body.String(), `"`+id+`"`) {
				t.Fatalf("can resurrect %s: %s", id, w.Body)
			}
		}
	}
}

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

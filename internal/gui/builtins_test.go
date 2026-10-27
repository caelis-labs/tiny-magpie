package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/gateway"
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

// boot.js says where the gateway listens when this process serves it, so
// the sidebar's line is drawn before the first paint (#3429), and nothing
// when it doesn't: one served by another magpie is the providers' read to
// find, never this process's guess.
func TestBootSaysWhereItsGatewayListens(t *testing.T) {
	h := builtinHome(t)
	t.Setenv("MAGPIE_ADDR", "127.0.0.1:3481")
	boot := func() string {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/boot.js", nil))
		return rec.Body.String()
	}
	prev := served.Swap(nil)
	t.Cleanup(func() { served.Store(prev) })
	if b := boot(); strings.Contains(b, "sideGateway") {
		t.Fatalf("no gateway served here: %s", b)
	}
	served.Store(gateway.New())
	if b := boot(); !strings.Contains(b, `"sideGateway":{"boot":true,"running":true,"url":"http://127.0.0.1:3481"}`) {
		t.Fatalf("served here: %s", b)
	}
	t.Setenv("MAGPIE_ADDR", "0.0.0.0:3483")
	if b := boot(); !strings.Contains(b, `"sideGateway":{"boot":true,"running":true,"url":"http://127.0.0.1:3483"}`) {
		t.Fatalf("every interface, reached on loopback: %s", b)
	}
	// MAGPIE_PUBLIC_URL goes with it while other machines are let in, for
	// magpie web's page to show (noahhrreion on Discord, in Docker), and
	// not while the gateway is this computer's alone
	t.Setenv("MAGPIE_PUBLIC_URL", "http://nas.local:3425")
	if b := boot(); !strings.Contains(b, `"sideGateway":{"boot":true,"publicURL":"http://nas.local:3425","running":true,"url":"http://127.0.0.1:3483"}`) {
		t.Fatalf("on the network, with MAGPIE_PUBLIC_URL: %s", b)
	}
	t.Setenv("MAGPIE_ADDR", "127.0.0.1:3483")
	if b := boot(); strings.Contains(b, "nas.local") {
		t.Fatalf("on loopback, not shared: %s", b)
	}
	t.Setenv("MAGPIE_PUBLIC_URL", "")
	served.Store(nil)
	if b := boot(); strings.Contains(b, "sideGateway") {
		t.Fatalf("stopped: %s", b)
	}
}

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

package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
)

// searchAPIs is a fake of every search API magpie asks, each on its own
// path and checking its own way of carrying the key; tavily answers 500
// while broken is set.
type searchAPIs struct {
	*httptest.Server
	mu     sync.Mutex
	asked  []string // vendor: query
	broken bool
}

func newSearchAPIs(t *testing.T) *searchAPIs {
	f := &searchAPIs{}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var q struct {
			Query string `json:"query"`
		}
		json.Unmarshal(b, &q)
		vendor, key := "", ""
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/tavily/search":
			vendor, key = "tavily", r.Header.Get("Authorization")
		case r.Method == http.MethodGet && r.URL.Path == "/brave/res/v1/web/search":
			vendor, key, q.Query = "brave", r.Header.Get("X-Subscription-Token"), r.URL.Query().Get("q")
		case r.Method == http.MethodPost && r.URL.Path == "/exa/search":
			vendor, key = "exa", r.Header.Get("X-Api-Key")
		case r.Method == http.MethodPost && r.URL.Path == "/firecrawl/v2/search":
			vendor, key = "firecrawl", r.Header.Get("Authorization")
		case r.Method == http.MethodGet && r.URL.Path == "/searxng/search" && r.URL.Query().Get("format") == "json":
			vendor, key, q.Query = "searxng", "none", r.URL.Query().Get("q")
		default:
			http.NotFound(w, r)
			return
		}
		want := map[string]string{"tavily": "Bearer tvly-k", "brave": "brave-k", "exa": "exa-k", "firecrawl": "Bearer fc-k", "searxng": "none"}[vendor]
		if key != want {
			http.Error(w, `{"error":"bad key `+key+`"}`, 401)
			return
		}
		f.mu.Lock()
		f.asked = append(f.asked, vendor+": "+q.Query)
		broken := f.broken
		f.mu.Unlock()
		if vendor == "tavily" && broken {
			http.Error(w, `{"detail":{"error":"down"}}`, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch vendor {
		case "tavily":
			io.WriteString(w, `{"query":"x","results":[{"title":"Go downloads","url":"https://go.dev/dl/","content":"go1.27.1 is the latest release.","score":0.9}]}`)
		case "brave":
			io.WriteString(w, `{"type":"search","web":{"results":[{"title":"Go &amp; <strong>downloads</strong>","url":"https://go.dev/dl/","description":"The <strong>latest</strong> Go is 1.27.1"}]}}`)
		case "exa":
			io.WriteString(w, `{"results":[{"title":"Go downloads","url":"https://go.dev/dl/","text":"Go 1.27.1 was released."}]}`)
		case "firecrawl":
			io.WriteString(w, `{"success":true,"data":{"web":[{"title":"Go downloads","url":"https://go.dev/dl/","description":"Download Go 1.27.1"}]}}`)
		case "searxng":
			io.WriteString(w, `{"query":"x","results":[{"title":"Go downloads","url":"https://go.dev/dl/","content":"Go 1.27.1 is out"},{"title":"","url":""}]}`)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

func (f *searchAPIs) api(vendor, key string) provider.SearchAPI {
	return provider.SearchAPI{Vendor: vendor, Key: key, URL: f.URL + "/" + vendor}
}

// Each search API is asked its own way and read into the pages it found.
func TestSearchAPIsEachAsked(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := newSearchAPIs(t)
	s := New()
	for _, c := range []struct {
		api  provider.SearchAPI
		text string
	}{
		{f.api("tavily", "tvly-k"), "go1.27.1 is the latest release."},
		{f.api("brave", "brave-k"), "The latest Go is 1.27.1"},
		{f.api("exa", "exa-k"), "Go 1.27.1 was released."},
		{f.api("firecrawl", "fc-k"), "Download Go 1.27.1"},
		{f.api("searxng", ""), "Go 1.27.1 is out"},
	} {
		pages, err := s.askSearchAPI(context.Background(), c.api, "latest go")
		if err != nil {
			t.Errorf("%s: %v", c.api.Vendor, err)
			continue
		}
		if len(pages) != 1 || pages[0].URL != "https://go.dev/dl/" || pages[0].Text != c.text || !strings.HasPrefix(pages[0].Title, "Go ") {
			t.Errorf("%s: %+v", c.api.Vendor, pages)
		}
	}
	if pages, _ := s.askSearchAPI(context.Background(), f.api("brave", "brave-k"), "q"); len(pages) == 1 && pages[0].Title != "Go & downloads" {
		t.Errorf("brave's title kept its tags: %q", pages[0].Title)
	}
	if _, err := s.askSearchAPI(context.Background(), f.api("exa", "wrong"), "q"); err == nil || !strings.Contains(err.Error(), "bad key") {
		t.Errorf("a refused key: %v", err)
	}
}

// Without a provider that searches or a search API, the client's search
// tool is left out, as before.
func TestNoSearchAPINoTool(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	if canSearch() {
		t.Fatal("can search with nothing set up")
	}
	if err := provider.SetSearchAPI(provider.SearchAPI{Vendor: "brave", Key: "k"}); err != nil {
		t.Fatal(err)
	}
	if !canSearch() {
		t.Fatal("a search API set up, and magpie can't search")
	}
}

// DeepSeek's Search API (TJHHHH on Discord) is asked as its docs say
// (api-docs.deepseek.com/guides/search): POST /v1/tools/search with the
// key as a Bearer and {query, count}, a query cut to its 128 characters.
// Its reply's summary is the page's text, since content is empty without
// need_content; snippet stands in when there is no summary, and its date
// is the page's age. 402 (no balance) and 429 (5 QPS) are refusals that
// rest the key as out of credit and rate limited.
func TestDeepSeekSearchAPI(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var mu sync.Mutex
	var got []map[string]any
	var paths, auths []string
	status := 200
	f := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		json.Unmarshal(b, &body)
		mu.Lock()
		got = append(got, body)
		paths = append(paths, r.Method+" "+r.URL.Path+" "+r.Header.Get("Content-Type"))
		auths = append(auths, r.Header.Get("Authorization"))
		st := status
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch st {
		case 402:
			w.WriteHeader(402)
			io.WriteString(w, `{"error":{"message":"Insufficient Balance","type":"unknown_error","param":null,"code":"invalid_request_error"}}`)
			return
		case 429:
			w.WriteHeader(429)
			io.WriteString(w, `{"error":{"message":"Rate limit reached","type":"rate_limit_error","param":null,"code":"rate_limit_exceeded"}}`)
			return
		}
		io.WriteString(w, `{"results":[`+
			`{"rerank_score":2.1,"title":"Go downloads","url":"https://go.dev/dl/","snippet":"Download Go.","summary":"Go 1.27.1 is the latest stable release, with binaries for every platform.","content":"","site_name":"Go","site_icon":"https://go.dev/favicon.ico","date":"2026-10-01T00:00:00+00:00","authority":3},`+
			`{"rerank_score":1.2,"title":"Go release history","url":"https://go.dev/doc/devel/release","snippet":"Release history of Go.","summary":"","content":"","site_name":null,"site_icon":null,"date":null,"authority":-1}`+
			`]}`)
	}))
	t.Cleanup(f.Close)
	s := New()
	a := provider.SearchAPI{Vendor: "deepseek", Key: "sk-ds", URL: f.URL}
	long := strings.Repeat("搜", 200)
	for _, q := range []string{"latest go", long} {
		pages, err := s.askSearchAPI(context.Background(), a, q)
		if err != nil {
			t.Fatal(err)
		}
		if len(pages) != 2 || pages[0].URL != "https://go.dev/dl/" || pages[0].Title != "Go downloads" ||
			pages[0].Text != "Go 1.27.1 is the latest stable release, with binaries for every platform." ||
			pages[0].Age != "2026-10-01T00:00:00+00:00" ||
			pages[1].Text != "Release history of Go." || pages[1].Age != "" {
			t.Fatalf("pages %+v", pages)
		}
	}
	if paths[0] != "POST /v1/tools/search application/json" || auths[0] != "Bearer sk-ds" {
		t.Fatalf("asked %q with %q", paths[0], auths[0])
	}
	if got[0]["query"] != "latest go" || got[0]["count"] != float64(searchHits) || len(got[0]) != 2 {
		t.Fatalf("body %v", got[0])
	}
	if q, _ := got[1]["query"].(string); q != strings.Repeat("搜", deepseekQueryMax) {
		t.Fatalf("a long query went as %d characters", len([]rune(q)))
	}
	for _, c := range []struct {
		status int
		why    string
	}{{402, "out of credit"}, {429, "rate limited"}} {
		mu.Lock()
		status = c.status
		mu.Unlock()
		_, err := s.askSearchAPI(context.Background(), a, "q")
		var r searchRefused
		if !errors.As(err, &r) || r.status != c.status {
			t.Fatalf("%d: %v", c.status, err)
		}
		if _, why := searchRestFor("deepseek", c.status, "", time.Now()); why != c.why {
			t.Errorf("%d rests as %q", c.status, why)
		}
	}
	if v, ok := provider.SearchVendorOf("deepseek"); !ok || v.Base != "https://api.deepseek.com" || v.KeyOptional {
		t.Errorf("vendor %+v", v)
	}
}

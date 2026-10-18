package gateway

import (
	"fmt"
	"github.com/yetone/magpie/internal/provider"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
)

func grokSignedIn(t *testing.T) {
	t.Helper()
	t.Setenv("FAKE_RESPONSES", "1")
	movedFake(t, "grok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to, err := url.Parse(provider.GrokBase + strings.TrimPrefix(r.URL.Path, "/v1"))
		if err == nil && to.Hostname() != "127.0.0.1" {
			err = fmt.Errorf("GrokBase %s isn't the test's: a test never reaches the network", provider.GrokBase)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		to.RawQuery = r.URL.RawQuery
		(&httputil.ReverseProxy{Rewrite: func(pr *httputil.ProxyRequest) {
			pr.Out.URL, pr.Out.Host = to, to.Host
		}}).ServeHTTP(w, r)
	}))
	// the vendor's own host was the built-in's, whose video it fetched from
	// there; the plugin's host is the fake's, so the test's upstream on
	// loopback is named as a user names a private host
	allowPrivate(t, "127.0.0.1")
}

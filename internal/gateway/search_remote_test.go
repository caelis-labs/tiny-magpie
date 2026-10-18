package gateway

import (
	"net/http"
	"net/http/httptest"

	"testing"
)

// A request another magpie marks as a search is one the model searches
// for alone: this magpie runs no searcher of its own for it.
func TestSearchingFromAnotherMagpie(t *testing.T) {
	var seen bool
	h := withCaller(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = searching(r.Context()) && r.Header.Get(SearchingHeader) == ""
	}))
	for _, c := range []struct {
		ua   string
		want bool
	}{{"magpie/0.1.983", true}, {"curl/8", false}} {
		r := httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
		r.Header.Set("User-Agent", c.ua)
		r.Header.Set(SearchingHeader, "1")
		seen = false
		h.ServeHTTP(httptest.NewRecorder(), r)
		if seen != c.want {
			t.Errorf("%s: searching = %v", c.ua, seen)
		}
	}
}

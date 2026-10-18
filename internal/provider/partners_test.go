package provider

import (
	"net/http"
	"net/http/httptest"

	"strings"

	"sync/atomic"
	"testing"
)

// fetchPartnersNow fetches the list as the background refresh does, and
// waits for it. One fetch runs at a time, as Partners and PartnersNow
// keep it: a fetch Partners started is waited for first, so it can't
// store an older list over this one.
func fetchPartnersNow() {
	claimPartnerFetch()
	partnerFetches = true
	loadPartners()
	partnerMu.Unlock()
	refreshPartners()
}

func partnerFeedServer(t *testing.T, body *atomic.Value) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "magpie" {
			t.Errorf("User-Agent %q", r.Header.Get("User-Agent"))
		}
		_, _ = rw.Write([]byte(body.Load().(string)))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func partnerIDs(ps []Partner) string {
	var ids []string
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	return strings.Join(ids, ",")
}

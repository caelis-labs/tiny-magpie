package gateway

import (
	"github.com/yetone/magpie/internal/provider"

	"testing"
)

func postOK(t *testing.T, s *Server, session, body string) (string, Route) {
	t.Helper()
	code, out := postAs(t, s, session, body)
	if code != 200 {
		t.Fatalf("%d %s", code, out)
	}
	return out, lastRoute(s)
}

// A rule with an intent: the classifier is asked as the turn begins, with
// the user's words only, and the turn goes where its answer says; the
// turn's tool rounds don't ask again.
func mustFind(t *testing.T, id string) provider.Provider {
	t.Helper()
	p, err := provider.Find(id)
	if err != nil {
		t.Fatal(err)
	}
	return *p
}

// An id is put in the vendor's URL with the account's token: what would
// change the path is turned away before anything is sent.

var pngBytes = []byte("\x89PNG\r\n\x1a\n fake")

package gui

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGatewayMode(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	for _, web := range []bool{false, true} {
		if on, why := gatewayMode(web); !on || why != "tiny-magpie" {
			t.Fatalf("mode %v %q", on, why)
		}
	}
	h := Handler(nil, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/boot.js", nil))
	if !strings.Contains(w.Body.String(), `"gateway":true`) {
		t.Fatal(w.Body)
	}
}

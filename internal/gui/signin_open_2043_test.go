package gui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// browserDown is a desktop whose browser doesn't start, as Windows'
// rundll32 under a PATHEXT of ".CPL" (#2043, xiaobaifly7).
type browserDown struct{ webHost }

func (browserDown) OpenURL(string) error {
	return errors.New(`exec: "rundll32": executable file not found in %PATH%`)
}

// Open again on a sign-in says why the browser didn't open, where it said
// nothing (#2043).
func TestOpenSaysTheBrowserDidntOpen(t *testing.T) {
	sandboxHome(t)
	rec := httptest.NewRecorder()
	Handler(browserDown{}, nil).ServeHTTP(rec, httptest.NewRequest("POST", "/api/open", strings.NewReader(`{"url":"https://example.invalid/authorize"}`)))
	if rec.Code == http.StatusNoContent || !strings.Contains(rec.Body.String(), "rundll32") {
		t.Fatalf("open: %d %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	Handler(webHost{}, nil).ServeHTTP(rec, httptest.NewRequest("POST", "/api/open", strings.NewReader(`{"url":"https://example.invalid/authorize"}`)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("an open that worked: %d %s", rec.Code, rec.Body)
	}
}

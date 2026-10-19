package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"sync/atomic"
	"testing"
	"time"
)

// Only the week's window being used up counts: the five hours pass on
// their own, and a window whose end went by says nothing.
func fakeWeekUsedUp(t *testing.T) *atomic.Int32 {
	soon := time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339)
	var consumed atomic.Int32
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/backend-api/wham/usage":
			json.NewEncoder(w).Encode(map[string]any{"plan_type": "pro",
				"rate_limit": map[string]any{
					"primary_window":   map[string]any{"used_percent": 10, "limit_window_seconds": 18000},
					"secondary_window": map[string]any{"used_percent": 100, "limit_window_seconds": 604800, "reset_at": time.Now().Add(72 * time.Hour).Unix()},
				},
				"rate_limit_reset_credits": map[string]any{"available_count": 2}})
		case "/backend-api/wham/rate-limit-reset-credits":
			json.NewEncoder(w).Encode(map[string]any{"available_count": 2, "credits": []any{
				map[string]any{"id": "soon", "status": "available", "expires_at": soon},
			}})
		case "/backend-api/wham/rate-limit-reset-credits/consume":
			consumed.Add(1)
			json.NewEncoder(w).Encode(map[string]any{"code": "reset", "windows_reset": 2})
		default:
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(fake.Close)
	oldBase := CodexBase
	CodexBase = fake.URL + "/backend-api/codex"
	t.Cleanup(func() { CodexBase = oldBase })
	return &consumed
}

// unreadableAutoReset leaves codex-autoreset.json as #1505's crash left
// its files, zeroed at full length, for a new magpie to read.
func unreadableAutoReset(t *testing.T) {
	p := autoReset.where()
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, make([]byte, 64), 0o644); err != nil {
		t.Fatal(err)
	}
	autoReset.Lock()
	autoReset.path = ""
	autoReset.Unlock()
}

// A codex-autoreset.json that can't be read is not "none spent" (#2017):
// a reset auto-use spent this week, kept in codex-resets-used.json, still
// holds off a second one the same week, while one from a week ago, or one
// used by hand, doesn't.

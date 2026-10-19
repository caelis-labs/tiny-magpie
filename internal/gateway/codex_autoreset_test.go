package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
)

// resetBackend stands in for the ChatGPT backend of accounts that are out
// of their allowance: their usage (the week's and five hours'), the
// resets each holds, and spending one, which lets the account answer again.
type resetBackend struct {
	mu    sync.Mutex
	week  map[string]float64 // used, by account id; the five hours are full
	held  map[string]int     // resets held
	out   map[string]bool    // out of its allowance
	spent []string           // accounts a reset was spent on
	gone  bool               // every reset spent elsewhere since the usage was read
	tried []string           // accounts asked for a turn
}

func newResetBackend(t *testing.T, accounts ...string) *resetBackend {
	t.Helper()
	b := &resetBackend{week: map[string]float64{}, held: map[string]int{}, out: map[string]bool{}}
	for _, a := range accounts {
		b.week[a], b.held[a], b.out[a] = 100, 1, true
	}
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acct := r.Header.Get("chatgpt-account-id")
		b.mu.Lock()
		defer b.mu.Unlock()
		now := time.Now()
		switch r.URL.Path {
		case "/backend-api/codex/responses":
			io.ReadAll(r.Body)
			b.tried = append(b.tried, acct)
			if b.out[acct] {
				w.WriteHeader(429)
				io.WriteString(w, `{"error":{"type":"usage_limit_reached","message":"The usage limit has been reached","plan_type":"plus","resets_in_seconds":7200}}`)
				return
			}
			io.WriteString(w, sse(
				`data: {"type":"response.created","response":{"id":"r1","model":"gpt-5.5"}}`,
				`data: {"type":"response.output_text.delta","delta":"pong"}`,
				`data: {"type":"response.completed","response":{"id":"r1","usage":{"input_tokens":7,"output_tokens":1}}}`))
		case "/backend-api/wham/usage":
			five := 100.0
			if !b.out[acct] {
				five = 0
			}
			json.NewEncoder(w).Encode(map[string]any{"plan_type": "plus",
				"rate_limit": map[string]any{
					"primary_window":   map[string]any{"used_percent": five, "limit_window_seconds": 18000, "reset_at": now.Add(2 * time.Hour).Unix()},
					"secondary_window": map[string]any{"used_percent": b.week[acct], "limit_window_seconds": 604800, "reset_at": now.Add(3 * 24 * time.Hour).Unix()}},
				"rate_limit_reset_credits": map[string]any{"available_count": b.held[acct]}})
		case "/backend-api/wham/rate-limit-reset-credits":
			var credits []any
			for i := range b.held[acct] {
				credits = append(credits, map[string]any{"id": acct + "-" + string(rune('a'+i)), "reset_type": "codex_rate_limits",
					"status": "available", "granted_at": "2026-09-01T00:00:00Z", "expires_at": now.Add(20 * 24 * time.Hour).Format(time.RFC3339)})
			}
			json.NewEncoder(w).Encode(map[string]any{"available_count": b.held[acct], "credits": credits})
		case "/backend-api/wham/rate-limit-reset-credits/consume":
			if b.held[acct] == 0 || b.gone {
				io.WriteString(w, `{"code":"no_credit","windows_reset":0}`)
				return
			}
			b.held[acct]--
			b.spent = append(b.spent, acct)
			b.out[acct], b.week[acct] = false, 0
			io.WriteString(w, `{"code":"reset","windows_reset":2}`)
		default:
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(up.Close)
	was := provider.CodexBase
	provider.CodexBase = up.URL + "/backend-api/codex"
	t.Cleanup(func() { provider.CodexBase = was })
	return b
}

func (b *resetBackend) seen() (tried, spent string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.Join(b.tried, ","), strings.Join(b.spent, ",")
}

// resetPost posts a Codex turn to srv as codexPost does.
func resetPost(t *testing.T, srv *Server) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", CodexPath+"/responses", strings.NewReader(`{"model":"gpt-5.5","stream":true,"input":"ping"}`))
	req.Header.Set("Authorization", "Bearer chatgpt-token")
	req.Header.Set("chatgpt-account-id", "acct-1")
	srv.Handler().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

// resetsTraced is the resets the routing trace says were spent by
// themselves, as who: what.
func resetsTraced(srv *Server) []string {
	var out []string
	for _, r := range srv.Trace(context.Background(), 0, 0).Routes {
		for _, t := range r.Tries {
			if t.Reset != nil {
				out = append(out, t.Reset.Who+": "+t.Reset.Text)
			}
		}
	}
	return out
}

// Codex's one account, its week used up, with the user's leave to spend
// its resets by itself: one is spent and the turn goes through (StringKe
// on Discord: resets gifted to the account, used once the week is out).

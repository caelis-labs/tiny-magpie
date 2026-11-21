package gateway

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
)

func TestTinyCodexCreditsAfterPlanLimit(t *testing.T) {
	fresh(t)
	const user = "credits@example.test"
	dir := filepath.Join(os.Getenv("HOME"), ".codex")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	token := func(claims string) string {
		return "e30." + base64.RawURLEncoding.EncodeToString([]byte(claims)) + ".signature"
	}
	auth := fmt.Sprintf(`{"auth_mode":"chatgpt","tokens":{"id_token":%q,"access_token":%q,"refresh_token":"fixture-refresh","account_id":"credits-account"}}`,
		token(`{"email":"`+user+`"}`), token(fmt.Sprintf(`{"exp":%d}`, time.Now().Add(time.Hour).Unix())))
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(auth), 0600); err != nil {
		t.Fatal(err)
	}
	provider.ForgetAccounts()
	t.Cleanup(provider.ForgetAccounts)

	reset := time.Now().Add(time.Hour)
	reading := provider.Allowance{{Used: 100, Resets: reset, Span: 5 * time.Hour, Credits: true}}
	oldAllowances := allowances
	allowances = func(string) map[string]provider.Allowance {
		return map[string]provider.Allowance{user: reading}
	}
	t.Cleanup(func() { allowances = oldAllowances })

	var turns atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backend-api/codex/responses" || r.Header.Get("chatgpt-account-id") != "credits-account" {
			t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		turns.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"pong\"}\n\n"+
			"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"fixture\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\n")
	}))
	t.Cleanup(upstream.Close)
	oldBase := provider.CodexBase
	provider.CodexBase = upstream.URL + "/backend-api/codex"
	t.Cleanup(func() { provider.CodexBase = oldBase })

	post := func() (int, string) {
		return postRetrieval(t, New(), "/v1/responses", `{"model":"codex/gpt-5.5","stream":true,"input":"ping"}`)
	}
	if code, body := post(); code != http.StatusOK || !strings.Contains(body, "pong") || turns.Load() != 1 {
		t.Fatalf("account with credits: status %d, upstream turns %d, body %s", code, turns.Load(), body)
	}
	if err := provider.SetCodexCredits(user, false); err != nil {
		t.Fatal(err)
	}
	if code, body := post(); code != http.StatusTooManyRequests || turns.Load() != 1 {
		t.Fatalf("credits disabled: status %d, upstream turns %d, body %s", code, turns.Load(), body)
	}
	if err := provider.SetCodexCredits(user, true); err != nil {
		t.Fatal(err)
	}
	if err := provider.SetAccountCap("codex", user, 95); err != nil {
		t.Fatal(err)
	}
	if code, body := post(); code != http.StatusTooManyRequests || turns.Load() != 1 {
		t.Fatalf("account capped: status %d, upstream turns %d, body %s", code, turns.Load(), body)
	}
}

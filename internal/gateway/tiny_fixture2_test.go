package gateway

// Shared sandbox support for Tiny boundary tests.
import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/routes"
)

// forgetRouting clears what routing keeps for the gateway's life — each
// provider's turn, each candidate's failures and tokens, each weighted
// round — so a test run again (-count) starts where it did the first time
// rather than a turn on, or with a failure already counted.
func forgetRouting() {
	routed.Lock()
	routed.used, routed.failures, routed.slow = map[string]tokenUse{}, map[string]int{}, map[string]slowRun{}
	routed.Unlock()
	routes.Forget()
}

func fresh(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	restingUntil.Lock()
	restingUntil.m = map[string]time.Time{}
	restingUntil.Unlock()
	sticks.Lock()
	sticks.m = map[string]stick{}
	sticks.Unlock()
	refusedSeals.Lock()
	refusedSeals.m = map[string]sealsRefused{}
	refusedSeals.Unlock()
	overflowLearned.Lock()
	overflowLearned.m = map[string]map[string]refusedLong{}
	overflowLearned.Unlock()
	turnRules.Lock()
	turnRules.m = map[string]turnRule{}
	turnRules.Unlock()
	classified.Lock()
	classified.m, classified.failed = map[string]classifiedAs{}, map[string]classifyFailure{}
	classified.Unlock()
	levelled.Lock()
	levelled.m = map[string]levelling{}
	levelled.Unlock()
	provider.ForgetCopilotForTest()
	provider.ForgetRemoteCardsForTest()
	provider.ForgetKeptCardsForTest()
	remoteRefreshes.Lock()
	remoteRefreshes.at = map[string]time.Time{}
	remoteRefreshes.Unlock()
	resetAuthThrottle()
	forgetRouting()
}

func postRetrieval(t *testing.T, s *Server, path, body string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

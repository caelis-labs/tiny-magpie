package usage

import (
	"path/filepath"

	"testing"
	"time"

	"github.com/yetone/magpie/internal/sessions"
)

// effortFixture is one Codex thread's call in a sandbox home.
func effortFixture(t *testing.T) (now time.Time, rollout string) {
	t.Helper()
	pageHome(t)
	now = holdClock(t, time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC))
	rollout = filepath.Join(sessions.CodexDir(), "sessions", "2026", "09", "30", "rollout-2026-09-30T10-00-00-main-a.jsonl")
	writeCodexCall(t, rollout, "main-a", "main-a", `"vscode"`, "resp-1", "m", "low", now.Add(-time.Hour), 20, 3)
	return now, rollout
}

package agent

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTinyDriftDoesNotReadAgentHistory(t *testing.T) {
	home, _ := codexHome(t, "", "model = \"gpt-5.5\"\n")
	a := codex(home)
	if err := a.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	reads := 0
	a.LastUsed = func() time.Time { reads++; return time.Time{} }
	a.Reached = func(time.Time) (time.Time, string, bool) { reads++; return time.Time{}, "", false }
	if d := a.Drift(); d != nil {
		t.Fatalf("configured agent drifted: %+v", d)
	}
	if reads != 0 {
		t.Fatalf("drift read Agent history %d times", reads)
	}
}

func TestTinyCodexPickDoesNotImportSessionProviders(t *testing.T) {
	home, read := codexHome(t, "", "model = \"gpt-5.5\"\n")
	db, err := sql.Open("sqlite", filepath.Join(home, ".codex", "state_5.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`CREATE TABLE threads (model_provider TEXT); INSERT INTO threads VALUES ('legacy_session_only')`); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	a := codex(home)
	if err := a.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(read(), "legacy_session_only") {
		t.Fatalf("model pick imported a provider from session history:\n%s", read())
	}
	if !a.Wired() {
		t.Fatal("model pick did not connect Codex")
	}
}

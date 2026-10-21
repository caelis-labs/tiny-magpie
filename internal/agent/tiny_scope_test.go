package agent

import (
	"database/sql"
	"github.com/yetone/magpie/internal/provider"
	"os"
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

// A Magpie-compatible block is not evidence that Tiny owns this Agent.
func TestTinyCatalogSyncLeavesUnmanagedAgentAlone(t *testing.T) {
	home, read := codexHome(t, "", "model = \"fake/m1\"\nmodel_provider = \"magpie\"\nmodel_catalog_json = \"magpie-models.json\"\n[model_providers.magpie]\nbase_url = \"http://127.0.0.1:43210/v1\"\nwire_api = \"responses\"\n")
	path := filepath.Join(home, ".codex", "magpie-models.json")
	original := `{"models":[{"slug":"foreign/model"}]}`
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	before := read()
	SyncCatalog()
	if got, err := os.ReadFile(path); err != nil || string(got) != original || read() != before {
		t.Fatalf("catalog sync changed an unmanaged Agent: read error=%v, catalog changed=%v, config changed=%v", err, string(got) != original, read() != before)
	}
	a := codex(home)
	// An explicit pick authorizes Tiny to keep this Agent's model list current.
	if err := a.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	if err := provider.Save(provider.Provider{ID: "added", Name: "Added", Key: "fixture", Chat: "http://127.0.0.1:1/v1", Models: []string{"new"}}); err != nil {
		t.Fatal(err)
	}
	SyncCatalog()
	got, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(got), "added/new") {
		t.Fatalf("owned Agent catalog not synchronized: %s %v", got, err)
	}
}

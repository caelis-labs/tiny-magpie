package sessions

import (
	"os"
	"path/filepath"

	"testing"
	"time"
)

// manageSetup is setup with magpie's own folder in the temp dir too, and
// every session file a day old, so none counts as running.
func manageSetup(t *testing.T) (claude, codex string) {
	claude, codex = setup(t)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
	old := time.Now().Add(-24 * time.Hour)
	for _, root := range []string{claude, codex} {
		filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err == nil {
				os.Chtimes(p, old, old)
			}
			return nil
		})
	}
	return
}

const mgCC = "11111111-2222-3333-4444-555555555555"
const mgCX = "01a0bdc9-b5fd-7e63-9658-68bc8dce5ecd"

func findManaged(list []Managed, id string) (Managed, bool) {
	for _, m := range list {
		if m.ID == id {
			return m, true
		}
	}
	return Managed{}, false
}

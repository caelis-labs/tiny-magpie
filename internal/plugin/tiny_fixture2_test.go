package plugin

// Shared sandbox support for Tiny boundary tests.
import (
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

// sandbox gives the test its own magpie folders and the Bun on PATH; a
// machine without Bun skips it.
func sandbox(t *testing.T) {
	t.Helper()
	bun := testenv.Bun(t)
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	t.Setenv("MAGPIE_BUN", bun)
	t.Cleanup(Settle)
}

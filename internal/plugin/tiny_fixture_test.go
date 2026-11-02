package plugin

// Sandbox support for Tiny boundary tests. Vendor requests and real Agent state are never used.
import (
	"os"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

// The tests run in a home of their own, never the user's real files, and
// off the network (testenv.Offline): each sandbox is a new home with no
// models.dev catalog, so every host start fetched the real models.dev for
// up to 30s, and on a runner where it didn't answer two starts spent a
// test's whole minute (#1922).
func TestMain(m *testing.M) {
	os.Exit(testenv.RunIn(m, func(string) { testenv.Offline() }))
}

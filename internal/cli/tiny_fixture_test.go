package cli

// Sandbox support for Tiny boundary tests. Vendor requests and real Agent state are never used.
import (
	"os"
	"testing"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/testenv"
)

// The tests run in a home of their own, never the user's real files, and
// off the network (testenv.Offline): the plugin tests start plugin hosts,
// and a host started in a home without a models.dev catalog fetched the
// real one for up to 30s (#1922).
func TestMain(m *testing.M) {
	if len(os.Args) == 3 && os.Args[1] == agent.DryRunArg {
		if err := run(os.Args[1:]); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	os.Exit(testenv.RunIn(m, func(string) { testenv.Offline() }))
}

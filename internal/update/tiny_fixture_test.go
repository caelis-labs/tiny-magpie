package update

// Sandbox support for Tiny boundary tests. Vendor requests and real Agent state are never used.
import (
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

// The tests run in a home of their own, never the user's real files.
func TestMain(m *testing.M) { testenv.Main(m) }

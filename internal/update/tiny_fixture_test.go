package update

// Sandbox support for Tiny boundary tests. Vendor requests and real Agent state are never used.
import (
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

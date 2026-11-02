package gui

// Sandbox support for Tiny boundary tests. Vendor requests and real Agent state are never used.
import (
	"io"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/yetone/magpie/internal/fx"
	"github.com/yetone/magpie/internal/testenv"
)

// TestMain keeps the rate off the network: a state read with no rate cached
// asks for one behind it (#541), which would otherwise reach the real
// endpoint and land in whatever test runs next. And the tests run in a
// home of their own, never the user's real files.
func TestMain(m *testing.M) {

	if addr := os.Getenv("MAGPIE_TEST_FAKE_GATEWAY"); addr != "" {
		fakeOlderGateway(addr)
		return
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		addr := l.Addr().String()
		l.Close()
		fx.SetRateURL("http://" + addr + "/")
	}
	os.Exit(testenv.Run(m))
}

func fakeOlderGateway(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		os.Exit(3)
	}
	http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"name":"magpie","version":"0.1.550","models":3}`)
	}))
}

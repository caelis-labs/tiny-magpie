package gui

import (
	"net"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/gateway"
)

// Tiny reports an occupied port without attaching to any other version.
func TestGatewayDoesNotAttachToAnyMagpieVersion(t *testing.T) {
	h := t.TempDir()
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(h, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(h, ".cache"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(h, ".local", "share"))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAGPIE_ADDR", ln.Addr().String())
	var answer string
	other := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(answer))
	})}
	go other.Serve(ln)
	defer other.Close()
	was := gateway.Version
	gateway.Version = "0.1.630"
	defer func() { gateway.Version = was }()
	served.Store(nil)

	for _, c := range []struct {
		answer, version string
		older, window   bool
	}{
		{`{"name":"magpie","version":"0.1.550","models":3,"window":true,"apis":[]}`, "0.1.550", true, true},
		{`{"name":"magpie","version":"0.1.630","models":3,"window":false}`, "0.1.630", false, false},
		{`{"name":"magpie","version":"0.1.640","models":3}`, "0.1.640", false, false},
		{`{"name":"magpie","models":3}`, "", false, false}, // one that doesn't say
	} {
		answer = c.answer
		g := providersState().Gateway
		if g.Running || g.Mine || g.Version != "" || g.Older || g.Window || !g.Taken {
			t.Errorf("%s: running %v mine %v version %q older %v window %v", c.answer, g.Running, g.Mine, g.Version, g.Older, g.Window)
		}
	}
}

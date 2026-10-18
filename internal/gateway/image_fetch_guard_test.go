package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"strings"

	"sync/atomic"
	"testing"

	"github.com/yetone/magpie/internal/settings"
)

func TestPublicAddr(t *testing.T) {
	for s, want := range map[string]bool{
		"127.0.0.1": false, "127.9.9.9": false, "10.1.2.3": false, "172.16.0.1": false,
		"192.168.1.1": false, "169.254.169.254": false, "100.64.0.1": false, "100.100.100.100": false,
		"0.0.0.0": false, "224.0.0.1": false, "255.255.255.255": false,
		"::1": false, "::": false, "::ffff:127.0.0.1": false, "::ffff:10.0.0.1": false,
		"::ffff:169.254.169.254": false, "::ffff:100.64.0.1": false, "fe80::1": false, "fc00::1": false, "fd12::1": false,
		"ff02::1": false, "64:ff9b::7f00:1": false, "2002:7f00:1::": false, "::7f00:1": false,
		"8.8.8.8": true, "93.184.216.34": true, "::ffff:8.8.8.8": true, "2606:4700::1111": true,
		"64:ff9b::808:808": true,
	} {
		if got := publicAddr(netip.MustParseAddr(s)); got != want {
			t.Errorf("publicAddr(%s) = %v, want %v", s, got, want)
		}
	}
}

// secret is a service only this computer reaches; what it was asked is
// counted.
func secret(t *testing.T, body string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func editFrom(t *testing.T, s *Server, url string, edit func(*http.Request)) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/images/edits", strings.NewReader(`{"model":"art/gpt-image-1","prompt":"make it blue","image":"`+url+`"}`))
	req.Header.Set("Content-Type", "application/json")
	if edit != nil {
		edit(req)
	}
	s.Handler().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func resolveTo(t *testing.T, names map[string]string) {
	t.Helper()
	was := lookupIP
	lookupIP = func(_ context.Context, host string) ([]netip.Addr, error) {
		if a, ok := names[host]; ok {
			return []netip.Addr{netip.MustParseAddr(a)}, nil
		}
		return was(context.Background(), host)
	}
	t.Cleanup(func() { lookupIP = was })
}

// allowPrivate writes the user's settings.json "fetchPrivate" for the test.
func allowPrivate(t *testing.T, entries ...string) {
	t.Helper()
	was := settings.Load()
	st := was
	st.FetchPrivate = entries
	if err := settings.Save(st); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { settings.Save(was) })
}

// fetchPrivate's entries are a name, name:port, an IP or a range; a
// settings.json that can't be read lets nothing extra in.
func TestFetchPrivateEntries(t *testing.T) {
	allowPrivate(t, "nas.lan", "box.lan:8080", "10.1.2.3", "192.168.7.0/24", "fd00::/8", "  ", "not a range/99")
	a := allowFor([]string{"vendor.lan:9000"})
	for _, c := range []struct {
		host, port string
		ok         bool
	}{
		{"nas.lan", "443", true}, {"NAS.lan", "80", true}, {"box.lan", "8080", true}, {"box.lan", "80", false},
		{"vendor.lan", "9000", true}, {"vendor.lan", "9001", false}, {"other.lan", "80", false},
	} {
		if a.host(c.host, c.port) != c.ok {
			t.Errorf("host %s:%s: %v", c.host, c.port, !c.ok)
		}
	}
	for ip, ok := range map[string]bool{
		"10.1.2.3": true, "10.1.2.4": false, "192.168.7.200": true, "192.168.8.1": false,
		"fd12::1": true, "127.0.0.1": false, "169.254.169.254": false, "8.8.8.8": true,
	} {
		if a.addr(netip.MustParseAddr(ip)) != ok {
			t.Errorf("%s: %v", ip, !ok)
		}
	}
	if err := os.WriteFile(settings.Path(), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	a = allowFor(nil)
	if a.host("nas.lan", "443") || a.addr(netip.MustParseAddr("10.1.2.3")) {
		t.Fatal("an unreadable settings.json let a private host in")
	}
	if got := hostOf("http://Box.lan/x"); got != "box.lan:80" {
		t.Fatalf("hostOf: %q", got)
	}
	if got := hostOf("https://[fd00::1]/v1"); got != "[fd00::1]:443" {
		t.Fatalf("hostOf: %q", got)
	}
}

// The dialler checks the address it connects to, whatever name led there.
func TestPrivateDialerRefusesPrivateAddresses(t *testing.T) {
	srv, hits := secret(t, "")
	for _, addr := range []string{srv.Listener.Addr().String(), "[::1]:9"} {
		c, err := privateDialer.DialContext(context.Background(), "tcp", addr)
		if err == nil {
			c.Close()
		}
		if !errors.Is(err, errPrivateAddress) {
			t.Errorf("%s: %v", addr, err)
		}
	}
	if hits.Load() != 0 {
		t.Fatal("the private service was asked")
	}
}

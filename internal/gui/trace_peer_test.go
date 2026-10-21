package gui

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// A magpie here that does not serve the gateway shows the routing of the one
// that does (#2606), and names it: the console reads the serving magpie's
// trace over loopback and plays it as its own, and its state carries how that
// magpie was started and its pid.
func TestConsoleKeepsTheNoticeFromAnOlderGateway(t *testing.T) {
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
	other := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/" {
			w.Write([]byte(`{"name":"magpie","version":"0.1.550","models":3,"window":true,"apis":[]}`))
			return
		}
		http.NotFound(w, r)
	})}
	go other.Serve(ln)
	defer other.Close()
	served.Store(nil)

	mux := http.NewServeMux()
	traceRoutes(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/gateway/trace", nil))
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var got struct {
		Mine   bool            `json:"mine"`
		Routes json.RawMessage `json:"routes"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Mine {
		t.Fatalf("an older gateway's routing was played as this console's own: %s", w.Body)
	}
}

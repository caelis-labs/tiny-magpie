package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

// noVision turns describing off, for a test of images turned away.
func noVision(t *testing.T) {
	t.Helper()
	s := settings.Load()
	s.Vision = "off"
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
}

// eyeUp is an upstream with a text-only model and one that sees: the one
// that sees describes every image as a red square, or fails when fail is
// set; what each model is sent is kept.
type eyeUp struct {
	mu     sync.Mutex
	fail   bool
	bodies map[string][]string
}

func (u *eyeUp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	model := modelOf(body)
	u.mu.Lock()
	if u.bodies == nil {
		u.bodies = map[string][]string{}
	}
	u.bodies[model] = append(u.bodies[model], string(body))
	fail := u.fail
	u.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if strings.HasSuffix(model, "eye") && fail {
		w.WriteHeader(500)
		io.WriteString(w, `{"error":{"message":"eye is down"}}`)
		return
	}
	answer := "ok"
	if strings.HasSuffix(model, "eye") {
		answer = "A red square with the word HELLO in it."
	}
	if strings.Contains(string(body), `"stream":true`) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, `data: {"id":"x","choices":[{"index":0,"delta":{"role":"assistant","content":"`+answer+`"}}]}`+"\n\n")
		io.WriteString(w, `data: {"id":"x","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`+"\n\n")
		io.WriteString(w, "data: [DONE]\n\n")
		return
	}
	io.WriteString(w, `{"id":"x","choices":[{"index":0,"message":{"role":"assistant","content":"`+answer+`"},"finish_reason":"stop"}]}`)
}

func (u *eyeUp) sent(model string) []string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return append([]string(nil), u.bodies[model]...)
}

func eyed(t *testing.T) (*Server, *eyeUp) {
	t.Helper()
	fresh(t)
	u := &eyeUp{}
	up := httptest.NewServer(u)
	t.Cleanup(up.Close)
	if err := provider.Save(provider.Provider{ID: "probe", Name: "Probe", Chat: up.URL + "/v1", Key: "key", Models: []string{"text", "eye"}}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SaveLive("probe", up.URL+"/v1", []catalog.Model{
		{ID: "text", ImageInput: imageInputBool(false)},
		{ID: "eye", Images: true, ImageInput: imageInputBool(true)},
	}); err != nil {
		t.Fatal(err)
	}
	return New(), u
}

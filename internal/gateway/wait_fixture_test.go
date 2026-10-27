package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
)

// slowEye is eyeUp with the describing model's answer held: until gate is
// closed for a request that says "hold", and by think for a description.
type slowEye struct {
	eyeUp
	held  chan struct{} // a request that says "hold" has arrived
	gate  chan struct{}
	think time.Duration
	open  func() // closes gate, once
}

func (u *slowEye) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	if strings.Contains(string(body), "hold") {
		u.held <- struct{}{}
		<-u.gate
	} else if strings.HasSuffix(modelOf(body), "eye") {
		time.Sleep(u.think)
	}
	r.Body = io.NopCloser(strings.NewReader(string(body)))
	u.eyeUp.ServeHTTP(w, r)
}

// slowEyed is eyed with one request at a time to the provider, so a
// request waits in its lane while another is out.
func slowEyed(t *testing.T, think time.Duration) (*Server, *slowEye) {
	t.Helper()
	fresh(t)
	u := &slowEye{held: make(chan struct{}, 1), gate: make(chan struct{}), think: think}
	up := httptest.NewServer(u)
	t.Cleanup(up.Close)
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(u.gate) }) })
	u.open = func() { once.Do(func() { close(u.gate) }) }
	one := 1
	if err := provider.Save(provider.Provider{ID: "probe", Name: "Probe", Chat: up.URL + "/v1", Key: "key", Models: []string{"text", "eye"}, MaxConcurrency: &one, QueueWait: 60}); err != nil {
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

package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// unroutedLimits sets when a request held before routing shows, and when
// it is answered for.
func unroutedLimits(t *testing.T, show, most time.Duration) {
	t.Helper()
	s, m := unroutedShow, unroutedMost
	unroutedShow, unroutedMost = show, most
	t.Cleanup(func() { unroutedShow, unroutedMost = s, m })
}

// unknownLimit sets how long a request may be held before routing in no
// wait magpie knows of.
func unknownLimit(t *testing.T, d time.Duration) {
	t.Helper()
	was := unroutedUnknown
	unroutedUnknown = d
	t.Cleanup(func() { unroutedUnknown = was })
}

// watchedHold is the watch of s's one request held before routing, once
// it is there.
func watchedHold(t *testing.T, s *Server) *unroutedHold {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		s.active.mu.Lock()
		var u *unroutedHold
		for _, f := range s.active.m {
			f.mu.Lock()
			if f.unrouted != nil {
				u = f.unrouted
			}
			f.mu.Unlock()
		}
		s.active.mu.Unlock()
		if u != nil {
			return u
		}
		if time.Now().After(deadline) {
			t.Fatal("no request was held before routing")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// unroutedReply is a reply postWhole got.
type unroutedReply struct {
	code int
	body string
	err  error
}

// heldPast waits until u's request has been held in known waits for d,
// and answers nil then; or answers its reply, if it is answered first.
func heldPast(t *testing.T, u *unroutedHold, d time.Duration, got chan unroutedReply) *unroutedReply {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for u.excusedFor() < d {
		select {
		case r := <-got:
			return &r
		case <-time.After(5 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			t.Fatalf("held in known waits for %s only", u.excusedFor())
		}
	}
	return nil
}

// posted posts body to srv in the background, its reply to the channel.
func posted(srv *httptest.Server, body io.Reader) chan unroutedReply {
	got := make(chan unroutedReply, 1)
	go func() {
		code, b, err := postWhole(srv, body)
		got <- unroutedReply{code, b, err}
	}()
	return got
}

// postWhole posts body to srv's Chat API and answers its reply.
func postWhole(srv *httptest.Server, body io.Reader) (int, string, error) {
	c := &http.Client{Timeout: 30 * time.Second}
	resp, err := c.Post(srv.URL+"/v1/chat/completions", "application/json", body)
	if err != nil {
		return 0, "", err
	}
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, string(b), err
}

const unroutedHi = `{"model":"probe/text","messages":[{"role":"user","content":"hi"}]}`

// sharwowwow on X: two minutes of a spinner before the 504. A request
// held before routing in no wait magpie knows of (here a lock routing
// needs, held and never let go) is answered once it has been held
// unroutedUnknown, long before unroutedMost.
func TestUnroutedUnknownHoldAnsweredSooner(t *testing.T) {
	unroutedLimits(t, 20*time.Millisecond, time.Hour)
	unknownLimit(t, 300*time.Millisecond)
	s, _ := slowEyed(t, 0)
	srv := httptest.NewServer(s.Handler())
	t.Cleanup(srv.Close)
	sticks.Lock()
	var once sync.Once
	unlock := func() { once.Do(sticks.Unlock) }
	t.Cleanup(unlock)

	got := posted(srv, strings.NewReader(unroutedHi))
	var r unroutedReply
	select {
	case r = <-got:
	case <-time.After(20 * time.Second):
		unlock()
		t.Fatal("a request held in no known wait was not answered before unroutedMost")
	}
	unlock()
	if r.err != nil || r.code != http.StatusGatewayTimeout || !strings.Contains(r.body, "without routing it to any provider") {
		t.Fatalf("held request answered %d %q %v", r.code, r.body, r.err)
	}
}

// A body still arriving (an image over a slow link) is a known wait: a
// request whose body takes longer than unroutedUnknown is routed and
// answered, not cut off.
func TestUnroutedSlowBodyNotCutOff(t *testing.T) {
	unroutedLimits(t, 20*time.Millisecond, time.Hour)
	unknownLimit(t, 100*time.Millisecond)
	s, _ := slowEyed(t, 0)
	srv := httptest.NewServer(s.Handler())
	t.Cleanup(srv.Close)
	pr, pw := io.Pipe()
	t.Cleanup(func() { pw.Close() })
	got := posted(srv, pr)
	half := len(unroutedHi) / 2
	if _, err := pw.Write([]byte(unroutedHi[:half])); err != nil {
		t.Fatal(err)
	}
	if r := heldPast(t, watchedHold(t, s), 3*unroutedUnknown, got); r != nil {
		t.Fatalf("answered %d %q %v while its body was still arriving", r.code, r.body, r.err)
	}
	if _, err := pw.Write([]byte(unroutedHi[half:])); err != nil {
		t.Fatal(err)
	}
	pw.Close()
	r := <-got
	if r.err != nil || r.code != http.StatusOK {
		t.Fatalf("a slow body was answered %d %q %v", r.code, r.body, r.err)
	}
}

// A request waiting in a key's queue (MaxConcurrency) past
// unroutedUnknown is not cut off: the queue is a known wait.
func TestUnroutedSlotWaitNotCutOff(t *testing.T) {
	unroutedLimits(t, 20*time.Millisecond, time.Hour)
	unknownLimit(t, 100*time.Millisecond)
	s, up := slowEyed(t, 0)
	srv := httptest.NewServer(s.Handler())
	t.Cleanup(srv.Close)
	codes := make(chan int, 2)
	go func() {
		code, _, _ := postWhole(srv, strings.NewReader(`{"model":"probe/text","messages":[{"role":"user","content":"hold"}]}`))
		codes <- code
	}()
	<-up.held // the one slot is taken
	go func() {
		code, _, _ := postWhole(srv, strings.NewReader(unroutedHi))
		codes <- code
	}()
	deadline := time.Now().Add(20 * time.Second)
	for {
		queued := false
		s.active.mu.Lock()
		for _, f := range s.active.m {
			if p, _ := f.phase.Load().(string); p == "slot" && time.Since(f.start) > 3*unroutedUnknown {
				queued = true
			}
		}
		s.active.mu.Unlock()
		if queued {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the second request never waited in the queue past the limit")
		}
		time.Sleep(5 * time.Millisecond)
	}
	up.open()
	for range 2 {
		if code := <-codes; code != http.StatusOK {
			t.Fatalf("a request queued past the limit was answered %d", code)
		}
	}
}

// routesOf is the trace's routes for model now.
func routesOf(s *Server, model string) []Route {
	var out []Route
	for _, r := range s.Trace(context.Background(), 0, 0).Routes {
		if r.Model == model {
			out = append(out, r)
		}
	}
	return out
}

// A request answered before it was shown leaves nothing on the trace of
// its wait, and a timer that fires as its handler returns adds nothing.
func TestUnroutedQuickRequestLeavesNoTrace(t *testing.T) {
	unroutedLimits(t, time.Millisecond, time.Minute)
	s, _ := slowEyed(t, 0)
	for range 20 {
		code, _ := postAs(t, s, "", `{"model":"nowhere/none","messages":[{"role":"user","content":"hi"}]}`)
		if code < 400 {
			t.Fatalf("an unknown model answered %d", code)
		}
	}
	time.Sleep(50 * time.Millisecond)
	for _, r := range s.Trace(context.Background(), 0, 0).Routes {
		if r.Provider == "magpie" && !r.Done {
			t.Fatalf("a route left showing a request that was answered: %+v", r)
		}
	}
}

// #3682: masking's time counts toward unroutedMost no more, while what
// comes after it still does. A request masked for well past unroutedMost
// isn't answered; then held by an image's description that never ends,
// it is answered 504 once that wait has had unroutedMost.

package gateway

import (
	"context"

	"testing"
	"time"

	"github.com/yetone/magpie/internal/gwtool"
)

// The magpie serving the gateway, and only while it serves, offers
// magpie's tools only as their plugins serve them (gwtool.Gate): with
// plugins/search off, not its web search.
func TestServingGatesTools(t *testing.T) {
	fresh(t)
	t.Setenv("MAGPIE_ADDR", "127.0.0.1:0") // never where a magpie runs
	searchOn()
	defer func() { searchOn = gwtool.Serve(gwtool.WebSearch) }()
	if !gwtool.Served(gwtool.WebSearch) {
		t.Fatal("gated before serving")
	}
	s := New()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.ListenAndServe(ctx) }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		s.lnMu.Lock()
		ln := s.ln
		s.lnMu.Unlock()
		if ln != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("gateway did not bind")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if gwtool.Served(gwtool.WebSearch) {
		t.Error("serving, with the plugin off: web search still served")
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if !gwtool.Served(gwtool.WebSearch) {
		t.Error("gate held after serving ended")
	}
}

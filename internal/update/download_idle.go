package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// A minute per byte is still progress. Two minutes without a byte gives
// slow links room while bounding a stalled response, independently of size.
const downloadIdleTimeout = 2 * time.Minute

var errDownloadIdle = errors.New("update download made no progress")

// idleTransport gives every response (including redirect bodies and source
// fallback attempts) its own idle clock. Cancelling the request interrupts a
// blocked HTTP Read; checking elapsed time only between reads cannot do that.
type idleTransport struct {
	base http.RoundTripper
	idle time.Duration
}

func (t *idleTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancelCause(req.Context())
	res, err := t.base.RoundTrip(req.Clone(ctx))
	if err != nil {
		cancel(nil)
		return res, err
	}
	b := &idleBody{ReadCloser: res.Body, ctx: ctx, cancel: cancel, idle: t.idle, last: time.Now()}
	b.timer = time.AfterFunc(t.idle, b.expire)
	res.Body = b
	return res, nil
}

func (t *idleTransport) CloseIdleConnections() {
	if c, ok := t.base.(interface{ CloseIdleConnections() }); ok {
		c.CloseIdleConnections()
	}
}

type idleBody struct {
	io.ReadCloser
	ctx     context.Context
	cancel  context.CancelCauseFunc
	mu      sync.Mutex
	timer   *time.Timer
	idle    time.Duration
	last    time.Time
	stopped bool
}

func (b *idleBody) expire() {
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return
	}
	// Reset can race an already-started callback. Recheck the most recent
	// byte under the same lock before deciding that this response stalled.
	if left := b.idle - time.Since(b.last); left > 0 {
		b.timer.Reset(left)
		b.mu.Unlock()
		return
	}
	b.stopped = true
	b.mu.Unlock()
	b.cancel(fmt.Errorf("%w for %s", errDownloadIdle, b.idle))
}

func (b *idleBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.mu.Lock()
		if !b.stopped {
			b.last = time.Now()
			b.timer.Reset(b.idle)
		}
		b.mu.Unlock()
	}
	if err != nil {
		if cause := context.Cause(b.ctx); cause != nil {
			err = cause
		}
		b.stop()
	}
	return n, err
}

func (b *idleBody) stop() {
	b.mu.Lock()
	b.stopped = true
	b.timer.Stop()
	b.mu.Unlock()
	b.cancel(nil)
}

func (b *idleBody) Close() error {
	b.stop()
	return b.ReadCloser.Close()
}

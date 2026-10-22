package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

type downloadRoundTrip func(*http.Request) (*http.Response, error)

func (f downloadRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Each byte is real progress even at one byte per minute. Fake time avoids
// runner-speed assumptions while exercising the actual client's old 10m cap.
func TestTinyDownloadContinuousProgressPastTenMinutes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		data := strings.Repeat("x", 12)
		fixtureDownloadClient(t, downloadRoundTrip(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: &pacedDownload{ctx: r.Context(), left: len(data)}, ContentLength: int64(len(data))}, nil
		}), 0)
		path := filepath.Join(t.TempDir(), "candidate.new")
		hash := sha256.Sum256([]byte(data))
		start := time.Now()
		err := fetch(WithMirror(context.Background(), "off"), Asset{URL: "https://fixture.invalid/candidate", Size: int64(len(data)), SHA256: hex.EncodeToString(hash[:])}, path)
		if err != nil {
			t.Fatalf("continuous download interrupted after %s: %v", time.Since(start), err)
		}
		if time.Since(start) <= 10*time.Minute {
			t.Fatal("fixture did not cross the old total deadline")
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != data {
			t.Fatalf("candidate = %q, %v", got, err)
		}
	})
}

type pacedDownload struct {
	ctx  context.Context
	left int
}

func (b *pacedDownload) Read(p []byte) (int, error) {
	if b.left == 0 {
		return 0, io.EOF
	}
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	select {
	case <-b.ctx.Done():
		return 0, b.ctx.Err()
	case <-timer.C:
		p[0] = 'x'
		b.left--
		return 1, nil
	}
}
func (b *pacedDownload) Close() error { return nil }

func fixtureDownloadClient(t *testing.T, base http.RoundTripper, idle time.Duration) {
	t.Helper()
	old := client
	c := *client
	installed, ok := client.Transport.(*idleTransport)
	if !ok {
		t.Fatal("update client has no download idle watchdog")
	}
	transport := *installed
	transport.base = base
	if idle > 0 {
		transport.idle = idle
	}
	c.Transport = &transport
	client = &c
	t.Cleanup(func() { client = old; c.CloseIdleConnections() })
}

func fixtureAsset(data string) Asset {
	hash := sha256.Sum256([]byte(data))
	return Asset{URL: "https://fixture.invalid/candidate", Size: int64(len(data)), SHA256: hex.EncodeToString(hash[:])}
}

type downloadReadFunc func([]byte) (int, error)

func (f downloadReadFunc) Read(p []byte) (int, error) { return f(p) }
func (f downloadReadFunc) Close() error               { return nil }

func TestTinyDownloadIdleAndCancellation(t *testing.T) {
	for _, which := range []string{"blocked read", "zero byte reads", "user cancel", "caller deadline"} {
		t.Run(which, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(WithMirror(context.Background(), "off"))
				defer cancel()
				if which == "user cancel" {
					time.AfterFunc(10*time.Second, cancel)
				}
				if which == "caller deadline" {
					var stop context.CancelFunc
					ctx, stop = context.WithTimeout(ctx, 10*time.Second)
					defer stop()
				}
				readStart := time.Now()
				fixtureDownloadClient(t, downloadRoundTrip(func(r *http.Request) (*http.Response, error) {
					body := downloadReadFunc(func(p []byte) (int, error) {
						if which == "zero byte reads" {
							if time.Since(readStart) > 4*time.Minute {
								return 0, io.ErrNoProgress
							}
							select {
							case <-time.After(10 * time.Second):
								return 0, nil
							case <-r.Context().Done():
								return 0, r.Context().Err()
							}
						}
						<-r.Context().Done()
						return 0, r.Context().Err()
					})
					return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body, ContentLength: -1}, nil
				}), 0)
				path := filepath.Join(t.TempDir(), "candidate.new")
				start := time.Now()
				err := fetch(ctx, fixtureAsset("x"), path)
				want, elapsed := errDownloadIdle, downloadIdleTimeout
				if which == "user cancel" {
					want, elapsed = context.Canceled, 10*time.Second
				}
				if which == "caller deadline" {
					want, elapsed = context.DeadlineExceeded, 10*time.Second
				}
				if !errors.Is(err, want) || time.Since(start) != elapsed {
					t.Fatalf("fetch = %v after %s, want %v after %s", err, time.Since(start), want, elapsed)
				}
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("partial candidate survived: %v", err)
				}
			})
		})
	}
}

func TestTinyDownloadRejectsInvalidCandidate(t *testing.T) {
	for _, which := range []string{"checksum", "size"} {
		t.Run(which, func(t *testing.T) {
			fixtureDownloadClient(t, downloadRoundTrip(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("bad")), ContentLength: -1}, nil
			}), 0)
			asset := fixtureAsset("bad")
			if which == "checksum" {
				asset.SHA256 = strings.Repeat("a", 64)
			} else {
				asset.Size++
			}
			path := filepath.Join(t.TempDir(), "candidate.new")
			err := fetch(WithMirror(context.Background(), "off"), asset, path)
			if err == nil || !strings.Contains(err.Error(), which) {
				t.Fatalf("invalid candidate accepted: %v", err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("invalid candidate survived: %v", err)
			}
		})
	}
}

func TestTinyDownloadRetryHasFreshIdleClock(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		attempts := 0
		fixtureDownloadClient(t, downloadRoundTrip(func(r *http.Request) (*http.Response, error) {
			attempts++
			var body io.ReadCloser
			if attempts == 1 {
				body = downloadReadFunc(func([]byte) (int, error) { <-r.Context().Done(); return 0, r.Context().Err() })
			} else {
				body = &pacedDownload{ctx: r.Context(), left: 12}
			}
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body, ContentLength: 12}, nil
		}), 0)
		path := filepath.Join(t.TempDir(), "candidate.new")
		err := download(WithMirror(context.Background(), "off"), fixtureAsset(strings.Repeat("x", 12)), path)
		if err != nil || attempts != 2 {
			t.Fatalf("retry = %v (%d attempts)", err, attempts)
		}
	})
}

func TestTinyDownloadCancelDuringRetryPause(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(WithMirror(context.Background(), "off"))
		defer cancel()
		attempts := 0
		fixtureDownloadClient(t, downloadRoundTrip(func(r *http.Request) (*http.Response, error) {
			attempts++
			time.AfterFunc(time.Second, cancel)
			return nil, io.ErrUnexpectedEOF
		}), 0)
		err := download(ctx, fixtureAsset("x"), filepath.Join(t.TempDir(), "candidate.new"))
		if !errors.Is(err, context.Canceled) || attempts != 1 {
			t.Fatalf("canceled retry = %v (%d attempts)", err, attempts)
		}
	})
}

func TestTinyDownloadRedirectAndNextAssetHaveFreshIdleClocks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		requests := 0
		fixtureDownloadClient(t, downloadRoundTrip(func(r *http.Request) (*http.Response, error) {
			requests++
			res := &http.Response{StatusCode: 200, Header: make(http.Header), Body: &pacedDownload{ctx: r.Context(), left: 12}, ContentLength: 12}
			if r.URL.Path == "/candidate" {
				res.StatusCode = http.StatusFound
				res.Header.Set("Location", "/asset")
				res.Body = &pacedDownload{ctx: r.Context(), left: 1}
				res.ContentLength = 1
			}
			return res, nil
		}), 0)
		for range 2 {
			if err := fetch(WithMirror(context.Background(), "off"), fixtureAsset(strings.Repeat("x", 12)), filepath.Join(t.TempDir(), "candidate.new")); err != nil {
				t.Fatal(err)
			}
		}
		if requests != 4 {
			t.Fatalf("requests = %d", requests)
		}
		// All redirect/final bodies must stop their timers on EOF/Close.
		time.Sleep(3 * downloadIdleTimeout)
	})
}

// An actual HTTP body blocks in the transport, rather than a fixture's Read.
// Only the watchdog's request cancellation can release it. No timing assertion
// depends on scheduling: the server sends one byte then deliberately stalls.
func TestTinyDownloadHTTPStallInterruptsBlockedRead(t *testing.T) {
	ended := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(ended)
		w.Write([]byte("x"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	fixtureDownloadClient(t, base, 50*time.Millisecond)
	asset := fixtureAsset("xy")
	asset.URL = srv.URL
	path := filepath.Join(t.TempDir(), "candidate.new")
	err := fetch(WithMirror(context.Background(), "off"), asset, path)
	if !errors.Is(err, errDownloadIdle) {
		t.Fatalf("blocked HTTP read = %v", err)
	}
	<-ended
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("partial HTTP candidate survived: %v", err)
	}
}

func TestTinyDownloadSourceFallbackHasFreshIdleClock(t *testing.T) {
	t.Setenv("MAGPIE_MIRRORS", "on")
	t.Setenv("MAGPIE_GITHUB_MIRROR", "https://mirror.fixture.invalid")
	synctest.Test(t, func(t *testing.T) {
		requests := 0
		fixtureDownloadClient(t, downloadRoundTrip(func(r *http.Request) (*http.Response, error) {
			requests++
			if r.URL.Host == "github.com" {
				return &http.Response{StatusCode: http.StatusServiceUnavailable, Status: "503 unavailable", Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
			}
			if r.URL.Host != "mirror.fixture.invalid" {
				t.Fatalf("unexpected host: %s", r.URL.Host)
			}
			return &http.Response{StatusCode: 200, Header: make(http.Header), ContentLength: 12, Body: &pacedDownload{ctx: r.Context(), left: 12}}, nil
		}), 0)
		asset := fixtureAsset(strings.Repeat("x", 12))
		asset.URL = Site + "/releases/download/tiny-v2.0.0/candidate"
		if err := fetch(WithMirror(context.Background(), "off"), asset, filepath.Join(t.TempDir(), "candidate.new")); err != nil {
			t.Fatal(err)
		}
		if requests != 2 {
			t.Fatalf("source attempts = %d", requests)
		}
	})
}

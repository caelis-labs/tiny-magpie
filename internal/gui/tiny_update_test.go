package gui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/settings"
	"github.com/yetone/magpie/internal/update"
)

func TestTinyUpdaterChecksWithoutDownloading(t *testing.T) {
	old := Version
	Version = "1.0.0"
	t.Cleanup(func() { Version = old })
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := settings.Load()
	s.NoAutoDownload = false
	s.AutoRestart = true
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"tag_name": "tiny-v2.0.0", "html_url": update.ReleasePage("2.0.0"), "prerelease": false, "body": "fixture"},
			{"tag_name": "tiny-v3.0.0", "html_url": update.ReleasePage("3.0.0"), "prerelease": true, "body": "preview fixture"},
		})
	}))
	defer srv.Close()
	t.Setenv("TINY_MAGPIE_UPDATE_FEED", srv.URL)
	u := &updater{exe: "fixture-that-must-never-be-replaced"}
	u.checkByClock()
	if j := u.json(); j.State != "found" || j.Latest != "2.0.0" || u.staged != "" {
		t.Fatalf("clock downloaded or failed to find: %+v", j)
	}
	if !u.begin() || u.begin() {
		t.Fatal("checks are not serialized")
	}
	if !updateDue(settings.Settings{UpdateEvery: 360}, time.Now().Add(-7*time.Hour), time.Now()) {
		t.Fatal("six-hour check never due")
	}
	if updateDue(settings.Settings{NoAutoUpdate: true, UpdateEvery: 360}, time.Time{}, time.Now()) {
		t.Fatal("checks run while disabled")
	}
}

func TestTinyUpdateAPIUsesOwnFeed(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { json.NewEncoder(w).Encode([]any{}) }))
	defer srv.Close()
	t.Setenv("TINY_MAGPIE_UPDATE_FEED", srv.URL)
	old := updates
	updates = &updater{}
	t.Cleanup(func() { updates = old })
	h := Handler(nil, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/update/check", strings.NewReader(`{}`)))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"current"`) {
		t.Fatalf("own update API unavailable: %d %s", w.Code, w.Body)
	}
}

func TestTinyUpdaterClockAndShutdown(t *testing.T) {
	old := Version
	Version = "1.0.0"
	t.Cleanup(func() { Version = old })
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	asked := make(chan struct{}, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked <- struct{}{}
		json.NewEncoder(w).Encode([]any{})
	}))
	defer srv.Close()
	t.Setenv("TINY_MAGPIE_UPDATE_FEED", srv.URL)
	u := &updater{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { u.watch(ctx, 0, time.Hour); close(done) }()
	select {
	case <-asked:
	case <-time.After(3 * time.Second):
		t.Fatal("startup never checked Tiny releases")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("update clock outlived shutdown")
	}
	if j := u.json(); j.State != "latest" {
		t.Fatalf("startup check: %+v", j)
	}
}

func TestTinyChannelChangeRejectsStagedPreview(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := settings.Load()
	s.UpdateChannel = "stable"
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "candidate.new")
	if err := os.WriteFile(file, []byte("preview bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	u := &updater{channel: "preview", staged: file, exe: filepath.Join(t.TempDir(), "installed")}
	if u.install(false) {
		t.Fatal("installed old-channel preview after selecting stable")
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatal("channel refusal destroyed staged bytes", err)
	}
}

func TestTinyDefaultChannelSaveRespectsChecksOff(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { json.NewEncoder(w).Encode([]any{}) }))
	defer srv.Close()
	t.Setenv("TINY_MAGPIE_UPDATE_FEED", srv.URL)
	s := settings.Load()
	s.NoAutoUpdate = true
	s.UpdateChannel = ""
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	old := updates
	updates = &updater{}
	t.Cleanup(func() { updates = old })
	s.UpdateChannel = "stable"
	s.NoUpdatePill = true
	body, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(w, httptest.NewRequest("POST", "/api/settings", strings.NewReader(string(body))))
	if w.Code != 200 {
		t.Fatalf("save: %d %s", w.Code, w.Body)
	}
	updates.mu.Lock()
	state := updates.state
	updates.mu.Unlock()
	if state != "" {
		// A broken implementation starts a check. Keep its failure entirely
		// on the fixture and let it finish before restoring the environment.
		for range 100 {
			updates.mu.Lock()
			busy := updates.state == "checking" || updates.state == "downloading"
			updates.mu.Unlock()
			if !busy {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("ordinary save started an update check while disabled: %s", state)
	}
}

func TestTinyChannelChangeDuringCheckRechecksWithClockOff(t *testing.T) {
	old := Version
	Version = "1.0.0"
	t.Cleanup(func() { Version = old })
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := settings.Load()
	s.NoAutoUpdate = true
	s.UpdateChannel = "preview"
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	started, release, second := make(chan struct{}), make(chan struct{}), make(chan struct{}, 1)
	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count.Add(1) == 1 {
			close(started)
			<-release
		} else {
			select {
			case second <- struct{}{}:
			default:
			}
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{"tag_name": "tiny-v2.0.0", "html_url": update.ReleasePage("2.0.0"), "prerelease": true},
			{"tag_name": "tiny-v1.1.0", "html_url": update.ReleasePage("1.1.0")},
		})
	}))
	defer srv.Close()
	t.Setenv("TINY_MAGPIE_UPDATE_FEED", srv.URL)
	u := &updater{exe: "fixture-no-download"}
	done := make(chan struct{})
	go func() { u.check(); close(done) }()
	<-started
	s.UpdateChannel = "stable"
	if err := settings.Save(s); err != nil {
		close(release)
		t.Fatal(err)
	}
	close(release)
	<-done
	select {
	case <-second:
	case <-time.After(5 * time.Second):
		t.Fatal("channel change never rechecked with the clock off")
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		j := u.json()
		if j.State == "found" && j.Latest == "1.1.0" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("wrong channel after in-flight check: %+v", u.json())
}

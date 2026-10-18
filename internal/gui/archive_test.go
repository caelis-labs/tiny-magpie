package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"os"
	"path/filepath"

	"testing"

	"github.com/yetone/magpie/internal/davsync"
	"github.com/yetone/magpie/internal/gateway"
)

// fakeArchive is a bucket of archived calls in memory; size, when set,
// is what it says an object's size is.
type fakeArchive struct {
	objs map[string][]byte
	size int64
}

func (f fakeArchive) Open(_ context.Context, name string) (io.ReadCloser, int64, error) {
	b, ok := f.objs[name]
	if !ok {
		return nil, 0, davsync.ErrNoObject
	}
	if f.size != 0 {
		return io.NopCloser(bytes.NewReader(b)), f.size, nil
	}
	return io.NopCloser(bytes.NewReader(b)), int64(len(b)), nil
}

func archiveHome(t *testing.T, f fakeArchive) http.Handler {
	t.Helper()
	h := t.TempDir()
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(h, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(h, ".cache"))
	os.MkdirAll(filepath.Join(h, "Downloads"), 0o700)
	was := archiveBucket
	archiveBucket = func() (archiveStore, bool) { return f, true }
	t.Cleanup(func() { archiveBucket = was })
	return Handler(nil, nil)
}

func archivedCall(t *testing.T, req, res string) []byte {
	t.Helper()
	b, err := json.Marshal(gateway.Archived{ID: "2026-10-02/101010-00000000000000aa", Call: gateway.Call{Model: "p/m", Status: 200},
		Request:  gateway.ArchivePart{Method: "POST", Path: "/v1/messages", Body: req, Size: int64(len(req))},
		Response: gateway.ArchivePart{Status: 200, Body: res, Size: int64(len(res))}})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

const archiveQuery = "?date=2026-10-02&id=101010-00000000000000aa"
const archiveObject = "archive/2026-10-02/101010-00000000000000aa.json"

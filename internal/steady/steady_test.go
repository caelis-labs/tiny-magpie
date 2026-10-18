package steady

import (
	"errors"
	"io/fs"
	"os"
	"runtime"
	"testing"
	"testing/synctest"
	"time"
)

// Probing an empty profile must not spend the longer replacement budget on
// every file that has never existed.
func TestMissingProbeStaysShort(t *testing.T) {
	windows = true
	defer func() { windows, stat = runtime.GOOS == "windows", os.Stat }()
	stat = func(string) (fs.FileInfo, error) { return nil, fs.ErrNotExist }
	synctest.Test(t, func(t *testing.T) {
		started := time.Now()
		if _, err := Stat("never-created"); !errors.Is(err, fs.ErrNotExist) {
			t.Fatal(err)
		}
		if elapsed := time.Since(started); elapsed >= 100*time.Millisecond {
			t.Fatalf("empty-profile probe took %v", elapsed)
		}
	})
}

// as Windows has it: a file being renamed over is, for a moment, not there
func TestStatWaitsOutAReplace(t *testing.T) {
	windows = true
	defer func() { windows, stat, readFile = runtime.GOOS == "windows", os.Stat, os.ReadFile }()
	missing := 3
	stat = func(p string) (fs.FileInfo, error) {
		if missing > 0 {
			missing--
			return nil, fs.ErrNotExist
		}
		return os.Stat(p)
	}
	p := t.TempDir()
	if _, err := Stat(p); err != nil {
		t.Fatalf("a file being replaced taken for gone: %v", err)
	}
	// A known file can disappear beyond the short Stat probe budget.
	missingUntil := time.Now().Add(100 * time.Millisecond)
	readFile = func(p string) ([]byte, error) {
		if time.Now().Before(missingUntil) {
			return nil, &fs.PathError{Op: "open", Path: p, Err: fs.ErrNotExist}
		}
		return []byte("accounts"), nil
	}
	if b, err := ReadThere(p); err != nil || string(b) != "accounts" {
		t.Fatalf("read %q, %v", b, err)
	}
	// one really gone is said so, after the wait
	stat = os.Stat
	if _, err := Stat(p + "-none"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a file not there: %v", err)
	}
}

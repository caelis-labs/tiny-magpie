package usage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/sessions"
)

// The file is of the code that makes a call's rows, not of the binary: an
// upgrade that changed no source of this package reads it as the last run
// left it. The name is of every source file here but the tests, and of
// the call shards' version.
func TestRequestDiskNamesItsSource(t *testing.T) {
	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	io.WriteString(h, sessions.CallCacheVersion)
	sources := 0
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(h, "\x00%s\x00%d\x00", name, len(b))
		h.Write(b)
		sources++
	}
	if want := fmt.Sprintf("source %x", h.Sum(nil)); requestDiskBinary() != want || sources < 10 {
		t.Fatalf("named %q, want %q of %d source files", requestDiskBinary(), want, sources)
	}
}

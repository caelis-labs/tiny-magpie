package gui

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

// A future upstream merge must not restore the old header or Windows icon
// while the app/tray still use Tiny's separate identity.
func TestTinyBrandAssetConsistency(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	read := func(name string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	source := regexp.MustCompile(`d="([^"]+)"`).FindSubmatch(read("build/icon/magpie-small.svg"))
	header := regexp.MustCompile(`<path id="bird" fill-rule="evenodd" d="([^"]+)"`).FindSubmatch(read("internal/gui/assets/index.html"))
	if len(source) != 2 || len(header) != 2 || !bytes.Equal(source[1], header[1]) {
		t.Fatal("control bar bird diverged from Tiny SVG source")
	}
	if !bytes.Contains(read("internal/gui/assets/icons/favicon.svg"), source[1]) {
		t.Fatal("favicon diverged from Tiny SVG source")
	}
	if !bytes.Equal(read("internal/gui/icon.png"), read("build/windows/icon-64.png")) {
		t.Fatal("Windows resource still uses a different app icon")
	}
}

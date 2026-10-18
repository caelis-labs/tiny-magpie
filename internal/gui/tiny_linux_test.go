//go:build linux && !nogui

package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTinyWebDataDoesNotMigrateAnotherProgram(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_DATA_HOME", root)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	old := filepath.Base(exe)
	if strings.HasPrefix(old, programName+"-") {
		t.Skip("test executable uses tiny's own legacy namespace")
	}
	from := filepath.Join(root, old, "localstorage")
	if err := os.MkdirAll(from, 0700); err != nil {
		t.Fatal(err)
	}
	keepWebData()
	if _, err := os.Stat(from); err != nil {
		t.Fatal("another program's WebKit data moved", err)
	}
	if _, err := os.Stat(filepath.Join(root, programName)); !os.IsNotExist(err) {
		t.Fatalf("imported another program's WebKit data: %v", err)
	}
}

//go:build !darwin && !windows && !android

package autostart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A record a write cut short is not one that opens magpie at login: the
// desktop starts no entry with no Exec, and none at all with nothing in it,
// so Open at login reads off rather than saying on for a record that never
// starts anything.
func TestSetThenTruncatedRecordReadsAsOff(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", home+"/.config")
	t.Setenv("APPIMAGE", "/opt/my apps/magpie.AppImage")
	if err := Set(true); err != nil {
		t.Fatal(err)
	}
	if Enabled() == Off {
		t.Fatal("not on after Set(true)")
	}
	p := record()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	// the head of the entry, cut off before the Exec that names the
	// program, which is what a write that goes no further leaves behind
	head := b[:strings.Index(string(b), "Exec=")]
	if err := os.WriteFile(p, head, 0o644); err != nil {
		t.Fatal(err)
	}
	if Enabled() != Off {
		t.Errorf("a desktop entry with no Exec reads as on:\n%s", head)
	}
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if Enabled() != Off {
		t.Error("an empty desktop entry reads as on")
	}
	// an Exec key with nothing after it names nothing to run either, and the
	// desktop starts no entry that names nothing
	if err := os.WriteFile(p, []byte("[Desktop Entry]\nType=Application\nName=magpie\nExec=\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Enabled() != Off {
		t.Error("a desktop entry with an empty Exec reads as on")
	}
}

func TestTinyAutostartPreservesOriginalMagpie(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	original := filepath.Join(root, "autostart", "magpie.desktop")
	os.MkdirAll(filepath.Dir(original), 0700)
	before := []byte("[Desktop Entry]\nType=Application\nExec=/original/magpie\n")
	if err := os.WriteFile(original, before, 0600); err != nil {
		t.Fatal(err)
	}
	if Enabled() != Off {
		t.Fatal("original entry was mistaken for tiny")
	}
	if err := enable("/fixture/tiny-magpie"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "autostart", "tiny-magpie.desktop")); err != nil {
		t.Fatal(err)
	}
	if err := disable(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(original)
	if err != nil || string(after) != string(before) {
		t.Fatalf("original autostart changed: %v %q", err, after)
	}
}

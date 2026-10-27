//go:build unix

package usage

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

// A file cut short under its mapping by another program is a read that
// failed, not a crash.
func TestRequestDiskMappingCutShort(t *testing.T) {
	path := filepath.Join(t.TempDir(), "requests-v1.bin")
	if err := os.WriteFile(path, make([]byte, 1<<20), 0o600); err != nil {
		t.Fatal(err)
	}
	m, ok := mapDiskFile(path)
	if !ok {
		t.Fatal("not mapped")
	}
	var last byte
	if !readMapped(m, func() { last = m.data[len(m.data)-1] }) || last != 0 {
		t.Fatal("a whole file's read failed")
	}
	if err := os.Truncate(path, 0); err != nil {
		t.Fatal(err)
	}
	if readMapped(m, func() { last = m.data[len(m.data)-1] }) {
		t.Fatal("a read past the cut passed")
	}
	runtime.KeepAlive(m)
}

// A spill that can't be written has each of its chunks made again: none is
// left without its archive.
func TestArchiveSpillFailedMadeAgain(t *testing.T) {
	pageHome(t)
	sp := newArchiveSpill()
	c := &rowChunk{Archive: []byte("archive")}
	if !sp.put(c, 7) {
		t.Fatal("not put")
	}
	if c.Archive != nil {
		t.Fatal("archive kept on the heap")
	}
	kept, next := *c, *c
	sp.hand(&kept, &next)
	sp.f.Close() // the flush fails
	var made []int
	sp.done(func(i int) *rowChunk {
		made = append(made, i)
		return &rowChunk{Archive: []byte("again")}
	})
	if !slices.Equal(made, []int{7}) {
		t.Fatalf("made again %v, want [7]", made)
	}
	for _, c := range []*rowChunk{&kept, &next} {
		if string(c.Archive) != "again" || c.mapped != nil {
			t.Errorf("archive %q mapped %v, want the one made again, of the heap", c.Archive, c.mapped)
		}
	}
	if _, err := os.Stat(sp.f.Name()); !os.IsNotExist(err) {
		t.Errorf("spill left: %v", err)
	}
}

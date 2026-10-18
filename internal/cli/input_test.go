package cli

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func piped(t *testing.T, lines ...string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) > 0 {
		w.WriteString(strings.Join(lines, "\n") + "\n")
	}
	w.Close()
	oldFile, oldReader := os.Stdin, stdin
	os.Stdin, stdin = r, bufio.NewReader(r)
	t.Cleanup(func() {
		os.Stdin, stdin = oldFile, oldReader
		r.Close()
	})
}

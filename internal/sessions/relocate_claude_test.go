package sessions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func relocationFixture(t *testing.T) (string, string, ClaudeRelocation) {
	t.Helper()
	root := t.TempDir()
	in := ClaudeRelocation{From: filepath.Join(root, "old_app"), To: filepath.Join(root, "new_app")}
	if err := os.Mkdir(in.To, 0o700); err != nil {
		t.Fatal(err)
	}
	name, _ := claudeProjectName(in.From)
	source := filepath.Join(root, "projects", name)
	row := map[string]any{
		"type": "user", "cwd": in.From, "sessionId": mgCC, "uuid": "unchanged",
		"message": map[string]any{"role": "user", "content": "Historical path: " + in.From},
	}
	b, _ := json.Marshal(row)
	for p, content := range map[string][]byte{
		mgCC + ".jsonl": b,
		filepath.Join(mgCC, "subagents", "agent-a.jsonl"): b,
		filepath.Join(mgCC, "tool-results", "result.txt"): []byte(in.From),
		filepath.Join("memory", "MEMORY.md"):              []byte(in.From),
	} {
		target := filepath.Join(source, p)
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ageRelocation(t, source)
	return root, source, in
}

func ageRelocation(t *testing.T, root string) {
	t.Helper()
	old := time.Now().Add(-time.Hour)
	if err := filepath.WalkDir(root, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return os.Chtimes(p, old, old)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestClaudeProjectName(t *testing.T) {
	got, err := claudeProjectName("/home/user/my_app")
	if err != nil || got != "-home-user-my-app" {
		t.Fatalf("%q %v", got, err)
	}
	if _, err := claudeProjectName("/" + strings.Repeat("a", 200)); err == nil {
		t.Fatal("unsupported hashed path accepted")
	}
	if _, _, err := relocationPaths(ClaudeRelocation{From: "/a_b", To: "/a-b", WSL: "Ubuntu"}); err == nil {
		t.Fatal("encoding collision accepted")
	}
}

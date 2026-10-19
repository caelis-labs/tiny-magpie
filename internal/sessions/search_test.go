package sessions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// claudeFile writes a Claude Code session of the given lines and returns it
// as a session to read.
func claudeFile(t *testing.T, lines ...string) Session {
	t.Helper()
	path := filepath.Join(t.TempDir(), "11111111-2222-3333-4444-555555555555.jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return Session{Agent: "claude", ID: "11111111-2222-3333-4444-555555555555", Path: path}
}

const (
	claudeAsked = `{"type":"user","message":{"role":"user","content":"Fix the   login\nbug in Auth.go please"},"uuid":"u1"}`
	claudeTool  = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"grep -r rainbow ."}}]},"uuid":"a1"}`
	claudeSaid  = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"The rainbow table lookup in auth.go skips the salt."}]},"uuid":"a2"}`
)

func TestContainsFold(t *testing.T) {
	for _, c := range []struct {
		b, needle string
		want      bool
	}{
		{"The Redis cache", "redis", true},
		{"THE REDIS CACHE", "redis", true},
		{"rRedis", "redis", true},
		{"rediS", "redis", true},
		{"redi", "redis", false},
		{"Rxdis redix", "redis", false},
		{"登录失败", "登录", true},
		{"登", "登录", false},
		{"", "a", false},
		{"abc", "", true},
	} {
		if got := containsFold([]byte(c.b), []byte(c.needle)); got != c.want {
			t.Errorf("containsFold(%q, %q) = %v", c.b, c.needle, got)
		}
	}
}

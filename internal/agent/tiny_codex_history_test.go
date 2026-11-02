package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/gateway"
)

// Disconnect restores model configuration without rewriting Agent history.
func TestTinyCodexDisconnectLeavesHistoryUntouched(t *testing.T) {
	base := gateway.URL() + gateway.CodexPath
	home, read := codexHome(t, `{"tokens":{"access_token":"x","id_token":"x.e30.x"}}`,
		"model = \"gpt-5.5\"\nopenai_base_url = \""+base+"\"\n")
	if codexLeft(home) {
		t.Fatalf("Codex through magpie read as let go:\n%s", read())
	}
	dir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(filepath.Join(dir, "thread-writer-locks"), 0o755); err != nil {
		t.Fatal(err)
	}
	day := filepath.Join(dir, "sessions", "2026", "10", "30")
	os.MkdirAll(day, 0o755)
	path := filepath.Join(day, "rollout-2026-10-30T15-05-20-01a15820-ff81-7840-bb3f-e18771c53aa8.jsonl")
	line := `{"timestamp":"2026-10-30T07:05:40.551Z","type":"response_item","payload":{"type":"reasoning","id":"rs_18b3f2a9c1d2e3f40000002a","summary":[{"type":"summary_text","text":"先看 go.mod"}],"encrypted_content":null}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	os.Chtimes(path, old, old)

	if err := codex(home).Unwire(); err != nil {
		t.Fatal(err)
	}
	if !codexLeft(home) {
		t.Fatalf("Codex after Disconnect still read as through magpie:\n%s", read())
	}
	b, _ := os.ReadFile(path)
	if string(b) != line {
		t.Fatalf("after Disconnect:\n%s", b)
	}
}

package usage

// Shared sandbox support for Tiny boundary tests.
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/sessions"
)

// holdClock stops Clock at at for the test and gives at back, for the test
// to stamp its calls by: midnight then never falls between a call and the
// period it is asked for in.
func holdClock(t *testing.T, at time.Time) time.Time {
	t.Helper()
	old := Clock
	Clock = func() time.Time { return at }
	t.Cleanup(func() { Clock = old })
	return at
}

func pageHome(t *testing.T) {
	sessionHome(t)
	sessions.Reset()
	catalog.Reset()
	t.Cleanup(func() { sessions.Reset(); catalog.Reset() })
	os.MkdirAll(filepath.Dir(catalog.CachePath()), 0700)
	os.WriteFile(catalog.CachePath(), []byte(`{"openai":{"id":"openai","models":{"m":{"id":"m","cost":{"input":2,"output":8,"cache_read":0.5,"cache_write":2.5}}}}}`), 0600)
}

func sessionHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	for k, v := range map[string]string{"HOME": home, "USERPROFILE": home, "XDG_CONFIG_HOME": filepath.Join(home, ".config"), "XDG_CACHE_HOME": filepath.Join(home, ".cache"), "CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude"), "CODEX_HOME": filepath.Join(home, ".codex"), "OPENAI_BASE_URL": ""} {
		t.Setenv(k, v)
	}
	return home
}

func writeCodexCall(t *testing.T, path, id, session, source, responseID, model, effort string, at time.Time, input, output int) {
	t.Helper()
	var sourceJSON json.RawMessage
	if err := json.Unmarshal([]byte(source), &sourceJSON); err != nil {
		t.Fatal(err)
	}
	lines := []any{
		map[string]any{"timestamp": "2026-09-30T10:00:00Z", "type": "session_meta", "payload": map[string]any{"id": id, "session_id": session, "source": sourceJSON}},
		map[string]any{"timestamp": at.Add(-time.Second).Format(time.RFC3339Nano), "type": "turn_context", "payload": map[string]any{"turn_id": "turn-" + responseID, "model": model, "effort": effort}},
		map[string]any{"timestamp": at.Format(time.RFC3339Nano), "type": "token_usage_record", "payload": map[string]any{
			"response_id": responseID, "session_id": session, "thread_id": id, "turn_id": "turn-" + responseID,
			"usage": map[string]int{"input_tokens": input, "cached_input_tokens": 0, "cache_write_input_tokens": 0, "output_tokens": output, "reasoning_output_tokens": 0, "total_tokens": input + output},
		}},
	}
	var data []byte
	for _, line := range lines {
		b, err := json.Marshal(line)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, b...)
		data = append(data, '\n')
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

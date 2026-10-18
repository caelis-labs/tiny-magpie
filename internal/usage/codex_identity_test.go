package usage

import (
	"encoding/json"

	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/sessions"
)

func writeCodexIdentity(t *testing.T, id, session string, source any) string {
	t.Helper()
	payload := map[string]any{"id": id, "session_id": session}
	if source != nil {
		payload["source"] = source
	}
	line, err := json.Marshal(map[string]any{"timestamp": "2026-09-30T10:00:00Z", "type": "session_meta", "payload": payload})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sessions.CodexDir(), "sessions", "2026", "09", "30", "rollout-2026-09-30T10-00-00-"+id+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(line, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	return path
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

func codexSpawn(parent string) map[string]any {
	return map[string]any{"subagent": map[string]any{"thread_spawn": map[string]any{"parent_thread_id": parent}}}
}

func findRowByResponse(rows []Row, responseID string) (Row, bool) {
	for _, row := range rows {
		if row.ResponseID == responseID {
			return row, true
		}
	}
	return Row{}, false
}

func findModelEffort(rows []sessions.ModelEffort, model, effort string) (sessions.ModelEffort, bool) {
	for _, row := range rows {
		if row.Model == model && row.Effort == effort {
			return row, true
		}
	}
	return sessions.ModelEffort{}, false
}

func TestCodexRequestPageFoldsCopiesBeforeResponseMatching(t *testing.T) {
	pageHome(t)
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	identities := map[string]sessions.CodexIdentity{
		"main-a":      {ID: "main-a", Role: "main", RootSession: "main-a"},
		"child-b":     {ID: "child-b", Role: "subagent", ParentSession: "main-a", RootSession: "main-a"},
		"other-child": {ID: "other-child", Role: "subagent", ParentSession: "main-a", RootSession: "main-a"},
		"nested-c":    {ID: "nested-c", Role: "subagent", ParentSession: "child-b", RootSession: "main-a"},
	}
	makePage := func(actors ...string) RequestPage {
		gateway := &rowChunk{}
		gateway.add(Row{Record: Record{Time: now.Add(-time.Minute), Agent: "codex", Provider: "openai", Model: "m", Effort: "high", Input: 500, Output: 50, Status: 200, ResponseID: "resp"}}, "", 1, false)
		local := &rowChunk{}
		for i, actor := range actors {
			local.add(Row{Record: Record{Time: now.Add(-time.Hour), Agent: "codex", Model: "native-model", Effort: "low", Input: 80, Output: 8,
				Session: "native-session-" + actor, ResponseID: "resp"}, Source: "log", codexActor: actor}, "", int64(i), false)
		}
		return buildRequestBlocksWithCodex(All, Filter{}, 0, 100, now, []*rowChunk{gateway}, []*rowChunk{local}, nil, nil, identities)
	}

	page := makePage("child-b", "child-b")
	if len(page.Rows) != 1 || page.Sum.Calls != 1 || page.Sum.Input != 500 || page.Rows[0].Model != "m" || page.Rows[0].Effort != "high" || page.Rows[0].SessionRole != "subagent" || page.Rows[0].Session != "main-a" || page.Rows[0].Subagent != "child-b" {
		t.Fatalf("identical native history copies were not folded into the gateway call: %+v", page)
	}
	page = makePage("child-b", "other-child")
	if len(page.Rows) != 1 || page.Sum.Calls != 1 || page.Sum.Input != 500 || page.Rows[0].SessionRole != "unknown" || page.Rows[0].Subagent != "" {
		t.Fatalf("conflicting native actors were guessed: %+v", page)
	}

	gateway, local := &rowChunk{}, &rowChunk{}
	gateway.add(Row{Record: Record{Time: now.Add(-time.Minute), Agent: "codex", Provider: "openai", Model: "m", Input: 300, Output: 30, Status: 200, ResponseID: "resp-parent"}}, "", 1, false)
	local.add(Row{Record: Record{Time: now.Add(-time.Hour), Agent: "codex", Model: "native-model", Input: 30, Output: 3,
		Session: "native-session", ResponseID: "resp-parent"}, Source: "log", codexActor: "nested-c"}, "", 1, false)
	page = buildRequestBlocksWithCodex(All, Filter{Query: "child-b"}, 0, 100, now, []*rowChunk{gateway}, []*rowChunk{local}, nil, nil, identities)
	if len(page.Rows) != 1 || page.Rows[0].ResponseID != "resp-parent" || page.Rows[0].ParentAgent != "child-b" {
		t.Fatalf("request-page parent-id filter = %+v", page.Rows)
	}
}

func TestLogRecordResponseIDRemainsCodexOnly(t *testing.T) {
	row := Row{Record: logRecord(sessions.Call{Agent: "claude", ResponseID: "claude-response"})}
	if row.ResponseID != "" || row.SessionRole != "" {
		t.Fatalf("non-Codex call gained Codex identity data: %+v", row)
	}
	encoded, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	if containsJSONField(encoded, "response_id") || containsJSONField(encoded, "session_role") {
		t.Fatalf("non-Codex row JSON changed: %s", encoded)
	}
	identities := map[string]sessions.CodexIdentity{"thread": {ID: "thread", Role: "main", RootSession: "thread"}}
	call := sessions.Call{Agent: "claude", ThreadID: "thread"}
	if got := codexIdentityRow(Row{Record: logRecord(call)}, codexActor(call), identities); got.SessionRole != "" || got.Subagent != "" {
		t.Fatalf("non-Codex ThreadID changed identity fields: %+v", got)
	}
}

func containsJSONField(data []byte, field string) bool {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return false
	}
	_, ok := object[field]
	return ok
}

func TestCodexIdentityFixturePath(t *testing.T) {
	pageHome(t)
	path := filepath.Join(sessions.CodexDir(), "sessions", "2026", "09", "30", "rollout-2026-09-30T10-00-00-main-a.jsonl")
	writeCodexCall(t, path, "main-a", "native-main", `"vscode"`, "resp-fixture", "m", "low", time.Date(2026, 9, 30, 11, 0, 0, 0, time.UTC), 1, 1)
	calls := sessions.Calls(time.Time{})
	if len(calls) != 1 || calls[0].ThreadID != "main-a" || calls[0].ResponseID != "resp-fixture" || calls[0].Model != "m" || calls[0].Effort != "low" {
		t.Fatalf("fixture did not produce a Codex call: %s (%d calls): %+v", path, len(calls), calls)
	}
	if got := logRecord(calls[0]).ResponseID; got != "resp-fixture" {
		t.Fatalf("local Codex response ID was lost from its usage row: %q", got)
	}
}

func TestCodexActorSurvivesPackedRequestChunk(t *testing.T) {
	identities := map[string]sessions.CodexIdentity{"child": {ID: "child", Role: "subagent", ParentSession: "main", RootSession: "main"}}
	chunk := &rowChunk{}
	for i := range 64 {
		chunk.add(Row{Record: Record{Agent: "codex", Model: "model", Session: "old", ResponseID: "response", Status: 200}, Source: "log", codexActor: "child"}, "", int64(i), false)
	}
	packed := chunk.pack()
	if len(packed.Archive) == 0 {
		t.Fatal("fixture did not use packed chunk storage")
	}
	decoded := packed.unpack()
	if decoded == nil {
		t.Fatal("packed chunk could not be decoded")
	}
	decoded.codexIdentities = identities
	row := decoded.row(0)
	if row.SessionRole != "subagent" || row.Session != "main" || row.Subagent != "child" {
		t.Fatalf("packed Codex actor identity was lost: %+v", row)
	}
	if len(chunk.codexActors) != len(chunk.Rows) || chunk.codexActors[0] != "child" {
		t.Fatal("query identity changed the source chunk")
	}
}

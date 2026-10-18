package usage

import (
	"encoding/json"
	"math"
	"os"

	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/sessions"
)

func codexCompactionCall(at time.Time, responseID string, input, output int, compacted bool) sessions.Call {
	return sessions.Call{Time: at, Agent: "codex", Session: "fixture-session", ThreadID: "fixture-thread", ResponseID: responseID,
		Model: "native-model", Effort: "native-effort", Tokens: sessions.Tokens{Input: input, Output: output}, Millis: 1000, Compaction: compacted}
}

func compactionValue(row Row) (bool, bool) {
	if row.Compaction == nil {
		return false, false
	}
	return *row.Compaction, true
}

func TestCodexCompactionLedgerCorrelationAndDuplicateMerge(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "late marker last", true: "late marker first"}[reverse], func(t *testing.T) {
			pageHome(t)
			now := holdClock(t, time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC))
			at := now.Add(-time.Hour)
			for i, row := range []struct {
				id     string
				input  int
				output int
				kind   string
				effort string
			}{
				{"resp-true", 11, 1, "chat", "gw-high"},
				{"resp-duplicate", 13, 2, "tool", "gw-low"},
				{"resp-false", 17, 3, "analysis", "gw-medium"},
				{"resp-unpaired", 19, 4, "chat", "gw-high"},
			} {
				Append(Record{Time: at.Add(time.Duration(i) * time.Second), Agent: "codex", Provider: "openai", Model: "m", Served: "m",
					Effort: row.effort, Kind: row.kind, Input: row.input, Output: row.output, Status: 200, ResponseID: row.id})
			}
			duplicateFalse := codexCompactionCall(at.Add(time.Second), "resp-duplicate", 900, 90, false)
			duplicateTrue := duplicateFalse
			duplicateTrue.Compaction = true
			duplicateTrue.Compact = 25 // internal call timing stays out of the HTTP usage row
			duplicates := []sessions.Call{duplicateFalse, duplicateTrue}
			if reverse {
				duplicates[0], duplicates[1] = duplicates[1], duplicates[0]
			}
			native := []sessions.Call{
				codexCompactionCall(at, "resp-true", 800, 80, true),
				duplicates[0], duplicates[1],
				codexCompactionCall(at.Add(2*time.Second), "resp-false", 700, 70, false),
				codexCompactionCall(at.Add(4*time.Second), "resp-direct-true", 23, 4, true),
				codexCompactionCall(at.Add(5*time.Second), "resp-direct-false", 29, 5, false),
			}
			old := LogCalls
			LogCalls = func(time.Time) []sessions.Call { return native }
			defer func() { LogCalls = old }()

			ledger := LedgerOf(All, Filter{})
			if ledger.Sum.Calls != 6 || len(ledger.Rows) != 6 || ledger.Sum.Input != 112 || ledger.Sum.Output != 19 {
				t.Fatalf("matched native calls were counted or billed again: totals=%+v rows=%+v", ledger.Sum, ledger.Rows)
			}
			if math.Abs(ledger.Sum.Cost-0.0002) > 1e-12 {
				t.Fatalf("gateway price changed or native call was billed twice: %v", ledger.Sum.Cost)
			}
			for id, want := range map[string]bool{"resp-true": true, "resp-duplicate": true, "resp-false": false} {
				row, ok := findRowByResponse(ledger.Rows, id)
				got, known := compactionValue(row)
				if !ok || !known || got != want {
					t.Errorf("%s compaction = %v, known=%v, found=%v; want %v", id, got, known, ok, want)
				}
			}
			row, ok := findRowByResponse(ledger.Rows, "resp-true")
			if !ok || row.Model != "m" || row.Effort != "gw-high" || row.Kind != "chat" || row.Input != 11 || row.Output != 1 || !row.Priced {
				t.Fatalf("native metadata replaced gateway request details: %+v", row)
			}
			if _, known := compactionValue(mustResponseRow(t, ledger.Rows, "resp-unpaired")); known {
				t.Fatal("an unpaired gateway row claimed native compaction evidence")
			}
			if compacted, known := compactionValue(mustResponseRow(t, ledger.Rows, "resp-direct-true")); !known || !compacted {
				t.Fatal("native-only compacted call lost its marker")
			}
			if compacted, known := compactionValue(mustResponseRow(t, ledger.Rows, "resp-direct-false")); !known || compacted {
				t.Fatal("native-only non-compacted call lost its false marker")
			}
			trueJSON, _ := json.Marshal(mustResponseRow(t, ledger.Rows, "resp-direct-true"))
			falseJSON, _ := json.Marshal(mustResponseRow(t, ledger.Rows, "resp-false"))
			orphanJSON, _ := json.Marshal(mustResponseRow(t, ledger.Rows, "resp-unpaired"))
			if string(trueJSON) == "" || !strings.Contains(string(trueJSON), `"compaction":true`) || strings.Contains(string(trueJSON), `"compact_ms"`) {
				t.Fatalf("Codex usage JSON marker: %s", trueJSON)
			}
			if !strings.Contains(string(falseJSON), `"compaction":false`) || strings.Contains(string(orphanJSON), `"compaction"`) {
				t.Fatalf("Codex three-state usage JSON: false=%s orphan=%s", falseJSON, orphanJSON)
			}
		})
	}
}

func mustResponseRow(t *testing.T, rows []Row, responseID string) Row {
	t.Helper()
	row, ok := findRowByResponse(rows, responseID)
	if !ok {
		t.Fatalf("response %s missing from %+v", responseID, rows)
	}
	return row
}

func TestCodexRequestPageDoesNotMutateSharedCompactionChunks(t *testing.T) {
	pageHome(t)
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	yes, no := true, false
	makeChunks := func(responseID string, compacted ...bool) (*rowChunk, *rowChunk) {
		gateway, native := &rowChunk{}, &rowChunk{}
		gateway.add(Row{Record: Record{Time: now.Add(-time.Minute), Agent: "codex", Provider: "openai", Model: "gateway-model",
			Input: 12, Output: 2, Status: 200, ResponseID: responseID}}, "", 1, false)
		for i, marker := range compacted {
			marker := marker
			native.add(Row{Record: Record{Time: now.Add(-time.Hour), Agent: "codex", Model: "native-model", ResponseID: responseID,
				Compaction: &marker}, Source: "log"}, "", int64(i), false)
		}
		return gateway, native
	}
	build := func(gateway, native *rowChunk) RequestPage {
		var natives []*rowChunk
		if native != nil {
			natives = []*rowChunk{native}
		}
		return buildRequestBlocksWithCodex(All, Filter{}, 0, 100, now, []*rowChunk{gateway}, natives, nil, nil, nil)
	}
	assertMarker := func(t *testing.T, page RequestPage, want bool) {
		t.Helper()
		if len(page.Rows) != 1 {
			t.Fatalf("page rows = %d, want one gateway row: %+v", len(page.Rows), page.Rows)
		}
		got, known := compactionValue(page.Rows[0])
		if !known || got != want {
			t.Fatalf("page compaction = %v, known=%v; want %v", got, known, want)
		}
	}
	assertNoMarker := func(t *testing.T, page RequestPage) {
		t.Helper()
		if len(page.Rows) != 1 {
			t.Fatalf("page rows = %d, want one gateway row: %+v", len(page.Rows), page.Rows)
		}
		if _, known := compactionValue(page.Rows[0]); known {
			t.Fatalf("unpaired page has compaction evidence: %+v", page.Rows[0])
		}
	}

	t.Run("removing unique match clears evidence", func(t *testing.T) {
		gateway, native := makeChunks("resp-unique", yes)
		assertMarker(t, build(gateway, native), true)
		if gateway.Rows[0].Flags&(rowCompactionKnown|rowCompaction) != 0 {
			t.Fatal("request build wrote the native compaction marker into the cached gateway chunk")
		}
		assertNoMarker(t, build(gateway, nil))
		if gateway.Rows[0].Flags&(rowCompactionKnown|rowCompaction) != 0 {
			t.Fatal("unpaired request retained compaction evidence from an earlier page build")
		}
	})

	t.Run("changed duplicate copies recompute marker", func(t *testing.T) {
		gateway, native := makeChunks("resp-duplicate", no, yes)
		assertMarker(t, build(gateway, native), true)
		if native.Rows[0].Flags&(rowCompactionKnown|rowCompaction) != rowCompactionKnown {
			t.Fatal("duplicate folding wrote the merged marker into the cached native chunk")
		}
		if native.Rows[1].Flags&(rowCompactionKnown|rowCompaction) != rowCompactionKnown|rowCompaction {
			t.Fatal("request build changed the second cached native copy")
		}
		_, changedNative := makeChunks("resp-duplicate", no, no)
		assertMarker(t, build(gateway, changedNative), false)
	})

	t.Run("concurrent page builds share immutable chunks", func(t *testing.T) {
		gateway, native := makeChunks("resp-concurrent", no, yes)
		var wg sync.WaitGroup
		errs := make(chan string, 8)
		for i := 0; i < cap(errs); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				page := build(gateway, native)
				if len(page.Rows) != 1 {
					errs <- "page did not contain exactly one gateway row"
					return
				}
				got, known := compactionValue(page.Rows[0])
				if !known || !got {
					errs <- "page lost duplicate-merged compaction marker"
				}
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			t.Error(err)
		}
		if gateway.Rows[0].Flags&(rowCompactionKnown|rowCompaction) != 0 || native.Rows[0].Flags&(rowCompactionKnown|rowCompaction) != rowCompactionKnown {
			t.Fatal("concurrent page builds mutated shared source rows")
		}
	})
}

func hasEffortInput(rows []sessions.ModelEffort, model, effort string, input int) bool {
	row, ok := findModelEffort(rows, model, effort)
	return ok && row.Input == input
}

func appendCompactionLine(t *testing.T, path, responseID, threadID, sessionID, turnID string, at time.Time, input, output int) {
	t.Helper()
	usage := map[string]int{"input_tokens": input, "cached_input_tokens": 0, "cache_write_input_tokens": 0,
		"output_tokens": output, "reasoning_output_tokens": 0, "total_tokens": input + output}
	line, err := json.Marshal(map[string]any{"timestamp": at.Format(time.RFC3339Nano), "type": "compacted", "payload": map[string]any{
		"compaction_response_id": responseID,
		"latest_token_usage_record": map[string]any{"response_id": responseID, "session_id": sessionID, "thread_id": threadID, "turn_id": turnID,
			"usage": usage, "thread_token_usage": usage},
	}})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		t.Fatal(err)
	}
}

// A Codex call told from token_count alone (a rollout from before Codex wrote
// token_usage_record) has no response_id and can't carry the native marker:
// its compaction is unknown, never false.
func TestCodexCallWithoutResponseIDHasNoCompactionState(t *testing.T) {
	pageHome(t)
	now := holdClock(t, time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC))
	old := LogCalls
	LogCalls = func(time.Time) []sessions.Call {
		return []sessions.Call{
			codexCompactionCall(now.Add(-time.Hour), "", 31, 6, false),
			codexCompactionCall(now.Add(-time.Hour+time.Second), "resp-recorded", 37, 7, false),
		}
	}
	defer func() { LogCalls = old }()
	for name, rows := range map[string][]Row{"ledger": LedgerOf(All, Filter{}).Rows, "requests": QueryPage(All, Filter{}, 0, 50).Rows} {
		var legacy, recorded *Row
		for i := range rows {
			switch rows[i].Input {
			case 31:
				legacy = &rows[i]
			case 37:
				recorded = &rows[i]
			}
		}
		if legacy == nil || recorded == nil {
			t.Fatalf("%s: native rows missing: %+v", name, rows)
		}
		if b, _ := json.Marshal(legacy); strings.Contains(string(b), `"compaction"`) {
			t.Errorf("%s: a call with no response_id claimed a compaction state: %s", name, b)
		}
		if compacted, known := compactionValue(*recorded); !known || compacted {
			t.Errorf("%s: a recorded call lost its false marker: %+v", name, recorded.Compaction)
		}
	}
}

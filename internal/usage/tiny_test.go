package usage

import (
	"github.com/yetone/magpie/internal/sessions"
	"path/filepath"
	"testing"
	"time"
)

func TestTinyUsageCountsGatewayWithoutImportingAgentLogs(t *testing.T) {
	pageHome(t)
	before := LogCalls
	LogCalls = nil
	t.Cleanup(func() { LogCalls = before })
	now := holdClock(t, time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC))
	path := filepath.Join(sessions.CodexDir(), "sessions", "2026", "09", "30", "rollout-native.jsonl")
	writeCodexCall(t, path, "native", "native", `"vscode"`, "response-native", "native-model", "high", now.Add(-time.Hour), 9000, 1000)
	Append(Record{Time: now, Agent: "fixture", Provider: "fixture", Model: "gateway-model", Input: 7, Output: 3, Status: 200})
	ledger := LedgerOf(All, Filter{})
	if len(ledger.Rows) != 1 || ledger.Sum.Input != 7 || ledger.Sum.Output != 3 {
		t.Fatalf("imported native calls: %+v", ledger)
	}
	reads := 0
	page := queryPage(All, Filter{}, 0, 10, func(sessions.CallSource) []sessions.Call { reads++; return nil })
	if page.Total != 1 || reads != 0 {
		t.Fatalf("read Agent files: total=%d reads=%d", page.Total, reads)
	}
	if direct := Direct(All); direct.Calls != 0 {
		t.Fatalf("imported direct calls: %+v", direct)
	}
	if modelEffort := CodexModelEffort(All); len(modelEffort) != 0 {
		t.Fatalf("imported session consumption: %+v", modelEffort)
	}
}

package gateway

import (
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// masaka on Discord: clicking a part of the context window card shows the
// text the request sent there — the system prompt, each tool's definition,
// each turn — masked as the vendor got it, even where the request went to
// a model the user let have it unmasked. The route, which goes in the
// routing history on disk, still holds no text.
func TestPromptTextChatPieces(t *testing.T) {
	body := []byte(`{"model":"m","messages":[
		{"role":"system","content":"be brief"},
		{"role":"user","content":"read it"},
		{"role":"assistant","reasoning_content":"I should read it.","content":"Reading.","tool_calls":[{"id":"c1","type":"function","function":{"name":"read","arguments":"{\"path\":\"a.go\"}"}}]},
		{"role":"tool","tool_call_id":"c1","content":"package a"}]}`)
	b, ok := promptPieces(provider.Chat, body)
	if !ok {
		t.Fatal("not read")
	}
	var got []string
	for _, pc := range b.pieces[PartChat+"\x00"+itemKey("turn", "1")] {
		pc = pc.read()
		got = append(got, pc.Role+"="+pc.Text)
	}
	want := "user=read it|thinking=I should read it.|assistant=Reading.|call={\n  \"path\": \"a.go\"\n}"
	if strings.Join(got, "|") != want {
		t.Fatalf("got %q", strings.Join(got, "|"))
	}
	// the sizes are what they were before the text was kept
	if p := promptOf(provider.Chat, body); partOf(p, PartChat).Tokens != func() int {
		n := 0
		for _, it := range b.parts[PartChat] {
			n += it.Tokens
		}
		return n
	}() {
		t.Fatalf("sizes differ")
	}
}

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

// An image the request sent inline is shown as its picture in the text of
// the context window's Files (the owner: "这里要是能够显示图片就更牛逼了"),
// in each protocol's shape, a tool's result's image too. A linked image,
// an SVG or a source that isn't base64 stays a name: the GUI draws only
// what the request itself carried.
func TestPromptTextImagesArePictures(t *testing.T) {
	const png = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	const url = "data:image/png;base64," + png
	cases := []struct {
		from provider.Protocol
		body string
		want []string
	}{
		{provider.Anthropic, `{"messages":[{"role":"user","content":[
			{"type":"text","text":"look"},
			{"type":"image","source":{"type":"base64","media_type":"image/png","data":"` + png + `"}},
			{"type":"image","source":{"type":"url","url":"https://example.com/a.png"}},
			{"type":"image","source":{"type":"base64","media_type":"image/svg+xml","data":"PHN2Zz4="}}]},
			{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"Shot","input":{}}]},
			{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":[
				{"type":"image","source":{"type":"base64","media_type":"image/jpeg","data":"` + png + `"}}]}]}]}`,
			[]string{url, "", "", "data:image/jpeg;base64," + png}},
		{provider.Responses, `{"input":[{"role":"user","content":[
			{"type":"input_text","text":"look"},
			{"type":"input_image","image_url":"` + url + `"},
			{"type":"input_image","image_url":"https://example.com/a.png"}]}]}`,
			[]string{url, ""}},
		{provider.Chat, `{"messages":[{"role":"user","content":[
			{"type":"text","text":"look"},
			{"type":"image_url","image_url":{"url":"` + url + `"}},
			{"type":"image_url","image_url":{"url":"data:image/png;base64,<script>"}}]}]}`,
			[]string{url, ""}},
		{provider.Gemini, `{"contents":[{"role":"user","parts":[
			{"text":"look"},
			{"inlineData":{"mimeType":"image/webp","data":"` + png + `"}},
			{"fileData":{"mimeType":"image/png","fileUri":"https://example.com/a.png"}}]}]}`,
			[]string{"data:image/webp;base64," + png, ""}},
	}
	for _, c := range cases {
		b, ok := promptPieces(c.from, []byte(c.body))
		if !ok {
			t.Fatalf("%s: not read", c.from)
		}
		var got []string
		for key, pcs := range b.pieces {
			if !strings.HasPrefix(key, PartFiles+"\x00") {
				continue
			}
			for _, pc := range pcs {
				if pc = pc.read(); pc.Role == "image" {
					got = append(got, pc.Src)
				}
			}
		}
		if strings.Join(got, " | ") != strings.Join(c.want, " | ") {
			t.Fatalf("%s: srcs %q, want %q", c.from, got, c.want)
		}
	}
}

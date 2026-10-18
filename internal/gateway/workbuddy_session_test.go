package gateway

import (
	"net/http"

	"strings"

	"testing"
)

// WorkBuddy 5.8.2's bundled ModelProvider sends the session as
// X-Conversation-ID, not its per-turn request or per-message id.
func TestWorkBuddySessionHeader(t *testing.T) {
	const body = `{"messages":[{"role":"user","content":"hello"}]}`
	h := http.Header{}
	h.Set("X-Conversation-ID", "workbuddy-session")
	h.Set("X-Conversation-Request-ID", "turn-1")
	h.Set("X-Conversation-Message-ID", "message-1")
	for name, got := range map[string]string{
		"session": sessionOf(h), "native": nativeSessionOf(h),
		"conversation": conversationID(h, []byte(body)), "otel": observedSession(h),
	} {
		if got != "workbuddy-session" {
			t.Errorf("%s = %q, want WorkBuddy's actual session", name, got)
		}
	}
	if got := convOf(h, []byte(body)); got != "" {
		t.Errorf("an explicit WorkBuddy session still has an inferred conversation %q", got)
	}
	h.Set(SessionHeader, "override")
	if sessionOf(h) != "override" || nativeSessionOf(h) != "workbuddy-session" || observedSession(h) != "workbuddy-session" || conversationID(h, []byte(body)) != "workbuddy-session" {
		t.Fatal("the explicit override lost WorkBuddy's native session")
	}
	for _, header := range []string{"x-opencode-session", "x-session-affinity", "x-session-id", "session_id", "session-id", "x-claude-code-session-id"} {
		t.Run(header, func(t *testing.T) {
			h := http.Header{}
			h.Set(header, "existing-session")
			h.Set("X-Conversation-ID", "workbuddy-session")
			if sessionOf(h) != "existing-session" || nativeSessionOf(h) != "existing-session" || conversationID(h, []byte(body)) != "existing-session" {
				t.Fatal("WorkBuddy's fallback header overrode an existing session header")
			}
		})
	}
	h = http.Header{}
	h.Set("X-Conversation-Request-ID", "turn-only")
	h.Set("X-Conversation-Message-ID", "message-only")
	h.Set("X-Parent-Conversation-ID", "parent-only")
	if sessionOf(h) != "" || nativeSessionOf(h) != "" || conversationID(h, []byte(body)) != conversationID(http.Header{}, []byte(body)) {
		t.Fatal("a turn, message or parent id became the current session")
	}
	h.Set("x-conversation-id", "  ")
	if sessionOf(h) != "" || nativeSessionOf(h) != "" {
		t.Fatal("an empty conversation header became a session")
	}
	h.Set("x-conversation-id", "  "+strings.Repeat("s", 140)+"  ")
	if sessionOf(h) != strings.Repeat("s", 128) || nativeSessionOf(h) != strings.Repeat("s", 128) {
		t.Fatal("the new header bypassed the existing session bounds")
	}
}

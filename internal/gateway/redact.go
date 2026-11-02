package gateway

import (
	"context"
	"net/http"
	"sync"

	"github.com/yetone/magpie/internal/gwmw"
	"github.com/yetone/magpie/internal/provider"
)

// Tiny keeps upstream call sites without request-content masking.
// Credential scrubbing in logs is a separate retained security boundary.
func redacted(ctx context.Context, w http.ResponseWriter, body []byte) (http.ResponseWriter, []byte, func(), error) {
	return w, body, func() {}, nil
}
func redactedPrompt(ctx context.Context, w http.ResponseWriter, prompt string) (http.ResponseWriter, string, func(), error) {
	return w, prompt, func() {}, nil
}
func maskedText(ctx context.Context, s string) (string, int, error) { return s, 0, nil }
func restored(s string, escaped bool) string                        { return s }
func restoring(w http.ResponseWriter) (http.ResponseWriter, func()) { return w, func() {} }
func redactionWanted() bool                                         { return false }

// Tiny does not retain an archive copy of a request that was not sent.
func maskedCopy(ctx context.Context, body []byte) []byte { return nil }

// laterMasked is the body of a request that went as written, masked the
// first time an observer is handed it (withBodies): after its try, never
// on its way (#3973). nil when masking fails: nothing unmasked is handed
// on in its place.
type laterMasked struct {
	once  sync.Once
	ctx   context.Context
	plain []byte
	b     []byte
}

func (l *laterMasked) body() []byte {
	l.once.Do(func() {
		l.b, l.plain = maskedCopy(l.ctx, l.plain), nil
	})
	return l.b
}

// unredactedRoute says a request resolved to p, or to the group whose
// models are ms, goes only where the user set requests to go unmasked
// (provider.SkipsRedaction): p and each of its fallbacks, or every model of
// the group (a member's own fallbacks are not the group's, see planGroup).
// One that may go on to anyone else stays masked for all of them.
func unredactedRoute(p provider.Provider, isGroup bool, ms []provider.Member) bool {
	if isGroup {
		if len(ms) == 0 {
			return false
		}
		for _, m := range ms {
			if !m.Provider.SkipsRedaction() {
				return false
			}
		}
		return true
	}
	if !p.SkipsRedaction() {
		return false
	}
	// one that resolves to nothing isn't tried (plan); a routing group
	// among them is each of its members
	for _, m := range provider.FallbackModels(p) {
		if !m.Provider.SkipsRedaction() {
			return false
		}
	}
	return true
}

func maskedQuery(ctx context.Context, query string, p *provider.Provider) (string, error) {
	return query, nil
}

func redactedHits(ctx context.Context, w http.ResponseWriter, body []byte) (http.ResponseWriter, []byte, func(), maskHits, error) {
	return w, body, func() {}, maskHits{}, nil
}
func redactedPromptHits(ctx context.Context, w http.ResponseWriter, prompt string) (http.ResponseWriter, string, func(), maskHits, error) {
	return w, prompt, func() {}, maskHits{}, nil
}
func maskTextHits(ctx context.Context, s string) (gwmw.Masker, string, maskHits, error) {
	return nil, s, maskHits{}, nil
}

func maskStatus(err error) int { return http.StatusServiceUnavailable }

// Tiny never selects a different provider from the contents of a request.
func maskedFor(ctx context.Context, p provider.Provider, isGroup bool, ms []provider.Member) context.Context {
	return ctx
}

type agentTurnKey struct{}

func agentTurn(r *http.Request) *http.Request                         { return r }
func isAgentTurn(r *http.Request) bool                                { return false }
func MaskRoute(id string) bool                                        { return false }
func maskRouteFor(r *http.Request, hits maskHits) string              { return "" }
func codexMaskRoute(r *http.Request, rest string, body []byte) string { return "" }

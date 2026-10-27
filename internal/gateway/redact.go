package gateway

import (
	"context"
	"net/http"

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
	for _, id := range p.Fallback {
		// one that resolves to nothing isn't tried (plan)
		if fp, _, ok := provider.Resolve(id); ok && !fp.SkipsRedaction() {
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

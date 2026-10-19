package gateway

import (
	"net/http"

	"github.com/yetone/magpie/internal/provider"
)

// Tiny keeps upstream call sites without request-content masking.
// Credential scrubbing in logs is a separate retained security boundary.
func redacted(w http.ResponseWriter, body []byte) (http.ResponseWriter, []byte, func(), error) {
	return w, body, func() {}, nil
}
func redactedPrompt(w http.ResponseWriter, prompt string) (http.ResponseWriter, string, func(), error) {
	return w, prompt, func() {}, nil
}
func maskedText(s string) (string, int, error)                      { return s, 0, nil }
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

func maskedQuery(query string, p *provider.Provider) (string, error) { return query, nil }

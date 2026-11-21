package gateway

import (
	"context"
	"net/http"
	"slices"
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

// unredactedRoute says a request resolved to model of p, or to the group
// whose models are ms, goes only where the user set requests to go
// unmasked (provider.SkipsRedaction, or a group's Redact): p and each of
// its fallbacks, the groups built around the model that it goes on to when
// it can't take the request (provider.GroupsAfter, for a model named), or
// every model of the group (a member's own fallbacks are not the group's,
// see planGroup). One that may go on to anyone else stays masked for all
// of them. model is "" for a route with none to name.
func unredactedRoute(p provider.Provider, model string, isGroup bool, ms []provider.Member) bool {
	if isGroup {
		return unmaskedMembers(ms)
	}
	if !p.SkipsRedaction() {
		return false
	}
	after := afterMembers(p, model)
	return len(after) == 0 || unmaskedMembers(after)
}

// afterMembers are the models a request to model of p may go on to: its
// fallbacks (a routing group among them is each of its members; one that
// resolves to nothing isn't tried, see plan) and the members of the groups
// built around it (provider.GroupsAfter), each as its groups' Redact has
// it.
func afterMembers(p provider.Provider, model string) []provider.Member {
	ms := provider.FallbackModels(p)
	if model == "" {
		return ms
	}
	for _, gid := range provider.GroupsAfter(p.ID + "/" + model) {
		if !slices.Contains(p.Fallback, gid) {
			_, gms, _ := provider.FallbackGroup(gid)
			ms = append(ms, gms...)
		}
	}
	return ms
}

// unmaskedMembers says every model in ms, the ones a picture, a video, a
// group or a vendor's own route may go to, goes as written: its provider
// is one the user set to go unmasked (provider.SkipsRedaction), or its
// group says so (Group.Redact), so the request goes as written, as a
// chat's does on such a route (unredactedRoute). None is masked.
func unmaskedMembers(ms []provider.Member) bool {
	return len(ms) > 0 && !slices.ContainsFunc(ms, func(m provider.Member) bool { return !m.Unmasked() })
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
func maskedFor(ctx context.Context, p provider.Provider, model string, isGroup bool, ms []provider.Member) context.Context {
	return ctx
}

type agentTurnKey struct{}

func agentTurn(r *http.Request) *http.Request                         { return r }
func isAgentTurn(r *http.Request) bool                                { return false }
func MaskRoute(id string) bool                                        { return false }
func maskRouteFor(r *http.Request, hits maskHits) string              { return "" }
func codexMaskRoute(r *http.Request, rest string, body []byte) string { return "" }

// Tiny does not run a request masker or mark requests as masked.
func redactionWantedFor(context.Context) bool              { return false }
func maskedBefore(*http.Request) (string, bool)            { return "", false }
func withMaskedAs(r *http.Request, _ string) *http.Request { return r }
func maskedAsOf(context.Context) string                    { return "" }
func ownMaskedAs(context.Context) string                   { return "" }

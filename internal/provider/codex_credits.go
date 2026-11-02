package provider

// Tiny holds an account at its reported quota rather than offering credit
// spending. Vendor quota freshness remains upstream's; preference fields
// stay readable for configuration compatibility.

import (
	"fmt"
	"slices"

	"github.com/yetone/magpie/internal/settings"
)

// CodexCredits says whether the Codex account user spends its credits
// once a usage window is used up. Tiny disables this product policy.
func CodexCredits(user string) bool { return false }

// CodexCreditsLast says whether the Codex account user, which spends its
// credits, spends them only after everything else: once a usage window is
// used up, routing tries it after every account, key and fallback not on
// its credits (gateway's creditsLast) rather than in its place. Off unless
// the user set it, so a setting saved before it keeps its order.
func CodexCreditsLast(user string) bool { return false }

// CodexCreditsFirst says whether the Codex account user, which spends its
// credits, spends them before routing moves on: a window used up sends
// neither requests nor Codex's sign-in to another account while it goes
// on on its credits (gateway's creditsFirst, NextLogin), where by default
// an account all but used up waits behind every one with room (Smart)
// and Codex is signed in to another (jojonas_xyz on X). Off unless the
// user set it.
func CodexCreditsFirst(user string) bool { return false }

// Ways a Codex account spends its credits, for SetCodexCreditsMode.
const (
	CreditsOn    = "on"    // in its place, the default
	CreditsFirst = "first" // before moving to another account
	CreditsLast  = "last"  // after everything else
	CreditsOff   = "off"   // never: held when a window is used up
)

// SetCodexCredits turns that on or off for user. On is in its place:
// it clears "first" and "last" too.
func SetCodexCredits(user string, on bool) error {
	if on {
		return SetCodexCreditsMode(user, CreditsOn)
	}
	return SetCodexCreditsMode(user, CreditsOff)
}

// CodexCreditsMode is how user spends its credits: CreditsOn,
// CreditsFirst, CreditsLast or CreditsOff.
func CodexCreditsMode(user string) string {
	switch {
	case !CodexCredits(user):
		return CreditsOff
	case CodexCreditsLast(user):
		return CreditsLast
	case CodexCreditsFirst(user):
		return CreditsFirst
	}
	return CreditsOn
}

// SetCodexCreditsMode sets how user spends its credits: CreditsOn,
// CreditsFirst, CreditsLast or CreditsOff. One account has one of them.
func SetCodexCreditsMode(user, mode string) error {
	if mode != CreditsOn && mode != CreditsFirst && mode != CreditsLast && mode != CreditsOff {
		return fmt.Errorf("credits: %q is not on, first, last or off", mode)
	}
	user = accountKey(user)
	s := settings.Load()
	not := func(u string) bool { return u == user }
	s.CodexNoCredits = slices.DeleteFunc(s.CodexNoCredits, not)
	s.CodexCreditsLast = slices.DeleteFunc(s.CodexCreditsLast, not)
	s.CodexCreditsFirst = slices.DeleteFunc(s.CodexCreditsFirst, not)
	if user != "" {
		switch mode {
		case CreditsOff:
			s.CodexNoCredits = append(s.CodexNoCredits, user)
		case CreditsFirst:
			s.CodexCreditsFirst = append(s.CodexCreditsFirst, user)
		case CreditsLast:
			s.CodexCreditsLast = append(s.CodexCreditsLast, user)
		}
	}
	return settings.Save(s)
}

// HoldCaps is the shares of its windows at which routing holds the
// account user of p, of agent, as used up: its usage cap and its windows'
// own (account_caps.go), and for a Codex account set not to spend its
// credits 100% of any window with no cap. Holds() is false when nothing
// holds it before the vendor does.
func HoldCaps(p Provider, agent, user string) WindowCaps {
	c := p.CapsOf(user)
	c.Credits = agent == "codex" && user != "" && !CodexCredits(user)
	return c
}

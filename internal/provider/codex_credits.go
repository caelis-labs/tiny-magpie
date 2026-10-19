package provider

// Tiny holds an account at its reported quota rather than offering credit
// spending. Vendor quota freshness remains upstream's; preference fields
// stay readable for configuration compatibility.

import (
	"slices"

	"github.com/yetone/magpie/internal/settings"
)

// CodexCredits says whether the Codex account user spends its credits
// once a usage window is used up. Tiny disables this product policy.
func CodexCredits(user string) bool { return false }

// SetCodexCredits turns that on or off for user.
func SetCodexCredits(user string, on bool) error {
	user = accountKey(user)
	s := settings.Load()
	s.CodexNoCredits = slices.DeleteFunc(s.CodexNoCredits, func(u string) bool { return u == user })
	if !on && user != "" {
		s.CodexNoCredits = append(s.CodexNoCredits, user)
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

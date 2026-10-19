package provider

import (
	"testing"
	"time"
)

// An account that spends a Codex reset about to run out by itself has its
// windows started again then, resetExpiryLead before the reset runs out:
// Weekly pace goes by the hours until that, when it is sooner than the week
// renews, and Smart's renewals do too (#717, #718). A reset that runs out
// after the week renews, or has run out, changes nothing.
func TestPaceUntilAutoUsedReset(t *testing.T) {
	now := time.Now()
	near := func(got, want float64) bool { return got > want*0.99 && got < want*1.01 }
	week := func(used float64, renews, runsOut time.Duration) Allowance {
		a := Allowance{
			{Used: 10, Resets: now.Add(2 * time.Hour), Span: 5 * time.Hour},
			{Used: used, Resets: now.Add(renews), Span: 7 * 24 * time.Hour},
		}
		if runsOut != 0 {
			a = a.restartedBy(now.Add(runsOut))
		}
		return a
	}
	// the issue's A: 80% left, the week renewing in five days, a reset
	// running out in an hour and a half — spent in one
	a := week(20, 120*time.Hour, 90*time.Minute)
	if p, due := a.Pace("m", now); !near(p, 80) || !due.Equal(now.Add(time.Hour)) {
		t.Fatalf("by the reset spent: %v %v", p, due)
	}
	if r := a.Renewal("m", now); len(r) != 2 || !r[0].Equal(now.Add(time.Hour)) || !r[1].Equal(now.Add(time.Hour)) {
		t.Fatalf("renewals by the reset spent: %v", r)
	}
	// none to spend: by the week, as before
	if p, due := week(20, 120*time.Hour, 0).Pace("m", now); !near(p, 80.0/120) || !due.Equal(now.Add(120*time.Hour)) {
		t.Fatalf("no reset: %v %v", p, due)
	}
	// the week renews first: by the week
	b := week(50, 48*time.Hour, 20*24*time.Hour)
	if p, due := b.Pace("m", now); !near(p, 50.0/48) || !due.Equal(now.Add(48*time.Hour)) {
		t.Fatalf("week sooner: %v %v", p, due)
	}
	if r := b.Renewal("m", now); !r[0].Equal(now.Add(48 * time.Hour)) {
		t.Fatalf("week sooner, renewals: %v", r)
	}
	// within the lead already: spent on the next look
	if p, due := week(20, 120*time.Hour, 20*time.Minute).Pace("m", now); !near(p, 80) || !due.Equal(now) {
		t.Fatalf("within the lead: %v %v", p, due)
	}
	// run out already: nothing to spend
	if p, _ := week(20, 120*time.Hour, -time.Minute).Pace("m", now); !near(p, 80.0/120) {
		t.Fatalf("ran out: %v", p)
	}
	// five hours alone are cut short too: the reset runs out in two and a
	// half hours, spent in two
	five := Allowance{{Used: 40, Resets: now.Add(4 * time.Hour), Span: 5 * time.Hour}}.restartedBy(now.Add(150 * time.Minute))
	if p, due := five.Pace("m", now); !near(p, 60.0/2) || !due.Equal(now.Add(2*time.Hour)) {
		t.Fatalf("five hours alone: %v %v", p, due)
	}
}

// A reset counts only for an account the user lets spend its resets, that
// holds one that runs out, and whose windows have been used — as the
// spending itself (check). The allowances routing reads carry it.
func TestRestartsWhenHeldUp(t *testing.T) {
	now := time.Now()
	acct := func(fiveUsed float64, fiveBack time.Duration) Allowance {
		five := Limit{Used: fiveUsed, Span: 5 * time.Hour}
		if fiveBack != 0 {
			five.Resets = now.Add(fiveBack)
		}
		return Allowance{five, {Used: 32, Resets: now.Add(120 * time.Hour), Span: 7 * 24 * time.Hour}}.restartedBy(now.Add(3 * time.Hour))
	}
	for name, c := range map[string]struct {
		a    Allowance
		want time.Time
	}{
		"free again before it is spent":  {acct(100, 110*time.Minute), now.Add(150 * time.Minute)},
		"held up past it":                {acct(100, 160*time.Minute), now},
		"held up, not saying until when": {acct(100, 0), now},
		"not held up":                    {acct(60, 110*time.Minute), now.Add(150 * time.Minute)},
	} {
		if got := c.a.Restarts(now); !got.Equal(c.want) {
			t.Errorf("%s: restarts %v, want %v", name, got, c.want)
		}
	}
}

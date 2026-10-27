package filememo

import (
	"testing"
	"testing/synctest"
)

func TestForgetAcrossIndependentClocks(t *testing.T) {
	// No subscriber exists in these isolated request clocks. Invalidating
	// parsed data must not retain a notification owned by a finished clock.
	for range 2 {
		synctest.Test(t, func(t *testing.T) { Forget() })
	}
	Forget()
	first := Written()
	Forget()
	select {
	case <-first:
	default:
		t.Fatal("subscriber not notified")
	}
	second := Written()
	select {
	case <-second:
		t.Fatal("new subscriber got an old notification")
	default:
	}
	Forget()
}

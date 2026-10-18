package agent

import (
	"github.com/yetone/magpie/internal/settings"
	"testing"
)

func visionOff(t *testing.T) {
	t.Helper()
	s := settings.Load()
	s.Vision = "off"
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
}

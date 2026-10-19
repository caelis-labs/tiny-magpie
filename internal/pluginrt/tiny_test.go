package pluginrt

import "testing"

func TestTinyCoreCannotBeDisabled(t *testing.T) {
	off := false
	for _, id := range []string{"gateway", "groups", "routing", "usage", "budget", "middleware", "rotation"} {
		c := Compose([]Manifest{{ID: id}}, File{Preset: PresetEssentials, Plugins: []Row{{ID: id, Enabled: &off}}})
		if len(c) != 1 || !c[0].Enabled {
			t.Fatalf("core disabled by old config: %+v", c)
		}
		if err := Set(id, false); err == nil {
			t.Fatalf("disabled %s", id)
		}
	}
}

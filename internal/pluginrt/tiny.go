package pluginrt

// Fixed names Tiny's compiled gateway assembly. The upstream lifecycle and
// config schemas still run; copied presets cannot turn core pages off.
func Fixed(id string) bool {
	switch id {
	case "budget", "gateway", "groups", "middleware", "rotation", "routing", "usage":
		return true
	}
	return false
}

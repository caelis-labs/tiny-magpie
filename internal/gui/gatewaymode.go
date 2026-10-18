package gui

import "sync/atomic"

// WebGateway keeps the upstream web CLI flag compatible. All tiny windows
// use the gateway view, regardless of an imported upstream mode setting.
var WebGateway atomic.Bool
var webPage atomic.Bool

func gatewayMode(web bool) (bool, string) {
	return true, "tiny-magpie"
}

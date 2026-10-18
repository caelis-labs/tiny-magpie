package gui

import (
	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/provider"
)

// Provider edits affect this gateway only; agents keep their own configuration.
func providerChange(change func() error) ([]agent.Move, error) { return nil, change() }
func renameProvider(from, to string) ([]string, error)         { return nil, provider.Rename(from, to) }

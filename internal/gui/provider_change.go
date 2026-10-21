package gui

import "github.com/yetone/magpie/internal/agent"

// Explicit Provider edits keep connected Agents on models the gateway serves.
func providerChange(change func() error) ([]agent.Move, error) { return agent.Reseat(change) }
func renameProvider(from, to string) ([]string, error)         { return agent.RenameProvider(from, to) }

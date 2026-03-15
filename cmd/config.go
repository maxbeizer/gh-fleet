package cmd

import (
	"github.com/maxbeizer/gh-fleet/internal/fleet"
)

// loadConfig finds and loads fleet.toml, searching the hint dir, CWD, and ~/code/gh-fleet.
func loadConfig(hint string) (*fleet.Config, error) {
	dir, err := fleet.FindConfigDir(hint)
	if err != nil {
		return nil, err
	}
	return fleet.LoadConfig(dir)
}

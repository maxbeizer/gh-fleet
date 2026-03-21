package fleet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the fleet.toml configuration.
type Config struct {
	Owner     string          `toml:"owner"`
	Discovery DiscoveryConfig `toml:"discovery"`
	Sync      SyncConfig      `toml:"sync"`
	Catalog   CatalogConfig   `toml:"catalog"`
	Settings  SettingsConfig  `toml:"settings"`
	Dir       string          `toml:"-"` // directory containing fleet.toml
}

type DiscoveryConfig struct {
	Exclude []string `toml:"exclude"`
}

type SyncFile struct {
	Canon        string `toml:"canon"`
	Target       string `toml:"target"`
	Template     bool   `toml:"template"`
	SkipIfExists bool   `toml:"skip_if_exists"`
}

type SyncConfig struct {
	Files        []SyncFile        `toml:"files"`
	TemplateVars map[string]string `toml:"template_vars"`
}

type CatalogConfig struct {
	Output string `toml:"output"`
	Header string `toml:"header"`
}

type SettingsConfig struct {
	HasWiki             *bool `toml:"has_wiki"`
	DeleteBranchOnMerge *bool `toml:"delete_branch_on_merge"`
	AllowSquashMerge    *bool `toml:"allow_squash_merge"`
	AllowMergeCommit    *bool `toml:"allow_merge_commit"`
	AllowRebaseMerge    *bool `toml:"allow_rebase_merge"`
}

// RepoSettings returns the desired settings, using defaults for unset fields.
func (s SettingsConfig) RepoSettings() (hasWiki, deleteBranch, squash, mergeCommit, rebase bool) {
	boolVal := func(p *bool, def bool) bool {
		if p != nil {
			return *p
		}
		return def
	}
	return boolVal(s.HasWiki, false),
		boolVal(s.DeleteBranchOnMerge, true),
		boolVal(s.AllowSquashMerge, true),
		boolVal(s.AllowMergeCommit, false),
		boolVal(s.AllowRebaseMerge, false)
}

// FindConfigDir locates fleet.toml by checking: the given dir, CWD, then
// ~/code/gh-fleet. Returns the directory containing fleet.toml.
func FindConfigDir(hint string) (string, error) {
	candidates := []string{hint}

	if cwd, err := os.Getwd(); err == nil && cwd != hint {
		candidates = append(candidates, cwd)
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "code", "gh-fleet"))
	}

	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "fleet.toml")); err == nil {
			return dir, nil
		}
	}

	return "", fmt.Errorf("fleet.toml not found (checked %v)", candidates)
}

// LoadConfig reads fleet.toml from the given directory (or cwd).
func LoadConfig(dir string) (*Config, error) {
	path := filepath.Join(dir, "fleet.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading fleet.toml: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing fleet.toml: %w", err)
	}

	cfg.Dir = dir
	return &cfg, nil
}

// IsExcluded checks if a repo name is in the exclude list.
func (c *Config) IsExcluded(name string) bool {
	for _, ex := range c.Discovery.Exclude {
		if ex == name {
			return true
		}
	}
	return false
}

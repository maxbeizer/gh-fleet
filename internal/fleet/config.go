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
	Dir       string          `toml:"-"` // directory containing fleet.toml
}

type DiscoveryConfig struct {
	Auto    bool     `toml:"auto"`
	Exclude []string `toml:"exclude"`
}

type SyncFile struct {
	Canon    string `toml:"canon"`
	Target   string `toml:"target"`
	Template bool   `toml:"template"`
}

type SyncConfig struct {
	Files        []SyncFile        `toml:"files"`
	TemplateVars map[string]string `toml:"template_vars"`
}

type CatalogConfig struct {
	Output string `toml:"output"`
	Header string `toml:"header"`
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

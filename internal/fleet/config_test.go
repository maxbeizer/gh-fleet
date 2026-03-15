package fleet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
owner = "testuser"

[discovery]
auto = true
exclude = ["gh-skip-me"]

[[sync.files]]
canon = "canon/ci.yml"
target = ".github/workflows/ci.yml"

[[sync.files]]
canon = "canon/Makefile"
target = "Makefile"
template = true

[sync.template_vars]
GO_MIN_MAJOR = "1"
GO_MIN_MINOR = "24"

[catalog]
output = "README.md"
header = "canon/readme-header.md"
`
	if err := os.WriteFile(filepath.Join(dir, "fleet.toml"), []byte(tomlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Owner != "testuser" {
		t.Errorf("Owner = %q, want %q", cfg.Owner, "testuser")
	}
	if !cfg.Discovery.Auto {
		t.Error("Discovery.Auto = false, want true")
	}
	if len(cfg.Discovery.Exclude) != 1 || cfg.Discovery.Exclude[0] != "gh-skip-me" {
		t.Errorf("Exclude = %v, want [gh-skip-me]", cfg.Discovery.Exclude)
	}
	if len(cfg.Sync.Files) != 2 {
		t.Errorf("Sync.Files length = %d, want 2", len(cfg.Sync.Files))
	}
	if cfg.Sync.Files[1].Template != true {
		t.Error("Sync.Files[1].Template = false, want true")
	}
	if cfg.Sync.TemplateVars["GO_MIN_MAJOR"] != "1" {
		t.Errorf("TemplateVars[GO_MIN_MAJOR] = %q, want %q", cfg.Sync.TemplateVars["GO_MIN_MAJOR"], "1")
	}
}

func TestIsExcluded(t *testing.T) {
	cfg := &Config{
		Discovery: DiscoveryConfig{
			Exclude: []string{"gh-skip", "gh-ignore"},
		},
	}

	tests := []struct {
		name string
		want bool
	}{
		{"gh-skip", true},
		{"gh-ignore", true},
		{"gh-keep", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.IsExcluded(tt.name); got != tt.want {
				t.Errorf("IsExcluded(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

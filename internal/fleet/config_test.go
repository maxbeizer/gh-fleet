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
exclude = ["gh-skip-me"]

[[sync.files]]
canon = "canon/ci.yml"
target = ".github/workflows/ci.yml"

[[sync.files]]
canon = "canon/copilot-instructions.md"
target = ".github/copilot-instructions.md"
skip_if_exists = true

[[sync.files]]
canon = "canon/Makefile"
target = "Makefile"
template = true

[sync.template_vars]
GO_MIN_MAJOR = "1"
GO_MIN_MINOR = "24"
GO_MIN_PATCH = "0"

[catalog]
output = "README.md"
header = "canon/readme-header.md"

[settings]
has_wiki = false
delete_branch_on_merge = true
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
	if len(cfg.Discovery.Exclude) != 1 || cfg.Discovery.Exclude[0] != "gh-skip-me" {
		t.Errorf("Exclude = %v, want [gh-skip-me]", cfg.Discovery.Exclude)
	}
	if len(cfg.Sync.Files) != 3 {
		t.Errorf("Sync.Files length = %d, want 3", len(cfg.Sync.Files))
	}
	if !cfg.Sync.Files[1].SkipIfExists {
		t.Error("Sync.Files[1].SkipIfExists = false, want true")
	}
	if !cfg.Sync.Files[2].Template {
		t.Error("Sync.Files[2].Template = false, want true")
	}
	if cfg.Sync.TemplateVars["GO_MIN_MAJOR"] != "1" {
		t.Errorf("TemplateVars[GO_MIN_MAJOR] = %q, want %q", cfg.Sync.TemplateVars["GO_MIN_MAJOR"], "1")
	}
	if cfg.Sync.TemplateVars["GO_MIN_MINOR"] != "24" {
		t.Errorf("TemplateVars[GO_MIN_MINOR] = %q, want %q", cfg.Sync.TemplateVars["GO_MIN_MINOR"], "24")
	}
	if cfg.Sync.TemplateVars["GO_MIN_PATCH"] != "0" {
		t.Errorf("TemplateVars[GO_MIN_PATCH] = %q, want %q", cfg.Sync.TemplateVars["GO_MIN_PATCH"], "0")
	}
	if cfg.Settings.HasWiki == nil || *cfg.Settings.HasWiki != false {
		t.Errorf("Settings.HasWiki = %v, want false", cfg.Settings.HasWiki)
	}
	if cfg.Settings.DeleteBranchOnMerge == nil || *cfg.Settings.DeleteBranchOnMerge != true {
		t.Errorf("Settings.DeleteBranchOnMerge = %v, want true", cfg.Settings.DeleteBranchOnMerge)
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

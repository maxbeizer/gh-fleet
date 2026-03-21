package cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-fleet/internal/fleet"
	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

var (
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#d29922"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#3fb950"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f85149"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	boldStyle  = lipgloss.NewStyle().Bold(true)
)

func runDrift(args []string) error {
	fs := flag.NewFlagSet("drift", flag.ContinueOnError)
	configDir := fs.String("config", ".", "directory containing fleet.toml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(*configDir)
	if err != nil {
		return err
	}

	repos, err := discoverRepos(cfg)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Checking drift across %d repos...\n\n", len(repos))

	// Check Go versions
	gh.FetchGoVersions(cfg.Owner, repos)

	goVersions := map[string][]string{}
	for _, r := range repos {
		v := r.GoVersion
		if v == "" {
			v = "(no go.mod)"
		}
		goVersions[v] = append(goVersions[v], r.Name)
	}

	fmt.Println(boldStyle.Render("Go Versions"))
	versions := sortedKeys(goVersions)
	for _, v := range versions {
		names := goVersions[v]
		sort.Strings(names)
		style := okStyle
		if v != versions[len(versions)-1] && v != "(no go.mod)" {
			style = warnStyle
		}
		if v == "(no go.mod)" {
			style = dimStyle
		}
		fmt.Printf("  %s %s\n", style.Render(v), dimStyle.Render(fmt.Sprintf("(%d)", len(names))))
		for _, n := range names {
			fmt.Printf("    %s\n", n)
		}
	}
	fmt.Println()

	// Check for missing synced files
	fmt.Println(boldStyle.Render("Synced Files"))
	for _, sf := range cfg.Sync.Files {
		missing := []string{}
		for _, r := range repos {
			if !gh.FileExists(cfg.Owner, r.Name, sf.Target) {
				missing = append(missing, r.Name)
			}
		}
		if len(missing) == 0 {
			fmt.Printf("  %s %s\n", okStyle.Render("✅"), sf.Target)
		} else {
			fmt.Printf("  %s %s %s\n", errStyle.Render("❌"), sf.Target,
				warnStyle.Render(fmt.Sprintf("missing in %d repos", len(missing))))
			for _, m := range missing {
				fmt.Printf("    %s\n", m)
			}
		}
	}

	return nil
}

func discoverRepos(cfg *fleet.Config) ([]gh.Repo, error) {
	fmt.Fprintf(os.Stderr, "Discovering extensions for %s...\n", cfg.Owner)
	repos, err := gh.ListGHRepos(cfg.Owner)
	if err != nil {
		return nil, err
	}

	var filtered []gh.Repo
	for _, r := range repos {
		if cfg.IsExcluded(r.Name) || r.IsArchived || r.IsFork || r.IsPrivate {
			continue
		}
		filtered = append(filtered, r)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})
	return filtered, nil
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		// Put "(no go.mod)" last
		if keys[i] == "(no go.mod)" {
			return false
		}
		if keys[j] == "(no go.mod)" {
			return true
		}
		return compareVersions(keys[i], keys[j]) < 0
	})
	return keys
}

func compareVersions(a, b string) int {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	for i := 0; i < len(pa) && i < len(pb); i++ {
		na, _ := strconv.Atoi(pa[i])
		nb, _ := strconv.Atoi(pb[i])
		if na < nb {
			return -1
		}
		if na > nb {
			return 1
		}
	}
	return len(pa) - len(pb)
}

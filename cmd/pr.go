package cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

func runPR(args []string) error {
	fs := flag.NewFlagSet("pr", flag.ContinueOnError)
	configDir := fs.String("config", ".", "directory containing fleet.toml")
	fileFilter := fs.String("file", "", "only show/act on PRs for this synced file")
	merge := fs.Bool("merge", false, "squash-merge all listed PRs")
	admin := fs.Bool("admin", false, "use --admin to bypass branch protection (with --merge)")
	close := fs.Bool("close", false, "close all listed PRs with a comment")
	dryRun := fs.Bool("dry-run", false, "preview merge/close actions without executing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *merge && *close {
		return fmt.Errorf("cannot use --merge and --close together")
	}

	cfg, err := loadConfig(*configDir)
	if err != nil {
		return err
	}

	repos, err := discoverRepos(cfg)
	if err != nil {
		return err
	}

	repoNames := make([]string, len(repos))
	for i, r := range repos {
		repoNames[i] = r.Name
	}

	fmt.Fprintf(os.Stderr, "Fetching fleet PRs across %d repos...\n", len(repoNames))
	allPRs := gh.FetchFleetPRs(cfg.Owner, repoNames)

	if len(allPRs) == 0 {
		fmt.Println(okStyle.Render("No open fleet PRs found."))
		return nil
	}

	// Group PRs by synced file (derived from branch name "fleet/sync-<file>")
	grouped := groupPRsByFile(allPRs)

	// Apply file filter
	if *fileFilter != "" {
		filtered := make(map[string][]gh.FleetPR)
		for file, prs := range grouped {
			if strings.EqualFold(file, *fileFilter) {
				filtered[file] = prs
			}
		}
		grouped = filtered
		if len(grouped) == 0 {
			fmt.Printf("No fleet PRs found for file %s\n", dimStyle.Render(*fileFilter))
			return nil
		}
	}

	// Display grouped PRs
	files := sortedFileKeys(grouped)
	total := 0
	for _, file := range files {
		prs := grouped[file]
		total += len(prs)
		fmt.Printf("\n%s %s\n", boldStyle.Render("●"), boldStyle.Render(file))
		for _, pr := range prs {
			fmt.Printf("  %s  %s  %s\n",
				warnStyle.Render(fmt.Sprintf("#%d", pr.Number)),
				pr.Repo,
				dimStyle.Render(pr.URL),
			)
		}
	}
	fmt.Printf("\n%s\n", dimStyle.Render(fmt.Sprintf("%d open fleet PR(s) across %d file(s)", total, len(files))))

	// Handle --merge
	if *merge {
		fmt.Println()
		for _, file := range files {
			for _, pr := range grouped[file] {
				label := fmt.Sprintf("%s/%s#%d", cfg.Owner, pr.Repo, pr.Number)
				if *dryRun {
					fmt.Printf("  %s %s\n", dimStyle.Render("[dry-run] would merge"), label)
					continue
				}
				fmt.Fprintf(os.Stderr, "  Merging %s...\n", label)
				if err := gh.MergePR(cfg.Owner, pr.Repo, pr.Number, *admin); err != nil {
					fmt.Printf("  %s %s: %v\n", errStyle.Render("✗"), label, err)
				} else {
					fmt.Printf("  %s %s\n", okStyle.Render("✓ merged"), label)
				}
			}
		}
	}

	// Handle --close
	if *close {
		comment := "Closed by gh-fleet pr --close"
		fmt.Println()
		for _, file := range files {
			for _, pr := range grouped[file] {
				label := fmt.Sprintf("%s/%s#%d", cfg.Owner, pr.Repo, pr.Number)
				if *dryRun {
					fmt.Printf("  %s %s\n", dimStyle.Render("[dry-run] would close"), label)
					continue
				}
				fmt.Fprintf(os.Stderr, "  Closing %s...\n", label)
				if err := gh.ClosePR(cfg.Owner, pr.Repo, pr.Number, comment); err != nil {
					fmt.Printf("  %s %s: %v\n", errStyle.Render("✗"), label, err)
				} else {
					fmt.Printf("  %s %s\n", okStyle.Render("✓ closed"), label)
				}
			}
		}
	}

	return nil
}

// groupPRsByFile groups PRs by synced file name derived from the branch name.
// Branch "fleet/sync-makefile" → file "Makefile" (best-effort reconstruction).
func groupPRsByFile(prs []gh.FleetPR) map[string][]gh.FleetPR {
	grouped := make(map[string][]gh.FleetPR)
	for _, pr := range prs {
		file := fileFromBranch(pr.Branch)
		grouped[file] = append(grouped[file], pr)
	}
	// Sort PRs within each group by repo name
	for file := range grouped {
		sort.Slice(grouped[file], func(i, j int) bool {
			return grouped[file][i].Repo < grouped[file][j].Repo
		})
	}
	return grouped
}

// fileFromBranch extracts the synced file name from a branch like "fleet/sync-makefile".
func fileFromBranch(branch string) string {
	name := strings.TrimPrefix(branch, "fleet/sync-")
	// The sync command replaces dots with hyphens, but we can't perfectly
	// reverse that. Return as-is for grouping.
	return name
}

func sortedFileKeys(m map[string][]gh.FleetPR) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

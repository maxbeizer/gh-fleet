package cmd

import (
	"flag"
	"fmt"
	"os"
	"sync"

	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

// desiredSettings defines the enforced repo settings for the fleet.
var desiredSettings = gh.RepoSettings{
	HasWiki:             false,
	DeleteBranchOnMerge: true,
	AllowSquashMerge:    true,
	AllowMergeCommit:    false,
	AllowRebaseMerge:    false,
}

func runSettings(args []string) error {
	fs := flag.NewFlagSet("settings", flag.ContinueOnError)
	configDir := fs.String("config", ".", "directory containing fleet.toml")
	dryRun := fs.Bool("dry-run", false, "preview changes without applying")
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

	fmt.Println(boldStyle.Render("Repo Settings"))
	fmt.Printf("  wiki: %s  delete-head-branch: %s  squash-only: %s\n\n",
		errStyle.Render("off"),
		okStyle.Render("on"),
		okStyle.Render("on"),
	)

	// Fetch current settings concurrently
	type repoResult struct {
		repo     gh.Repo
		current  *gh.RepoCurrentSettings
		err      error
		needsfix bool
	}
	results := make([]repoResult, len(repos))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i, r := range repos {
		results[i].repo = r
		wg.Add(1)
		go func(idx int, repo gh.Repo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s, err := gh.GetRepoSettings(cfg.Owner, repo.Name)
			results[idx].current = s
			results[idx].err = err
		}(i, r)
	}
	wg.Wait()

	// Check and apply
	totalChanged := 0
	totalOK := 0
	totalErrors := 0

	for i := range results {
		r := &results[i]
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s %s\n", errStyle.Render("❌"), r.repo.Name, dimStyle.Render(r.err.Error()))
			totalErrors++
			continue
		}

		r.needsfix = needsSettingsUpdate(r.current)
		if !r.needsfix {
			fmt.Printf("  %s %s\n", okStyle.Render("✅"), r.repo.Name)
			totalOK++
			continue
		}

		diffs := settingsDiff(r.current)
		fmt.Printf("  %s %s %s\n", warnStyle.Render("⇄"), r.repo.Name, dimStyle.Render(diffs))
		totalChanged++

		if !*dryRun {
			if err := gh.UpdateRepoSettings(cfg.Owner, r.repo.Name, desiredSettings); err != nil {
				fmt.Fprintf(os.Stderr, "    %s %v\n", errStyle.Render("❌"), err)
			} else {
				fmt.Printf("    %s updated\n", okStyle.Render("✅"))
			}
		}
	}

	fmt.Println()
	if *dryRun && totalChanged > 0 {
		fmt.Printf("%s repos need updates. Run without --dry-run to apply.\n",
			warnStyle.Render(fmt.Sprintf("%d", totalChanged)))
	} else if totalChanged == 0 {
		fmt.Printf("%s All %d repos have correct settings!\n", okStyle.Render("✅"), totalOK)
	}
	if totalErrors > 0 {
		fmt.Fprintf(os.Stderr, "%s %d repos had errors\n", errStyle.Render("⚠️"), totalErrors)
	}

	return nil
}

func needsSettingsUpdate(s *gh.RepoCurrentSettings) bool {
	return s.HasWiki != desiredSettings.HasWiki ||
		s.DeleteBranchOnMerge != desiredSettings.DeleteBranchOnMerge ||
		s.AllowSquashMerge != desiredSettings.AllowSquashMerge ||
		s.AllowMergeCommit != desiredSettings.AllowMergeCommit ||
		s.AllowRebaseMerge != desiredSettings.AllowRebaseMerge
}

func settingsDiff(s *gh.RepoCurrentSettings) string {
	var diffs []string
	if s.HasWiki != desiredSettings.HasWiki {
		diffs = append(diffs, "wiki")
	}
	if s.DeleteBranchOnMerge != desiredSettings.DeleteBranchOnMerge {
		diffs = append(diffs, "delete-branch")
	}
	if s.AllowMergeCommit != desiredSettings.AllowMergeCommit {
		diffs = append(diffs, "merge-commit")
	}
	if s.AllowRebaseMerge != desiredSettings.AllowRebaseMerge {
		diffs = append(diffs, "rebase-merge")
	}
	if s.AllowSquashMerge != desiredSettings.AllowSquashMerge {
		diffs = append(diffs, "squash-merge")
	}

	result := ""
	for i, d := range diffs {
		if i > 0 {
			result += ", "
		}
		result += d
	}
	return "(" + result + ")"
}

package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

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

	hasWiki, deleteBranch, squash, mergeCommit, rebase := cfg.Settings.RepoSettings()
	desiredSettings := gh.RepoSettings{
		HasWiki:             hasWiki,
		DeleteBranchOnMerge: deleteBranch,
		AllowSquashMerge:    squash,
		AllowMergeCommit:    mergeCommit,
		AllowRebaseMerge:    rebase,
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

		r.needsfix = needsSettingsUpdate(r.current, desiredSettings)
		if !r.needsfix {
			fmt.Printf("  %s %s\n", okStyle.Render("✅"), r.repo.Name)
			totalOK++
			continue
		}

		diffs := settingsDiff(r.current, desiredSettings)
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

func needsSettingsUpdate(s *gh.RepoCurrentSettings, desired gh.RepoSettings) bool {
	return s.HasWiki != desired.HasWiki ||
		s.DeleteBranchOnMerge != desired.DeleteBranchOnMerge ||
		s.AllowSquashMerge != desired.AllowSquashMerge ||
		s.AllowMergeCommit != desired.AllowMergeCommit ||
		s.AllowRebaseMerge != desired.AllowRebaseMerge
}

func settingsDiff(s *gh.RepoCurrentSettings, desired gh.RepoSettings) string {
	var diffs []string
	if s.HasWiki != desired.HasWiki {
		diffs = append(diffs, "wiki")
	}
	if s.DeleteBranchOnMerge != desired.DeleteBranchOnMerge {
		diffs = append(diffs, "delete-branch")
	}
	if s.AllowMergeCommit != desired.AllowMergeCommit {
		diffs = append(diffs, "merge-commit")
	}
	if s.AllowRebaseMerge != desired.AllowRebaseMerge {
		diffs = append(diffs, "rebase-merge")
	}
	if s.AllowSquashMerge != desired.AllowSquashMerge {
		diffs = append(diffs, "squash-merge")
	}
	return "(" + strings.Join(diffs, ", ") + ")"
}

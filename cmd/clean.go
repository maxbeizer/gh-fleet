package cmd

import (
	"flag"
	"fmt"
	"os"
	"sync"

	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

func runClean(args []string) error {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	configDir := fs.String("config", ".", "directory containing fleet.toml")
	dryRun := fs.Bool("dry-run", false, "preview deletions without applying")
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

	fmt.Println(boldStyle.Render("Fleet Clean — stale sync branches"))
	if *dryRun {
		fmt.Println(dimStyle.Render("  (dry-run mode)"))
	}
	fmt.Println()

	type branchResult struct {
		repo     string
		branch   string
		hasOpenPR bool
		deleted  bool
		err      error
	}

	// Collect branches concurrently
	type repoBranches struct {
		repo     string
		branches []string
		err      error
	}
	rb := make([]repoBranches, len(repos))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i, r := range repos {
		rb[i].repo = r.Name
		wg.Add(1)
		go func(idx int, repo gh.Repo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			branches, err := gh.ListSyncBranches(cfg.Owner, repo.Name)
			rb[idx].branches = branches
			rb[idx].err = err
		}(i, r)
	}
	wg.Wait()

	// Process each branch
	var results []branchResult
	for _, r := range rb {
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s %s\n",
				errStyle.Render("❌"), r.repo, dimStyle.Render(r.err.Error()))
			continue
		}
		for _, b := range r.branches {
			results = append(results, branchResult{repo: r.repo, branch: b})
		}
	}

	if len(results) == 0 {
		fmt.Printf("  %s No stale sync branches found.\n", okStyle.Render("✅"))
		return nil
	}

	// Check open PRs and delete concurrently
	var wg2 sync.WaitGroup
	sem2 := make(chan struct{}, 10)

	for i := range results {
		wg2.Add(1)
		go func(idx int) {
			defer wg2.Done()
			sem2 <- struct{}{}
			defer func() { <-sem2 }()

			r := &results[idx]
			hasPR, err := gh.BranchHasOpenPR(cfg.Owner, r.repo, r.branch)
			if err != nil {
				r.err = err
				return
			}
			r.hasOpenPR = hasPR

			if !hasPR && !*dryRun {
				gh.DeleteBranch(cfg.Owner, r.repo, r.branch)
				r.deleted = true
			}
		}(i)
	}
	wg2.Wait()

	// Print results
	totalDeleted := 0
	totalSkipped := 0
	totalErrors := 0

	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s %s %s\n",
				errStyle.Render("❌"), r.repo, r.branch, dimStyle.Render(r.err.Error()))
			totalErrors++
			continue
		}

		if r.hasOpenPR {
			fmt.Printf("  %s %s %s %s\n",
				warnStyle.Render("⏭"), r.repo, r.branch, dimStyle.Render("(has open PR)"))
			totalSkipped++
			continue
		}

		if *dryRun {
			fmt.Printf("  %s %s %s %s\n",
				warnStyle.Render("🗑"), r.repo, r.branch, dimStyle.Render("(would delete)"))
			totalDeleted++
		} else {
			fmt.Printf("  %s %s %s %s\n",
				okStyle.Render("🗑"), r.repo, r.branch, dimStyle.Render("(deleted)"))
			totalDeleted++
		}
	}

	fmt.Println()
	if *dryRun && totalDeleted > 0 {
		fmt.Printf("%s stale branches to delete. Run without --dry-run to apply.\n",
			warnStyle.Render(fmt.Sprintf("%d", totalDeleted)))
	} else if !*dryRun && totalDeleted > 0 {
		fmt.Printf("%s Deleted %d stale branches.\n",
			okStyle.Render("✅"), totalDeleted)
	}
	if totalSkipped > 0 {
		fmt.Printf("%s Skipped %d branches with open PRs.\n",
			warnStyle.Render("⏭"), totalSkipped)
	}
	if totalErrors > 0 {
		fmt.Fprintf(os.Stderr, "%s %d branches had errors.\n",
			errStyle.Render("⚠️"), totalErrors)
	}
	if totalDeleted == 0 && totalSkipped == 0 && totalErrors == 0 {
		fmt.Printf("  %s No stale sync branches found.\n", okStyle.Render("✅"))
	}

	return nil
}

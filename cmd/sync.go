package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxbeizer/gh-fleet/internal/fleet"
	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	configDir := fs.String("config", ".", "directory containing fleet.toml")
	fileFilter := fs.String("file", "", "sync only this file (basename of canon path)")
	dryRun := fs.Bool("dry-run", false, "preview changes without creating PRs")
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

	// Filter sync files if --file is specified
	syncFiles := cfg.Sync.Files
	if *fileFilter != "" {
		var matched []fleet.SyncFile
		for _, sf := range syncFiles {
			base := filepath.Base(sf.Canon)
			if base == *fileFilter || sf.Target == *fileFilter {
				matched = append(matched, sf)
			}
		}
		if len(matched) == 0 {
			return fmt.Errorf("no sync file matching %q", *fileFilter)
		}
		syncFiles = matched
	}

	totalChanges := 0

	for _, sf := range syncFiles {
		canonPath := filepath.Join(cfg.Dir, sf.Canon)
		canonContent, err := os.ReadFile(canonPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Skipping %s: %v\n", sf.Canon, err)
			continue
		}

		fmt.Printf("\n%s → %s\n", boldStyle.Render(sf.Canon), sf.Target)

		for _, r := range repos {
			content := string(canonContent)

			// Apply template variables
			if sf.Template {
				extName := strings.TrimPrefix(r.Name, "gh-")
				content = strings.ReplaceAll(content, "extension-template", extName)
				for k, v := range cfg.Sync.TemplateVars {
					content = strings.ReplaceAll(content, fmt.Sprintf("${%s}", k), v)
				}
			}

			// Fetch current content
			remote, err := gh.FetchFileContent(cfg.Owner, r.Name, sf.Target)
			if err != nil {
				// File doesn't exist — needs sync
				fmt.Printf("  %s %s %s\n", errStyle.Render("➕"), r.Name, dimStyle.Render("(missing)"))
				totalChanges++

				if !*dryRun {
					if err := syncFile(cfg.Owner, r.Name, sf.Target, content); err != nil {
						fmt.Fprintf(os.Stderr, "    ❌ %v\n", err)
					} else {
						fmt.Printf("    %s PR created\n", okStyle.Render("✅"))
					}
				}
				continue
			}

			// Compare
			if strings.TrimSpace(remote) != strings.TrimSpace(content) {
				fmt.Printf("  %s %s %s\n", warnStyle.Render("⇄"), r.Name, dimStyle.Render("(differs)"))
				totalChanges++

				if !*dryRun {
					if err := syncFile(cfg.Owner, r.Name, sf.Target, content); err != nil {
						fmt.Fprintf(os.Stderr, "    ❌ %v\n", err)
					} else {
						fmt.Printf("    %s PR created\n", okStyle.Render("✅"))
					}
				}
			} else {
				fmt.Printf("  %s %s\n", okStyle.Render("✅"), r.Name)
			}
		}
	}

	if *dryRun && totalChanges > 0 {
		fmt.Printf("\n%s changes detected. Run without --dry-run to create PRs.\n",
			warnStyle.Render(fmt.Sprintf("%d", totalChanges)))
	} else if totalChanges == 0 {
		fmt.Printf("\n%s Everything in sync!\n", okStyle.Render("✅"))
	}

	return nil
}

func syncFile(owner, repo, targetPath, content string) error {
	branch := fmt.Sprintf("fleet/sync-%s", strings.ReplaceAll(filepath.Base(targetPath), ".", "-"))
	commitMsg := fmt.Sprintf("chore: sync %s via gh-fleet", targetPath)
	prTitle := fmt.Sprintf("chore: sync %s", filepath.Base(targetPath))
	prBody := fmt.Sprintf("Automated sync from [gh-fleet](https://github.com/%s/gh-fleet) canonical files.\n\nFile: `%s`",
		owner, targetPath)

	return gh.CreateBranchAndPR(owner, repo, branch, targetPath, content, commitMsg, prTitle, prBody)
}

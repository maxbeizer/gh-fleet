package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/maxbeizer/gh-fleet/internal/fleet"
	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

type syncResult struct {
	lines   []string
	changes int
}

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

		results := make([]syncResult, len(repos))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5)

		for i, r := range repos {
			wg.Add(1)
			go func(idx int, r gh.Repo, sf fleet.SyncFile) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				res := &results[idx]
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
					res.lines = append(res.lines, fmt.Sprintf("  %s %s %s", errStyle.Render("➕"), r.Name, dimStyle.Render("(missing)")))
					res.changes++

					if sf.SkipIfExists && *dryRun {
						preview := scaffoldCopilotInstructions(r)
						previewLines := strings.SplitN(preview, "\n", 6)
						if len(previewLines) > 5 {
							previewLines = previewLines[:5]
						}
						for _, pl := range previewLines {
							res.lines = append(res.lines, fmt.Sprintf("    %s", dimStyle.Render(pl)))
						}
					}

					if !*dryRun {
						fileContent := content
						if sf.SkipIfExists {
							fileContent = scaffoldCopilotInstructions(r)
						}
						if err := syncFile(cfg.Owner, r.Name, sf.Target, fileContent); err != nil {
							res.lines = append(res.lines, fmt.Sprintf("    ❌ %v", err))
						} else {
							res.lines = append(res.lines, fmt.Sprintf("    %s PR created", okStyle.Render("✅")))
						}
					}
					return
				}

				// Compare
				if strings.TrimSpace(remote) != strings.TrimSpace(content) {
					if sf.SkipIfExists {
						res.lines = append(res.lines, fmt.Sprintf("  %s %s %s", dimStyle.Render("⊘"), r.Name, dimStyle.Render("(exists, skipped)")))
						return
					}

					res.lines = append(res.lines, fmt.Sprintf("  %s %s %s", warnStyle.Render("⇄"), r.Name, dimStyle.Render("(differs)")))
					res.changes++

					if !*dryRun {
						if err := syncFile(cfg.Owner, r.Name, sf.Target, content); err != nil {
							res.lines = append(res.lines, fmt.Sprintf("    ❌ %v", err))
						} else {
							res.lines = append(res.lines, fmt.Sprintf("    %s PR created", okStyle.Render("✅")))
						}
					}
				} else {
					res.lines = append(res.lines, fmt.Sprintf("  %s %s", okStyle.Render("✅"), r.Name))
				}
			}(i, r, sf)
		}
		wg.Wait()

		// Print results in order and tally changes
		for _, res := range results {
			for _, line := range res.lines {
				fmt.Println(line)
			}
			totalChanges += res.changes
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

func scaffoldCopilotInstructions(r gh.Repo) string {
	extName := strings.TrimPrefix(r.Name, "gh-")

	var b strings.Builder
	fmt.Fprintf(&b, "# Copilot Instructions for %s\n\n", r.Name)

	if r.Description != "" {
		fmt.Fprintf(&b, "## Project Overview\n%s\n\n", r.Description)
	} else {
		fmt.Fprintf(&b, "## Project Overview\n`%s` is a GitHub CLI extension.\n\n", extName)
	}

	b.WriteString("## Technology Stack\n")
	if r.PrimaryLanguage != "" {
		fmt.Fprintf(&b, "- **Language**: %s\n", r.PrimaryLanguage)
	}
	b.WriteString("- **Framework**: GitHub CLI (`gh`) extension\n")
	if r.GoVersion != "" {
		fmt.Fprintf(&b, "- **Go version**: %s+\n", r.GoVersion)
	}
	b.WriteString("\n")

	b.WriteString("## Development\n")
	fmt.Fprintf(&b, "- Build: `go build -o %s` or use `make`\n", r.Name)
	fmt.Fprintf(&b, "- Run: `gh %s`\n", extName)
	b.WriteString("- Test: `go test ./...`\n")

	return b.String()
}

# Copilot Instructions for gh-fleet

## Project Overview
gh-fleet is a command center for managing all of maxbeizer's `gh` CLI extensions. It auto-discovers `gh-*` repos, detects configuration drift, syncs canonical files, and generates a living catalog README.

## Architecture
- `fleet.toml` — config defining which repos to manage and which files to sync
- `canon/` — golden-standard versions of files to sync across all extension repos
- `cmd/` — CLI command implementations (catalog, drift, sync, status)
- `internal/fleet/` — config loading and repo discovery
- `internal/github/` — gh CLI API wrappers

## Key Patterns
- All GitHub API calls go through `gh` CLI (`gh api`, `gh repo list`, `gh pr create`)
- Concurrent operations use goroutines with semaphore for rate limiting
- Sync creates PRs rather than pushing directly to repos
- Templates support variable substitution (e.g., EXTENSION_NAME derived from repo name)

## Technology Stack
- **Language**: Go 1.24+
- **Config**: TOML via `github.com/BurntSushi/toml`
- **Styling**: `github.com/charmbracelet/lipgloss`
- **CLI**: Standard library `flag` package
- **GitHub**: `gh` CLI for all API operations (no direct HTTP)

## Testing
- Table-driven tests following Go conventions
- Config loading has unit tests
- Commands rely on `gh` CLI so integration tests need GitHub access

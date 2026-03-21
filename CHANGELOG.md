# Changelog

All notable changes to gh-fleet are documented here.

## [Unreleased]

### Added
- `skip_if_exists` flag for sync files — skips repos that already have their own version of a file
- Scaffold mode for copilot-instructions — generates repo-specific instructions from metadata (name, description, language) instead of copying the generic template
- `[settings]` section in fleet.toml — repo settings (wiki, merge methods, branch deletion) are now configurable instead of hardcoded
- `DeleteBranch` helper for cleaning up stale sync branches
- Test coverage for `skip_if_exists`, `GO_MIN_PATCH` template var, and settings config

### Fixed
- **Sync is now idempotent** — existing `fleet/sync-*` branches are deleted before re-creation, preventing "branch already exists" failures on repeated runs
- Goreleaser canonical file no longer hardcodes `gh-extension-template` as project name; marked as `template = true` so binary name is substituted per-repo
- Goreleaser-action upgraded from v5 to v6 in canonical release.yml
- Makefile `GO_MIN_*` values now use template variable placeholders (`${GO_MIN_MAJOR}`, etc.) instead of hardcoded defaults
- `compareVersions` uses numeric comparison so `1.3 < 1.24` sorts correctly
- `UpdateRepoSettings` captures API error output instead of suppressing with `--silent`
- Removed dead stdin-piping code path in `CreateBranchAndPR` that always fell through
- Removed redundant `-q`/`--jq` flag duplication on file SHA lookup

### Changed
- Repo list limit increased from 200 to 1000
- Removed unused `discovery.auto` config field
- `settingsDiff` simplified with `strings.Join`

## [0.1.0] — 2026-03-18

### Added
- Initial release
- `catalog` command — auto-generates README with extension table
- `drift` command — detects Go version and file drift across repos
- `sync` command — pushes canonical files to out-of-sync repos via PRs
- `status` command — health matrix with language, Go version, last push
- `settings` command — enforces repo settings (wiki, merge methods) across the fleet
- Auto-discovery of `gh-*` repos with exclude list
- Template variable substitution for canonical files
- Concurrent GitHub API calls with semaphore rate limiting

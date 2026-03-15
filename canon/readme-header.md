# gh-fleet

Command center for [maxbeizer](https://github.com/maxbeizer)'s `gh` CLI extensions.

## Install

```bash
gh extension install maxbeizer/gh-fleet
```

## Usage

```bash
gh fleet              # show help
gh fleet catalog      # regenerate this README with extension catalog
gh fleet drift        # detect configuration drift across repos
gh fleet sync         # push canonical files to out-of-sync repos
gh fleet status       # quick health matrix across all extension repos
```

## How it works

`gh-fleet` uses a `fleet.toml` config to auto-discover all `gh-*` repos and a `canon/` directory of golden-standard files. It can detect drift (Go versions, missing CI, missing goreleaser) and sync canonical files across all repos via PRs.


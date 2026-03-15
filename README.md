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


## Extensions

| Extension | Description | Language | Go | Stars | Install |
|-----------|-------------|----------|----|-------|---------|
| [gh-agent-viz](https://github.com/maxbeizer/gh-agent-viz) | a gh CLI extension to see what your agents are up to | Go | 1.24.2 | ⭐1 | `gh ext install maxbeizer/gh-agent-viz` |
| [gh-art](https://github.com/maxbeizer/gh-art) | Famous art in your terminal | Go | 1.21 | ⭐1 | `gh ext install maxbeizer/gh-art` |
| [gh-atc](https://github.com/maxbeizer/gh-atc) | Private Copilot Air Traffic Control analytics platform (u... | Go | 1.24.13 | ⭐0 | `gh ext install maxbeizer/gh-atc` |
| [gh-branch-breaker](https://github.com/maxbeizer/gh-branch-breaker) |  | Go | 1.22 | ⭐0 | `gh ext install maxbeizer/gh-branch-breaker` |
| [gh-contrib](https://github.com/maxbeizer/gh-contrib) | A CLI extension for understanding your contributions on G... | Go | 1.23.0 | ⭐3 | `gh ext install maxbeizer/gh-contrib` |
| [gh-dotcom-train-approval](https://github.com/maxbeizer/gh-dotcom-train-approval) |  | Shell | — | ⭐0 | `gh ext install maxbeizer/gh-dotcom-train-approval` |
| [gh-dotfiles](https://github.com/maxbeizer/gh-dotfiles) | Back up your gh CLI config to your dotfiles repo | Go | 1.22.5 | ⭐0 | `gh ext install maxbeizer/gh-dotfiles` |
| [gh-extension-precompile](https://github.com/maxbeizer/gh-extension-precompile) | Action for publishing binary GitHub CLI extensions | Shell | — | ⭐0 | `gh ext install maxbeizer/gh-extension-precompile` |
| [gh-extension-template](https://github.com/maxbeizer/gh-extension-template) | Template repo for Go-based gh CLI extensions | Makefile | 1.24.13 | ⭐0 | `gh ext install maxbeizer/gh-extension-template` |
| [gh-exts](https://github.com/maxbeizer/gh-exts) | Your extensions, extended | Go | 1.24.2 | ⭐0 | `gh ext install maxbeizer/gh-exts` |
| [gh-fleet](https://github.com/maxbeizer/gh-fleet) | Command center for managing all your gh CLI extensions | — | — | ⭐0 | `gh ext install maxbeizer/gh-fleet` |
| [gh-ghostty](https://github.com/maxbeizer/gh-ghostty) | Change your ghostty theme via the gh cli | Go | 1.22 | ⭐0 | `gh ext install maxbeizer/gh-ghostty` |
| [gh-habitat](https://github.com/maxbeizer/gh-habitat) | Tamagotchi for your Copilot CLI agents — watch them hat... | Go | 1.22 | ⭐0 | `gh ext install maxbeizer/gh-habitat` |
| [gh-hearth](https://github.com/maxbeizer/gh-hearth) | 🔥 A roaring fire for your terminal — a GitHub CLI ex... | Go | 1.24.0 | ⭐0 | `gh ext install maxbeizer/gh-hearth` |
| [gh-helm](https://github.com/maxbeizer/gh-helm) |  | Go | 1.22 | ⭐0 | `gh ext install maxbeizer/gh-helm` |
| [gh-memex](https://github.com/maxbeizer/gh-memex) |  | Go | 1.16 | ⭐0 | `gh ext install maxbeizer/gh-memex` |
| [gh-not](https://github.com/maxbeizer/gh-not) | GitHub rule-based notifications management | Go | 1.25.0 | ⭐0 | `gh ext install maxbeizer/gh-not` |
| [gh-onion](https://github.com/maxbeizer/gh-onion) | America's finest news source in your terminal | Go | 1.24.13 | ⭐0 | `gh ext install maxbeizer/gh-onion` |
| [gh-oss](https://github.com/maxbeizer/gh-oss) | Find open source issues to work on — right from your te... | Go | 1.22 | ⭐2 | `gh ext install maxbeizer/gh-oss` |
| [gh-pagerduty](https://github.com/maxbeizer/gh-pagerduty) | GitHub CLI extension for PagerDuty — check on-call stat... | Go | 1.24.13 | ⭐2 | `gh ext install maxbeizer/gh-pagerduty` |
| [gh-planning](https://github.com/maxbeizer/gh-planning) |  | Go | 1.24.2 | ⭐0 | `gh ext install maxbeizer/gh-planning` |
| [gh-rdm](https://github.com/maxbeizer/gh-rdm) | Remote Development Manager - gh CLI extension for clipboa... | Go | 1.24.13 | ⭐0 | `gh ext install maxbeizer/gh-rdm` |
| [gh-repo-peek](https://github.com/maxbeizer/gh-repo-peek) | Quick repo stats at a glance — a GitHub CLI extension | Shell | — | ⭐0 | `gh ext install maxbeizer/gh-repo-peek` |
| [gh-screenshot-to-codespace](https://github.com/maxbeizer/gh-screenshot-to-codespace) | gh extension to send screenshots to a GitHub Codespace fo... | Shell | — | ⭐0 | `gh ext install maxbeizer/gh-screenshot-to-codespace` |
| [gh-slack](https://github.com/maxbeizer/gh-slack) | Utility for archiving a slack conversation as markdown | Go | 1.24.0 | ⭐0 | `gh ext install maxbeizer/gh-slack` |
| [gh-slim-vtt](https://github.com/maxbeizer/gh-slim-vtt) | A tiny GH CLI extension for slimming down vtt files | Go | 1.19 | ⭐1 | `gh ext install maxbeizer/gh-slim-vtt` |
| [gh-sportsball](https://github.com/maxbeizer/gh-sportsball) | Watch live games in your terminal. Scores, play-by-play, ... | Go | 1.24.13 | ⭐2 | `gh ext install maxbeizer/gh-sportsball` |
| [gh-spotify](https://github.com/maxbeizer/gh-spotify) | Spotify in your terminal | Go | 1.25.0 | ⭐0 | `gh ext install maxbeizer/gh-spotify` |
| [gh-til](https://github.com/maxbeizer/gh-til) | Today I Learned — a developer second brain in your term... | Go | 1.21 | ⭐0 | `gh ext install maxbeizer/gh-til` |
| [gh-train-approval](https://github.com/maxbeizer/gh-train-approval) |  | Go | 1.16 | ⭐0 | `gh ext install maxbeizer/gh-train-approval` |
| [gh-uncle-max](https://github.com/maxbeizer/gh-uncle-max) |  | Go | 1.22 | ⭐0 | `gh ext install maxbeizer/gh-uncle-max` |
| [gh-wut](https://github.com/maxbeizer/gh-wut) | What do I need to know right now? — a GitHub CLI extension | Go | 1.25.0 | ⭐0 | `gh ext install maxbeizer/gh-wut` |
| [gh-yt2md](https://github.com/maxbeizer/gh-yt2md) | YouTube to Markdown - gh CLI extension | Go | 1.24.13 | ⭐0 | `gh ext install maxbeizer/gh-yt2md` |

---

*Catalog generated by [gh-fleet](https://github.com/maxbeizer/gh-fleet) on 2026-03-15.*

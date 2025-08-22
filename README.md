# PR Pilot ✈️

[![CI Status](https://github.com/bjess9/pr-pilot/workflows/CI/badge.svg)](https://github.com/bjess9/pr-pilot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjess9/pr-pilot)](https://goreportcard.com/report/github.com/bjess9/pr-pilot)
[![License](https://img.shields.io/github/license/bjess9/pr-pilot)](LICENSE)

TUI for tracking PRs across teams and repos. Auto-filters bot noise.

## Install

```bash
git clone https://github.com/bjess9/pr-pilot.git
cd pr-pilot  
make build
```

## Setup

**1. Auth (pick one):**
```bash
gh auth login                    # Use GitHub CLI (recommended)
export GITHUB_TOKEN="ghp_xxx"    # Or set env var
```

**2. Config:**
Create `~/.prpilot_config.yaml`:
```yaml
# Track repos by topic (recommended)
mode: "topics"
topics: ["backend", "frontend"]
topic_org: "your-org"

# Filter out bot spam
exclude_bots: true
include_drafts: true
```

See [example_config.yaml](example_config.yaml) for all options.

**3. Run:**
```bash
./pr-pilot
```

## Usage

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `Enter` | Open PR |
| `r` | Refresh |
| `f/s/d` | Filter |
| `c` | Clear filters |
| `h` | Help |
| `q` | Quit |

PRs are sorted by **most recent activity** (last updated), not creation date.

## Config Modes

```yaml
# By topic (recommended for teams)
mode: "topics"
topics: ["team-backend"]
topic_org: "company"

# By organization
mode: "organization" 
organization: "company"

# By specific repos
mode: "repos"
repos: ["company/api", "company/web"]

# By teams
mode: "teams"
organization: "company"
teams: ["backend-team"]

# Custom search
mode: "search"
search_query: "org:company is:pr is:open author:@me"
```

## Filtering

```yaml
exclude_bots: true              # Filters renovate/dependabot (default)
include_drafts: true            # Show draft PRs
exclude_authors: ["ci-bot"]     # Custom author exclusions  
exclude_titles: ["chore:", "docs:"]  # Title pattern exclusions
```

## Development

```bash
make test     # Run tests  
make build    # Build binary
make help     # Show all commands
```

That's it. No releases, no packages. Just clone, build, configure, run.

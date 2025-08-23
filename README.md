# PR Compass 🧭

[![CI Status](https://github.com/bjess9/pr-compass/workflows/CI/badge.svg)](https://github.com/bjess9/pr-compass/actions)
[![Docker Builds](https://github.com/bjess9/pr-compass/workflows/Docker%20Build%20and%20Push/badge.svg)](https://github.com/bjess9/pr-compass/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjess9/pr-compass)](https://goreportcard.com/report/github.com/bjess9/pr-compass)
[![License](https://img.shields.io/github/license/bjess9/pr-compass)](LICENSE)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/bjess9/pr-compass/pkgs/container/pr-compass)
[![Coverage](https://coveralls.io/repos/github/bjess9/pr-compass/badge.svg?branch=main)](https://coveralls.io/github/bjess9/pr-compass?branch=main)

TUI for tracking PRs across teams and repos. Auto-filters bot noise.

## 📋 Table of Contents

- [Install](#install)
- [Setup](#setup)
- [Usage](#usage)
- [Config Modes](#config-modes)
- [Filtering](#filtering)
- [Docker Support](#docker-support)
- [Security](#security)
- [Development](#development)

## Install

### Option 1: Docker (Recommended)

**📚 See**: [DOCKER.md](DOCKER.md)

### Option 2: Build from Source

```bash
git clone https://github.com/bjess9/pr-compass.git
cd pr-compass
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
mode: 'topics'
topics: ['backend', 'frontend']
topic_org: 'your-org'

# Filter out bot spam
exclude_bots: true
include_drafts: true
```

**📚 Detailed config options**: See [docs/configuration.md](docs/configuration.md)

**3. Run:**

```bash
./pr-compass        # Linux/macOS
# or pr-compass.exe # Windows
```

## Usage

| Key     | Action        |
| ------- | ------------- |
| `↑/↓`   | Navigate      |
| `Enter` | Open PR       |
| `r`     | Refresh       |
| `f/s/d` | Filter        |
| `c`     | Clear filters |
| `h`     | Help          |
| `q`     | Quit          |

PRs are sorted by **most recent activity** (last updated), not creation date.

## Config Modes

Five modes available: `topics` (recommended), `organization`, `repos`, `teams`, `search`.

**📚 Full configuration guide**: [docs/configuration.md](docs/configuration.md)

## 🐳 Docker Support

**📚 Docker usage**: [DOCKER.md](DOCKER.md)

## Security

PR Compass handles GitHub authentication tokens securely:

- ✅ **No token persistence by app** - PR Compass never writes tokens to files or databases
- ✅ **External token management** - uses environment variables or GitHub CLI's secure storage
- ✅ **Minimal permissions** - requires only `repo` and `read:org` scopes
- ✅ **Secure API communication** - all requests use HTTPS with proper validation

**Token Security**: When using `GITHUB_TOKEN` environment variable, you are responsible for securing it in your shell configuration. GitHub CLI (`gh auth login`) is recommended as it manages tokens securely.

## Development

**📚 Development guide**: [CONTRIBUTING.md](CONTRIBUTING.md)

**CI/CD**: Linux testing, security scanning, Docker builds, automated releases.

**📚 Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md)

# PR Pilot ✈️

[![CI Status](https://github.com/bjess9/pr-pilot/workflows/CI/badge.svg)](https://github.com/bjess9/pr-pilot/actions)
[![Docker Builds](https://github.com/bjess9/pr-pilot/workflows/Docker%20Build%20and%20Push/badge.svg)](https://github.com/bjess9/pr-pilot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjess9/pr-pilot)](https://goreportcard.com/report/github.com/bjess9/pr-pilot)
[![License](https://img.shields.io/github/license/bjess9/pr-pilot)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/bjess9/pr-pilot)](https://hub.docker.com/r/bjess9/pr-pilot)
[![Coverage](https://codecov.io/gh/bjess9/pr-pilot/branch/main/graph/badge.svg)](https://codecov.io/gh/bjess9/pr-pilot)

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

```bash
# Pull and run directly
docker run --rm -e GITHUB_TOKEN=your_token bjess9/pr-pilot:latest

# Or using GitHub CLI authentication
docker run --rm -v ~/.config/gh:/root/.config/gh:ro bjess9/pr-pilot:latest
```

### Option 2: Build from Source

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
mode: 'topics'
topics: ['backend', 'frontend']
topic_org: 'your-org'

# Filter out bot spam
exclude_bots: true
include_drafts: true
```

See [example_config.yaml](example_config.yaml) for all options.

**3. Run:**

```bash
./pr-pilot        # Linux/macOS
# or pr-pilot.exe # Windows
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
exclude_bots: true # Filters renovate/dependabot (default)
include_drafts: true # Show draft PRs
exclude_authors: ['ci-bot'] # Custom author exclusions
exclude_titles: ['chore:', 'docs:'] # Title pattern exclusions
```

## 🐳 Docker Support

PR Pilot offers full Docker support for easy deployment and development:

### Quick Start

```bash
# Using environment token
docker run --rm -e GITHUB_TOKEN=ghp_your_token bjess9/pr-pilot:latest

# Using GitHub CLI authentication (recommended)
docker run --rm -v ~/.config/gh:/root/.config/gh:ro bjess9/pr-pilot:latest

# With custom configuration
docker run --rm -v $(pwd)/config:/root/.config -e GITHUB_TOKEN=your_token bjess9/pr-pilot:latest
```

### Development with Docker Compose

```bash
# Start development environment
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# Access development container
docker-compose exec pr-pilot sh

# Run tests inside container
docker-compose exec pr-pilot go test ./...
```

### Available Images

- `bjess9/pr-pilot:latest` - Latest stable release
- `bjess9/pr-pilot:v1.0.0` - Specific version
- `bjess9/pr-pilot:main` - Latest development build

**Multi-architecture support**: Images are available for `linux/amd64` and `linux/arm64`.

📚 **For detailed Docker usage**, see [DOCKER.md](DOCKER.md)

## Security

PR Pilot handles GitHub authentication tokens securely:

- ✅ **No token persistence by app** - PR Pilot never writes tokens to files or databases
- ✅ **External token management** - uses environment variables or GitHub CLI's secure storage
- ✅ **Minimal permissions** - requires only `repo` and `read:org` scopes
- ✅ **Secure API communication** - all requests use HTTPS with proper validation

**Token Security**: When using `GITHUB_TOKEN` environment variable, you are responsible for securing it in your shell configuration. GitHub CLI (`gh auth login`) is recommended as it manages tokens securely.

## Development

### Local Development

```bash
make test            # Run all tests
make test-coverage   # Generate coverage report
make build           # Build binary
make clean           # Clean build artifacts
make help            # Show all commands
```

### Docker Development

```bash
# Start development environment
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# Access development container
docker-compose exec pr-pilot sh

# Or use the development tools container
docker-compose exec dev-tools sh
```

### CI/CD Pipeline

The project includes a comprehensive CI/CD pipeline with:

- ✅ **Multi-platform testing**: Ubuntu, Windows, macOS
- ✅ **Multiple Go versions**: 1.20, 1.21, 1.22
- ✅ **Security scanning**: Gosec, govulncheck, Nancy
- ✅ **Coverage reporting**: Codecov, Coveralls
- ✅ **Docker builds**: Multi-architecture (amd64/arm64)
- ✅ **Automated releases**: Docker Hub integration

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `make test`
5. Check security: `make security-scan` (if available)
6. Submit a pull request

That's it. Simple setup, comprehensive testing, Docker-ready deployment.

<div align="center">

# 🧭 PR Compass

<h3>Terminal-based pull request monitoring for developers and teams</h3>

<p>
  <a href="https://github.com/bjess9/pr-compass/actions"><img src="https://github.com/bjess9/pr-compass/workflows/CI/badge.svg" alt="CI Status"></a>
  <a href="https://github.com/bjess9/pr-compass/actions"><img src="https://github.com/bjess9/pr-compass/workflows/Docker%20Build%20and%20Push/badge.svg" alt="Docker Builds"></a>
  <a href="https://goreportcard.com/report/github.com/bjess9/pr-compass"><img src="https://goreportcard.com/badge/github.com/bjess9/pr-compass" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/bjess9/pr-compass" alt="License"></a>
  <br>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/bjess9/pr-compass" alt="Go Version"></a>
  <a href="https://github.com/bjess9/pr-compass/pkgs/container/pr-compass"><img src="https://img.shields.io/badge/docker-ghcr.io-blue" alt="Docker Image"></a>
  <a href="https://coveralls.io/github/bjess9/pr-compass?branch=main"><img src="https://coveralls.io/repos/github/bjess9/pr-compass/badge.svg?branch=main" alt="Coverage"></a>
</p>

<p>
  <a href="#quick-start">Quick Start</a> •
  <a href="#installation">Installation</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#usage">Usage</a> •
  <a href="#documentation">Documentation</a>
</p>

</div>

Terminal-based GitHub pull request monitoring across multiple repositories.

**Features:** Multi-repo view, keyboard navigation, flexible configuration, Docker support.

---

## Quick Start

```bash
# Install
docker pull ghcr.io/bjess9/pr-compass:latest

# Authenticate
gh auth login

# Configure
echo "mode: 'topics'" > ~/.prcompass_config.yaml
echo "topics: ['backend']" >> ~/.prcompass_config.yaml
echo "topic_org: 'your-org'" >> ~/.prcompass_config.yaml

# Run
docker run --rm -v ~/.config/gh:/root/.config/gh:ro ghcr.io/bjess9/pr-compass:latest
```

## Installation

```bash
# Docker
docker pull ghcr.io/bjess9/pr-compass:latest

# Build from source
git clone https://github.com/bjess9/pr-compass.git && cd pr-compass && make build
```

## Configuration

Supports `topics`, `organization`, `repos`, `teams`, `search` modes.

**Details:** [docs/configuration.md](docs/configuration.md)

## Usage

|   Key   |    Action     | Description         |
| :-----: | :-----------: | :------------------ |
| `↑` `k` |  Navigate up  | Move selection up   |
| `↓` `j` | Navigate down | Move selection down |
| `Enter` |    Open PR    | Open in browser     |
|   `r`   |    Refresh    | Fetch latest data   |
|   `f`   |    Filter     | Draft/Open/All      |
|   `q`   |     Quit      | Exit                |

**More shortcuts:** `h` for help

## Documentation

[Configuration](docs/configuration.md) • [Docker](DOCKER.md) • [Contributing](CONTRIBUTING.md) • [Troubleshooting](docs/troubleshooting.md)

## License

MIT

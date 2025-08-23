<div align="center">

# 🧭 PR Compass

**Navigate Your GitHub Pull Requests with Confidence**

_A terminal-based pull request monitoring tool for developers and teams_

[![CI Status](https://github.com/bjess9/pr-compass/workflows/CI/badge.svg)](https://github.com/bjess9/pr-compass/actions)
[![Docker Builds](https://github.com/bjess9/pr-compass/workflows/Docker%20Build%20and%20Push/badge.svg)](https://github.com/bjess9/pr-compass/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjess9/pr-compass)](https://goreportcard.com/report/github.com/bjess9/pr-compass)
[![License](https://img.shields.io/github/license/bjess9/pr-compass)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bjess9/pr-compass)](go.mod)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/bjess9/pr-compass/pkgs/container/pr-compass)
[![Coverage](https://coveralls.io/repos/github/bjess9/pr-compass/badge.svg?branch=main)](https://coveralls.io/github/bjess9/pr-compass?branch=main)

[Features](#-features) • [Quick Start](#-quick-start) • [Installation](#-installation) • [Configuration](#-configuration) • [Docker](#-docker) • [Contributing](#-contributing)

---

</div>

## 🎯 What is PR Compass?

PR Compass is a **terminal-based interface** for monitoring GitHub pull requests across multiple repositories. Built for developers who need to stay on top of code reviews and track PR activity.

### 🚀 Why PR Compass?

- **🔍 Multi-Repository View** - See PRs from multiple repos in one place
- **⚡ Efficient Navigation** - Keyboard shortcuts and clean interface
- **🎛️ Flexible Configuration** - Track by repositories, topics, teams, or organizations
- **🧹 Clean Interface** - Simple TUI that focuses on what matters
- **🐳 Docker Support** - Run with Docker if preferred
- **🔒 Secure** - Uses existing GitHub authentication, no token storage

---

## ✨ Features

### 📋 **PR Management**

- **Multi-Repository View** - Monitor PRs across multiple repositories
- **Flexible Organization** - Group by topics, organizations, or teams
- **Bot Filtering** - Hide automated PRs (Dependabot, etc.)
- **Draft Support** - Show or hide draft pull requests

### ⚡ **Interface & Navigation**

- **Keyboard Navigation** - Vim-style shortcuts for efficient browsing
- **PR Information** - View status, reviewers, labels, and summaries
- **Sorting Options** - Sort by activity, creation date, or comments
- **Quick Actions** - Open PRs in browser with a keypress

### 🔧 **Configuration**

- **Multiple Auth Methods** - GitHub CLI, environment variables, or tokens
- **Flexible Modes** - Track by repos, topics, teams, organizations, or search
- **Docker Support** - Run locally or in containers
- **YAML Configuration** - Simple config file setup

---

## 🚀 Quick Start

Get up and running quickly:

### 1️⃣ **Install**

```bash
# Docker (Recommended)
docker pull ghcr.io/bjess9/pr-compass:latest

# Or build from source
git clone https://github.com/bjess9/pr-compass.git && cd pr-compass && make build
```

### 2️⃣ **Authenticate**

```bash
# Using GitHub CLI (recommended)
gh auth login

# Or set environment variable
export GITHUB_TOKEN="ghp_your_token_here"
```

### 3️⃣ **Configure**

Create `~/.prcompass_config.yaml`:

```yaml
# Track repositories by topics (recommended for teams)
mode: 'topics'
topics: ['backend', 'frontend', 'infrastructure']
topic_org: 'your-organization'

# Filter configuration
exclude_bots: true # Hide bot PRs (Dependabot, etc.)
include_drafts: true # Show draft PRs
max_age_days: 30 # Only show PRs from last 30 days
```

### 4️⃣ **Launch**

```bash
./pr-compass
```

**🎉 That's it!** You'll see a clean interface showing PRs from your configured repositories.

---

## 📦 Installation

### 🐳 Docker (Recommended)

```bash
docker pull ghcr.io/bjess9/pr-compass:latest
```

**📚 Full Docker setup**: [DOCKER.md](DOCKER.md)

### 🔧 Build from Source

**Requirements**: Go 1.21+

```bash
# Clone repository
git clone https://github.com/bjess9/pr-compass.git
cd pr-compass

# Build
make build

# Install globally (optional)
sudo cp pr-compass /usr/local/bin/
```

### 📦 Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/bjess9/pr-compass/releases).

---

## ⚙️ Configuration

PR Compass supports **5 flexible configuration modes**: `topics` (recommended), `organization`, `repos`, `teams`, `search`.

**📚 Complete configuration guide**: [docs/configuration.md](docs/configuration.md)

---

## 🎮 Usage

### ⌨️ **Keyboard Shortcuts**

| Key         | Action           | Description              |
| ----------- | ---------------- | ------------------------ |
| `↑` `k`     | Navigate up      | Move selection up        |
| `↓` `j`     | Navigate down    | Move selection down      |
| `Enter` `o` | Open PR          | Open in default browser  |
| `r`         | Refresh          | Fetch latest PR data     |
| `f`         | Filter by status | Draft/Open/All           |
| `s`         | Sort options     | Updated/Created/Comments |
| `d`         | Toggle drafts    | Show/hide draft PRs      |
| `c`         | Clear filters    | Reset all filters        |
| `h` `?`     | Help             | Show help screen         |
| `q` `Esc`   | Quit             | Exit application         |

### 📊 **PR Information Display**

Each PR shows:

- **Status indicators** (🟢 approved, 🟡 pending, 🔴 changes requested)
- **Author and repository** information
- **Labels and assignees**
- **Review status** and comment counts
- **Last update time** (sorted by most recent activity)

---

## 🐳 Docker

Full Docker support with multi-architecture builds (AMD64/ARM64).

**📚 Complete Docker documentation**: [DOCKER.md](DOCKER.md)

---

## 🔒 Security

PR Compass never stores GitHub tokens and uses existing authentication (GitHub CLI or environment variables). Requires only `repo` and `read:org` scopes.

---

## 🛠️ Development

**📚 Development setup**: [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 🤝 Contributing

We welcome contributions! See our detailed guide for setup, testing, and PR guidelines.

**📚 Contributing guide**: [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📚 Documentation

| Document                                     | Description                                      |
| -------------------------------------------- | ------------------------------------------------ |
| [Configuration Guide](docs/configuration.md) | Detailed configuration options and examples      |
| [Docker Guide](DOCKER.md)                    | Docker deployment and development setup          |
| [Architecture](docs/architecture.md)         | Technical design decisions and project structure |
| [Troubleshooting](docs/troubleshooting.md)   | Common issues and solutions                      |
| [Contributing](CONTRIBUTING.md)              | Development setup and contribution guidelines    |

---

## 🌟 Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/bjess9/pr-compass/issues)  
- 📖 **Documentation**: [docs/](docs/) folder

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built for developers who work with pull requests**

[⭐ Star this repository](https://github.com/bjess9/pr-compass) if PR Compass helps you stay on top of your pull requests!

</div>

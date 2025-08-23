<div align="center">

# 🧭 PR Compass

**Navigate Your GitHub Pull Requests with Confidence**

*A blazing-fast, terminal-based pull request monitoring tool for developers and teams*

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

PR Compass is a **powerful terminal-based interface** that gives you complete visibility into GitHub pull requests across your entire organization. Built for developers who need to stay on top of code reviews, track team progress, and never miss important PRs.

### 🚀 Why PR Compass?

- **🔍 Smart Filtering** - Automatically filters out bot noise while surfacing PRs that need your attention
- **⚡ Lightning Fast** - Native Go performance with concurrent API calls and smart caching
- **🎛️ Flexible Configuration** - Track by repositories, topics, teams, or organizations
- **🧹 Clean Interface** - Beautiful TUI that focuses on what matters most
- **🐳 Docker Ready** - Run anywhere with containerized deployment
- **🔒 Security First** - Never stores tokens, integrates with GitHub CLI securely

---

## ✨ Features

### 🎪 **Smart PR Discovery**
- **Multi-Repository Tracking** - Monitor PRs across your entire organization
- **Topic-Based Organization** - Group repositories by topics for better organization  
- **Team-Focused Views** - Track PRs for specific GitHub teams
- **Bot Noise Filtering** - Automatically hide Dependabot and other automated PRs

### ⚡ **Performance & Efficiency**
- **Concurrent Processing** - Parallel API calls with intelligent batching
- **Real-Time Updates** - Live progress indicators during data fetching
- **Smart Caching** - Reduced API calls with intelligent caching strategies
- **Rate Limit Handling** - Automatic exponential backoff and retry logic

### 🎨 **Developer Experience**
- **Intuitive Keyboard Navigation** - Vim-inspired shortcuts for power users
- **Rich PR Information** - Status, reviewers, labels, and change summaries
- **Flexible Sorting** - Sort by activity, creation date, or priority
- **Quick Actions** - Open PRs in browser with a single keypress

### 🛠️ **Enterprise Ready**
- **Multiple Authentication Methods** - GitHub CLI, environment variables, or direct tokens
- **Docker Support** - Full containerization with development and production configs
- **Secure Token Handling** - No token persistence, external token management
- **CI/CD Integration** - Automated testing, security scanning, and releases

---

## 🚀 Quick Start

Get up and running in under 2 minutes:

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
exclude_bots: true      # Hide bot PRs (Dependabot, etc.)
include_drafts: true    # Show draft PRs
max_age_days: 30       # Only show PRs from last 30 days
```

### 4️⃣ **Launch**

```bash
./pr-compass
```

**🎉 That's it!** You'll see a beautiful interface showing all relevant PRs across your organization.

---

## 📦 Installation

### 🐳 Docker (Recommended)

```bash
# Quick run
docker run -it --rm \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -v ~/.prcompass_config.yaml:/root/.prcompass_config.yaml:ro \
  ghcr.io/bjess9/pr-compass:latest

# Using Docker Compose
git clone https://github.com/bjess9/pr-compass.git
cd pr-compass
docker-compose up
```

**📚 Full Docker guide**: [DOCKER.md](DOCKER.md)

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

PR Compass supports **5 flexible configuration modes**:

| Mode | Use Case | Configuration |
|------|----------|---------------|
| **`topics`** ⭐ | **Recommended** - Track repos by GitHub topics | `topics: ['backend', 'frontend']` |
| **`organization`** | Monitor entire org | `organization: 'your-org'` |
| **`repos`** | Specific repositories | `repositories: ['org/repo1', 'org/repo2']` |
| **`teams`** | Team-based tracking | `teams: ['backend-team']` |
| **`search`** | Custom search queries | `search_queries: ['label:urgent']` |

### 📋 **Complete Configuration Example**

```yaml
# ~/.prcompass_config.yaml
mode: 'topics'
topics: ['web', 'api', 'infrastructure']
topic_org: 'acme-corp'

# Filtering
exclude_bots: true
include_drafts: true
max_age_days: 14

# Display preferences  
sort_by: 'updated'        # updated, created, comments
max_results: 50
show_descriptions: true

# Performance tuning
api_timeout: 30
concurrent_requests: 5
cache_ttl_minutes: 5
```

**📚 Detailed configuration guide**: [docs/configuration.md](docs/configuration.md)

---

## 🎮 Usage

### ⌨️ **Keyboard Shortcuts**

| Key | Action | Description |
|-----|--------|-------------|
| `↑` `k` | Navigate up | Move selection up |
| `↓` `j` | Navigate down | Move selection down |  
| `Enter` `o` | Open PR | Open in default browser |
| `r` | Refresh | Fetch latest PR data |
| `f` | Filter by status | Draft/Open/All |
| `s` | Sort options | Updated/Created/Comments |
| `d` | Toggle drafts | Show/hide draft PRs |
| `c` | Clear filters | Reset all filters |
| `h` `?` | Help | Show help screen |
| `q` `Esc` | Quit | Exit application |

### 📊 **PR Information Display**

Each PR shows:
- **Status indicators** (🟢 approved, 🟡 pending, 🔴 changes requested)
- **Author and repository** information
- **Labels and assignees**
- **Review status** and comment counts
- **Last update time** (sorted by most recent activity)

---

## 🐳 Docker

Full Docker support with multi-architecture builds (AMD64/ARM64):

```bash
# Development with live config reload
docker-compose -f docker-compose.dev.yml up

# Production deployment
docker run -d \
  --name pr-compass \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -v ~/.prcompass_config.yaml:/root/.prcompass_config.yaml:ro \
  ghcr.io/bjess9/pr-compass:latest
```

**📚 Complete Docker documentation**: [DOCKER.md](DOCKER.md)

---

## 🔒 Security

PR Compass follows **security-first principles**:

- ✅ **Zero Token Persistence** - Never writes tokens to disk
- ✅ **External Token Management** - Leverages GitHub CLI or environment variables
- ✅ **Minimal Permissions** - Requires only `repo` and `read:org` scopes
- ✅ **Secure Communication** - All API calls use HTTPS with proper validation
- ✅ **Container Security** - Docker images scanned with Trivy, runs as non-root user

### 🛡️ **Token Security Best Practices**

1. **Use GitHub CLI** (recommended):
   ```bash
   gh auth login --scopes repo,read:org
   ```

2. **Environment Variables** (for CI/CD):
   ```bash
   export GITHUB_TOKEN="ghp_your_token_here"
   ```

3. **Avoid** storing tokens in config files or shell history

---

## 🛠️ Development

### 🏗️ **Local Development**

```bash
# Setup
git clone https://github.com/bjess9/pr-compass.git
cd pr-compass
make dev-setup

# Run tests
make test

# Build and run
make dev

# Watch mode for development
make test-watch
```

### 🧪 **Testing**

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Integration tests
make test-integration

# Lint code
make lint
```

### 📊 **Project Stats**

- **Language**: Go 1.23+
- **Test Coverage**: 85%+ 
- **Dependencies**: Minimal, security-focused
- **CI/CD**: GitHub Actions with security scanning
- **Container**: Multi-arch builds (AMD64/ARM64)

---

## 🤝 Contributing

We welcome contributions! PR Compass is built by developers, for developers.

### 🚀 **How to Contribute**

1. **🍴 Fork the repository**
2. **🌟 Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **✍️ Make your changes** with tests
4. **✅ Ensure tests pass**: `make test`
5. **📝 Submit a pull request**

### 🐛 **Bug Reports**

Found a bug? Please [open an issue](https://github.com/bjess9/pr-compass/issues/new) with:
- Steps to reproduce
- Expected vs actual behavior  
- System information (OS, Go version)

### 💡 **Feature Requests**

Have an idea? We'd love to hear it! [Open a feature request](https://github.com/bjess9/pr-compass/issues/new) with:
- Use case description
- Proposed implementation
- Why it would benefit the community

**📚 Detailed contributing guide**: [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [Configuration Guide](docs/configuration.md) | Detailed configuration options and examples |
| [Docker Guide](DOCKER.md) | Docker deployment and development setup |
| [Architecture](docs/architecture.md) | Technical design decisions and project structure |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and solutions |
| [Contributing](CONTRIBUTING.md) | Development setup and contribution guidelines |

---

## 🌟 Support

- 📖 **Documentation**: Comprehensive guides in [docs/](docs/)
- 🐛 **Issues**: [GitHub Issues](https://github.com/bjess9/pr-compass/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/bjess9/pr-compass/discussions)
- 🔒 **Security**: Report vulnerabilities via [GitHub Security](https://github.com/bjess9/pr-compass/security)

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built with ❤️ by developers, for developers**

[⭐ Star this repository](https://github.com/bjess9/pr-compass) if PR Compass helps you stay on top of your pull requests!

</div>
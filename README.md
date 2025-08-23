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
  <a href="#-features">Features</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-configuration">Configuration</a> •
  <a href="#-docker">Docker</a> •
  <a href="#-contributing">Contributing</a>
</p>

</div>

## 🎯 What is PR Compass?

PR Compass is a **terminal-based interface** for monitoring GitHub pull requests across multiple repositories. Built for developers who need to stay on top of code reviews and track PR activity.

<details>
<summary><strong>🚀 Why choose PR Compass?</strong></summary>

<br>

| Feature                       | Benefit                                                |
| ----------------------------- | ------------------------------------------------------ |
| 🔍 **Multi-Repository View**  | See PRs from multiple repos in one place               |
| ⚡ **Efficient Navigation**   | Keyboard shortcuts and clean interface                 |
| 🎛️ **Flexible Configuration** | Track by repositories, topics, teams, or organizations |
| 🧹 **Clean Interface**        | Simple TUI that focuses on what matters                |
| 🐳 **Docker Support**         | Run with Docker if preferred                           |
| 🔒 **Secure**                 | Uses existing GitHub authentication, no token storage  |

</details>

<div align="center">

**─────────────────────────────────────────**

</div>

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

<div align="center">

|     Key     |      Action      | Description              |
| :---------: | :--------------: | :----------------------- |
|   `↑` `k`   |   Navigate up    | Move selection up        |
|   `↓` `j`   |  Navigate down   | Move selection down      |
| `Enter` `o` |     Open PR      | Open in default browser  |
|     `r`     |     Refresh      | Fetch latest PR data     |
|     `f`     | Filter by status | Draft/Open/All           |
|     `s`     |   Sort options   | Updated/Created/Comments |
|     `d`     |  Toggle drafts   | Show/hide draft PRs      |
|     `c`     |  Clear filters   | Reset all filters        |
|   `h` `?`   |       Help       | Show help screen         |
|  `q` `Esc`  |       Quit       | Exit application         |

</div>

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

<div align="center">

**─────────────────────────────────────────**

<br>

**Built for developers who work with pull requests**

<br>

<a href="https://github.com/bjess9/pr-compass">
  <img src="https://img.shields.io/badge/⭐_Star_this_repo-black?style=for-the-badge&logoColor=yellow" alt="Star this repository">
</a>

<br><br>

```ascii
Thanks for checking out PR Compass! 🧭
```

</div>

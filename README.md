# PR Pilot ✈️

[![CI Status](https://github.com/bjess9/pr-pilot/workflows/CI/badge.svg)](https://github.com/bjess9/pr-pilot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjess9/pr-pilot)](https://goreportcard.com/report/github.com/bjess9/pr-pilot)
[![License](https://img.shields.io/github/license/bjess9/pr-pilot)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bjess9/pr-pilot)](go.mod)

A powerful CLI tool for tracking and managing Pull Requests across teams and organizations with a clean terminal interface.

## Features

- **🎯 Team-Based Tracking** - Monitor PRs by organization, teams, repository topics, or custom search queries
- **📊 Rich PR Details** - View status, reviews, labels, age, and conflicts at a glance  
- **🚀 Multiple Tracking Modes** - Support for repos, organization-wide, team-based, topic-based, and custom search
- **🎨 Intuitive TUI** - Clean terminal interface with keyboard navigation
- **⚡ Auto-Refresh** - Automatic PR list updates every 60 seconds
- **🔍 Smart Filtering** - Filter by drafts, status, author, and more
- **🌐 Browser Integration** - Open PRs directly in your default browser
- **📱 Cross-Platform** - Works on macOS, Linux, and Windows

## Installation

### Quick Install

```bash
# Download latest release
curl -L -o pr-pilot.tar.gz https://github.com/bjess9/pr-pilot/releases/latest/download/pr-pilot_$(uname -s)_$(uname -m).tar.gz
tar -xzf pr-pilot.tar.gz
sudo mv pr-pilot /usr/local/bin/

# Or use Go install
go install github.com/bjess9/pr-pilot/cmd/pr-pilot@latest
```

### Platform-Specific

<details>
<summary>macOS</summary>

```bash
# Intel Macs
curl -L -o pr-pilot.tar.gz https://github.com/bjess9/pr-pilot/releases/latest/download/pr-pilot_Darwin_x86_64.tar.gz

# Apple Silicon Macs  
curl -L -o pr-pilot.tar.gz https://github.com/bjess9/pr-pilot/releases/latest/download/pr-pilot_Darwin_arm64.tar.gz

tar -xzf pr-pilot.tar.gz && sudo mv pr-pilot /usr/local/bin/
```

</details>

<details>
<summary>Linux</summary>

```bash
# x86_64
curl -L -o pr-pilot.tar.gz https://github.com/bjess9/pr-pilot/releases/latest/download/pr-pilot_Linux_x86_64.tar.gz

# ARM64
curl -L -o pr-pilot.tar.gz https://github.com/bjess9/pr-pilot/releases/latest/download/pr-pilot_Linux_arm64.tar.gz

tar -xzf pr-pilot.tar.gz && sudo mv pr-pilot /usr/local/bin/
```

</details>

<details>
<summary>Windows</summary>

```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/bjess9/pr-pilot/releases/latest/download/pr-pilot_Windows_x86_64.zip" -OutFile "pr-pilot.zip"
Expand-Archive pr-pilot.zip -DestinationPath "C:\tools\"
# Add C:\tools to your PATH
```

</details>

## Quick Start

1. **Authenticate with GitHub:**
   ```bash
   pr-pilot auth login
   ```

2. **Configure PR tracking:**
   ```bash
   pr-pilot configure
   ```

3. **Start monitoring PRs:**
   ```bash
   pr-pilot
   ```

## Configuration

PR Pilot supports multiple tracking modes to fit different team workflows:

### Repository Topics (Recommended)
Track all PRs from repositories tagged with specific topics - perfect for team-based workflows:

```yaml
mode: "topics"
topic_org: "your-org"
topics:
  - "backend"
  - "infrastructure" 
  - "api"
```

### Organization-Wide
Monitor all PRs across your GitHub organization:

```yaml
mode: "organization"  
organization: "your-org"
```

### Specific Repositories
Track individual repositories:

```yaml
mode: "repos"
repos:
  - "your-org/repo1"
  - "your-org/repo2"
```

### Custom Search
Use GitHub's search API for advanced filtering:

```yaml
mode: "search"
search_query: "org:myorg is:pr is:open author:@me"
```

Configuration is stored in `~/.prpilot_config.yaml`. See [example_config.yaml](example_config.yaml) for all options.

## Usage

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate PR list |
| `Enter` | Open PR in browser |
| `r` | Refresh PR list |
| `d` | Filter to draft PRs |
| `c` | Clear all filters |
| `h` or `?` | Show help |
| `q` or `Ctrl+C` | Quit |

### Command Line Options

```bash
pr-pilot                  # Start with interactive TUI
pr-pilot configure        # Set up configuration
pr-pilot auth login       # Authenticate with GitHub
pr-pilot --version        # Show version
pr-pilot --help          # Show help
```

## Testing

PR Pilot includes comprehensive testing without external dependencies:

```bash
# Run all tests
make test

# Unit tests only
make test-unit

# Integration tests  
make test-integration

# With coverage
make test-coverage
```

All tests use mock GitHub clients and don't require API tokens or network access. See [TESTING.md](TESTING.md) for details.

## Development

### Prerequisites

- Go 1.21+
- Make (optional)

### Setup

```bash
git clone https://github.com/bjess9/pr-pilot.git
cd pr-pilot
go mod download
make test
make build
```

### Project Structure

```
pr-pilot/
├── cmd/pr-pilot/          # Main application
├── internal/              # Internal packages
│   ├── auth/             # GitHub authentication
│   ├── config/           # Configuration management
│   ├── github/           # GitHub API client + mocks
│   └── ui/               # Terminal UI components
├── .github/workflows/    # CI/CD pipelines  
└── test/                 # Test utilities
```

### Build Commands

```bash
make build          # Build binary
make test           # Run tests
make lint           # Lint code
make clean          # Clean artifacts
make dev            # Run in development mode
```

## Architecture

PR Pilot follows a clean architecture pattern:

- **CLI Layer** (`cmd/`) - Command-line interface and argument parsing
- **UI Layer** (`internal/ui/`) - Terminal user interface using Bubble Tea
- **Business Logic** (`internal/config/`, `internal/github/`) - Core functionality
- **External APIs** (`internal/github/client.go`) - GitHub API integration

The application is designed to be testable with comprehensive mocks for external dependencies.

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Built for developers, by developers.** ⭐ Star this project if you find it useful!

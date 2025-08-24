# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PR Compass is a terminal-based GitHub pull request monitoring tool built in Go. It provides a TUI for tracking PRs across multiple repositories with flexible configuration modes including repos, organization, teams, search queries, and topics.

## Key Commands

### Build and Run
- `make build` - Build the pr-compass binary
- `make dev` - Build and run in development mode
- `./pr-compass` - Run the built binary

### Testing
- `make test` - Run all tests (unit + integration)
- `make test-unit` - Run unit tests only (`go test -v ./internal/...`)
- `make test-integration` - Run integration tests using custom test runner
- `make test-coverage` - Generate HTML coverage report
- `make test-ci` - Run tests in CI mode (quiet output)

### Code Quality
- `make fmt` - Format code (`go fmt ./...`)
- `make lint` - Lint code (`golangci-lint run`)
- `make security` - Security check (`gosec ./...`)
- `make check` - Run fmt + lint + unit tests

### Development Setup
- `make setup` - Full development environment setup (installs tools + creates config)
- `make dev-config` - Copy example config to `~/.prcompass_config.yaml`
- `make dev-deps` - Install golangci-lint and gosec

## Architecture

### Core Components

**Authentication (`internal/auth/`)**
- Token resolution hierarchy: `GITHUB_TOKEN` env var → GitHub CLI token (`gh auth token`) → fallback
- Uses GitHub CLI integration for seamless authentication

**Configuration (`internal/config/`)**
- Config file: `~/.prcompass_config.yaml`
- Five modes: `repos`, `organization`, `teams`, `search`, `topics`
- Auto-mode detection for backward compatibility
- Filtering options: exclude bots, authors, title patterns

**GitHub Integration (`internal/github/`)**
- Strategy pattern with `PRFetcher` interface for different data sources
- Implementations: `ReposFetcher`, `OrganizationFetcher`, `TeamsFetcher`, `SearchFetcher`, `TopicsFetcher`
- GraphQL client for efficient API usage
- Cached and optimized fetchers for performance

**UI Layer (`internal/ui/`)**
- Built with Charm's Bubble Tea TUI framework
- Table-based PR display with keyboard navigation
- Progressive enhancement of PR data (comments, reviews, checks)
- Real-time refresh capabilities

**Performance & Caching (`internal/cache/`, `internal/batch/`)**
- Generic cache system with TTL support for PR data
- Batch manager with worker pools for concurrent API calls
- File-based caching to reduce GitHub API usage

### Data Flow

1. **Config Loading**: Viper loads YAML config with mode detection
2. **Authentication**: Token resolution via auth hierarchy  
3. **Fetcher Selection**: Factory pattern selects appropriate `PRFetcher`
4. **Data Fetching**: Concurrent API calls with caching and batch processing
5. **UI Rendering**: Bubble Tea renders table with progressive enhancement
6. **Background Updates**: Periodic refresh and real-time data enhancement

### Key Dependencies

- **Bubble Tea**: TUI framework for terminal interface
- **Viper**: Configuration management with YAML support  
- **go-github/v55**: GitHub API client library
- **OAuth2**: GitHub authentication handling

### Testing Strategy

- **Unit Tests**: Test individual packages in `internal/` 
- **Integration Tests**: Custom test runner in `test/integration/`
- **Behavior Tests**: Component interaction testing across packages
- **Mock Infrastructure**: GitHub API mocking for reliable tests

### Configuration Modes

- **`repos`**: Explicit repository list (legacy mode)
- **`organization`**: All repos in an organization 
- **`teams`**: Team-specific repository access
- **`search`**: Custom GitHub search queries
- **`topics`**: Track repositories by topic tags (recommended for teams)

## Development Notes

- Uses Go 1.23.10 with modules
- Main executable in `cmd/pr-compass/main.go`
- Configuration error handling via custom error types in `internal/errors/`
- Concurrent processing patterns throughout for GitHub API efficiency
- File path: Configuration at `~/.prcompass_config.yaml` (note: main.go references legacy `~/.prpilot_config.yaml`)
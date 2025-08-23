# Architecture

## Design Decisions

**Authentication**: Never store tokens. Use external auth (env vars, GitHub CLI) for security.

**Concurrency**: Parallel repository fetching with goroutines. Context cancellation prevents resource leaks.

**Error Handling**: Fail gracefully. Skip broken repos, continue operation. Structured errors with user-friendly messages.

**UI Pattern**: Bubble Tea MVC. Single state machine, immutable updates.

**Configuration**: YAML file. Mode-based fetcher selection via factory pattern.

## Key Abstractions

**PRFetcher Interface**: Strategy pattern for different GitHub query types (repos, orgs, teams, search, topics).

**PRFilter**: Pipeline pattern. API results → bot filtering → user filtering → UI display.

**Context Propagation**: All GitHub API calls respect cancellation. Clean shutdown guaranteed.

## Non-obvious Choices

**Why no caching?** PRs change frequently. Fresh data preferred over stale cache complexity.

**Why concurrent per-repo?** GitHub API limits per-endpoint, not total. Parallel fetching 3-5x faster.

**Why factory pattern for fetchers?** Config determines strategy at runtime. Clean separation of concerns.

**Why structured errors?** User vs developer messaging. Prevents sensitive data leaks in errors.

# Contributing

## Quick Start

```bash
git clone https://github.com/bjess9/pr-pilot.git
cd pr-pilot
make build
make test
```

## Development

### Requirements

- Go 1.20+
- GitHub CLI (`gh`) or GitHub token

### Setup

```bash
make dev-setup    # Install dependencies
make test-watch   # TDD mode
```

### Code Style

- Run `gofmt` before committing
- Follow Go best practices
- Add tests for new features
- Keep functions small and focused

## Testing

```bash
make test          # All tests
make test-coverage # Coverage report
make test-ci       # CI simulation
```

### Test Requirements

- Unit tests for all new functions
- Integration tests for main flows
- Coverage > 80% for new packages
- All tests must pass

## Pull Requests

### Process

1. Fork → feature branch
2. Add tests
3. Run `make test lint`
4. Submit PR with clear description
5. Address review feedback

### PR Guidelines

- Clear, descriptive title
- Reference related issues
- Include test coverage
- Update docs if needed
- Small, focused changes preferred

## Issues

### Bug Reports

Include:

- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, Go version)
- Configuration (sanitized)

### Feature Requests

Include:

- Use case description
- Proposed solution
- Alternative approaches considered

## Architecture

- `cmd/` - CLI entry points
- `internal/auth/` - GitHub authentication
- `internal/config/` - Configuration management
- `internal/github/` - GitHub API client
- `internal/ui/` - Terminal UI components
- `internal/errors/` - Error handling

## Code Quality

### Required Checks

- All tests pass
- No linter warnings
- Security scan clean
- Dependencies up to date

### Automation

- CI runs on all PRs
- Security scanning enabled
- Coverage tracking active
- Linux testing (Docker deployment)

That's it. Keep it simple, test thoroughly, submit clean PRs.

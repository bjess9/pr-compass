# Testing Guide for PR Pilot

This document explains how to run and write tests for PR Pilot without requiring external dependencies like GitHub OAuth.

## Overview

PR Pilot uses a comprehensive testing strategy that includes:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions with mocked dependencies  
- **UI Behavior Tests**: Test TUI interactions without external dependencies
- **Mock Data Tests**: Validate test data integrity

## Running Tests

### Quick Start

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests  
make test-integration

# Run tests with coverage report
make test-coverage
```

### Manual Test Commands

```bash
# Unit tests
go test -v ./internal/...

# Integration tests with custom runner
go run test_runner.go

# Specific package tests
go test -v ./internal/config
go test -v ./internal/github
```

## Test Architecture

### 1. Mock GitHub Client (`internal/github/mock.go`)

The mock client simulates GitHub API responses without making real API calls:

```go
client := github.NewMockClient()

// Test different configuration modes
prs, err := client.FetchPRsFromConfig(config)

// Simulate API errors
client.SetError(errors.New("rate limit exceeded"))

// Add custom test data
client.AddPR(customPR)
```

**Features:**
- Realistic test PR data with proper GitHub API structure
- Support for all configuration modes (repos, organization, teams, topics, search)
- Error simulation capabilities
- Customizable test data

### 2. Configuration Tests (`internal/config/config_test.go`)

Tests configuration loading and mode detection:

```go
func TestLoadConfig(t *testing.T) {
    // Create temporary config file
    tempDir := t.TempDir()
    configPath := filepath.Join(tempDir, "test_config.yaml")
    
    // Test config loading
    cfg, err := LoadConfig()
    // Verify config values...
}
```

### 3. UI Model Tests (`internal/model_test.go`)

Tests TUI behavior using Bubble Tea's testing patterns:

```go
func TestModelKeyboardNavigation(t *testing.T) {
    model := InitialModel("test-token")
    
    // Simulate key presses
    keyMsg := tea.KeyMsg{
        Type:  tea.KeyRunes,
        Runes: []rune("h"), // Help key
    }
    
    updatedModel, _ := model.Update(keyMsg)
    // Verify UI state changes...
}
```

### 4. Utility Tests (`internal/utils_test.go`)

Tests table creation, formatting, and helper functions:

```go
func TestCreateTableRows(t *testing.T) {
    testPRs := github.NewMockClient().PRs
    rows := createTableRows(testPRs)
    
    // Verify table structure
    if len(rows) != len(testPRs) {
        t.Errorf("Expected %d rows, got %d", len(testPRs), len(rows))
    }
}
```

## Testing Without External Dependencies

### No GitHub API Calls

All tests use the mock GitHub client:

- **Mock data** simulates real GitHub API responses
- **No authentication** required
- **No rate limiting** concerns
- **Consistent test data** across runs

### No OAuth Required

Tests bypass the OAuth flow entirely:

```go
// Tests use a simple token string
model := InitialModel("test-token")

// Mock client doesn't validate tokens
client := github.NewMockClient()
```

### Isolated File System

Configuration tests use temporary directories:

```go
tempDir := t.TempDir() // Automatically cleaned up
configPath := filepath.Join(tempDir, "test_config.yaml")
```

## Writing New Tests

### Adding Unit Tests

1. Create `*_test.go` files alongside your source code
2. Use the `testing` package conventions
3. Leverage mock clients for external dependencies

```go
func TestNewFeature(t *testing.T) {
    client := github.NewMockClient()
    
    // Test your feature
    result, err := yourFunction(client)
    
    // Assert expected behavior
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Adding Integration Tests

Add tests to `test_runner.go` for complex scenarios:

```go
func testNewIntegrationScenario() bool {
    // Set up test environment
    // Run end-to-end scenario
    // Validate results
    return success
}
```

### Adding UI Tests

Test TUI interactions using Bubble Tea patterns:

```go
func TestNewUIBehavior(t *testing.T) {
    model := InitialModel("test-token")
    
    // Load test data
    testPRs := github.NewMockClient().PRs
    updatedModel, _ := model.Update(testPRs)
    
    // Simulate user interaction
    keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("your-key")}
    finalModel, _ := updatedModel.Update(keyMsg)
    
    // Verify behavior
    result := finalModel.(model)
    // Assert state changes...
}
```

## Test Data

### Mock PR Structure

The mock client generates realistic PR data:

```go
type MockPR struct {
    Number:    int
    Title:     string
    Author:    string
    Repository: string
    Draft:     bool
    Mergeable: bool
    CreatedAt: time.Time
    Labels:    []string
    Reviewers: []string
}
```

### Available Test PRs

The mock client includes diverse test scenarios:

1. **Ready PR**: Not draft, mergeable, with labels
2. **Draft PR**: Work in progress, different author
3. **PR with conflicts**: Mergeable = false
4. **PR with reviewers**: Has requested reviewers
5. **Labeled PR**: Various label types for testing

## Continuous Integration

### CI-Friendly Commands

```bash
# Fast tests for CI
make test-ci

# Generate coverage for CI
make test-coverage
```

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: make test-ci
      
      - name: Generate coverage
        run: make test-coverage
```

## Debugging Tests

### Verbose Output

```bash
# Detailed test output
go test -v ./internal/...

# Specific test function
go test -v -run TestSpecificFunction ./internal/config
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage
open coverage.html  # View in browser
```

### Custom Test Runner Debug

The test runner provides detailed output for debugging integration issues:

```bash
go run test_runner.go
# Shows step-by-step test execution
# Detailed error messages
# Mock data validation
```

## Best Practices

1. **Use mock clients** instead of real API calls
2. **Test error conditions** with `client.SetError()`
3. **Use temporary directories** for file operations
4. **Test both happy path and edge cases**
5. **Keep tests fast** - no external dependencies
6. **Use descriptive test names** that explain the scenario
7. **Clean up resources** with `t.TempDir()` and defer statements

This testing approach ensures reliable, fast tests that can run anywhere without external dependencies or authentication requirements.

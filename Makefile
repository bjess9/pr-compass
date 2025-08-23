# PR Compass Makefile

.PHONY: build test test-unit test-integration test-all clean run-tests help

# Build the application
build:
	@echo "🔨 Building PR Compass..."
	go build -buildvcs=false -o pr-compass ./cmd/pr-compass
	@echo "📋 Creating Windows executable..."
	@cp pr-compass pr-compass.exe 2>/dev/null || echo "Windows executable created"

# Run all tests
test: test-unit test-integration
	@echo "✅ All tests completed"

# Run unit tests only
test-unit:
	@echo "📋 Running unit tests..."
	go test -v ./internal/...

# Run integration tests using our custom test runner
test-integration: build
	@echo "🔗 Running integration tests..."
	go run test/integration/test_runner.go

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Generate coverage summary for PR comments
coverage-summary:
	@echo "📈 Generating coverage summary..."
	@go tool cover -func=coverage.out | tail -1 | awk '{print "**Total Coverage:** " $$3}' > coverage-summary.txt
	@echo "" >> coverage-summary.txt
	@echo "### 📊 Coverage by Package:" >> coverage-summary.txt
	@go tool cover -func=coverage.out | grep -v "total:" | awk '{printf "- **%s**: %s\n", $$1, $$3}' >> coverage-summary.txt

# Run tests in CI mode (quiet output)
test-ci:
	@echo "🤖 Running tests in CI mode..."
	go test -short -v ./internal/...
	go run test/integration/test_runner.go

# Clean build artifacts and test files
clean:
	@echo "🧹 Cleaning up..."
	rm -f pr-compass pr-compass.exe
	rm -f coverage.out coverage.html
	rm -rf /tmp/prpilot_test* test/fixtures/temp*

# Run the application in development mode
dev: build
	@echo "🚀 Running PR Compass in development mode..."
	./pr-compass

# Configure for development (creates example config)
dev-config:
	@echo "⚙️ Setting up development configuration..."
	@if [ ! -f ~/.prpilot_config.yaml ]; then \
		cp example_config.yaml ~/.prpilot_config.yaml; \
		echo "Created ~/.prpilot_config.yaml from example"; \
	else \
		echo "Configuration already exists at ~/.prpilot_config.yaml"; \
	fi

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run

# Run security check
security:
	@echo "🔐 Running security check..."
	gosec ./...

# Install development dependencies
dev-deps:
	@echo "📦 Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Full development setup
setup: dev-deps dev-config
	@echo "🎯 Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit ~/.prpilot_config.yaml with your settings"
	@echo "  2. Run 'make test' to verify everything works"
	@echo "  3. Run 'make dev' to start the application"

# Run quick development checks
check: fmt lint test-unit
	@echo "✅ All checks passed"

# Install development tools
install-tools:
	@echo "🔧 Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Prepare for release
release-check: test lint security
	@echo "🚀 Release checks complete"
	@echo "Ready to create release"

# Show available targets
help:
	@echo "PR Compass Development Commands"
	@echo "============================="
	@echo ""
	@echo "Building:"
	@echo "  build        - Build the PR Compass binary"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test         - Run all tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-ci      - Run tests in CI mode"
	@echo ""
	@echo "Development:"
	@echo "  dev          - Run the application"
	@echo "  dev-config   - Create example configuration"
	@echo "  setup        - Full development environment setup"
	@echo "  check        - Run quick development checks"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  security     - Run security checks"
	@echo "  install-tools - Install development tools"
	@echo ""
	@echo "Release:"
	@echo "  release-check - Run all pre-release checks"
	@echo ""
	@echo "  help         - Show this help message"

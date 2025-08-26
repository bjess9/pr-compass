package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Authentication errors
var (
	ErrAuthTokenInvalid     = errors.New("GitHub token format is invalid - check that your token starts with 'ghp_', 'gho_', 'ghu_', or 'ghs_' and is complete")
	ErrAuthTokenMissing     = errors.New("no GitHub token found - set GITHUB_TOKEN environment variable or use 'gh auth login'")
	ErrAuthStorageFailed    = errors.New("failed to store authentication token - check file permissions in your home directory")
	ErrAuthPermissionDenied = errors.New("GitHub API permission denied - ensure your token has 'repo' and 'read:org' scopes")
)

// Configuration errors
func NewConfigNotFoundError(configPath string) error {
	return fmt.Errorf("configuration file not found: %s - create it or copy example_config.yaml to get started", configPath)
}

func NewConfigInvalidError(cause error) error {
	return fmt.Errorf("configuration file has invalid syntax or structure - check your YAML syntax and compare with example_config.yaml: %w", cause)
}

func NewConfigModeInvalidError(mode string) error {
	return fmt.Errorf("configuration mode '%s' is not supported - use one of: 'repos', 'organization', 'teams', 'search', or 'topics'", mode)
}

// GitHub API errors
func NewGitHubRateLimitError(resetTime string, cause error) error {
	msg := "GitHub API rate limit exceeded - wait for the rate limit to reset"
	if resetTime != "" {
		msg = fmt.Sprintf("GitHub API rate limit exceeded - wait until %s for rate limit reset, or use a different token", resetTime)
	}
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

func NewGitHubNetworkError(cause error) error {
	return fmt.Errorf("unable to connect to GitHub - check your internet connection and try again: %w", cause)
}

func NewGitHubNotFoundError(resource string, cause error) error {
	msg := fmt.Sprintf("GitHub resource not found: %s - check that the repository/organization name is correct and you have access to it", resource)
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

func NewGitHubForbiddenError(resource string, cause error) error {
	msg := fmt.Sprintf("access denied to %s - ensure your token has sufficient permissions and you're a member of the organization/team", resource)
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

func NewGitHubUnknownError(statusCode int, cause error) error {
	msg := fmt.Sprintf("unexpected GitHub API error (HTTP %d) - please try again later", statusCode)
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

// Context errors
func NewTimeoutError(operation string, cause error) error {
	msg := fmt.Sprintf("operation timed out: %s - try again with a more specific query or check your network connection", operation)
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

func NewCancelledError(operation string, cause error) error {
	msg := fmt.Sprintf("operation cancelled: %s", operation)
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

// Repository errors
func NewRepositoryInvalidError(repo string, cause error) error {
	msg := fmt.Sprintf("invalid repository format: %s - repository names should be in 'owner/repo' format (e.g., 'microsoft/vscode')", repo)
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

func NewOrganizationNotFoundError(org string, cause error) error {
	msg := fmt.Sprintf("organization not found: %s - check the organization name and ensure you have access to it", org)
	if cause != nil {
		return fmt.Errorf("%s: %w", msg, cause)
	}
	return errors.New(msg)
}

// Helper function to convert HTTP status codes to appropriate GitHub errors
func NewGitHubErrorFromHTTPStatus(statusCode int, resource string, cause error) error {
	switch statusCode {
	case http.StatusNotFound:
		return NewGitHubNotFoundError(resource, cause)
	case http.StatusForbidden:
		return NewGitHubForbiddenError(resource, cause)
	case http.StatusUnauthorized:
		return fmt.Errorf("GitHub API permission denied - ensure your token has 'repo' and 'read:org' scopes: %w", cause)
	case http.StatusTooManyRequests:
		return NewGitHubRateLimitError("", cause)
	default:
		return NewGitHubUnknownError(statusCode, cause)
	}
}

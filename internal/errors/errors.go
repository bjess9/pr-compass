package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error that occurred
type ErrorType string

const (
	// Authentication error types
	ErrorTypeAuthTokenInvalid     ErrorType = "auth_token_invalid"
	ErrorTypeAuthTokenMissing     ErrorType = "auth_token_missing"
	ErrorTypeAuthStorageFailed    ErrorType = "auth_storage_failed"
	ErrorTypeAuthPermissionDenied ErrorType = "auth_permission_denied"

	// Configuration error types
	ErrorTypeConfigNotFound    ErrorType = "config_not_found"
	ErrorTypeConfigInvalid     ErrorType = "config_invalid"
	ErrorTypeConfigModeInvalid ErrorType = "config_mode_invalid"

	// GitHub API error types
	ErrorTypeGitHubRateLimit    ErrorType = "github_rate_limit"
	ErrorTypeGitHubNetworkError ErrorType = "github_network_error"
	ErrorTypeGitHubNotFound     ErrorType = "github_not_found"
	ErrorTypeGitHubForbidden    ErrorType = "github_forbidden"
	ErrorTypeGitHubUnknown      ErrorType = "github_unknown"

	// Context error types
	ErrorTypeTimeout   ErrorType = "timeout"
	ErrorTypeCancelled ErrorType = "cancelled"

	// Repository/Resource error types
	ErrorTypeRepositoryInvalid    ErrorType = "repository_invalid"
	ErrorTypeOrganizationNotFound ErrorType = "organization_not_found"
)

// PRPilotError represents a domain-specific error in the PR Pilot application
type PRPilotError struct {
	Type        ErrorType
	Message     string
	UserMessage string // User-friendly message
	Suggestion  string // Suggestion for user action
	Cause       error  // Underlying error
}

// Error implements the error interface
func (e *PRPilotError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// UserFriendlyError returns a user-friendly error message with suggestions
func (e *PRPilotError) UserFriendlyError() string {
	msg := e.UserMessage
	if msg == "" {
		msg = e.Message
	}

	if e.Suggestion != "" {
		return fmt.Sprintf("%s\n\nSuggestion: %s", msg, e.Suggestion)
	}
	return msg
}

// Unwrap returns the underlying cause error for error wrapping compatibility
func (e *PRPilotError) Unwrap() error {
	return e.Cause
}

// IsType checks if the error is of a specific type
func (e *PRPilotError) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// Authentication error constructors
func NewAuthTokenInvalidError(cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeAuthTokenInvalid,
		Message:     "GitHub token format is invalid",
		UserMessage: "Your GitHub token appears to be invalid or malformed",
		Suggestion:  "Please check that your token starts with 'ghp_', 'gho_', 'ghu_', or 'ghs_' and is complete",
		Cause:       cause,
	}
}

func NewAuthTokenMissingError() *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeAuthTokenMissing,
		Message:     "No GitHub token found",
		UserMessage: "No GitHub authentication token found",
		Suggestion:  "Set GITHUB_TOKEN environment variable or use 'gh auth login' to authenticate with GitHub CLI",
		Cause:       nil,
	}
}

func NewAuthStorageError(cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeAuthStorageFailed,
		Message:     "Failed to store authentication token",
		UserMessage: "Unable to save your GitHub token for future use",
		Suggestion:  "You may need to re-enter your token next time. Check file permissions in your home directory",
		Cause:       cause,
	}
}

func NewAuthPermissionDeniedError(cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeAuthPermissionDenied,
		Message:     "GitHub API permission denied",
		UserMessage: "Access denied by GitHub API - insufficient permissions",
		Suggestion:  "Ensure your token has 'repo' and 'read:org' scopes, and you have access to the requested resources",
		Cause:       cause,
	}
}

// Configuration error constructors
func NewConfigNotFoundError(configPath string, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeConfigNotFound,
		Message:     fmt.Sprintf("Configuration file not found: %s", configPath),
		UserMessage: "Configuration file not found",
		Suggestion:  fmt.Sprintf("Create %s or copy example_config.yaml to get started", configPath),
		Cause:       cause,
	}
}

func NewConfigInvalidError(cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeConfigInvalid,
		Message:     "Configuration file is invalid",
		UserMessage: "Your configuration file has invalid syntax or structure",
		Suggestion:  "Check your YAML syntax and compare with example_config.yaml",
		Cause:       cause,
	}
}

func NewConfigModeInvalidError(mode string) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeConfigModeInvalid,
		Message:     fmt.Sprintf("Invalid configuration mode: %s", mode),
		UserMessage: fmt.Sprintf("Configuration mode '%s' is not supported", mode),
		Suggestion:  "Use one of: 'repos', 'organization', 'teams', 'search', or 'topics'",
		Cause:       nil,
	}
}

// GitHub API error constructors
func NewGitHubRateLimitError(resetTime string, cause error) *PRPilotError {
	suggestion := "Wait for the rate limit to reset"
	if resetTime != "" {
		suggestion = fmt.Sprintf("Wait until %s for rate limit reset, or use a different token", resetTime)
	}

	return &PRPilotError{
		Type:        ErrorTypeGitHubRateLimit,
		Message:     "GitHub API rate limit exceeded",
		UserMessage: "You've hit GitHub's API rate limit",
		Suggestion:  suggestion,
		Cause:       cause,
	}
}

func NewGitHubNetworkError(cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeGitHubNetworkError,
		Message:     "Network error connecting to GitHub API",
		UserMessage: "Unable to connect to GitHub",
		Suggestion:  "Check your internet connection and try again. GitHub might be temporarily unavailable",
		Cause:       cause,
	}
}

func NewGitHubNotFoundError(resource string, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeGitHubNotFound,
		Message:     fmt.Sprintf("GitHub resource not found: %s", resource),
		UserMessage: fmt.Sprintf("Could not find %s on GitHub", resource),
		Suggestion:  "Check that the repository/organization name is correct and you have access to it",
		Cause:       cause,
	}
}

func NewGitHubForbiddenError(resource string, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeGitHubForbidden,
		Message:     fmt.Sprintf("Access forbidden to GitHub resource: %s", resource),
		UserMessage: fmt.Sprintf("Access denied to %s", resource),
		Suggestion:  "Ensure your token has sufficient permissions and you're a member of the organization/team",
		Cause:       cause,
	}
}

func NewGitHubUnknownError(statusCode int, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeGitHubUnknown,
		Message:     fmt.Sprintf("Unknown GitHub API error (HTTP %d)", statusCode),
		UserMessage: "An unexpected error occurred with the GitHub API",
		Suggestion:  "Please try again later. If the problem persists, check GitHub's status page",
		Cause:       cause,
	}
}

// Context error constructors
func NewTimeoutError(operation string, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeTimeout,
		Message:     fmt.Sprintf("Operation timed out: %s", operation),
		UserMessage: fmt.Sprintf("The %s operation took too long and was cancelled", operation),
		Suggestion:  "Try again with a more specific query or check your network connection",
		Cause:       cause,
	}
}

func NewCancelledError(operation string, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeCancelled,
		Message:     fmt.Sprintf("Operation cancelled: %s", operation),
		UserMessage: "The operation was cancelled",
		Suggestion:  "The operation was stopped, likely due to app shutdown or user interruption",
		Cause:       cause,
	}
}

// Repository error constructors
func NewRepositoryInvalidError(repo string, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeRepositoryInvalid,
		Message:     fmt.Sprintf("Invalid repository format: %s", repo),
		UserMessage: fmt.Sprintf("Repository '%s' has invalid format", repo),
		Suggestion:  "Repository names should be in 'owner/repo' format (e.g., 'microsoft/vscode')",
		Cause:       cause,
	}
}

func NewOrganizationNotFoundError(org string, cause error) *PRPilotError {
	return &PRPilotError{
		Type:        ErrorTypeOrganizationNotFound,
		Message:     fmt.Sprintf("Organization not found: %s", org),
		UserMessage: fmt.Sprintf("Organization '%s' could not be found or accessed", org),
		Suggestion:  "Check the organization name and ensure you have access to it",
		Cause:       cause,
	}
}

// Helper function to convert HTTP status codes to appropriate GitHub errors
func NewGitHubErrorFromHTTPStatus(statusCode int, resource string, cause error) *PRPilotError {
	switch statusCode {
	case http.StatusNotFound:
		return NewGitHubNotFoundError(resource, cause)
	case http.StatusForbidden:
		return NewGitHubForbiddenError(resource, cause)
	case http.StatusUnauthorized:
		return NewAuthPermissionDeniedError(cause)
	case http.StatusTooManyRequests:
		return NewGitHubRateLimitError("", cause)
	default:
		return NewGitHubUnknownError(statusCode, cause)
	}
}

// IsPRPilotError checks if an error is a PRPilotError
func IsPRPilotError(err error) (*PRPilotError, bool) {
	if prErr, ok := err.(*PRPilotError); ok {
		return prErr, true
	}
	return nil, false
}

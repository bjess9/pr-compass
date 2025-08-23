package errors

import (
	"errors"
	"net/http"
	"testing"
)

// TestPRPilotError_Error tests the Error() method
func TestPRPilotError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *PRPilotError
		expected string
	}{
		{
			name: "error without cause",
			err: &PRPilotError{
				Type:    ErrorTypeAuthTokenInvalid,
				Message: "Token is invalid",
			},
			expected: "auth_token_invalid: Token is invalid",
		},
		{
			name: "error with cause",
			err: &PRPilotError{
				Type:    ErrorTypeGitHubNetworkError,
				Message: "Network failure",
				Cause:   errors.New("connection timeout"),
			},
			expected: "github_network_error: Network failure (caused by: connection timeout)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestPRPilotError_UserFriendlyError tests the UserFriendlyError() method
func TestPRPilotError_UserFriendlyError(t *testing.T) {
	tests := []struct {
		name     string
		err      *PRPilotError
		expected string
	}{
		{
			name: "uses UserMessage when provided",
			err: &PRPilotError{
				Message:     "Technical error message",
				UserMessage: "User-friendly message",
			},
			expected: "User-friendly message",
		},
		{
			name: "falls back to Message when UserMessage is empty",
			err: &PRPilotError{
				Message:     "Technical error message",
				UserMessage: "",
			},
			expected: "Technical error message",
		},
		{
			name: "includes suggestion when provided",
			err: &PRPilotError{
				Message:     "Error occurred",
				UserMessage: "Something went wrong",
				Suggestion:  "Try doing this",
			},
			expected: "Something went wrong\n\nSuggestion: Try doing this",
		},
		{
			name: "no suggestion provided",
			err: &PRPilotError{
				Message:     "Error occurred",
				UserMessage: "Something went wrong",
				Suggestion:  "",
			},
			expected: "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.UserFriendlyError()
			if result != tt.expected {
				t.Errorf("UserFriendlyError() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestPRPilotError_Unwrap tests the Unwrap() method
func TestPRPilotError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		err      *PRPilotError
		expected error
	}{
		{
			name: "returns cause when present",
			err: &PRPilotError{
				Type:  ErrorTypeTimeout,
				Cause: originalErr,
			},
			expected: originalErr,
		},
		{
			name: "returns nil when no cause",
			err: &PRPilotError{
				Type:  ErrorTypeTimeout,
				Cause: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Unwrap()
			if result != tt.expected {
				t.Errorf("Unwrap() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestPRPilotError_IsType tests the IsType() method
func TestPRPilotError_IsType(t *testing.T) {
	err := &PRPilotError{
		Type: ErrorTypeAuthTokenInvalid,
	}

	tests := []struct {
		name      string
		checkType ErrorType
		expected  bool
	}{
		{
			name:      "matches correct type",
			checkType: ErrorTypeAuthTokenInvalid,
			expected:  true,
		},
		{
			name:      "does not match different type",
			checkType: ErrorTypeAuthTokenMissing,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := err.IsType(tt.checkType)
			if result != tt.expected {
				t.Errorf("IsType(%s) = %v, expected %v", tt.checkType, result, tt.expected)
			}
		})
	}
}

// TestAuthErrorConstructors tests all authentication error constructors
func TestAuthErrorConstructors(t *testing.T) {
	cause := errors.New("underlying cause")

	t.Run("NewAuthTokenInvalidError", func(t *testing.T) {
		err := NewAuthTokenInvalidError(cause)
		if err.Type != ErrorTypeAuthTokenInvalid {
			t.Errorf("Expected type %s, got %s", ErrorTypeAuthTokenInvalid, err.Type)
		}
		if err.Cause != cause {
			t.Errorf("Expected cause to be set")
		}
		if err.UserMessage == "" {
			t.Error("UserMessage should not be empty")
		}
		if err.Suggestion == "" {
			t.Error("Suggestion should not be empty")
		}
	})

	t.Run("NewAuthTokenMissingError", func(t *testing.T) {
		err := NewAuthTokenMissingError()
		if err.Type != ErrorTypeAuthTokenMissing {
			t.Errorf("Expected type %s, got %s", ErrorTypeAuthTokenMissing, err.Type)
		}
		if err.UserMessage == "" {
			t.Error("UserMessage should not be empty")
		}
		if err.Suggestion == "" {
			t.Error("Suggestion should not be empty")
		}
	})

	t.Run("NewAuthStorageError", func(t *testing.T) {
		err := NewAuthStorageError(cause)
		if err.Type != ErrorTypeAuthStorageFailed {
			t.Errorf("Expected type %s, got %s", ErrorTypeAuthStorageFailed, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewAuthPermissionDeniedError", func(t *testing.T) {
		err := NewAuthPermissionDeniedError(cause)
		if err.Type != ErrorTypeAuthPermissionDenied {
			t.Errorf("Expected type %s, got %s", ErrorTypeAuthPermissionDenied, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})
}

// TestConfigErrorConstructors tests all configuration error constructors
func TestConfigErrorConstructors(t *testing.T) {
	cause := errors.New("underlying cause")

	t.Run("NewConfigNotFoundError", func(t *testing.T) {
		configPath := "/path/to/config.yaml"
		err := NewConfigNotFoundError(configPath, cause)
		if err.Type != ErrorTypeConfigNotFound {
			t.Errorf("Expected type %s, got %s", ErrorTypeConfigNotFound, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
		if err.Message == "" {
			t.Error("Message should contain config path")
		}
	})

	t.Run("NewConfigInvalidError", func(t *testing.T) {
		err := NewConfigInvalidError(cause)
		if err.Type != ErrorTypeConfigInvalid {
			t.Errorf("Expected type %s, got %s", ErrorTypeConfigInvalid, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewConfigModeInvalidError", func(t *testing.T) {
		mode := "invalid_mode"
		err := NewConfigModeInvalidError(mode)
		if err.Type != ErrorTypeConfigModeInvalid {
			t.Errorf("Expected type %s, got %s", ErrorTypeConfigModeInvalid, err.Type)
		}
		if err.Message == "" {
			t.Error("Message should contain invalid mode")
		}
	})
}

// TestGitHubErrorConstructors tests all GitHub error constructors
func TestGitHubErrorConstructors(t *testing.T) {
	cause := errors.New("underlying cause")

	t.Run("NewGitHubRateLimitError", func(t *testing.T) {
		resetTime := "2023-12-01T12:00:00Z"
		err := NewGitHubRateLimitError(resetTime, cause)
		if err.Type != ErrorTypeGitHubRateLimit {
			t.Errorf("Expected type %s, got %s", ErrorTypeGitHubRateLimit, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewGitHubNetworkError", func(t *testing.T) {
		err := NewGitHubNetworkError(cause)
		if err.Type != ErrorTypeGitHubNetworkError {
			t.Errorf("Expected type %s, got %s", ErrorTypeGitHubNetworkError, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewGitHubNotFoundError", func(t *testing.T) {
		resource := "repository owner/repo"
		err := NewGitHubNotFoundError(resource, cause)
		if err.Type != ErrorTypeGitHubNotFound {
			t.Errorf("Expected type %s, got %s", ErrorTypeGitHubNotFound, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewGitHubForbiddenError", func(t *testing.T) {
		resource := "private repository"
		err := NewGitHubForbiddenError(resource, cause)
		if err.Type != ErrorTypeGitHubForbidden {
			t.Errorf("Expected type %s, got %s", ErrorTypeGitHubForbidden, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewGitHubUnknownError", func(t *testing.T) {
		statusCode := 500
		err := NewGitHubUnknownError(statusCode, cause)
		if err.Type != ErrorTypeGitHubUnknown {
			t.Errorf("Expected type %s, got %s", ErrorTypeGitHubUnknown, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})
}

// TestContextErrorConstructors tests context error constructors
func TestContextErrorConstructors(t *testing.T) {
	cause := errors.New("underlying cause")

	t.Run("NewTimeoutError", func(t *testing.T) {
		operation := "fetch PRs"
		err := NewTimeoutError(operation, cause)
		if err.Type != ErrorTypeTimeout {
			t.Errorf("Expected type %s, got %s", ErrorTypeTimeout, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewCancelledError", func(t *testing.T) {
		operation := "fetch PRs"
		err := NewCancelledError(operation, cause)
		if err.Type != ErrorTypeCancelled {
			t.Errorf("Expected type %s, got %s", ErrorTypeCancelled, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})
}

// TestResourceErrorConstructors tests resource error constructors
func TestResourceErrorConstructors(t *testing.T) {
	cause := errors.New("underlying cause")

	t.Run("NewRepositoryInvalidError", func(t *testing.T) {
		repo := "invalid-repo-format"
		err := NewRepositoryInvalidError(repo, cause)
		if err.Type != ErrorTypeRepositoryInvalid {
			t.Errorf("Expected type %s, got %s", ErrorTypeRepositoryInvalid, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})

	t.Run("NewOrganizationNotFoundError", func(t *testing.T) {
		org := "nonexistent-org"
		err := NewOrganizationNotFoundError(org, cause)
		if err.Type != ErrorTypeOrganizationNotFound {
			t.Errorf("Expected type %s, got %s", ErrorTypeOrganizationNotFound, err.Type)
		}
		if err.Cause != cause {
			t.Error("Cause should be set")
		}
	})
}

// TestNewGitHubErrorFromHTTPStatus tests the HTTP status code to error mapping
func TestNewGitHubErrorFromHTTPStatus(t *testing.T) {
	cause := errors.New("http error")
	resource := "test resource"

	tests := []struct {
		name         string
		statusCode   int
		expectedType ErrorType
	}{
		{
			name:         "404 Not Found",
			statusCode:   http.StatusNotFound,
			expectedType: ErrorTypeGitHubNotFound,
		},
		{
			name:         "403 Forbidden",
			statusCode:   http.StatusForbidden,
			expectedType: ErrorTypeGitHubForbidden,
		},
		{
			name:         "429 Too Many Requests",
			statusCode:   http.StatusTooManyRequests,
			expectedType: ErrorTypeGitHubRateLimit,
		},
		{
			name:         "500 Internal Server Error",
			statusCode:   http.StatusInternalServerError,
			expectedType: ErrorTypeGitHubUnknown,
		},
		{
			name:         "Unknown status code",
			statusCode:   418, // I'm a teapot
			expectedType: ErrorTypeGitHubUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewGitHubErrorFromHTTPStatus(tt.statusCode, resource, cause)
			if err.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, err.Type)
			}
			if err.Cause != cause {
				t.Error("Cause should be set")
			}
		})
	}
}

// TestIsPRPilotError tests the error type checking utility function
func TestIsPRPilotError(t *testing.T) {
	prPilotErr := &PRPilotError{
		Type:    ErrorTypeAuthTokenInvalid,
		Message: "Test error",
	}
	standardErr := errors.New("standard error")

	tests := []struct {
		name        string
		err         error
		expectedErr *PRPilotError
		expectedOk  bool
	}{
		{
			name:        "PRPilotError returns correct error and true",
			err:         prPilotErr,
			expectedErr: prPilotErr,
			expectedOk:  true,
		},
		{
			name:        "standard error returns nil and false",
			err:         standardErr,
			expectedErr: nil,
			expectedOk:  false,
		},
		{
			name:        "nil error returns nil and false",
			err:         nil,
			expectedErr: nil,
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, ok := IsPRPilotError(tt.err)
			if err != tt.expectedErr {
				t.Errorf("IsPRPilotError() err = %v, expected %v", err, tt.expectedErr)
			}
			if ok != tt.expectedOk {
				t.Errorf("IsPRPilotError() ok = %v, expected %v", ok, tt.expectedOk)
			}
		})
	}
}

// TestErrorConstants tests that all error type constants are properly defined
func TestErrorConstants(t *testing.T) {
	// Verify that all error types are non-empty strings
	errorTypes := []ErrorType{
		ErrorTypeAuthTokenInvalid,
		ErrorTypeAuthTokenMissing,
		ErrorTypeAuthStorageFailed,
		ErrorTypeAuthPermissionDenied,
		ErrorTypeConfigNotFound,
		ErrorTypeConfigInvalid,
		ErrorTypeConfigModeInvalid,
		ErrorTypeGitHubRateLimit,
		ErrorTypeGitHubNetworkError,
		ErrorTypeGitHubNotFound,
		ErrorTypeGitHubForbidden,
		ErrorTypeGitHubUnknown,
		ErrorTypeTimeout,
		ErrorTypeCancelled,
		ErrorTypeRepositoryInvalid,
		ErrorTypeOrganizationNotFound,
	}

	for _, errorType := range errorTypes {
		if string(errorType) == "" {
			t.Errorf("Error type %v should not be empty", errorType)
		}
	}
}

// TestErrorIntegration tests that errors work properly with Go's error handling patterns
func TestErrorIntegration(t *testing.T) {
	originalCause := errors.New("original problem")
	prErr := NewAuthTokenInvalidError(originalCause)

	// Test that it satisfies the error interface
	var err error = prErr
	if err.Error() == "" {
		t.Error("Error() should return non-empty string")
	}

	// Test error unwrapping
	if !errors.Is(prErr, originalCause) {
		t.Error("errors.Is should find the wrapped error")
	}

	// Test error type checking with errors.As
	var targetErr *PRPilotError
	if !errors.As(prErr, &targetErr) {
		t.Error("errors.As should work with PRPilotError")
	}
	if targetErr.Type != ErrorTypeAuthTokenInvalid {
		t.Error("Type should be preserved through errors.As")
	}
}

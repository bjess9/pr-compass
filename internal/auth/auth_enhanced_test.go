package auth

import (
	"os"
	"testing"

	"github.com/bjess9/pr-compass/internal/errors"
)

// TestAuthenticate tests the main authentication function
func TestAuthenticate(t *testing.T) {
	// Save original env var to restore later
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	t.Run("with valid GITHUB_TOKEN env var", func(t *testing.T) {
		// Set a valid test token
		testToken := "ghp_1234567890123456789012345678901234567890"
		os.Setenv("GITHUB_TOKEN", testToken)

		token, err := Authenticate()
		if err != nil {
			t.Errorf("Expected no error with valid token, got: %v", err)
		}

		if token != testToken {
			t.Errorf("Expected token %s, got %s", testToken, maskToken(token))
		}
	})

	t.Run("with invalid GITHUB_TOKEN env var", func(t *testing.T) {
		// Set an invalid test token (too short)
		os.Setenv("GITHUB_TOKEN", "invalid")

		_, err := Authenticate()
		if err == nil {
			t.Error("Expected error with invalid token, got none")
		}

		if err != errors.ErrAuthTokenMissing {
			t.Errorf("Expected ErrAuthTokenMissing, got: %v", err)
		}
	})

	t.Run("without GITHUB_TOKEN and no gh CLI", func(t *testing.T) {
		// Unset the environment variable
		os.Unsetenv("GITHUB_TOKEN")

		_, err := Authenticate()
		if err == nil {
			// This might succeed if the user actually has gh CLI set up
			t.Log("No error - user may have valid gh CLI token")
		} else {
			if err != errors.ErrAuthTokenMissing {
				t.Errorf("Expected ErrAuthTokenMissing, got: %v", err)
			}
		}
	})
}

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"valid ghp token", "ghp_1234567890123456789012345678901234567890", true},
		{"valid gho token", "gho_1234567890123456789012345678901234567890", true},
		{"valid ghu token", "ghu_1234567890123456789012345678901234567890", true},
		{"valid ghs token", "ghs_1234567890123456789012345678901234567890", true},
		{"too short", "ghp_123", false},
		{"wrong prefix", "abc_1234567890123456789012345678901234567890", false},
		{"empty string", "", false},
		{"no prefix", "1234567890123456789012345678901234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateToken(tt.token)
			if result != tt.expected {
				t.Errorf("validateToken(%s) = %v, expected %v",
					maskToken(tt.token), result, tt.expected)
			}
		})
	}
}

// maskToken masks a token for safe logging/testing
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

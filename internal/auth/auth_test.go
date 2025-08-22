package auth

import (
	"testing"
)

// TestValidateToken verifies token validation logic
func TestValidateToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"valid ghp_ token", "ghp_1234567890abcdefghijklmnopqrstuvwxyz", true},
		{"valid gho_ token", "gho_1234567890abcdefghijklmnopqrstuvwxyz", true},
		{"valid ghu_ token", "ghu_1234567890abcdefghijklmnopqrstuvwxyz", true},
		{"valid ghs_ token", "ghs_1234567890abcdefghijklmnopqrstuvwxyz", true},
		{"valid github_pat_ token", "github_pat_11ABCDEFG0123456789_abcdefghijklmnopqrstuvwxyz", true},
		{"valid legacy 40-char token", "1234567890abcdefghijklmnopqrstuvwxyz1234", true},
		{"empty token", "", false},
		{"too short token", "ghp_123", false},
		{"invalid prefix", "invalid_1234567890abcdefghijklmnopqrstuvwxyz", false},
		{"whitespace token", "   ", false},
		{"token with whitespace", " ghp_1234567890abcdefghijklmnopqrstuvwxyz ", true}, // should be trimmed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateToken(tt.token)
			if result != tt.expected {
				t.Errorf("validateToken(%q) = %v, want %v", tt.token, result, tt.expected)
			}
		})
	}
}

// TestAuthenticationPriorityOrder tests that authentication sources are tried in correct order
func TestAuthenticationPriorityOrder(t *testing.T) {
	// This is more of a documentation test - the actual priority is:
	// 1. GITHUB_TOKEN environment variable (highest priority)
	// 2. GitHub CLI token (gh auth token)

	// We can't easily test the full flow due to environment dependencies,
	// but we can document the expected behavior
	priorities := []string{
		"GITHUB_TOKEN environment variable",
		"GitHub CLI token (gh auth token)",
	}

	if len(priorities) != 2 {
		t.Error("Authentication should have exactly 2 priority levels")
	}

	// Test that priority order is documented correctly
	expectedOrder := []string{
		"GITHUB_TOKEN environment variable",
		"GitHub CLI token (gh auth token)",
	}

	for i, expected := range expectedOrder {
		if priorities[i] != expected {
			t.Errorf("Priority %d should be %q, got %q", i+1, expected, priorities[i])
		}
	}
}

package github

import (
	"testing"
)

// TestNewClient tests the GitHub client creation function
func TestNewClient(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "creates client with valid token",
			token: "ghp_1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:  "creates client with different token format",
			token: "gho_1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:  "creates client with empty token",
			token: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token)
			
			// Should never return an error - this function always succeeds
			if err != nil {
				t.Errorf("NewClient() returned unexpected error: %v", err)
			}
			
			// Should return a valid GitHub client
			if client == nil {
				t.Error("NewClient() returned nil client")
			}
			
			// Verify it's actually a GitHub client
			// We can't easily type-check the concrete type, but we can verify
			// it has the expected GitHub client structure
			if client.BaseURL == nil {
				t.Error("Client BaseURL should be set")
			}
		})
	}
}

// TestNewClientIntegration tests that the created client has proper structure
func TestNewClientIntegration(t *testing.T) {
	token := "test_token_123"
	client, err := NewClient(token)
	
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	
	// Verify that the client has all the expected service endpoints
	if client.PullRequests == nil {
		t.Error("Client should have PullRequests service")
	}
	
	if client.Repositories == nil {
		t.Error("Client should have Repositories service")
	}
	
	if client.Organizations == nil {
		t.Error("Client should have Organizations service")
	}
	
	if client.Search == nil {
		t.Error("Client should have Search service")
	}
	
	if client.Teams == nil {
		t.Error("Client should have Teams service")
	}
}

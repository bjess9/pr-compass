package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Create test config content
	configContent := `mode: "topics"
topic_org: "testorg"
topics:
  - "backend"
  - "api"
repos: []
organization: ""
teams: []
search_query: ""`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load config from specific path
	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFromPath failed: %v", err)
	}

	// Verify config values
	if cfg.Mode != "topics" {
		t.Errorf("Expected mode 'topics', got '%s'", cfg.Mode)
	}

	if cfg.TopicOrg != "testorg" {
		t.Errorf("Expected topic_org 'testorg', got '%s'", cfg.TopicOrg)
	}

	if len(cfg.Topics) != 2 {
		t.Errorf("Expected 2 topics, got %d", len(cfg.Topics))
	}

	if len(cfg.Topics) >= 2 {
		expectedTopics := []string{"backend", "api"}
		for i, expected := range expectedTopics {
			if i < len(cfg.Topics) && cfg.Topics[i] != expected {
				t.Errorf("Expected topic[%d] '%s', got '%s'", i, expected, cfg.Topics[i])
			}
		}
	}
}

func TestAutoDetectMode(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		expected string
	}{
		{
			name: "repos mode",
			config: `
repos:
  - "owner/repo1"
  - "owner/repo2"
`,
			expected: "repos",
		},
		{
			name: "topics mode",
			config: `
topics:
  - "backend"
topic_org: "testorg"
`,
			expected: "topics",
		},
		{
			name: "organization mode",
			config: `
organization: "testorg"
`,
			expected: "organization",
		},
		{
			name: "teams mode",
			config: `
organization: "testorg"
teams:
  - "backend-team"
`,
			expected: "teams",
		},
		{
			name: "search mode",
			config: `
search_query: "org:testorg is:pr is:open"
`,
			expected: "search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "test_config.yaml")

			err := os.WriteFile(configPath, []byte(tt.config), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Load config from specific path
			cfg, err := LoadConfigFromPath(configPath)
			if err != nil {
				t.Fatalf("LoadConfigFromPath failed: %v", err)
			}

			if cfg.Mode != tt.expected {
				t.Errorf("Expected mode '%s', got '%s'", tt.expected, cfg.Mode)
			}
		})
	}
}

func TestConfigExists(t *testing.T) {
	// Test with non-existent config path
	_, err := LoadConfigFromPath("/non/existent/path/config.yaml")
	if err == nil {
		t.Error("LoadConfigFromPath should return error for non-existent config")
	}
	
	// Test with valid config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	configContent := `mode: "repos"`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Errorf("LoadConfigFromPath should succeed for valid config: %v", err)
		return
	}
	
	if cfg == nil {
		t.Error("Config should not be nil for valid config file")
		return
	}
	
	if cfg.Mode != "repos" {
		t.Errorf("Expected mode 'repos', got '%s'", cfg.Mode)
	}
}

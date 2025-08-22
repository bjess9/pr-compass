package config

import (
	"os"
	"path/filepath"
	"testing"
)

// User Story: As a team lead, I want to configure PR tracking by repository topics
// so I can see all PRs from repositories tagged with my team's topic
func TestUserCanConfigureTopicsMode(t *testing.T) {
	// Given a user wants to track PRs from repositories tagged "backend" and "api"
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "team_config.yaml")

	configContent := `mode: "topics"
topic_org: "mycompany"
topics:
  - "backend"  
  - "api"`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to set up user config: %v", err)
	}

	// When the user loads their configuration
	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("User cannot load their configuration: %v", err)
	}

	// Then the system should be ready to track PRs from topic-tagged repositories
	if cfg.Mode != "topics" {
		t.Errorf("System not configured for topics tracking. Expected 'topics', got '%s'", cfg.Mode)
	}

	if cfg.TopicOrg != "mycompany" {
		t.Errorf("System not configured for correct organization. Expected 'mycompany', got '%s'", cfg.TopicOrg)
	}

	if len(cfg.Topics) != 2 {
		t.Errorf("System not configured for correct number of topics. Expected 2, got %d", len(cfg.Topics))
	}
}

// User Story: As a user, I want the system to automatically detect my preferred tracking mode
// so I don't have to explicitly specify it when migrating from old configurations
func TestSystemAutoDetectsUserIntent(t *testing.T) {
	scenarios := []struct {
		userIntent   string
		config       string
		expectedMode string
	}{
		{
			userIntent: "user wants to track specific repositories",
			config: `repos:
  - "company/repo1"
  - "company/repo2"`,
			expectedMode: "repos",
		},
		{
			userIntent: "user wants to track repositories by topics",
			config: `topics:
  - "backend"
topic_org: "company"`,
			expectedMode: "topics",
		},
		{
			userIntent:   "user wants to track entire organization",
			config:       `organization: "company"`,
			expectedMode: "organization",
		},
		{
			userIntent: "user wants to track specific teams",
			config: `organization: "company"
teams:
  - "backend-team"`,
			expectedMode: "teams",
		},
		{
			userIntent:   "user wants custom search tracking",
			config:       `search_query: "org:company is:pr is:open"`,
			expectedMode: "search",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.userIntent, func(t *testing.T) {
			// Given a user has configured their preferences
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "user_config.yaml")

			err := os.WriteFile(configPath, []byte(scenario.config), 0644)
			if err != nil {
				t.Fatalf("Failed to set up user config: %v", err)
			}

			// When the system loads the configuration
			cfg, err := LoadConfigFromPath(configPath)
			if err != nil {
				t.Fatalf("System failed to understand user intent: %v", err)
			}

			// Then it should automatically detect the correct tracking mode
			if cfg.Mode != scenario.expectedMode {
				t.Errorf("System misunderstood user intent. Expected '%s' mode, got '%s'", scenario.expectedMode, cfg.Mode)
			}
		})
	}
}

// User Story: As a user, I want clear error messages when my configuration is invalid
// so I can quickly fix setup issues
func TestUserGetsHelpfulErrorsForInvalidConfiguration(t *testing.T) {
	// Given a user provides an invalid configuration path
	_, err := LoadConfigFromPath("/this/path/does/not/exist/config.yaml")

	// When the system tries to load it
	// Then it should provide a clear error message
	if err == nil {
		t.Error("System should tell user when configuration file doesn't exist")
	}

	// The error should be helpful for troubleshooting
	if err != nil && !containsAnyString(err.Error(), []string{"not found", "file", "config"}) {
		t.Errorf("Error message not helpful for user: %v", err)
	}
}

// User Story: As a user with a valid configuration, I want the system to load it successfully
// so I can start using the tool immediately
func TestUserCanSuccessfullyLoadValidConfiguration(t *testing.T) {
	// Given a user has created a valid configuration
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "valid_config.yaml")

	configContent := `mode: "repos"
repos:
  - "company/important-repo"`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid config: %v", err)
	}

	// When the user starts the application
	cfg, err := LoadConfigFromPath(configPath)

	// Then the system should load successfully and be ready to use
	if err != nil {
		t.Errorf("System failed to load user's valid configuration: %v", err)
	}

	if cfg == nil {
		t.Error("System should provide loaded configuration to user")
	}

	if cfg != nil && cfg.Mode != "repos" {
		t.Errorf("System loaded incorrect configuration mode. Expected 'repos', got '%s'", cfg.Mode)
	}
}

// Helper function to check if error contains helpful keywords
func containsAnyString(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(text) > 0 && len(keyword) > 0 {
			// Simple contains check (case insensitive would be better but this works for our test)
			for i := 0; i <= len(text)-len(keyword); i++ {
				if text[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

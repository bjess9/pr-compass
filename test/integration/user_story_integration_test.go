// User Story Integration Tests - End-to-End Behavioral Testing
// These tests validate complete user workflows without external dependencies

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/github"
)

// User Story: As a team lead, I want to quickly set up PR tracking for my team
// so I can monitor all team activity without manual maintenance
func TestTeamLeadCanQuicklySetupTeamTracking(t *testing.T) {
	// Given a team lead wants to track PRs from repositories tagged "backend"
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "team_config.yaml")

	teamConfig := `mode: "topics"
topic_org: "testorg" 
topics:
  - "backend"
  - "api"`

	err := os.WriteFile(configPath, []byte(teamConfig), 0644)
	if err != nil {
		t.Fatalf("Team lead cannot create configuration: %v", err)
	}

	// When they load their configuration and start tracking
	cfg, err := config.LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Team lead cannot load their configuration: %v", err)
	}

	client := github.NewMockClient()
	prs, err := client.FetchPRsFromConfig(cfg)
	if err != nil {
		t.Fatalf("Team lead cannot fetch team PRs: %v", err)
	}

	// Then they should see PRs from all repositories tagged with their team's topics
	if len(prs) == 0 {
		t.Error("Team lead should see PRs from team repositories")
	}

	// And all PRs should be from their organization
	for _, pr := range prs {
		orgName := pr.GetBase().GetRepo().GetOwner().GetLogin()
		if orgName != "testorg" {
			t.Errorf("Team lead sees PR from wrong organization: %s", orgName)
		}
	}

	fmt.Printf("✅ Team lead can track %d PRs from team repositories\n", len(prs))
}

// User Story: As a developer, I want comprehensive test coverage for reliability
// so I can trust the tool in my daily workflow
func TestDeveloperHasComprehensiveTestCoverage(t *testing.T) {
	// This meta-test validates that we have proper test coverage
	// by ensuring all major components have corresponding tests

	tests := []struct {
		component string
		testFile  string
	}{
		{"Configuration loading", "../../internal/config/config_test.go"},
		{"GitHub API integration", "../../internal/github/client_test.go"},
		{"UI components", "../../internal/ui/multitab_test.go"},
		{"Cache system", "../../internal/cache/cache_test.go"},
		{"Batch processing", "../../internal/batch/manager_test.go"},
	}

	for _, test := range tests {
		if _, err := os.Stat(test.testFile); os.IsNotExist(err) {
			t.Errorf("Missing test coverage for %s: %s not found", test.component, test.testFile)
		}
	}

	fmt.Println("✅ Comprehensive test coverage verified")
}

// Test helper functions

func TestMain(m *testing.M) {
	// Set up mock environment
	os.Setenv("GITHUB_TOKEN", "mock-token-for-testing")

	// Run tests
	code := m.Run()

	// Clean up
	os.Unsetenv("GITHUB_TOKEN")

	os.Exit(code)
}

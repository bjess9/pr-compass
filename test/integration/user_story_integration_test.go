// User Story Integration Tests - End-to-End Behavioral Testing
// These tests validate complete user workflows without external dependencies

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/github"
	"github.com/bjess9/pr-compass/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
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

// User Story: As a developer, I want a simple workflow to check PRs and open them in browser
// so I can quickly review and take action on relevant PRs
func TestDeveloperCanReviewAndOpenPRs(t *testing.T) {
	// Given a developer has PRs to review
	app := ui.InitialModel("dev-token")

	// When they start the application
	initialView := app.View()
	if !strings.Contains(initialView, "⏳ Loading") {
		t.Error("Developer should see loading indicator on startup")
	}

	// And their PRs load
	testPRs := github.NewMockClient().PRs
	prMsg := ui.PrsWithConfigMsg{Prs: testPRs, Cfg: nil}
	loadedApp, _ := app.Update(prMsg)
	loadedView := loadedApp.View()

	if strings.Contains(loadedView, "⏳ Loading") {
		t.Errorf("Developer should not see loading after PRs load. View content: %s", loadedView)
	}

	// When they navigate through PRs
	downKey := createKeyMsg("down")
	navigatedApp, _ := loadedApp.Update(downKey)

	// And request to open a PR
	enterKey := createKeyMsg("enter")
	_, cmd := navigatedApp.Update(enterKey)

	// Then the system should initiate opening the PR in browser
	if cmd == nil {
		t.Error("Developer should be able to open PR in browser")
	}

	fmt.Println("✅ Developer can navigate PRs and open in browser")
}

// User Story: As a user with many PRs, I want to filter and focus on specific types
// so I can prioritize my review time effectively
func TestUserCanFilterAndFocusOnSpecificPRTypes(t *testing.T) {
	// Given a user has many PRs of different types
	app := ui.InitialModel("user-token")
	testPRs := github.NewMockClient().PRs // Contains mix of drafts, ready PRs, etc.
	prMsg := ui.PrsWithConfigMsg{Prs: testPRs, Cfg: nil}
	loadedApp, _ := app.Update(prMsg)

	// When they filter to drafts only
	draftFilterKey := createKeyMsg("d")
	draftApp, _ := loadedApp.Update(draftFilterKey)
	draftView := draftApp.View()

	if !strings.Contains(draftView, "draft") && !strings.Contains(draftView, "Draft") && !strings.Contains(draftView, "Filter:") {
		t.Error("User should see draft filter is active")
	}

	// And then clear filters to see everything again
	clearKey := createKeyMsg("c")
	clearedApp, _ := draftApp.Update(clearKey)
	clearedView := clearedApp.View()

	if strings.Contains(clearedView, "Filtered by") {
		t.Error("User should not see filter indicators after clearing")
	}

	// When they access help to learn more features
	helpKey := createKeyMsg("h")
	helpApp, _ := clearedApp.Update(helpKey)
	helpView := helpApp.View()

	if !strings.Contains(helpView, "Commands") && !strings.Contains(helpView, "Column Guide") {
		t.Error("User should see help information")
	}

	fmt.Println("✅ User can filter PRs and access help")
}

// User Story: As a user encountering issues, I want clear error messages and recovery options
// so I can resolve problems and continue using the tool
func TestUserReceivesHelpfulErrorHandlingAndRecovery(t *testing.T) {
	// Given GitHub API is temporarily unavailable
	client := github.NewMockClient()
	client.SetError(fmt.Errorf("GitHub API rate limit exceeded - please try again in 30 minutes"))

	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"company/important-repo"},
	}

	// When user attempts to fetch PRs
	prs, err := client.FetchPRsFromConfig(cfg)

	// Then they should receive helpful error information
	if err == nil {
		t.Error("User should be informed about GitHub unavailability")
	}

	if len(prs) > 0 {
		t.Error("User should not receive stale data during errors")
	}

	if err != nil && !strings.Contains(err.Error(), "rate limit") {
		t.Error("User should understand the specific error (rate limit)")
	}

	// And when they try the UI with an error
	app := ui.InitialModel("user-token")
	errorMsg := createErrorMsg("GitHub API temporarily unavailable")
	errorApp, _ := app.Update(errorMsg)
	errorView := errorApp.View()

	if !strings.Contains(errorView, "❌ Error") && !strings.Contains(errorView, "Error") {
		t.Error("User should see clear error indication in UI")
	}

	if !strings.Contains(errorView, "q") {
		t.Error("User should know they can quit when there's an error")
	}

	fmt.Println("✅ User receives helpful error messages and recovery options")
}

// User Story: As a new user, I want to quickly validate that the tool works
// so I can confirm it's set up correctly before relying on it
func TestNewUserCanQuicklyValidateSetup(t *testing.T) {
	// Given a new user wants to test the tool

	// When they run the unit tests (simulate this by checking if mock data works)
	client := github.NewMockClient()
	if len(client.PRs) == 0 {
		t.Error("New user cannot validate setup - no test data available")
	}

	// Simulate successful test validation
	if len(client.PRs) < 3 {
		t.Error("New user should see realistic test data (at least 3 PRs)")
	}

	// Verify test data quality
	for i, pr := range client.PRs {
		if pr.GetNumber() == 0 || pr.GetTitle() == "" || pr.GetHTMLURL() == "" {
			t.Errorf("Test PR %d missing essential attributes for realistic testing", i)
		}
	}

	fmt.Printf("✅ New user can validate setup with %d realistic test PRs\n", len(client.PRs))
}

// User Story: As a developer extending the tool, I want comprehensive test coverage
// so I can make changes confidently without breaking existing functionality
func TestDeveloperHasComprehensiveTestCoverage(t *testing.T) {
	// Given a developer wants to extend the tool

	// When they check test coverage across different scenarios
	scenarios := []struct {
		name        string
		description string
		test        func() error
	}{
		{
			name:        "Configuration Loading",
			description: "Multiple config modes work correctly",
			test: func() error {
				modes := []string{"repos", "topics", "organization", "teams", "search"}
				for _, mode := range modes {
					tempDir, _ := os.MkdirTemp("", "test")
					defer os.RemoveAll(tempDir)

					configPath := filepath.Join(tempDir, "test.yaml")
					testConfig := fmt.Sprintf(`mode: "%s"`, mode)
					if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
						return fmt.Errorf("failed to write test config: %v", err)
					}

					cfg, err := config.LoadConfigFromPath(configPath)
					if err != nil {
						return fmt.Errorf("config loading failed for %s mode: %v", mode, err)
					}
					if cfg.Mode != mode {
						return fmt.Errorf("config mode detection failed: expected %s, got %s", mode, cfg.Mode)
					}
				}
				return nil
			},
		},
		{
			name:        "PR Fetching",
			description: "All fetching modes return appropriate results",
			test: func() error {
				client := github.NewMockClient()
				configs := []*config.Config{
					{Mode: "repos", Repos: []string{"testorg/api-service", "testorg/frontend"}},
					{Mode: "organization", Organization: "testorg"},
					{Mode: "topics", TopicOrg: "testorg", Topics: []string{"backend"}},
					{Mode: "search", SearchQuery: "org:testorg is:pr is:open"},
				}

				for _, cfg := range configs {
					prs, err := client.FetchPRsFromConfig(cfg)
					if err != nil {
						return fmt.Errorf("PR fetching failed for %s mode: %v", cfg.Mode, err)
					}
					if len(prs) == 0 {
						return fmt.Errorf("no PRs returned for %s mode", cfg.Mode)
					}
				}
				return nil
			},
		},
		{
			name:        "UI Interactions",
			description: "Key user interactions work correctly",
			test: func() error {
				app := ui.InitialModel("test-token")

				// Test key interactions
				keys := []string{"h", "r", "d", "c", "q"}
				for _, key := range keys {
					keyMsg := createKeyMsg(key)
					_, cmd := app.Update(keyMsg)

					// Some keys should generate commands, others just update UI
					if key == "q" && cmd == nil {
						return fmt.Errorf("quit key should generate command")
					}
				}
				return nil
			},
		},
	}

	// Then all scenarios should pass
	for _, scenario := range scenarios {
		if err := scenario.test(); err != nil {
			t.Errorf("%s failed: %v", scenario.name, err)
		} else {
			fmt.Printf("✅ %s: %s\n", scenario.name, scenario.description)
		}
	}
}

// Helper functions for creating test messages
func createKeyMsg(key string) interface{} {
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}

func createErrorMsg(message string) tea.Msg {
	// Return error directly - the UI handles error interface
	return fmt.Errorf("%s", message)
}

// Run this test file with: go test -v user_story_integration_test.go
func TestMain(m *testing.M) {
	fmt.Println("🎭 Running User Story Integration Tests")
	fmt.Println("======================================")
	code := m.Run()
	if code == 0 {
		fmt.Println("✅ All user stories validated successfully!")
	} else {
		fmt.Println("❌ Some user stories need attention")
	}
	os.Exit(code)
}

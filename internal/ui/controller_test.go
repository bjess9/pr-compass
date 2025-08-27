package ui

import (
	"testing"

	"github.com/bjess9/pr-compass/internal/ui/services"
	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

// createTestRegistry creates a test service registry
func createTestRegistry(token string) *services.Registry {
	return services.NewRegistry(token, nil)
}

// TestNewUIController tests controller creation
func TestNewUIController(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	if controller == nil {
		t.Fatal("NewUIController returned nil")
	}

	if controller.services != registry {
		t.Error("Expected controller to have registry")
	}
}

// TestControllerApplyFilter tests the filtering functionality
func TestControllerApplyFilter(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	// Create test PRs
	testPRs := []*gh.PullRequest{
		createTestPR(1, "alice", "Fix bug in auth", false, ""),
		createTestPR(2, "bob", "Add new feature", true, ""),
		createTestPR(3, "alice", "Update documentation", false, "dirty"),
		createTestPR(4, "charlie", "Refactor code", false, ""),
	}

	tests := []struct {
		name           string
		mode           string
		value          string
		expectedCount  int
		expectedStatus string
	}{
		{
			name:           "filter by author",
			mode:           "author",
			value:          "alice",
			expectedCount:  2,
			expectedStatus: "Filter applied: author=alice (2 results)",
		},
		{
			name:           "filter by draft status",
			mode:           "status",
			value:          "draft",
			expectedCount:  1,
			expectedStatus: "Filter applied: status=draft (1 results)",
		},
		{
			name:           "filter by conflicts",
			mode:           "status",
			value:          "conflicts",
			expectedCount:  1,
			expectedStatus: "Filter applied: status=conflicts (1 results)",
		},
		{
			name:           "filter drafts boolean",
			mode:           "draft",
			value:          "true",
			expectedCount:  1,
			expectedStatus: "Filter applied: draft=true (1 results)",
		},
		{
			name:           "empty filter returns all",
			mode:           "",
			value:          "",
			expectedCount:  4,
			expectedStatus: "",
		},
		{
			name:           "no matches returns empty",
			mode:           "author",
			value:          "nonexistent",
			expectedCount:  0,
			expectedStatus: "Filter applied: author=nonexistent (0 results)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to PRData format
			prData := make([]*types.PRData, len(testPRs))
			for i, pr := range testPRs {
				prData[i] = &types.PRData{PullRequest: pr}
			}

			// Create filter options
			filter := types.FilterOptions{Mode: tt.mode, Value: tt.value}
			result := controller.ApplyFilter(prData, filter)

			if len(result.FilteredPRs) != tt.expectedCount {
				t.Errorf("Expected %d filtered PRs, got %d", tt.expectedCount, len(result.FilteredPRs))
			}

			if result.StatusMsg != tt.expectedStatus {
				t.Errorf("Expected status message '%s', got '%s'", tt.expectedStatus, result.StatusMsg)
			}
		})
	}
}

// TestControllerFilterDraftPRs tests draft PR filtering
func TestControllerFilterDraftPRs(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	testPRs := []*gh.PullRequest{
		createTestPR(1, "alice", "Regular PR", false, ""),
		createTestPR(2, "bob", "Draft PR", true, ""),
		createTestPR(3, "charlie", "Another draft", true, ""),
		createTestPR(4, "dave", "Regular PR 2", false, ""),
	}

	// Convert to PRData format
	prData := make([]*types.PRData, len(testPRs))
	for i, pr := range testPRs {
		prData[i] = &types.PRData{PullRequest: pr}
	}

	filteredData := controller.FilterDraftPRs(prData)

	// Convert back to check results
	drafts := make([]*gh.PullRequest, len(filteredData))
	for i, pr := range filteredData {
		drafts[i] = pr.PullRequest
	}

	expectedCount := 2
	if len(drafts) != expectedCount {
		t.Errorf("Expected %d draft PRs, got %d", expectedCount, len(drafts))
	}

	// Verify all returned PRs are drafts
	for _, pr := range drafts {
		if !pr.GetDraft() {
			t.Error("Non-draft PR returned in draft filter results")
		}
	}
}

// TestControllerCalculateTableHeight tests table height calculations
func TestControllerCalculateTableHeight(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	tests := []struct {
		name           string
		terminalHeight int
		expectedHeight int
	}{
		{"minimum height for small terminal", 10, 3},  // 10 - 16 = -6, returns minHeight = 3
		{"normal height for medium terminal", 30, 14}, // 30 - 16 = 14
		{"large height for big terminal", 50, 20},     // 50 - 16 = 34, capped to maxHeight = 20
		{"very small terminal", 5, 3},                 // 5 - 16 = -11, returns minHeight = 3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			height := controller.CalculateTableHeight(tt.terminalHeight)
			if height != tt.expectedHeight {
				t.Errorf("Expected height %d, got %d", tt.expectedHeight, height)
			}
		})
	}
}

// TestControllerValidateTabSwitch tests tab switching validation
func TestControllerValidateTabSwitch(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	tests := []struct {
		name       string
		currentIdx int
		targetIdx  int
		totalTabs  int
		expected   bool
	}{
		{"valid switch", 0, 1, 3, true},
		{"switch to same tab", 1, 1, 3, false},
		{"negative target", 0, -1, 3, false},
		{"target beyond range", 0, 5, 3, false},
		{"valid switch to last tab", 0, 2, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controller.ValidateTabSwitch(tt.currentIdx, tt.targetIdx, tt.totalTabs)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestControllerGetSpinnerFrame tests spinner animation
func TestControllerGetSpinnerFrame(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	// Test that spinner cycles through frames
	frames := make([]string, 12)
	for i := 0; i < 12; i++ {
		frames[i] = controller.GetSpinnerFrame(i)
	}

	// Should have 10 different spinner characters cycling
	if frames[0] != frames[10] {
		t.Error("Spinner should cycle after 10 frames")
	}

	if frames[0] == frames[1] {
		t.Error("Consecutive spinner frames should be different")
	}
}

// TestControllerFilterPRsEdgeCases tests edge cases in filtering
func TestControllerFilterPRsEdgeCases(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	// Test with nil user
	prWithNilUser := &gh.PullRequest{
		Number: gh.Int(1),
		Title:  gh.String("PR with nil user"),
		Draft:  gh.Bool(false),
		User:   nil, // This could happen in real data
	}

	// Convert to PRData format
	prData := []*types.PRData{{PullRequest: prWithNilUser}}
	filter := types.FilterOptions{Mode: "author", Value: "alice"}

	// Should not crash and should handle gracefully
	result := controller.ApplyFilter(prData, filter)
	if len(result.FilteredPRs) != 0 {
		t.Error("Should not match PRs with nil user")
	}

	// Test with empty PR list
	emptyResult := controller.ApplyFilter([]*types.PRData{}, filter)
	if len(emptyResult.FilteredPRs) != 0 {
		t.Error("Should return empty list for empty input")
	}
}

// Helper function to create test PRs
func createTestPR(number int, author, title string, isDraft bool, mergeableState string) *gh.PullRequest {
	pr := &gh.PullRequest{
		Number: gh.Int(number),
		Title:  gh.String(title),
		Draft:  gh.Bool(isDraft),
		User: &gh.User{
			Login: gh.String(author),
		},
	}

	if mergeableState != "" {
		pr.MergeableState = gh.String(mergeableState)
	}

	return pr
}

// TestControllerIntegration tests controller with realistic scenarios
func TestControllerIntegration(t *testing.T) {
	registry := createTestRegistry("test-token")
	controller := NewUIController(registry)

	// Create a realistic set of PRs mimicking a real project
	realisticPRs := []*gh.PullRequest{
		createTestPR(123, "developer1", "feat: add user authentication", false, "clean"),
		createTestPR(124, "developer2", "fix: resolve memory leak in cache", false, "dirty"),
		createTestPR(125, "developer1", "docs: update API documentation", true, "clean"),
		createTestPR(126, "renovate[bot]", "chore: update dependencies", false, "clean"),
		createTestPR(127, "developer3", "refactor: simplify user service", false, "blocked"),
	}

	// Test realistic filtering scenarios
	scenarios := []struct {
		name          string
		mode          string
		value         string
		expectedCount int
	}{
		{"find developer1's PRs", "author", "developer1", 2},
		{"find PRs with conflicts", "status", "conflicts", 1},
		{"find draft PRs", "status", "draft", 1},
		{"find ready PRs", "status", "ready", 3}, // PRs that aren't drafts or dirty are considered ready
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Convert to PRData format
			prData := make([]*types.PRData, len(realisticPRs))
			for i, pr := range realisticPRs {
				prData[i] = &types.PRData{PullRequest: pr}
			}

			filter := types.FilterOptions{Mode: scenario.mode, Value: scenario.value}
			result := controller.ApplyFilter(prData, filter)
			if len(result.FilteredPRs) != scenario.expectedCount {
				t.Errorf("Scenario '%s': expected %d results, got %d",
					scenario.name, scenario.expectedCount, len(result.FilteredPRs))
			}
		})
	}
}

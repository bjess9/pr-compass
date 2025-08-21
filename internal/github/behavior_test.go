package github

import (
	"errors" 
	"testing"
	"time"

	"github.com/bjess9/pr-pilot/internal/config"
)

// User Story: As a team lead, I want to see all PRs from repositories tagged with my team's topic
// so I can track all team activity including external contributions
func TestTeamLeadCanTrackAllTeamRepositoryActivity(t *testing.T) {
	// Given a team lead has configured tracking for repositories tagged "backend"
	client := NewMockClient()
	cfg := &config.Config{
		Mode:     "topics", 
		TopicOrg: "testorg",  // Use testorg to match mock data
		Topics:   []string{"backend", "api"},
	}

	// When they request to see current PR activity  
	prs, err := client.FetchPRsFromConfig(cfg)
	
	// Then they should see PRs from all repositories tagged with their topics
	if err != nil {
		t.Fatalf("Team lead cannot see team PR activity: %v", err)
	}
	
	if len(prs) == 0 {
		t.Error("Team lead should see PRs from topic-tagged repositories")
	}
	
	// And all returned PRs should be from repositories matching their topics
	for _, pr := range prs {
		repoName := pr.GetBase().GetRepo().GetName()
		repoOwner := pr.GetBase().GetRepo().GetOwner().GetLogin()
		
		if repoOwner != "testorg" {
			t.Errorf("PR from wrong organization. Expected 'testorg', got '%s' for PR #%d", 
				repoOwner, pr.GetNumber())
		}
		
		// PRs should be from repos that match the topic filtering logic
		if !repoMatchesTopics(repoName, []string{"backend", "api"}) {
			t.Errorf("PR #%d from repo '%s' doesn't match team topics", 
				pr.GetNumber(), repoName)
		}
	}
}

// User Story: As a developer, I want to track PRs from specific repositories I care about
// so I can focus on the work that affects me directly
func TestDeveloperCanTrackSpecificRepositories(t *testing.T) {
	// Given a developer wants to track PRs from specific repositories
	client := NewMockClient()
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"testorg/api-service", "testorg/frontend"},
	}

	// When they check for PR updates
	prs, err := client.FetchPRsFromConfig(cfg)
	
	// Then they should only see PRs from their specified repositories
	if err != nil {
		t.Fatalf("Developer cannot track their specific repositories: %v", err)
	}
	
	expectedRepos := map[string]bool{
		"testorg/api-service": false,
		"testorg/frontend":    false,
	}
	
	for _, pr := range prs {
		repoFullName := pr.GetBase().GetRepo().GetFullName()
		if _, expected := expectedRepos[repoFullName]; !expected {
			t.Errorf("Developer sees PR from untracked repository: %s", repoFullName)
		} else {
			expectedRepos[repoFullName] = true  // Mark as seen
		}
	}
}

// User Story: As an engineering manager, I want to see all PRs in my organization
// so I can get a high-level view of all development activity
func TestManagerCanViewOrganizationWideActivity(t *testing.T) {
	// Given an engineering manager wants organization-wide visibility
	client := NewMockClient()
	cfg := &config.Config{
		Mode:         "organization",
		Organization: "testorg",
	}

	// When they request to see all PR activity
	prs, err := client.FetchPRsFromConfig(cfg)
	
	// Then they should see PRs from all organization repositories
	if err != nil {
		t.Fatalf("Manager cannot see organization-wide activity: %v", err)
	}
	
	if len(prs) == 0 {
		t.Error("Manager should see PRs from organization repositories")
	}
	
	// All PRs should be from the organization
	for _, pr := range prs {
		repoOwner := pr.GetBase().GetRepo().GetOwner().GetLogin()
		if repoOwner != "testorg" {
			t.Errorf("Manager sees PR from outside organization: %s (expected: testorg)", repoOwner)
		}
	}
}

// User Story: As a user, I want to be notified when GitHub is unavailable
// so I understand why I'm not seeing updated PR information
func TestUserReceivesHelpfulErrorWhenGitHubUnavailable(t *testing.T) {
	// Given GitHub API is currently unavailable
	client := NewMockClient()
	client.SetError(errors.New("API rate limit exceeded"))
	
	cfg := &config.Config{
		Mode:  "repos", 
		Repos: []string{"testorg/api-service"},
	}

	// When user tries to fetch PR updates
	prs, err := client.FetchPRsFromConfig(cfg)
	
	// Then they should receive a clear error message
	if err == nil {
		t.Error("User should be informed when GitHub is unavailable")
	}
	
	if prs != nil && len(prs) > 0 {
		t.Error("User should not receive stale data when GitHub is unavailable")
	}
	
	// Error should be informative
	if err != nil && err.Error() == "" {
		t.Error("User should receive helpful error message about GitHub unavailability")
	}
}

// User Story: As a user with advanced search needs, I want to use custom queries
// so I can filter PRs based on complex criteria  
func TestUserCanUseCustomSearchQueries(t *testing.T) {
	// Given a user wants to use a custom search query
	client := NewMockClient()
	cfg := &config.Config{
		Mode:        "search",
		SearchQuery: "org:testorg is:pr is:open label:urgent",
	}

	// When they search for PRs using their custom criteria
	prs, err := client.FetchPRsFromConfig(cfg)
	
	// Then they should get results matching their search
	if err != nil {
		t.Fatalf("User cannot use custom search queries: %v", err)
	}
	
	// Should return some results for valid organization
	if len(prs) == 0 {
		t.Error("User should see PRs matching their search criteria")
	}
	
	// All results should be from the specified organization  
	for _, pr := range prs {
		repoOwner := pr.GetBase().GetRepo().GetOwner().GetLogin()
		if repoOwner != "testorg" {
			t.Errorf("Search results include PR from wrong organization: %s", repoOwner)
		}
	}
}

// User Story: As a user, I want to add custom PRs for testing
// so I can verify the tool works with my specific scenarios
func TestUserCanAddCustomPRsForTesting(t *testing.T) {
	// Given a user wants to test with custom PR data
	client := NewMockClient()
	initialCount := len(client.PRs)
	
	// When they add a custom test PR
	customPR := createTestPR(999, "Custom Test PR", "testuser", "testorg/test-repo", false, true, time.Now(), []string{"test"})
	client.AddPR(customPR)
	
	// Then the PR should be available in their test environment
	if len(client.PRs) != initialCount+1 {
		t.Errorf("User's custom PR not added. Expected %d PRs, got %d", initialCount+1, len(client.PRs))
	}
	
	// The custom PR should be retrievable
	found := false
	for _, pr := range client.PRs {
		if pr.GetNumber() == 999 && pr.GetTitle() == "Custom Test PR" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("User's custom PR not available for testing")
	}
}

// User Story: As a developer setting up the tool, I want to verify mock data is realistic
// so I can trust that my tests represent real-world scenarios
func TestMockDataRepresentsRealisticGitHubPRs(t *testing.T) {
	// Given the system provides mock PR data for testing
	client := NewMockClient()
	
	// When a developer inspects the mock data
	if len(client.PRs) == 0 {
		t.Fatal("System should provide realistic PR data for testing")
	}
	
	// Then each PR should have essential GitHub PR attributes
	for i, pr := range client.PRs {
		// PRs should have valid numbers
		if pr.GetNumber() == 0 {
			t.Errorf("Mock PR %d missing realistic PR number", i)
		}
		
		// PRs should have titles
		if pr.GetTitle() == "" {
			t.Errorf("Mock PR #%d missing realistic title", pr.GetNumber())
		}
		
		// PRs should have authors
		if pr.GetUser() == nil || pr.GetUser().GetLogin() == "" {
			t.Errorf("Mock PR #%d missing realistic author", pr.GetNumber())
		}
		
		// PRs should be associated with repositories
		if pr.GetBase() == nil || pr.GetBase().GetRepo() == nil || pr.GetBase().GetRepo().GetFullName() == "" {
			t.Errorf("Mock PR #%d missing realistic repository association", pr.GetNumber())
		}
		
		// PRs should have URLs
		if pr.GetHTMLURL() == "" {
			t.Errorf("Mock PR #%d missing realistic GitHub URL", pr.GetNumber())
		}
	}
}

// Helper function to check if repository name matches topics
func repoMatchesTopics(repoName string, topics []string) bool {
	for _, topic := range topics {
		// Simple matching logic for testing - in real implementation this would check GitHub topics
		if containsString(repoName, topic) || containsString(repoName, "service") {
			return true
		}
	}
	return false
}

// Helper function for string contains check
func containsString(text, substring string) bool {
	if len(substring) == 0 {
		return true
	}
	if len(text) < len(substring) {
		return false
	}
	for i := 0; i <= len(text)-len(substring); i++ {
		if text[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}

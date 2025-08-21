package github

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bjess9/pr-pilot/internal/config"
)

func TestMockClient_FetchPRsFromConfig_ReposMode(t *testing.T) {
	client := NewMockClient()
	
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"testorg/api-service", "testorg/frontend"},
	}

	prs, err := client.FetchPRsFromConfig(cfg)
	if err != nil {
		t.Fatalf("FetchPRsFromConfig failed: %v", err)
	}

	// Should return PRs that match the specified repos
	expectedCount := 2 // api-service and frontend PRs
	if len(prs) != expectedCount {
		t.Errorf("Expected %d PRs, got %d", expectedCount, len(prs))
	}

	// Verify returned PRs are from correct repos
	for _, pr := range prs {
		repoName := pr.GetBase().GetRepo().GetFullName()
		if repoName != "testorg/api-service" && repoName != "testorg/frontend" {
			t.Errorf("Unexpected PR from repo: %s", repoName)
		}
	}
}

func TestMockClient_FetchPRsFromConfig_OrganizationMode(t *testing.T) {
	client := NewMockClient()
	
	cfg := &config.Config{
		Mode:         "organization",
		Organization: "testorg",
	}

	prs, err := client.FetchPRsFromConfig(cfg)
	if err != nil {
		t.Fatalf("FetchPRsFromConfig failed: %v", err)
	}

	// Should return all PRs from testorg
	expectedCount := 5 // all test PRs are from testorg
	if len(prs) != expectedCount {
		t.Errorf("Expected %d PRs, got %d", expectedCount, len(prs))
	}

	// Verify all PRs are from testorg
	for _, pr := range prs {
		if pr.GetBase().GetRepo().GetOwner().GetLogin() != "testorg" {
			t.Errorf("Expected PR from testorg, got from: %s", pr.GetBase().GetRepo().GetOwner().GetLogin())
		}
	}
}

func TestMockClient_FetchPRsFromConfig_TopicsMode(t *testing.T) {
	client := NewMockClient()
	
	cfg := &config.Config{
		Mode:     "topics",
		TopicOrg: "testorg",
		Topics:   []string{"backend", "api"},
	}

	prs, err := client.FetchPRsFromConfig(cfg)
	if err != nil {
		t.Fatalf("FetchPRsFromConfig failed: %v", err)
	}

	// Should return PRs from repos that match the topics
	if len(prs) == 0 {
		t.Error("Expected some PRs to match topics filter")
	}

	// Verify PRs are from repos that contain topic keywords
	for _, pr := range prs {
		repoName := pr.GetBase().GetRepo().GetName()
		if !containsAny(repoName, []string{"backend", "api", "service"}) {
			t.Errorf("PR from unexpected repo: %s", repoName)
		}
	}
}

func TestMockClient_FetchPRsFromConfig_SearchMode(t *testing.T) {
	client := NewMockClient()
	
	cfg := &config.Config{
		Mode:        "search",
		SearchQuery: "org:testorg is:pr is:open",
	}

	prs, err := client.FetchPRsFromConfig(cfg)
	if err != nil {
		t.Fatalf("FetchPRsFromConfig failed: %v", err)
	}

	// Should return PRs that match the search query
	expectedCount := 5 // all test PRs match the basic search
	if len(prs) != expectedCount {
		t.Errorf("Expected %d PRs, got %d", expectedCount, len(prs))
	}
}

func TestMockClient_SetError(t *testing.T) {
	client := NewMockClient()
	testError := errors.New("API rate limit exceeded")
	
	client.SetError(testError)
	
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"testorg/api-service"},
	}

	_, err := client.FetchPRsFromConfig(cfg)
	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}
}

func TestMockClient_AddPR(t *testing.T) {
	client := NewMockClient()
	initialCount := len(client.PRs)
	
	// Create a new test PR
	now := time.Now()
	newPR := createTestPR(99, "Test PR", "testuser", "testorg/test-repo", false, true, now, []string{"test"})
	
	client.AddPR(newPR)
	
	if len(client.PRs) != initialCount+1 {
		t.Errorf("Expected %d PRs after adding, got %d", initialCount+1, len(client.PRs))
	}
	
	// Verify the added PR is present
	found := false
	for _, pr := range client.PRs {
		if pr.GetNumber() == 99 && pr.GetTitle() == "Test PR" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Added PR not found in client PRs")
	}
}

func TestGenerateTestPRs(t *testing.T) {
	prs := generateTestPRs()
	
	if len(prs) == 0 {
		t.Error("Expected test PRs to be generated")
	}
	
	// Verify PR structure
	for _, pr := range prs {
		if pr.GetNumber() == 0 {
			t.Error("PR number should not be 0")
		}
		
		if pr.GetTitle() == "" {
			t.Error("PR title should not be empty")
		}
		
		if pr.GetUser().GetLogin() == "" {
			t.Error("PR author should not be empty")
		}
		
		if pr.GetBase().GetRepo().GetFullName() == "" {
			t.Error("PR repository should not be empty")
		}
		
		if pr.GetHTMLURL() == "" {
			t.Error("PR HTML URL should not be empty")
		}
	}
}

func TestGenerateTestRepos(t *testing.T) {
	repos := generateTestRepos()
	
	if len(repos) == 0 {
		t.Error("Expected test repositories to be generated")
	}
	
	// Verify repo structure
	for _, repo := range repos {
		if repo.GetName() == "" {
			t.Error("Repository name should not be empty")
		}
		
		if repo.GetFullName() == "" {
			t.Error("Repository full name should not be empty")
		}
		
		if repo.GetOwner().GetLogin() == "" {
			t.Error("Repository owner should not be empty")
		}
		
		if len(repo.Topics) == 0 {
			t.Error("Repository should have topics")
		}
	}
}

// Helper function to check if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(s, substring) {
			return true
		}
	}
	return false
}

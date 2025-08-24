package services

import (
	"testing"

	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

func TestNewFilterService(t *testing.T) {
	service := NewFilterService()
	
	if service == nil {
		t.Fatal("NewFilterService returned nil")
	}
}

func TestFilterService_ApplyFilter_NoFilter(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Test PR 1"),
				User:   &gh.User{Login: gh.String("user1")},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(2),
				Title:  gh.String("Test PR 2"),
				User:   &gh.User{Login: gh.String("user2")},
			},
		},
	}
	
	filter := types.FilterOptions{
		Active: false, // Filter not active
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return all PRs when filter is not active
	if len(result) != 2 {
		t.Fatalf("Expected 2 PRs, got %d", len(result))
	}
}

func TestFilterService_ApplyFilter_AuthorFilter(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Test PR 1"),
				User:   &gh.User{Login: gh.String("testuser")},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(2),
				Title:  gh.String("Test PR 2"),
				User:   &gh.User{Login: gh.String("otheruser")},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(3),
				Title:  gh.String("Test PR 3"),
				User:   &gh.User{Login: gh.String("testuser")},
			},
		},
	}
	
	filter := types.FilterOptions{
		Mode:   "author",
		Value:  "testuser",
		Active: true,
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return only PRs by testuser
	if len(result) != 2 {
		t.Fatalf("Expected 2 PRs, got %d", len(result))
	}
	
	for _, pr := range result {
		if pr.GetUser().GetLogin() != "testuser" {
			t.Errorf("Expected author 'testuser', got '%s'", pr.GetUser().GetLogin())
		}
	}
}

func TestFilterService_ApplyFilter_RepoFilter(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Test PR 1"),
				User:   &gh.User{Login: gh.String("user1")},
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{Name: gh.String("test-repo")},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(2),
				Title:  gh.String("Test PR 2"),
				User:   &gh.User{Login: gh.String("user2")},
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{Name: gh.String("other-repo")},
				},
			},
		},
	}
	
	filter := types.FilterOptions{
		Mode:   "repo",
		Value:  "test-repo",
		Active: true,
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return only PRs from test-repo
	if len(result) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(result))
	}
	
	if result[0].GetBase().GetRepo().GetName() != "test-repo" {
		t.Errorf("Expected repo 'test-repo', got '%s'", result[0].GetBase().GetRepo().GetName())
	}
}

func TestFilterService_ApplyFilter_StatusFilter_Ready(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number:    gh.Int(1),
				Title:     gh.String("Ready PR"),
				User:      &gh.User{Login: gh.String("user1")},
				Draft:     gh.Bool(false),
				Mergeable: gh.Bool(true),
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number:    gh.Int(2),
				Title:     gh.String("Draft PR"),
				User:      &gh.User{Login: gh.String("user2")},
				Draft:     gh.Bool(true),
				Mergeable: gh.Bool(false),
			},
		},
	}
	
	filter := types.FilterOptions{
		Mode:   "status",
		Value:  "ready",
		Active: true,
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return only ready PRs
	if len(result) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(result))
	}
	
	if result[0].GetDraft() || !result[0].GetMergeable() {
		t.Error("Expected ready PR (not draft and mergeable)")
	}
}

func TestFilterService_ApplyFilter_StatusFilter_Draft(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Ready PR"),
				User:   &gh.User{Login: gh.String("user1")},
				Draft:  gh.Bool(false),
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(2),
				Title:  gh.String("Draft PR"),
				User:   &gh.User{Login: gh.String("user2")},
				Draft:  gh.Bool(true),
			},
		},
	}
	
	filter := types.FilterOptions{
		Mode:   "status",
		Value:  "draft",
		Active: true,
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return only draft PRs
	if len(result) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(result))
	}
	
	if !result[0].GetDraft() {
		t.Error("Expected draft PR")
	}
}

func TestFilterService_ApplyFilter_LabelFilter(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Test PR 1"),
				User:   &gh.User{Login: gh.String("user1")},
				Labels: []*gh.Label{
					{Name: gh.String("bug")},
					{Name: gh.String("urgent")},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(2),
				Title:  gh.String("Test PR 2"),
				User:   &gh.User{Login: gh.String("user2")},
				Labels: []*gh.Label{
					{Name: gh.String("feature")},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(3),
				Title:  gh.String("Test PR 3"),
				User:   &gh.User{Login: gh.String("user3")},
				Labels: []*gh.Label{
					{Name: gh.String("bug")},
				},
			},
		},
	}
	
	filter := types.FilterOptions{
		Mode:   "label",
		Value:  "bug",
		Active: true,
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return only PRs with "bug" label
	if len(result) != 2 {
		t.Fatalf("Expected 2 PRs, got %d", len(result))
	}
	
	for _, pr := range result {
		hasLabel := false
		for _, label := range pr.Labels {
			if label.GetName() == "bug" {
				hasLabel = true
				break
			}
		}
		if !hasLabel {
			t.Errorf("Expected PR #%d to have 'bug' label", pr.GetNumber())
		}
	}
}

func TestFilterService_ApplyFilter_SearchFilter(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Fix bug in authentication"),
				Body:   gh.String("This fixes a critical bug"),
				User:   &gh.User{Login: gh.String("user1")},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(2),
				Title:  gh.String("Add new feature"),
				Body:   gh.String("Adds awesome new functionality"),
				User:   &gh.User{Login: gh.String("user2")},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(3),
				Title:  gh.String("Update documentation"),
				Body:   gh.String("Fix typo in documentation"),
				User:   &gh.User{Login: gh.String("user3")},
			},
		},
	}
	
	filter := types.FilterOptions{
		Mode:   "search",
		Value:  "fix",
		Active: true,
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return PRs with "fix" in title or body (case insensitive)
	if len(result) != 2 {
		t.Fatalf("Expected 2 PRs, got %d", len(result))
	}
}

func TestFilterService_ApplyFilter_UnknownMode(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Test PR 1"),
				User:   &gh.User{Login: gh.String("user1")},
			},
		},
	}
	
	filter := types.FilterOptions{
		Mode:   "unknown",
		Value:  "test",
		Active: true,
	}
	
	result := service.ApplyFilter(prs, filter)
	
	// Should return all PRs when filter mode is unknown
	if len(result) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(result))
	}
}

func TestFilterService_GetFilterSuggestions_Author(t *testing.T) {
	service := NewFilterService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				User: &gh.User{Login: gh.String("user1")},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				User: &gh.User{Login: gh.String("user2")},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				User: &gh.User{Login: gh.String("user1")}, // Duplicate
			},
		},
	}
	
	suggestions := service.GetFilterSuggestions(prs, "author")
	
	// Should return unique authors, sorted
	if len(suggestions) != 2 {
		t.Fatalf("Expected 2 suggestions, got %d", len(suggestions))
	}
	
	// Should be sorted
	if suggestions[0] != "user1" || suggestions[1] != "user2" {
		t.Errorf("Expected sorted suggestions [user1, user2], got %v", suggestions)
	}
}

func TestFilterService_ValidateFilter_Valid(t *testing.T) {
	service := NewFilterService()
	
	validFilters := []types.FilterOptions{
		{Active: false}, // Inactive filter is always valid
		{Mode: "author", Value: "testuser", Active: true},
		{Mode: "repo", Value: "test-repo", Active: true},
		{Mode: "status", Value: "ready", Active: true},
		{Mode: "status", Value: "draft", Active: true},
		{Mode: "draft", Active: true}, // Draft mode doesn't need value
		{Mode: "label", Value: "bug", Active: true},
		{Mode: "search", Value: "test", Active: true},
	}
	
	for _, filter := range validFilters {
		err := service.ValidateFilter(filter)
		if err != nil {
			t.Errorf("Expected valid filter %+v, got error: %v", filter, err)
		}
	}
}

func TestFilterService_ValidateFilter_Invalid(t *testing.T) {
	service := NewFilterService()
	
	invalidFilters := []types.FilterOptions{
		{Mode: "", Active: true}, // Empty mode
		{Mode: "unknown", Value: "test", Active: true}, // Unknown mode
		{Mode: "status", Value: "invalid", Active: true}, // Invalid status value
		{Mode: "author", Value: "", Active: true}, // Empty value for non-draft mode
	}
	
	for _, filter := range invalidFilters {
		err := service.ValidateFilter(filter)
		if err == nil {
			t.Errorf("Expected invalid filter %+v to return error", filter)
		}
	}
}

func TestFilterService_GetActiveFiltersDescription(t *testing.T) {
	service := NewFilterService()
	
	tests := []struct {
		name     string
		filter   types.FilterOptions
		expected string
	}{
		{
			name: "Inactive filter",
			filter: types.FilterOptions{
				Mode:   "author",
				Value:  "testuser",
				Active: false,
			},
			expected: "",
		},
		{
			name: "Author filter",
			filter: types.FilterOptions{
				Mode:   "author",
				Value:  "testuser",
				Active: true,
			},
			expected: "by author: testuser",
		},
		{
			name: "Repo filter",
			filter: types.FilterOptions{
				Mode:   "repo",
				Value:  "test-repo",
				Active: true,
			},
			expected: "by repository: test-repo",
		},
		{
			name: "Draft filter",
			filter: types.FilterOptions{
				Mode:   "draft",
				Active: true,
			},
			expected: "drafts only",
		},
		{
			name: "Search filter",
			filter: types.FilterOptions{
				Mode:   "search",
				Value:  "bug fix",
				Active: true,
			},
			expected: "search: bug fix",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetActiveFiltersDescription(tt.filter)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
package services

import (
	"testing"

	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

func TestNewFilterService(t *testing.T) {
	service := NewFilterService()
	if service == nil {
		t.Fatal("NewFilterService should not return nil")
	}
}

func TestFilterService_FilterPRs(t *testing.T) {
	service := NewFilterService()

	// Create test PRs with various attributes
	testPRs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number:         gh.Int(1),
				Title:          gh.String("Fix authentication bug"),
				User:           &gh.User{Login: gh.String("alice")},
				Draft:          gh.Bool(false),
				MergeableState: gh.String("clean"),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{
						Name: gh.String("backend-api"),
					},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number:         gh.Int(2),
				Title:          gh.String("Add new user dashboard"),
				User:           &gh.User{Login: gh.String("bob")},
				Draft:          gh.Bool(true),
				MergeableState: gh.String("clean"),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{
						Name: gh.String("frontend-app"),
					},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number:         gh.Int(3),
				Title:          gh.String("Update documentation"),
				User:           &gh.User{Login: gh.String("alice")},
				Draft:          gh.Bool(false),
				MergeableState: gh.String("dirty"),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{
						Name: gh.String("docs"),
					},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number:         gh.Int(4),
				Title:          gh.String("Refactor database queries"),
				User:           &gh.User{Login: gh.String("charlie")},
				Draft:          gh.Bool(false),
				MergeableState: gh.String("clean"),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{
						Name: gh.String("backend-api"),
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		filter         types.FilterOptions
		expectedCount  int
		expectedPRNums []int
	}{
		{
			name:           "filter by author alice",
			filter:         types.FilterOptions{Mode: "author", Value: "alice"},
			expectedCount:  2,
			expectedPRNums: []int{1, 3},
		},
		{
			name:           "filter by author partial match",
			filter:         types.FilterOptions{Mode: "author", Value: "al"},
			expectedCount:  2,
			expectedPRNums: []int{1, 3},
		},
		{
			name:           "filter by author case insensitive",
			filter:         types.FilterOptions{Mode: "author", Value: "ALICE"},
			expectedCount:  2,
			expectedPRNums: []int{1, 3},
		},
		{
			name:           "filter by author no matches",
			filter:         types.FilterOptions{Mode: "author", Value: "nonexistent"},
			expectedCount:  0,
			expectedPRNums: []int{},
		},
		{
			name:           "filter by draft status true",
			filter:         types.FilterOptions{Mode: "draft", Value: "true"},
			expectedCount:  1,
			expectedPRNums: []int{2},
		},
		{
			name:           "filter by draft status false",
			filter:         types.FilterOptions{Mode: "draft", Value: "false"},
			expectedCount:  3,
			expectedPRNums: []int{1, 3, 4},
		},
		{
			name:           "filter by status draft",
			filter:         types.FilterOptions{Mode: "status", Value: "draft"},
			expectedCount:  1,
			expectedPRNums: []int{2},
		},
		{
			name:           "filter by status conflicts",
			filter:         types.FilterOptions{Mode: "status", Value: "conflicts"},
			expectedCount:  1,
			expectedPRNums: []int{3},
		},
		{
			name:           "filter by status ready",
			filter:         types.FilterOptions{Mode: "status", Value: "ready"},
			expectedCount:  2,
			expectedPRNums: []int{1, 4},
		},
		{
			name:           "filter by title partial match",
			filter:         types.FilterOptions{Mode: "title", Value: "auth"},
			expectedCount:  1,
			expectedPRNums: []int{1},
		},
		{
			name:           "filter by title case insensitive",
			filter:         types.FilterOptions{Mode: "title", Value: "DOCUMENTATION"},
			expectedCount:  1,
			expectedPRNums: []int{3},
		},
		{
			name:           "filter by repo name",
			filter:         types.FilterOptions{Mode: "repo", Value: "backend-api"},
			expectedCount:  2,
			expectedPRNums: []int{1, 4},
		},
		{
			name:           "filter by repo partial match",
			filter:         types.FilterOptions{Mode: "repo", Value: "backend"},
			expectedCount:  2,
			expectedPRNums: []int{1, 4},
		},
		{
			name:           "empty filter returns all",
			filter:         types.FilterOptions{Mode: "", Value: ""},
			expectedCount:  4,
			expectedPRNums: []int{1, 2, 3, 4},
		},
		{
			name:           "empty mode returns all",
			filter:         types.FilterOptions{Mode: "", Value: "alice"},
			expectedCount:  4,
			expectedPRNums: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FilterPRs(testPRs, tt.filter)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(result))
				return
			}

			// Check that we got the expected PR numbers
			gotPRNums := make([]int, len(result))
			for i, pr := range result {
				gotPRNums[i] = pr.GetNumber()
			}

			for _, expectedNum := range tt.expectedPRNums {
				found := false
				for _, gotNum := range gotPRNums {
					if gotNum == expectedNum {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected PR #%d in results, but it was missing", expectedNum)
				}
			}
		})
	}
}

func TestFilterService_FilterPRs_EdgeCases(t *testing.T) {
	service := NewFilterService()

	tests := []struct {
		name   string
		prs    []*types.PRData
		filter types.FilterOptions
		expect int
	}{
		{
			name:   "nil PRs slice",
			prs:    nil,
			filter: types.FilterOptions{Mode: "author", Value: "alice"},
			expect: 0,
		},
		{
			name:   "empty PRs slice",
			prs:    []*types.PRData{},
			filter: types.FilterOptions{Mode: "author", Value: "alice"},
			expect: 0,
		},
		{
			name: "PR with nil user",
			prs: []*types.PRData{
				{
					PullRequest: &gh.PullRequest{
						Number: gh.Int(1),
						Title:  gh.String("Test PR"),
						User:   nil,
					},
				},
			},
			filter: types.FilterOptions{Mode: "author", Value: "alice"},
			expect: 0,
		},
		{
			name: "PR with nil user login",
			prs: []*types.PRData{
				{
					PullRequest: &gh.PullRequest{
						Number: gh.Int(1),
						Title:  gh.String("Test PR"),
						User:   &gh.User{Login: nil},
					},
				},
			},
			filter: types.FilterOptions{Mode: "author", Value: "alice"},
			expect: 0,
		},
		{
			name: "PR with nil base",
			prs: []*types.PRData{
				{
					PullRequest: &gh.PullRequest{
						Number: gh.Int(1),
						Title:  gh.String("Test PR"),
						Base:   nil,
					},
				},
			},
			filter: types.FilterOptions{Mode: "repo", Value: "test"},
			expect: 0,
		},
		{
			name: "PR with nil base repo",
			prs: []*types.PRData{
				{
					PullRequest: &gh.PullRequest{
						Number: gh.Int(1),
						Title:  gh.String("Test PR"),
						Base:   &gh.PullRequestBranch{Repo: nil},
					},
				},
			},
			filter: types.FilterOptions{Mode: "repo", Value: "test"},
			expect: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FilterPRs(tt.prs, tt.filter)
			if len(result) != tt.expect {
				t.Errorf("Expected %d results, got %d", tt.expect, len(result))
			}
		})
	}
}

func TestFilterService_ValidateFilter(t *testing.T) {
	service := NewFilterService()

	tests := []struct {
		name    string
		filter  types.FilterOptions
		wantErr bool
	}{
		{
			name:    "valid author filter",
			filter:  types.FilterOptions{Mode: "author", Value: "alice"},
			wantErr: false,
		},
		{
			name:    "valid status filter",
			filter:  types.FilterOptions{Mode: "status", Value: "draft"},
			wantErr: false,
		},
		{
			name:    "valid draft filter",
			filter:  types.FilterOptions{Mode: "draft", Value: "true"},
			wantErr: false,
		},
		{
			name:    "valid title filter",
			filter:  types.FilterOptions{Mode: "title", Value: "bug fix"},
			wantErr: false,
		},
		{
			name:    "valid repo filter",
			filter:  types.FilterOptions{Mode: "repo", Value: "backend"},
			wantErr: false,
		},
		{
			name:    "empty mode is valid",
			filter:  types.FilterOptions{Mode: "", Value: "anything"},
			wantErr: false,
		},
		{
			name:    "invalid mode",
			filter:  types.FilterOptions{Mode: "invalid", Value: "test"},
			wantErr: true,
		},
		{
			name:    "unknown mode",
			filter:  types.FilterOptions{Mode: "unknown_mode", Value: "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateFilter(tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
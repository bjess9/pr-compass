package services

import (
	"context"
	"testing"
	"time"

	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

func TestNewEnhancementService(t *testing.T) {
	token := "test-token"
	service := NewEnhancementService(token)

	if service == nil {
		t.Fatal("NewEnhancementService returned nil")
	}
}

func TestEnhancementService_IsEnhanced(t *testing.T) {
	service := NewEnhancementService("fake-token")

	// PR not enhanced yet
	if service.IsEnhanced(123) {
		t.Error("Expected PR 123 to not be enhanced initially")
	}
}

func TestEnhancementService_GetEnhancedData_NotFound(t *testing.T) {
	service := NewEnhancementService("fake-token")

	// PR not enhanced yet
	enhanced, exists := service.GetEnhancedData(123)
	if exists {
		t.Error("Expected PR 123 to not exist in enhanced data")
	}
	if enhanced != nil {
		t.Error("Expected enhanced data to be nil for non-existent PR")
	}
}

func TestEnhancementService_EnhancePR_NilPR(t *testing.T) {
	service := NewEnhancementService("fake-token")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Test with nil PR
	enhanced, err := service.EnhancePR(ctx, nil)

	// Should handle nil gracefully
	if err == nil {
		t.Error("Expected error for nil PR")
	}
	if enhanced != nil {
		t.Error("Expected nil enhanced data for nil PR")
	}
}

func TestEnhancementService_EnhancePR_InvalidPR(t *testing.T) {
	service := NewEnhancementService("fake-token")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Test with invalid PR (missing base info)
	pr := &gh.PullRequest{
		Number: gh.Int(123),
		Title:  gh.String("Test PR"),
		// Missing Base which is required
	}

	enhanced, err := service.EnhancePR(ctx, pr)

	// Should handle invalid PR gracefully
	if err == nil {
		t.Error("Expected error for invalid PR")
	}
	if enhanced != nil {
		t.Error("Expected nil enhanced data for invalid PR")
	}
}

func TestEnhancementService_EnhancePRs_EmptyList(t *testing.T) {
	service := NewEnhancementService("fake-token")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	callbackCalled := false
	callback := func(enhanced *types.EnhancedData, err error) {
		callbackCalled = true
	}

	// Test with empty PR list
	err := service.EnhancePRs(ctx, []*gh.PullRequest{}, callback)

	if err != nil {
		t.Errorf("Expected no error for empty PR list, got: %v", err)
	}

	// Give a moment for any potential goroutines
	time.Sleep(100 * time.Millisecond)

	if callbackCalled {
		t.Error("Callback should not be called for empty PR list")
	}
}

func TestEnhancementService_EnhancePRs_WithFakeToken(t *testing.T) {
	service := NewEnhancementService("fake-token")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	callbackResults := []error{}
	callback := func(enhanced *types.EnhancedData, err error) {
		callbackResults = append(callbackResults, err)
	}

	// Create invalid PR that will fail to enhance
	prs := []*gh.PullRequest{
		{
			Number: gh.Int(123),
			Title:  gh.String("Test PR"),
			// Missing required fields will cause enhancement to fail
		},
	}

	err := service.EnhancePRs(ctx, prs, callback)

	if err != nil {
		t.Errorf("EnhancePRs should not return error immediately, got: %v", err)
	}

	// Wait a moment for goroutines to complete
	time.Sleep(200 * time.Millisecond)

	// Should have called callback with error
	if len(callbackResults) == 0 {
		t.Error("Expected callback to be called")
	}
}

func TestEnhancementService_EnhancePR_WithCancelledContext(t *testing.T) {
	service := NewEnhancementService("fake-token")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	pr := &gh.PullRequest{
		Number: gh.Int(123),
		Title:  gh.String("Test PR"),
		Base: &gh.PullRequestBranch{
			Repo: &gh.Repository{
				Name: gh.String("test-repo"),
				Owner: &gh.User{
					Login: gh.String("test-owner"),
				},
			},
		},
	}

	enhanced, err := service.EnhancePR(ctx, pr)

	// Should handle cancelled context gracefully
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
	if enhanced != nil {
		t.Error("Expected nil enhanced data for cancelled context")
	}
}

func TestEnhancementService_Concurrency(t *testing.T) {
	service := NewEnhancementService("fake-token")

	// Test concurrent access to IsEnhanced and GetEnhancedData
	// This mainly tests that there are no race conditions
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(prNum int) {
			defer func() { done <- true }()
			service.IsEnhanced(prNum)
			service.GetEnhancedData(prNum)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passed if no race conditions occurred
}

func TestFetchEnhancedPRData_ValidateInputs(t *testing.T) {
	// We can't test the actual fetchEnhancedPRData function directly since it's not exported,
	// but we can test the validation logic through EnhancePR

	service := NewEnhancementService("fake-token")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Test various invalid PR configurations
	testCases := []struct {
		name string
		pr   *gh.PullRequest
	}{
		{
			name: "nil PR",
			pr:   nil,
		},
		{
			name: "PR with nil base",
			pr: &gh.PullRequest{
				Number: gh.Int(123),
				Base:   nil,
			},
		},
		{
			name: "PR with nil repo",
			pr: &gh.PullRequest{
				Number: gh.Int(123),
				Base: &gh.PullRequestBranch{
					Repo: nil,
				},
			},
		},
		{
			name: "PR with nil owner",
			pr: &gh.PullRequest{
				Number: gh.Int(123),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{
						Owner: nil,
					},
				},
			},
		},
		{
			name: "PR with empty owner login",
			pr: &gh.PullRequest{
				Number: gh.Int(123),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{
						Owner: &gh.User{
							Login: gh.String(""),
						},
					},
				},
			},
		},
		{
			name: "PR with empty repo name",
			pr: &gh.PullRequest{
				Number: gh.Int(123),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{
						Name: gh.String(""),
						Owner: &gh.User{
							Login: gh.String("owner"),
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enhanced, err := service.EnhancePR(ctx, tc.pr)
			if err == nil {
				t.Error("Expected error for invalid PR configuration")
			}
			if enhanced != nil {
				t.Error("Expected nil enhanced data for invalid PR")
			}
		})
	}
}

func TestDetermineReviewStatus(t *testing.T) {
	tests := []struct {
		name     string
		reviews  []*gh.PullRequestReview
		expected string
	}{
		{
			name:     "no reviews",
			reviews:  []*gh.PullRequestReview{},
			expected: "no_review",
		},
		{
			name:     "nil reviews",
			reviews:  nil,
			expected: "no_review",
		},
		{
			name: "single approved review",
			reviews: []*gh.PullRequestReview{
				{
					User:  &gh.User{Login: gh.String("reviewer1")},
					State: gh.String("APPROVED"),
				},
			},
			expected: "approved",
		},
		{
			name: "single changes requested review",
			reviews: []*gh.PullRequestReview{
				{
					User:  &gh.User{Login: gh.String("reviewer1")},
					State: gh.String("CHANGES_REQUESTED"),
				},
			},
			expected: "changes_requested",
		},
		{
			name: "mixed reviews - changes requested wins",
			reviews: []*gh.PullRequestReview{
				{
					User:  &gh.User{Login: gh.String("reviewer1")},
					State: gh.String("APPROVED"),
				},
				{
					User:  &gh.User{Login: gh.String("reviewer2")},
					State: gh.String("CHANGES_REQUESTED"),
				},
			},
			expected: "changes_requested",
		},
		{
			name: "multiple reviews same user - latest wins",
			reviews: []*gh.PullRequestReview{
				{
					User:  &gh.User{Login: gh.String("reviewer1")},
					State: gh.String("CHANGES_REQUESTED"),
				},
				{
					User:  &gh.User{Login: gh.String("reviewer1")},
					State: gh.String("APPROVED"),
				},
			},
			expected: "approved",
		},
		{
			name: "commented reviews are pending",
			reviews: []*gh.PullRequestReview{
				{
					User:  &gh.User{Login: gh.String("reviewer1")},
					State: gh.String("COMMENTED"),
				},
			},
			expected: "pending",
		},
		{
			name: "all approved",
			reviews: []*gh.PullRequestReview{
				{
					User:  &gh.User{Login: gh.String("reviewer1")},
					State: gh.String("APPROVED"),
				},
				{
					User:  &gh.User{Login: gh.String("reviewer2")},
					State: gh.String("APPROVED"),
				},
			},
			expected: "approved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineReviewStatus(tt.reviews)
			if result != tt.expected {
				t.Errorf("determineReviewStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetermineChecksStatus(t *testing.T) {
	tests := []struct {
		name     string
		checks   []*gh.CheckRun
		expected string
	}{
		{
			name:     "no checks",
			checks:   []*gh.CheckRun{},
			expected: "none",
		},
		{
			name:     "nil checks",
			checks:   nil,
			expected: "none",
		},
		{
			name: "all successful",
			checks: []*gh.CheckRun{
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("success"),
				},
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("success"),
				},
			},
			expected: "success",
		},
		{
			name: "one failure",
			checks: []*gh.CheckRun{
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("success"),
				},
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("failure"),
				},
			},
			expected: "failure",
		},
		{
			name: "cancelled check counts as failure",
			checks: []*gh.CheckRun{
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("cancelled"),
				},
			},
			expected: "failure",
		},
		{
			name: "pending checks",
			checks: []*gh.CheckRun{
				{
					Status: gh.String("in_progress"),
				},
			},
			expected: "pending",
		},
		{
			name: "queued checks",
			checks: []*gh.CheckRun{
				{
					Status: gh.String("queued"),
				},
			},
			expected: "pending",
		},
		{
			name: "mixed status - failure takes precedence",
			checks: []*gh.CheckRun{
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("success"),
				},
				{
					Status: gh.String("in_progress"),
				},
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("failure"),
				},
			},
			expected: "failure",
		},
		{
			name: "mixed status - pending when no failures",
			checks: []*gh.CheckRun{
				{
					Status:     gh.String("completed"),
					Conclusion: gh.String("success"),
				},
				{
					Status: gh.String("in_progress"),
				},
			},
			expected: "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineChecksStatus(tt.checks)
			if result != tt.expected {
				t.Errorf("determineChecksStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

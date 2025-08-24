package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bjess9/pr-compass/internal/batch"
	"github.com/bjess9/pr-compass/internal/github"
	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

// enhancementService implements the EnhancementService interface
type enhancementService struct {
	token        string
	mutex        sync.RWMutex
	enhancedData map[int]*types.EnhancedData
	batchManager *batch.Manager[*gh.PullRequest, types.EnhancedData]
}

// NewEnhancementService creates a new enhancement service
func NewEnhancementService(token string) EnhancementService {
	// Create worker function for batch PR enhancement
	enhancePRWorker := func(batchCtx context.Context, pr *gh.PullRequest) (types.EnhancedData, error) {
		// Create timeout context for this specific PR (10 seconds)
		prCtx, prCancel := context.WithTimeout(batchCtx, 10*time.Second)
		defer prCancel()

		// Get GitHub client
		client, err := github.NewClient(token)
		if err != nil {
			return types.EnhancedData{Number: pr.GetNumber()}, err
		}

		// Fetch enhanced data for this PR
		return fetchEnhancedPRData(prCtx, client, pr)
	}

	// Create batch manager with 5 concurrent workers for optimal performance
	batchManager := batch.NewManager(5, enhancePRWorker)

	return &enhancementService{
		token:        token,
		enhancedData: make(map[int]*types.EnhancedData),
		batchManager: batchManager,
	}
}

// EnhancePR fetches detailed information for a single PR
func (s *enhancementService) EnhancePR(ctx context.Context, pr *gh.PullRequest) (*types.EnhancedData, error) {
	prNumber := pr.GetNumber()

	// Check if already enhanced
	s.mutex.RLock()
	if enhanced, exists := s.enhancedData[prNumber]; exists {
		s.mutex.RUnlock()
		return enhanced, nil
	}
	s.mutex.RUnlock()

	// Create timeout context for this specific PR
	prCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get GitHub client
	client, err := github.NewClient(s.token)
	if err != nil {
		return nil, err
	}

	// Fetch enhanced data
	enhancedData, err := fetchEnhancedPRData(prCtx, client, pr)
	if err != nil {
		return nil, err
	}

	// Store the enhanced data
	s.mutex.Lock()
	s.enhancedData[prNumber] = &enhancedData
	s.mutex.Unlock()

	return &enhancedData, nil
}

// EnhancePRs processes multiple PRs for enhancement
func (s *enhancementService) EnhancePRs(ctx context.Context, prs []*gh.PullRequest, callback func(*types.EnhancedData, error)) error {
	// Process PRs through batch manager
	for _, pr := range prs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Submit PR for enhancement
			go func(pr *gh.PullRequest) {
				// Simplified approach - directly enhance PR
				enhancedData, err := s.EnhancePR(ctx, pr)
				if err != nil {
					callback(nil, err)
				} else {
					callback(enhancedData, nil)
				}
			}(pr)
		}
	}

	return nil
}

// IsEnhanced checks if a PR has been enhanced
func (s *enhancementService) IsEnhanced(prNumber int) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.enhancedData[prNumber]
	return exists
}

// GetEnhancedData retrieves enhanced data for a PR if available
func (s *enhancementService) GetEnhancedData(prNumber int) (*types.EnhancedData, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	enhanced, exists := s.enhancedData[prNumber]
	return enhanced, exists
}

// fetchEnhancedPRData fetches detailed PR information from GitHub API
func fetchEnhancedPRData(ctx context.Context, client *gh.Client, pr *gh.PullRequest) (types.EnhancedData, error) {
	// Validate PR structure to avoid nil pointer panics
	if pr == nil {
		return types.EnhancedData{}, fmt.Errorf("PR is nil")
	}
	if pr.GetBase() == nil || pr.GetBase().GetRepo() == nil {
		return types.EnhancedData{}, fmt.Errorf("PR base or repository is nil for PR #%d", pr.GetNumber())
	}
	if pr.GetBase().GetRepo().GetOwner() == nil {
		return types.EnhancedData{}, fmt.Errorf("PR repository owner is nil for PR #%d", pr.GetNumber())
	}

	owner := pr.GetBase().GetRepo().GetOwner().GetLogin()
	repo := pr.GetBase().GetRepo().GetName()
	number := pr.GetNumber()

	// Additional validation for required fields
	if owner == "" {
		return types.EnhancedData{}, fmt.Errorf("PR owner is empty for PR #%d", number)
	}
	if repo == "" {
		return types.EnhancedData{}, fmt.Errorf("PR repository name is empty for PR #%d", number)
	}

	// Get detailed PR data
	detailedPR, _, err := client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return types.EnhancedData{}, err
	}

	// Get review status
	reviews, _, err := client.PullRequests.ListReviews(ctx, owner, repo, number, nil)
	reviewStatus := "unknown"
	if err == nil {
		reviewStatus = determineReviewStatus(reviews)
	}

	// Get checks status
	checksStatus := "unknown"
	if pr.GetHead() != nil {
		if sha := pr.GetHead().GetSHA(); sha != "" {
			checks, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
			if err == nil && checks != nil {
				checksStatus = determineChecksStatus(checks.CheckRuns)
			}
		}
	}

	// Determine mergeable status
	mergeableStatus := "unknown"
	if detailedPR.Mergeable != nil {
		if *detailedPR.Mergeable {
			mergeableStatus = "clean"
		} else {
			mergeableStatus = "conflicts"
		}
	}

	return types.EnhancedData{
		Number:         number,
		Comments:       detailedPR.GetComments(),
		ReviewComments: detailedPR.GetReviewComments(),
		ReviewStatus:   reviewStatus,
		ChecksStatus:   checksStatus,
		Mergeable:      mergeableStatus,
		Additions:      detailedPR.GetAdditions(),
		Deletions:      detailedPR.GetDeletions(),
		ChangedFiles:   detailedPR.GetChangedFiles(),
		EnhancedAt:     time.Now(),
	}, nil
}

// determineReviewStatus analyzes review data to determine overall status
func determineReviewStatus(reviews []*gh.PullRequestReview) string {
	if len(reviews) == 0 {
		return "no_review"
	}

	// Get latest review by each reviewer
	latestReviews := make(map[string]string)
	for _, review := range reviews {
		user := review.GetUser().GetLogin()
		state := review.GetState()
		latestReviews[user] = state
	}

	// Check for blocking states
	for _, state := range latestReviews {
		if state == "CHANGES_REQUESTED" {
			return "changes_requested"
		}
	}

	// Check if all reviews are approved
	approvedCount := 0
	for _, state := range latestReviews {
		if state == "APPROVED" {
			approvedCount++
		}
	}

	if approvedCount > 0 && approvedCount == len(latestReviews) {
		return "approved"
	}

	return "pending"
}

// determineChecksStatus analyzes check runs to determine overall status
func determineChecksStatus(checkRuns []*gh.CheckRun) string {
	if len(checkRuns) == 0 {
		return "none"
	}

	hasFailure := false
	hasPending := false

	for _, check := range checkRuns {
		switch check.GetStatus() {
		case "completed":
			if check.GetConclusion() == "failure" || check.GetConclusion() == "cancelled" {
				hasFailure = true
			}
		case "in_progress", "queued":
			hasPending = true
		}
	}

	if hasFailure {
		return "failure"
	}
	if hasPending {
		return "pending"
	}
	return "success"
}
package services

import (
	"context"
	"sort"

	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/github"
	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

// prService implements the PRService interface
type prService struct {
	token   string
	cache   *cache.PRCache
}

// NewPRService creates a new PR service
func NewPRService(token string, cache *cache.PRCache) PRService {
	return &prService{
		token: token,
		cache: cache,
	}
}

// FetchPRs retrieves PRs based on configuration
func (s *prService) FetchPRs(ctx context.Context, cfg *config.Config) ([]*types.PRData, error) {
	var ghPRs []*gh.PullRequest
	var err error

	// Use optimized fetching with caching if available
	if s.cache != nil {
		ghPRs, err = github.FetchPRsFromConfigOptimized(ctx, cfg, s.token, s.cache)
	} else {
		ghPRs, err = github.FetchPRsFromConfig(ctx, cfg, s.token)
	}

	if err != nil {
		return nil, err
	}

	// Convert to our internal type and sort
	prs := s.convertAndSort(ghPRs)
	return prs, nil
}

// RefreshPRs performs a background refresh of PR data
func (s *prService) RefreshPRs(ctx context.Context, cfg *config.Config) ([]*types.PRData, error) {
	// For background refresh, we can use the same logic as FetchPRs
	// but potentially with different timeout or error handling
	return s.FetchPRs(ctx, cfg)
}

// convertAndSort converts GitHub PRs to our internal type and sorts them
func (s *prService) convertAndSort(ghPRs []*gh.PullRequest) []*types.PRData {
	// Convert to our internal type
	prs := make([]*types.PRData, len(ghPRs))
	for i, pr := range ghPRs {
		prs[i] = &types.PRData{
			PullRequest: pr,
			Enhanced:    nil, // Will be populated later by enhancement service
		}
	}

	// Sort by most recently updated first
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].GetUpdatedAt().Time.After(prs[j].GetUpdatedAt().Time)
	})

	return prs
}
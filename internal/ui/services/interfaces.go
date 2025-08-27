package services

import (
	"context"

	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

// PRService handles all pull request related operations
type PRService interface {
	// FetchPRs retrieves PRs based on configuration
	FetchPRs(ctx context.Context, cfg *config.Config) ([]*types.PRData, error)

	// RefreshPRs performs a background refresh of PR data
	RefreshPRs(ctx context.Context, cfg *config.Config) ([]*types.PRData, error)
}

// EnhancementService handles PR data enhancement operations
type EnhancementService interface {
	// EnhancePR fetches detailed information for a single PR
	EnhancePR(ctx context.Context, pr *gh.PullRequest) (*types.EnhancedData, error)

	// EnhancePRs processes multiple PRs for enhancement
	EnhancePRs(ctx context.Context, prs []*gh.PullRequest, callback func(*types.EnhancedData, error)) error

	// IsEnhanced checks if a PR has been enhanced
	IsEnhanced(prNumber int) bool

	// GetEnhancedData retrieves enhanced data for a PR if available
	GetEnhancedData(prNumber int) (*types.EnhancedData, bool)
}

// StateService handles application state management
type StateService interface {
	// GetState returns the current application state
	GetState() *types.AppState

	// UpdateState updates the application state
	UpdateState(updater func(*types.AppState))

	// UpdatePRs updates the PR data in state
	UpdatePRs(prs []*types.PRData)

	// UpdateFilter updates the filter state
	UpdateFilter(filter types.FilterOptions)

	// SetError sets an error state
	SetError(err error)

	// ClearError clears the error state
	ClearError()

	// Additional methods for state management
	UpdateFilteredPRs(filteredPRs []*types.PRData)
	SetLoaded(loaded bool)
	SetBackgroundRefreshing(refreshing bool)
	SetEnhancing(enhancing bool)
	UpdateEnhancedCount(count int)
	AddToEnhancementQueue(prNumber int)
	RemoveFromEnhancementQueue(prNumber int)
	IsInEnhancementQueue(prNumber int) bool
	UpdateStatusMessage(message string)
	SetShowHelp(showHelp bool)
	UpdateSelectedPR(index int)
	UpdateTableCursor(cursor int)
	UpdatePREnhancement(prNumber int, enhanced *types.EnhancedData)
}

// FilterService handles PR filtering operations
type FilterService interface {
	// FilterPRs applies filtering to a list of PRs
	FilterPRs(prs []*types.PRData, filter types.FilterOptions) []*types.PRData

	// ValidateFilter validates filter options
	ValidateFilter(filter types.FilterOptions) error
}

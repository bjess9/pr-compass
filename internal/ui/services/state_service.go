package services

import (
	"sync"

	"github.com/bjess9/pr-compass/internal/ui/types"
)

// stateService implements the StateService interface
type stateService struct {
	mutex sync.RWMutex
	state *types.AppState
}

// NewStateService creates a new state service
func NewStateService() StateService {
	return &stateService{
		state: &types.AppState{
			PRs:              []*types.PRData{},
			FilteredPRs:      []*types.PRData{},
			UI:               types.UIState{},
			EnhancementQueue: make(map[int]bool),
		},
	}
}

// GetState returns the current application state
func (s *stateService) GetState() *types.AppState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a deep copy to prevent external mutations
	stateCopy := *s.state

	// Copy slices
	stateCopy.PRs = make([]*types.PRData, len(s.state.PRs))
	copy(stateCopy.PRs, s.state.PRs)

	stateCopy.FilteredPRs = make([]*types.PRData, len(s.state.FilteredPRs))
	copy(stateCopy.FilteredPRs, s.state.FilteredPRs)

	// Copy map
	stateCopy.EnhancementQueue = make(map[int]bool)
	for k, v := range s.state.EnhancementQueue {
		stateCopy.EnhancementQueue[k] = v
	}

	return &stateCopy
}

// UpdateState updates the application state
func (s *stateService) UpdateState(updater func(*types.AppState)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	updater(s.state)
}

// UpdatePRs updates the PR data in state
func (s *stateService) UpdatePRs(prs []*types.PRData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state.PRs = prs
	s.state.FilteredPRs = prs // Initially show all PRs
	s.state.Loaded = true
}

// UpdateFilter updates the filter state
func (s *stateService) UpdateFilter(filter types.FilterOptions) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.UI.Filter = filter
}

// SetError sets an error state
func (s *stateService) SetError(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.Error = err
}

// ClearError clears the error state
func (s *stateService) ClearError() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.Error = nil
}

// UpdateFilteredPRs updates the filtered PRs list
func (s *stateService) UpdateFilteredPRs(filteredPRs []*types.PRData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.FilteredPRs = filteredPRs
}

// SetLoaded sets the loaded state
func (s *stateService) SetLoaded(loaded bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.Loaded = loaded
}

// SetBackgroundRefreshing sets the background refreshing state
func (s *stateService) SetBackgroundRefreshing(refreshing bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.BackgroundRefreshing = refreshing
}

// SetEnhancing sets the enhancing state
func (s *stateService) SetEnhancing(enhancing bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.Enhancing = enhancing
}

// UpdateEnhancedCount updates the enhanced PR count
func (s *stateService) UpdateEnhancedCount(count int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.EnhancedCount = count
}

// AddToEnhancementQueue adds a PR to the enhancement queue
func (s *stateService) AddToEnhancementQueue(prNumber int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.EnhancementQueue[prNumber] = true
}

// RemoveFromEnhancementQueue removes a PR from the enhancement queue
func (s *stateService) RemoveFromEnhancementQueue(prNumber int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.state.EnhancementQueue, prNumber)
}

// IsInEnhancementQueue checks if a PR is in the enhancement queue
func (s *stateService) IsInEnhancementQueue(prNumber int) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.state.EnhancementQueue[prNumber]
	return exists
}

// UpdateStatusMessage updates the status message
func (s *stateService) UpdateStatusMessage(message string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.UI.StatusMsg = message
}

// SetShowHelp sets the help display state
func (s *stateService) SetShowHelp(showHelp bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.UI.ShowHelp = showHelp
}

// UpdateSelectedPR updates the currently selected PR
func (s *stateService) UpdateSelectedPR(index int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.UI.SelectedPR = index
}

// UpdateTableCursor updates the table cursor position
func (s *stateService) UpdateTableCursor(cursor int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.UI.TableCursor = cursor
}

// GetEnhancedPR returns an enhanced PR if available
func (s *stateService) GetEnhancedPR(prNumber int) (*types.PRData, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, pr := range s.state.PRs {
		if pr.GetNumber() == prNumber && pr.Enhanced != nil {
			return pr, true
		}
	}
	return nil, false
}

// UpdatePREnhancement updates the enhanced data for a specific PR
func (s *stateService) UpdatePREnhancement(prNumber int, enhanced *types.EnhancedData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update the PR in both PRs and FilteredPRs lists
	for _, pr := range s.state.PRs {
		if pr.GetNumber() == prNumber {
			pr.Enhanced = enhanced
		}
	}

	for _, pr := range s.state.FilteredPRs {
		if pr.GetNumber() == prNumber {
			pr.Enhanced = enhanced
		}
	}
}

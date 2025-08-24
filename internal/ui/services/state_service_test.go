package services

import (
	"testing"

	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

func TestNewStateService(t *testing.T) {
	service := NewStateService()
	
	if service == nil {
		t.Fatal("NewStateService returned nil")
	}
	
	state := service.GetState()
	if state == nil {
		t.Fatal("GetState returned nil")
	}
	
	// Check initial state
	if len(state.PRs) != 0 {
		t.Errorf("Expected empty PRs slice, got %d items", len(state.PRs))
	}
	if len(state.FilteredPRs) != 0 {
		t.Errorf("Expected empty FilteredPRs slice, got %d items", len(state.FilteredPRs))
	}
	if state.Loaded {
		t.Error("Expected Loaded to be false initially")
	}
	if state.EnhancementQueue == nil {
		t.Error("Expected EnhancementQueue to be initialized")
	}
}

func TestStateService_UpdatePRs(t *testing.T) {
	service := NewStateService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(1),
				Title:  gh.String("Test PR 1"),
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(2),
				Title:  gh.String("Test PR 2"),
			},
		},
	}
	
	service.UpdatePRs(prs)
	
	state := service.GetState()
	if len(state.PRs) != 2 {
		t.Fatalf("Expected 2 PRs, got %d", len(state.PRs))
	}
	if len(state.FilteredPRs) != 2 {
		t.Fatalf("Expected 2 FilteredPRs, got %d", len(state.FilteredPRs))
	}
	if !state.Loaded {
		t.Error("Expected Loaded to be true after UpdatePRs")
	}
	
	// Check PR content
	if state.PRs[0].GetNumber() != 1 {
		t.Errorf("Expected first PR number to be 1, got %d", state.PRs[0].GetNumber())
	}
	if state.PRs[1].GetTitle() != "Test PR 2" {
		t.Errorf("Expected second PR title to be 'Test PR 2', got %s", state.PRs[1].GetTitle())
	}
}

func TestStateService_UpdateFilter(t *testing.T) {
	service := NewStateService()
	
	filter := types.FilterOptions{
		Mode:   "author",
		Value:  "test-user",
		Active: true,
	}
	
	service.UpdateFilter(filter)
	
	state := service.GetState()
	if state.UI.Filter.Mode != "author" {
		t.Errorf("Expected filter mode 'author', got %s", state.UI.Filter.Mode)
	}
	if state.UI.Filter.Value != "test-user" {
		t.Errorf("Expected filter value 'test-user', got %s", state.UI.Filter.Value)
	}
	if !state.UI.Filter.Active {
		t.Error("Expected filter to be active")
	}
}

func TestStateService_ErrorHandling(t *testing.T) {
	service := NewStateService()
	
	// Test setting error
	testError := &testError{message: "test error"}
	service.SetError(testError)
	
	state := service.GetState()
	if state.Error == nil {
		t.Fatal("Expected error to be set")
	}
	if state.Error.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got %s", state.Error.Error())
	}
	
	// Test clearing error
	service.ClearError()
	state = service.GetState()
	if state.Error != nil {
		t.Error("Expected error to be cleared")
	}
}

func TestStateService_EnhancementQueue(t *testing.T) {
	service := NewStateService()
	
	// Test adding to queue
	service.AddToEnhancementQueue(123)
	if !service.IsInEnhancementQueue(123) {
		t.Error("Expected PR 123 to be in enhancement queue")
	}
	
	// Test removing from queue
	service.RemoveFromEnhancementQueue(123)
	if service.IsInEnhancementQueue(123) {
		t.Error("Expected PR 123 to be removed from enhancement queue")
	}
	
	// Test non-existent PR
	if service.IsInEnhancementQueue(456) {
		t.Error("Expected PR 456 to not be in enhancement queue")
	}
}

func TestStateService_StatusAndUI(t *testing.T) {
	service := NewStateService()
	
	// Test status message
	service.UpdateStatusMessage("Test status")
	state := service.GetState()
	if state.UI.StatusMsg != "Test status" {
		t.Errorf("Expected status message 'Test status', got %s", state.UI.StatusMsg)
	}
	
	// Test help display
	service.SetShowHelp(true)
	state = service.GetState()
	if !state.UI.ShowHelp {
		t.Error("Expected ShowHelp to be true")
	}
	
	// Test cursor updates
	service.UpdateTableCursor(5)
	service.UpdateSelectedPR(3)
	state = service.GetState()
	if state.UI.TableCursor != 5 {
		t.Errorf("Expected table cursor to be 5, got %d", state.UI.TableCursor)
	}
	if state.UI.SelectedPR != 3 {
		t.Errorf("Expected selected PR to be 3, got %d", state.UI.SelectedPR)
	}
}

func TestStateService_EnhancementData(t *testing.T) {
	service := NewStateService()
	
	// Create test PRs
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number: gh.Int(100),
				Title:  gh.String("Test PR"),
			},
		},
	}
	service.UpdatePRs(prs)
	
	// Update enhancement
	enhanced := &types.EnhancedData{
		Number:   100,
		Comments: 5,
	}
	service.UpdatePREnhancement(100, enhanced)
	
	// Check that enhancement was applied
	state := service.GetState()
	if state.PRs[0].Enhanced == nil {
		t.Fatal("Expected enhanced data to be set")
	}
	if state.PRs[0].Enhanced.Comments != 5 {
		t.Errorf("Expected 5 comments, got %d", state.PRs[0].Enhanced.Comments)
	}
	if state.FilteredPRs[0].Enhanced == nil {
		t.Fatal("Expected enhanced data to be set in filtered PRs too")
	}
}

func TestStateService_GetStateCopy(t *testing.T) {
	service := NewStateService()
	
	// Set up some state
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{Number: gh.Int(1)},
		},
	}
	service.UpdatePRs(prs)
	service.AddToEnhancementQueue(123)
	
	// Get two copies of state
	state1 := service.GetState()
	state2 := service.GetState()
	
	// They should have the same content
	if len(state1.PRs) != len(state2.PRs) {
		t.Error("State copies should have same PR count")
	}
	
	// But be different objects (deep copy test)
	if &state1.PRs == &state2.PRs {
		t.Error("State copies should not share the same PR slice reference")
	}
	if &state1.EnhancementQueue == &state2.EnhancementQueue {
		t.Error("State copies should not share the same EnhancementQueue reference")
	}
	
	// Modify one copy and ensure other is unaffected
	state1.EnhancementQueue[456] = true
	if _, exists := state2.EnhancementQueue[456]; exists {
		t.Error("Modifying one state copy should not affect another")
	}
}

// Helper type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
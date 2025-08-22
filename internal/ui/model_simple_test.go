package ui

import (
	"strings"
	"testing"

	"github.com/bjess9/pr-pilot/internal/github"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModelBasic(t *testing.T) {
	token := "test-token"
	m := InitialModel(token)

	// Test basic initialization
	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Initial view should show loading message")
	}
}

func TestModelWithPRsBasic(t *testing.T) {
	m := InitialModel("test-token")
	
	// Create test PRs
	testPRs := github.NewMockClient().PRs
	
	// Send PRs to model
	updatedModel, _ := m.Update(testPRs)
	
	// Test view after loading PRs
	view := updatedModel.View()
	if strings.Contains(view, "Loading") {
		t.Error("View should not show loading after PRs are loaded")
	}
	
	if !strings.Contains(view, "Navigate") {
		t.Error("View should show navigation help after PRs are loaded")
	}
}

func TestModelKeyHandling(t *testing.T) {
	m := InitialModel("test-token")
	
	// Test quit key
	quitMsg := tea.KeyMsg{
		Type:  tea.KeyCtrlC,
	}
	
	_, cmd := m.Update(quitMsg)
	if cmd == nil {
		t.Error("Quit key should return a command")
	}
}

func TestModelErrorHandling(t *testing.T) {
	m := InitialModel("test-token")
	
	// Test error message
	errorMsg := errMsg{err: &MockError{"test error"}}
	
	updatedModel, _ := m.Update(errorMsg)
	view := updatedModel.View()
	
	if !strings.Contains(view, "Error") {
		t.Error("View should show error when error message is received")
	}
}

// Mock error type for testing
type MockError struct {
	msg string
}

func (e *MockError) Error() string {
	return e.msg
}

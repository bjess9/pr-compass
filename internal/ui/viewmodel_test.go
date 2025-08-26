package ui

import (
	"testing"

	gh "github.com/google/go-github/v55/github"
)

// TestNewViewModel tests view model creation
func TestNewViewModel(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	if vm == nil {
		t.Fatal("NewViewModel returned nil")
	}
	
	if vm.controller != controller {
		t.Error("ViewModel controller not set correctly")
	}
}

// TestCreateTabViewModels tests tab view model creation
func TestCreateTabViewModels(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	// Create test tab manager
	tabManager := NewTabManager("test-token")
	
	// Add test tabs
	tab1 := &TabConfig{Name: "Tab 1", Mode: "repos", Repos: []string{"test/repo1"}}
	tab2 := &TabConfig{Name: "Tab 2", Mode: "repos", Repos: []string{"test/repo2"}}
	
	tabState1 := tabManager.AddTab(tab1)
	tabState2 := tabManager.AddTab(tab2)
	
	// Set some state
	tabState1.Loaded = true
	tabState1.PRs = []*gh.PullRequest{createTestPR(1, "user1", "Test PR", false, "")}
	tabState2.Loaded = false
	
	viewModels := vm.CreateTabViewModels(tabManager)
	
	if len(viewModels) != 2 {
		t.Fatalf("Expected 2 view models, got %d", len(viewModels))
	}
	
	// Test first tab
	vm1 := viewModels[0]
	if vm1.Name != "Tab 1" {
		t.Errorf("Expected tab name 'Tab 1', got '%s'", vm1.Name)
	}
	if !vm1.IsActive {
		t.Error("Expected first tab to be active")
	}
	if vm1.IsLoading {
		t.Error("Expected first tab not to be loading")
	}
	if vm1.PRCount != 1 {
		t.Errorf("Expected PR count 1, got %d", vm1.PRCount)
	}
	
	// Test second tab
	vm2 := viewModels[1]
	if vm2.Name != "Tab 2" {
		t.Errorf("Expected tab name 'Tab 2', got '%s'", vm2.Name)
	}
	if vm2.IsActive {
		t.Error("Expected second tab not to be active")
	}
	if !vm2.IsLoading {
		t.Error("Expected second tab to be loading")
	}
}

// TestCreateFilterViewModel tests filter view model creation
func TestCreateFilterViewModel(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	// Create test tab state
	tabConfig := &TabConfig{Name: "Test Tab", Mode: "repos", Repos: []string{"test/repo"}}
	tabState := NewTabState(tabConfig, "test-token")
	
	tests := []struct {
		name                 string
		filterMode           string
		filterValue          string
		expectedIsActive     bool
		expectedDescription  string
	}{
		{
			name:                "no filter",
			filterMode:          "",
			filterValue:         "",
			expectedIsActive:    false,
			expectedDescription: "",
		},
		{
			name:                "author filter",
			filterMode:          "author",
			filterValue:         "alice",
			expectedIsActive:    true,
			expectedDescription: "by author: alice",
		},
		{
			name:                "status filter",
			filterMode:          "status", 
			filterValue:         "draft",
			expectedIsActive:    true,
			expectedDescription: "by status: draft",
		},
		{
			name:                "draft filter",
			filterMode:          "draft",
			filterValue:         "true",
			expectedIsActive:    true,
			expectedDescription: "drafts only",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabState.FilterMode = tt.filterMode
			tabState.FilterValue = tt.filterValue
			
			filterVM := vm.CreateFilterViewModel(tabState)
			
			if filterVM.IsActive != tt.expectedIsActive {
				t.Errorf("Expected IsActive %v, got %v", tt.expectedIsActive, filterVM.IsActive)
			}
			
			if filterVM.Mode != tt.filterMode {
				t.Errorf("Expected Mode '%s', got '%s'", tt.filterMode, filterVM.Mode)
			}
			
			if filterVM.Value != tt.filterValue {
				t.Errorf("Expected Value '%s', got '%s'", tt.filterValue, filterVM.Value)
			}
			
			if filterVM.Description != tt.expectedDescription {
				t.Errorf("Expected Description '%s', got '%s'", tt.expectedDescription, filterVM.Description)
			}
		})
	}
}

// TestCreateTableViewModel tests table view model creation
func TestCreateTableViewModel(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	// Create test PRs
	testPRs := []*gh.PullRequest{
		createTestPR(1, "alice", "Fix bug", false, ""),
		createTestPR(2, "bob", "Add feature", true, ""),
	}
	
	enhancementQueue := map[int]bool{1: true}
	
	tableVM := vm.CreateTableViewModel(testPRs, enhancementQueue, 0, 20, 100)
	
	if len(tableVM.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(tableVM.Rows))
	}
	
	if tableVM.SelectedIndex != 0 {
		t.Errorf("Expected SelectedIndex 0, got %d", tableVM.SelectedIndex)
	}
	
	if tableVM.Height != 20 {
		t.Errorf("Expected Height 20, got %d", tableVM.Height)
	}
	
	if tableVM.Width != 100 {
		t.Errorf("Expected Width 100, got %d", tableVM.Width)
	}
}

// TestCreateStatusViewModel tests status view model creation  
func TestCreateStatusViewModel(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	tests := []struct {
		name              string
		statusMsg         string
		filterInfo        string
		expectedSpinner   bool
		expectedFilter    bool
	}{
		{
			name:            "loading state",
			statusMsg:       "Loading...",
			filterInfo:      "",
			expectedSpinner: true,
			expectedFilter:  false,
		},
		{
			name:            "with filter",
			statusMsg:       "Ready",
			filterInfo:      "author: alice",
			expectedSpinner: false,
			expectedFilter:  true,
		},
		{
			name:            "normal state",
			statusMsg:       "Ready",
			filterInfo:      "",
			expectedSpinner: false,
			expectedFilter:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusVM := vm.CreateStatusViewModel(0, tt.statusMsg, tt.filterInfo, "API: 5000/5000")
			
			if statusVM.Message != tt.statusMsg {
				t.Errorf("Expected Message '%s', got '%s'", tt.statusMsg, statusVM.Message)
			}
			
			if statusVM.ShowSpinner != tt.expectedSpinner {
				t.Errorf("Expected ShowSpinner %v, got %v", tt.expectedSpinner, statusVM.ShowSpinner)
			}
			
			if statusVM.HasActiveFilter != tt.expectedFilter {
				t.Errorf("Expected HasActiveFilter %v, got %v", tt.expectedFilter, statusVM.HasActiveFilter)
			}
			
			if statusVM.FilterInfo != tt.filterInfo {
				t.Errorf("Expected FilterInfo '%s', got '%s'", tt.filterInfo, statusVM.FilterInfo)
			}
		})
	}
}

// TestCreateHelpViewModel tests help view model creation
func TestCreateHelpViewModel(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	helpVM := vm.CreateHelpViewModel()
	
	if helpVM.Title != "PR Compass - Commands" {
		t.Errorf("Expected title 'PR Compass - Commands', got '%s'", helpVM.Title)
	}
	
	if !helpVM.ShowClose {
		t.Error("Expected ShowClose to be true")
	}
	
	if len(helpVM.Sections) == 0 {
		t.Error("Expected help sections to be created")
	}
	
	// Check that we have expected sections
	expectedSections := []string{"Navigation", "Filtering", "Actions"}
	if len(helpVM.Sections) != len(expectedSections) {
		t.Errorf("Expected %d sections, got %d", len(expectedSections), len(helpVM.Sections))
	}
	
	for i, section := range helpVM.Sections {
		if section.Title != expectedSections[i] {
			t.Errorf("Expected section %d title '%s', got '%s'", i, expectedSections[i], section.Title)
		}
		
		if len(section.Items) == 0 {
			t.Errorf("Expected section '%s' to have items", section.Title)
		}
	}
}

// TestValidateTabOperation tests tab operation validation
func TestValidateTabOperation(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	// Create test tab manager with multiple tabs
	tabManager := NewTabManager("test-token")
	tab1 := &TabConfig{Name: "Tab 1", Mode: "repos", Repos: []string{"test/repo1"}}
	tab2 := &TabConfig{Name: "Tab 2", Mode: "repos", Repos: []string{"test/repo2"}}
	tab3 := &TabConfig{Name: "Tab 3", Mode: "repos", Repos: []string{"test/repo3"}}
	
	tabManager.AddTab(tab1)
	tabManager.AddTab(tab2)
	tabManager.AddTab(tab3)
	
	tests := []struct {
		name          string
		operation     string
		targetIndex   int
		expectedValid bool
		expectedError string
	}{
		{
			name:          "valid tab switch",
			operation:     "switch",
			targetIndex:   1,
			expectedValid: true,
			expectedError: "",
		},
		{
			name:          "invalid tab switch - out of range",
			operation:     "switch",
			targetIndex:   5,
			expectedValid: false,
			expectedError: "Invalid tab index",
		},
		{
			name:          "valid tab close",
			operation:     "close",
			targetIndex:   1,
			expectedValid: true,
			expectedError: "",
		},
		{
			name:          "invalid operation",
			operation:     "unknown",
			targetIndex:   1,
			expectedValid: false,
			expectedError: "Unknown operation",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.ValidateTabOperation(tt.operation, tabManager, tt.targetIndex)
			
			if result.IsValid != tt.expectedValid {
				t.Errorf("Expected IsValid %v, got %v", tt.expectedValid, result.IsValid)
			}
			
			if tt.expectedError != "" && result.Error != tt.expectedError {
				t.Errorf("Expected Error '%s', got '%s'", tt.expectedError, result.Error)
			}
		})
	}
}

// TestValidateTabOperationEdgeCases tests edge cases for tab operations
func TestValidateTabOperationEdgeCases(t *testing.T) {
	controller := NewUIController("test-token")
	vm := NewViewModel(controller)
	
	// Create tab manager with only one tab
	tabManager := NewTabManager("test-token")
	tab1 := &TabConfig{Name: "Only Tab", Mode: "repos", Repos: []string{"test/repo1"}}
	tabManager.AddTab(tab1)
	
	// Try to close the last tab
	result := vm.ValidateTabOperation("close", tabManager, 0)
	
	if result.IsValid {
		t.Error("Should not allow closing the last tab")
	}
	
	if result.Error != "Cannot close the last tab" {
		t.Errorf("Expected error 'Cannot close the last tab', got '%s'", result.Error)
	}
}
package ui

import (
	"fmt"

	"github.com/bjess9/pr-compass/internal/ui/components"
	"github.com/bjess9/pr-compass/internal/ui/types"
	"github.com/charmbracelet/bubbles/table"
	gh "github.com/google/go-github/v55/github"
)

// ViewModel represents the presentation layer for the UI
// This separates UI state from business logic for better testability
type ViewModel struct {
	controller *UIController
}

// NewViewModel creates a new view model
func NewViewModel(controller *UIController) *ViewModel {
	return &ViewModel{
		controller: controller,
	}
}

// TabViewModel represents the view state for a single tab
type TabViewModel struct {
	Name            string
	IsActive        bool
	IsLoading       bool
	HasError        bool
	PRCount         int
	FilteredPRCount int
	StatusMessage   string
}

// CreateTabViewModels creates view models for all tabs
func (vm *ViewModel) CreateTabViewModels(tabManager *TabManager) []TabViewModel {
	viewModels := make([]TabViewModel, len(tabManager.Tabs))

	for i, tab := range tabManager.Tabs {
		viewModels[i] = TabViewModel{
			Name:            tab.Config.Name,
			IsActive:        i == tabManager.ActiveTabIdx,
			IsLoading:       !tab.Loaded,
			HasError:        tab.Error != nil,
			PRCount:         len(tab.PRs),
			FilteredPRCount: len(tab.FilteredPRs),
			StatusMessage:   tab.StatusMsg,
		}
	}

	return viewModels
}

// FilterViewModel represents filter state for the view
type FilterViewModel struct {
	IsActive    bool
	Mode        string
	Value       string
	Description string
}

// CreateFilterViewModel creates a filter view model
func (vm *ViewModel) CreateFilterViewModel(tab *TabState) FilterViewModel {
	return FilterViewModel{
		IsActive:    tab.FilterMode != "",
		Mode:        tab.FilterMode,
		Value:       tab.FilterValue,
		Description: vm.formatFilterDescription(tab.FilterMode, tab.FilterValue),
	}
}

// formatFilterDescription creates a human-readable filter description
func (vm *ViewModel) formatFilterDescription(mode, value string) string {
	if mode == "" {
		return ""
	}

	switch mode {
	case "author":
		return fmt.Sprintf("by author: %s", value)
	case "status":
		return fmt.Sprintf("by status: %s", value)
	case "draft":
		return "drafts only"
	default:
		return fmt.Sprintf("%s: %s", mode, value)
	}
}

// TableViewModel represents table rendering state
type TableViewModel struct {
	Rows          []table.Row
	SelectedIndex int
	Height        int
	Width         int
}

// CreateTableViewModel creates a table view model
func (vm *ViewModel) CreateTableViewModel(prs []*gh.PullRequest, enhancementQueue map[int]bool, selectedIndex, height, width int) TableViewModel {
	// Use the table component to create rows
	tableComponent := components.NewTableComponent()

	// Convert to PRData format for the table component
	prDataList := make([]*types.PRData, len(prs))
	for i, pr := range prs {
		prDataList[i] = &types.PRData{
			PullRequest: pr,
			Enhanced:    nil, // TODO: Add enhanced data if available
		}
	}

	rows := tableComponent.CreateRows(prDataList, enhancementQueue)

	return TableViewModel{
		Rows:          rows,
		SelectedIndex: selectedIndex,
		Height:        height,
		Width:         width,
	}
}

// StatusViewModel represents status bar information
type StatusViewModel struct {
	Message         string
	ShowSpinner     bool
	SpinnerFrame    string
	APILimitInfo    string
	HasActiveFilter bool
	FilterInfo      string
}

// CreateStatusViewModel creates a status view model
func (vm *ViewModel) CreateStatusViewModel(spinnerIndex int, statusMsg string, filterInfo string, apiInfo string) StatusViewModel {
	return StatusViewModel{
		Message:         statusMsg,
		ShowSpinner:     statusMsg != "" && (statusMsg == "Loading..." || statusMsg == "Refreshing..."),
		SpinnerFrame:    vm.controller.GetSpinnerFrame(spinnerIndex),
		APILimitInfo:    apiInfo,
		HasActiveFilter: filterInfo != "",
		FilterInfo:      filterInfo,
	}
}

// HelpViewModel represents help screen data
type HelpViewModel struct {
	Title     string
	Sections  []HelpSection
	ShowClose bool
}

// HelpSection represents a section in the help screen
type HelpSection struct {
	Title string
	Items []HelpItem
}

// HelpItem represents a single help item
type HelpItem struct {
	Key         string
	Description string
}

// CreateHelpViewModel creates a help view model
func (vm *ViewModel) CreateHelpViewModel() HelpViewModel {
	return HelpViewModel{
		Title:     "PR Compass - Commands",
		ShowClose: true,
		Sections: []HelpSection{
			{
				Title: "Navigation",
				Items: []HelpItem{
					{"↑/↓, j/k", "Navigate PRs"},
					{"Tab/Shift+Tab", "Switch tabs"},
					{"Ctrl+1-9", "Switch to tab number"},
					{"Enter", "Open PR in browser"},
				},
			},
			{
				Title: "Filtering",
				Items: []HelpItem{
					{"a", "Filter by author"},
					{"s", "Filter by status"},
					{"d", "Toggle draft filter"},
					{"c", "Clear filters"},
				},
			},
			{
				Title: "Actions",
				Items: []HelpItem{
					{"r", "Refresh PRs"},
					{"h, ?", "Show/hide help"},
					{"q, Ctrl+C", "Quit application"},
				},
			},
		},
	}
}

// ValidationResult represents validation results for UI operations
type ValidationResult struct {
	IsValid bool
	Error   string
}

// ValidateTabOperation validates tab operations
func (vm *ViewModel) ValidateTabOperation(operation string, tabManager *TabManager, targetIndex int) ValidationResult {
	switch operation {
	case "switch":
		if vm.controller.ValidateTabSwitch(tabManager.ActiveTabIdx, targetIndex, len(tabManager.Tabs)) {
			return ValidationResult{IsValid: true}
		}
		return ValidationResult{IsValid: false, Error: "Invalid tab index"}

	case "close":
		if len(tabManager.Tabs) <= 1 {
			return ValidationResult{IsValid: false, Error: "Cannot close the last tab"}
		}
		if targetIndex < 0 || targetIndex >= len(tabManager.Tabs) {
			return ValidationResult{IsValid: false, Error: "Invalid tab index"}
		}
		return ValidationResult{IsValid: true}

	default:
		return ValidationResult{IsValid: false, Error: "Unknown operation"}
	}
}

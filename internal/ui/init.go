package ui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// InitializedMultiTabModel wraps a multi-tab model that needs initial data fetching
type InitializedMultiTabModel struct {
	*MultiTabModel
}

// Init triggers the initial data fetch for tabs after they've been added
func (m *InitializedMultiTabModel) Init() tea.Cmd {
	// Send initialization message to trigger data fetching
	return func() tea.Msg {
		return initializeTabsMsg{}
	}
}

// InitialModelAuto automatically detects whether to use single-tab or multi-tab mode
func InitialModelAuto(token string) tea.Model {
	// Try to load multi-tab configuration
	multiConfig, err := LoadMultiTabConfig()
	if err != nil {
		log.Printf("Failed to load multi-tab config, falling back to single-tab: %v", err)
		singleModel := InitialModel(token)
		return &singleModel
	}

	// If we have multiple tabs, use multi-tab model
	if len(multiConfig.Tabs) > 1 {
		return InitialMultiTabModel(token, multiConfig)
	}

	// If we have only one tab, use single-tab model for better performance
	if len(multiConfig.Tabs) == 1 {
		// For single tab, we can use the existing optimized single-tab model
		singleModel := InitialModel(token)
		return &singleModel
	}

	// No tabs configured - shouldn't happen due to fallback logic, but handle gracefully
	log.Printf("No tabs found in config, creating default single-tab model")
	singleModel := InitialModel(token)
	return &singleModel
}

// InitialMultiTabModel creates a new multi-tab model with the given configuration
func InitialMultiTabModel(token string, multiConfig *MultiTabConfig) tea.Model {
	model := NewMultiTabModel(token)

	// Add all configured tabs
	for _, tabConfig := range multiConfig.Tabs {
		// Make a copy of the tab config to avoid pointer issues
		tabConfigCopy := tabConfig
		model.TabManager.AddTab(&tabConfigCopy)
	}

	// Set global refresh interval
	model.TabManager.GlobalRefreshInterval = multiConfig.RefreshIntervalMinutes
	if model.TabManager.GlobalRefreshInterval == 0 {
		model.TabManager.GlobalRefreshInterval = 5
	}

	return &InitializedMultiTabModel{
		MultiTabModel: model,
	}
}

// InitialModelMultiTab is the entry point for multi-tab mode
// It automatically detects single vs multi-tab configuration
func InitialModelMultiTab(token string) tea.Model {
	return InitialModelAuto(token)
}
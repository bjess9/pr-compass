package ui

import (
	"github.com/bjess9/pr-compass/internal/cache"
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
		// Create a default single tab configuration and use multi-tab model
		defaultConfig := &MultiTabConfig{
			RefreshIntervalMinutes: 5,
			Tabs: []TabConfig{
				{
					Name:                   "Main",
					Mode:                   "repos", // Will auto-detect actual mode
					RefreshIntervalMinutes: 5,
					MaxPRs:                 50,
				},
			},
		}
		return InitialMultiTabModel(token, defaultConfig)
	}

	// Always use multi-tab model for consistency, even with single tab
	if len(multiConfig.Tabs) >= 1 {
		return InitialMultiTabModel(token, multiConfig)
	}

	// No tabs configured - shouldn't happen due to fallback logic, but handle gracefully
	// Create a default single tab configuration and use multi-tab model
	defaultConfig := &MultiTabConfig{
		RefreshIntervalMinutes: 5,
		Tabs: []TabConfig{
			{
				Name:                   "Main",
				Mode:                   "repos", // Will auto-detect actual mode
				RefreshIntervalMinutes: 5,
				MaxPRs:                 50,
			},
		},
	}
	return InitialMultiTabModel(token, defaultConfig)
}

// InitialMultiTabModel creates a new multi-tab model with the given configuration
func InitialMultiTabModel(token string, multiConfig *MultiTabConfig) tea.Model {
	// Initialize cache (could be nil for simpler cases)
	var prCache *cache.PRCache = nil // For now, no caching in initial model

	model := NewMultiTabModel(token, prCache)

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

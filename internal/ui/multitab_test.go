package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTabConfig tests tab configuration creation and conversion
func TestTabConfig(t *testing.T) {
	config := &TabConfig{
		Name:                   "Test Tab",
		Mode:                   "repos",
		Repos:                  []string{"test/repo1", "test/repo2"},
		ExcludeBots:            true,
		IncludeDrafts:          false,
		RefreshIntervalMinutes: 10,
	}

	// Test conversion to standard config
	standardConfig := config.ConvertToConfig()
	if standardConfig.Mode != "repos" {
		t.Errorf("Expected mode 'repos', got '%s'", standardConfig.Mode)
	}
	if len(standardConfig.Repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(standardConfig.Repos))
	}
	if !standardConfig.ExcludeBots {
		t.Error("Expected ExcludeBots to be true")
	}
	if standardConfig.RefreshIntervalMinutes != 10 {
		t.Errorf("Expected refresh interval 10, got %d", standardConfig.RefreshIntervalMinutes)
	}
}

// TestTabManager tests tab manager functionality
func TestTabManager(t *testing.T) {
	manager := NewTabManager("test-token")

	// Test initial state
	if manager.GetTabCount() != 0 {
		t.Errorf("Expected 0 tabs initially, got %d", manager.GetTabCount())
	}
	if manager.GetActiveTab() != nil {
		t.Error("Expected no active tab initially")
	}

	// Add first tab
	tab1 := &TabConfig{
		Name:  "Tab 1",
		Mode:  "repos",
		Repos: []string{"test/repo"},
	}
	manager.AddTab(tab1)

	if manager.GetTabCount() != 1 {
		t.Errorf("Expected 1 tab after adding, got %d", manager.GetTabCount())
	}
	if manager.GetActiveTab() == nil {
		t.Error("Expected active tab after adding first tab")
	}
	if manager.GetActiveTab().Config.Name != "Tab 1" {
		t.Errorf("Expected active tab name 'Tab 1', got '%s'", manager.GetActiveTab().Config.Name)
	}

	// Add second tab
	tab2 := &TabConfig{
		Name:         "Tab 2",
		Mode:         "organization",
		Organization: "test-org",
	}
	manager.AddTab(tab2)

	if manager.GetTabCount() != 2 {
		t.Errorf("Expected 2 tabs after adding second, got %d", manager.GetTabCount())
	}

	// Test tab switching
	manager.NextTab()
	if manager.ActiveTabIdx != 1 {
		t.Errorf("Expected active tab index 1 after NextTab, got %d", manager.ActiveTabIdx)
	}
	if manager.GetActiveTab().Config.Name != "Tab 2" {
		t.Errorf("Expected active tab name 'Tab 2', got '%s'", manager.GetActiveTab().Config.Name)
	}

	manager.PrevTab()
	if manager.ActiveTabIdx != 0 {
		t.Errorf("Expected active tab index 0 after PrevTab, got %d", manager.ActiveTabIdx)
	}

	// Test direct tab switching
	if !manager.SwitchToTab(1) {
		t.Error("Expected successful switch to tab 1")
	}
	if manager.ActiveTabIdx != 1 {
		t.Errorf("Expected active tab index 1 after direct switch, got %d", manager.ActiveTabIdx)
	}
	if manager.SwitchToTab(5) {
		t.Error("Expected failed switch to non-existent tab 5")
	}

	// Test tab names
	names := manager.GetTabNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 tab names, got %d", len(names))
	}
	if names[0] != "Tab 1" || names[1] != "Tab 2" {
		t.Errorf("Expected tab names ['Tab 1', 'Tab 2'], got %v", names)
	}

	// Test tab closing
	if !manager.CloseTab(1) {
		t.Error("Expected successful close of tab 1")
	}
	if manager.GetTabCount() != 1 {
		t.Errorf("Expected 1 tab after closing, got %d", manager.GetTabCount())
	}
	if manager.ActiveTabIdx != 0 {
		t.Errorf("Expected active tab index 0 after closing tab 1, got %d", manager.ActiveTabIdx)
	}

	// Test can't close last tab
	if manager.CloseTab(0) {
		t.Error("Expected failed close of last remaining tab")
	}
	if manager.GetTabCount() != 1 {
		t.Errorf("Expected 1 tab still remaining, got %d", manager.GetTabCount())
	}
}

// TestMultiTabModel tests the multi-tab model functionality
func TestMultiTabModel(t *testing.T) {
	model := NewMultiTabModel("test-token", nil)

	// Test initial state
	if model.Width != 120 {
		t.Errorf("Expected default width 120, got %d", model.Width)
	}
	if model.Height != 30 {
		t.Errorf("Expected default height 30, got %d", model.Height)
	}
	if model.TabManager == nil {
		t.Fatal("Expected TabManager to be initialized")
	}

	// Test with no tabs
	view := model.View()
	if !strings.Contains(view, "No tabs configured") {
		t.Error("Expected 'No tabs configured' message when no tabs present")
	}

	// Add tabs
	tab1 := &TabConfig{Name: "Test1", Mode: "repos", Repos: []string{"test/repo"}}
	tab2 := &TabConfig{Name: "Test2", Mode: "organization", Organization: "test"}
	model.TabManager.AddTab(tab1)
	model.TabManager.AddTab(tab2)

	// Test view with tabs
	view = model.View()
	if !strings.Contains(view, "Test1") {
		t.Error("Expected tab name 'Test1' in view")
	}
	if !strings.Contains(view, "Test2") {
		t.Error("Expected tab name 'Test2' in view")
	}

	// Test window resize
	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 30}
	updatedModel, _ := model.Update(resizeMsg)
	multiTabModel := updatedModel.(*MultiTabModel)
	if multiTabModel.Width != 120 {
		t.Errorf("Expected width 120 after resize, got %d", multiTabModel.Width)
	}
	if multiTabModel.Height != 30 {
		t.Errorf("Expected height 30 after resize, got %d", multiTabModel.Height)
	}

	// Test tab switching via keyboard
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ = model.Update(tabMsg)
	multiTabModel = updatedModel.(*MultiTabModel)
	if multiTabModel.TabManager.ActiveTabIdx != 1 {
		t.Errorf("Expected active tab 1 after Tab key, got %d", multiTabModel.TabManager.ActiveTabIdx)
	}

	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updatedModel, _ = multiTabModel.Update(shiftTabMsg)
	multiTabModel = updatedModel.(*MultiTabModel)
	if multiTabModel.TabManager.ActiveTabIdx != 0 {
		t.Errorf("Expected active tab 0 after Shift+Tab, got %d", multiTabModel.TabManager.ActiveTabIdx)
	}
}

// TestMultiTabConfiguration tests loading multi-tab configurations
func TestMultiTabConfiguration(t *testing.T) {
	// Test multi-tab config structure
	multiConfig := &MultiTabConfig{
		RefreshIntervalMinutes: 5,
		Tabs: []TabConfig{
			{
				Name:          "Tab 1",
				Mode:          "repos",
				Repos:         []string{"test/repo1"},
				ExcludeBots:   true,
				IncludeDrafts: true,
			},
			{
				Name:                   "Tab 2",
				Mode:                   "organization",
				Organization:           "test-org",
				RefreshIntervalMinutes: 10,
			},
		},
	}

	if len(multiConfig.Tabs) != 2 {
		t.Errorf("Expected 2 tabs in config, got %d", len(multiConfig.Tabs))
	}

	// Test first tab
	tab1 := multiConfig.Tabs[0]
	if tab1.Name != "Tab 1" {
		t.Errorf("Expected tab name 'Tab 1', got '%s'", tab1.Name)
	}
	if tab1.Mode != "repos" {
		t.Errorf("Expected mode 'repos', got '%s'", tab1.Mode)
	}
	if len(tab1.Repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(tab1.Repos))
	}

	// Test second tab
	tab2 := multiConfig.Tabs[1]
	if tab2.Name != "Tab 2" {
		t.Errorf("Expected tab name 'Tab 2', got '%s'", tab2.Name)
	}
	if tab2.Mode != "organization" {
		t.Errorf("Expected mode 'organization', got '%s'", tab2.Mode)
	}
	if tab2.Organization != "test-org" {
		t.Errorf("Expected organization 'test-org', got '%s'", tab2.Organization)
	}
}

// TestTabStateInitialization tests tab state creation
func TestTabStateInitialization(t *testing.T) {
	config := &TabConfig{
		Name:                   "Test Tab",
		Mode:                   "repos",
		Repos:                  []string{"test/repo"},
		RefreshIntervalMinutes: 5,
	}

	tabState := NewTabState(config, "test-token")

	// Test basic initialization
	if tabState.Config.Name != "Test Tab" {
		t.Errorf("Expected tab name 'Test Tab', got '%s'", tabState.Config.Name)
	}
	if tabState.Loaded {
		t.Error("Expected tab to not be loaded initially")
	}
	if tabState.Error != nil {
		t.Errorf("Expected no error initially, got %v", tabState.Error)
	}

	// Test that required components are initialized
	if tabState.EnhancedData == nil {
		t.Error("Expected EnhancedData to be initialized")
	}
	if tabState.EnhancementQueue == nil {
		t.Error("Expected EnhancementQueue to be initialized")
	}
	if tabState.Ctx == nil {
		t.Error("Expected Context to be initialized")
	}
	if tabState.Cancel == nil {
		t.Error("Expected Cancel function to be initialized")
	}

	// Test that table is properly configured (height may vary based on implementation)
	if tabState.Table.Height() < 5 {
		t.Errorf("Expected table height >= 5, got %d", tabState.Table.Height())
	}

	// Test that load time is recent
	if time.Since(tabState.LoadTime) > time.Second {
		t.Error("Expected LoadTime to be very recent")
	}

	// Clean up context
	tabState.Cancel()
}

// TestTabCleanup tests proper resource cleanup
func TestTabCleanup(t *testing.T) {
	manager := NewTabManager("test-token")

	// Add multiple tabs
	for i := 0; i < 3; i++ {
		config := &TabConfig{
			Name:  "Tab " + string(rune('1'+i)),
			Mode:  "repos",
			Repos: []string{"test/repo"},
		}
		manager.AddTab(config)
	}

	if manager.GetTabCount() != 3 {
		t.Errorf("Expected 3 tabs, got %d", manager.GetTabCount())
	}

	// Test cleanup
	manager.Cleanup()

	// Verify all contexts are canceled (this is a bit tricky to test directly)
	// We'll just verify the cleanup method doesn't panic and runs successfully
	if manager.GetTabCount() != 3 {
		t.Error("Expected cleanup to not change tab count")
	}
}

// TestTabManagerEdgeCases tests edge cases and error conditions
func TestTabManagerEdgeCases(t *testing.T) {
	manager := NewTabManager("test-token")

	// Test operations on empty manager
	if manager.GetActiveTab() != nil {
		t.Error("Expected nil active tab when no tabs exist")
	}

	manager.NextTab() // Should not panic
	manager.PrevTab() // Should not panic

	if manager.ActiveTabIdx != 0 {
		t.Error("Expected ActiveTabIdx to remain 0 when no tabs exist")
	}

	// Test with single tab
	config := &TabConfig{Name: "Single", Mode: "repos", Repos: []string{"test/repo"}}
	manager.AddTab(config)

	manager.NextTab() // Should stay on same tab
	if manager.ActiveTabIdx != 0 {
		t.Error("Expected single tab to remain active after NextTab")
	}

	manager.PrevTab() // Should stay on same tab
	if manager.ActiveTabIdx != 0 {
		t.Error("Expected single tab to remain active after PrevTab")
	}

	// Test invalid tab operations
	if manager.CloseTab(-1) {
		t.Error("Expected failure when closing invalid tab index -1")
	}
	if manager.CloseTab(10) {
		t.Error("Expected failure when closing invalid tab index 10")
	}
	if manager.SwitchToTab(-1) {
		t.Error("Expected failure when switching to invalid tab index -1")
	}
	if manager.SwitchToTab(10) {
		t.Error("Expected failure when switching to invalid tab index 10")
	}
}

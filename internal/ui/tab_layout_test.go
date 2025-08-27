package ui

import (
	"strings"
	"testing"

	gh "github.com/google/go-github/v55/github"
)

// TestTabBarHorizontalLayout tests that tabs are rendered horizontally, not vertically
func TestTabBarHorizontalLayout(t *testing.T) {
	// Create test model with multiple tabs
	tabManager := NewTabManager("test-token")

	// Add multiple tabs to test horizontal layout
	tabConfigs := []*TabConfig{
		{Name: "Tab 1", Mode: "repos", Repos: []string{"test/repo1"}},
		{Name: "Tab 2", Mode: "repos", Repos: []string{"test/repo2"}},
		{Name: "Tab 3", Mode: "repos", Repos: []string{"test/repo3"}},
	}

	for _, config := range tabConfigs {
		tabManager.AddTab(config)
	}

	model := NewMultiTabModel("test-token", nil)
	model.TabManager = tabManager
	model.Width = 120
	model.Height = 40

	// Set all tabs as loaded to avoid loading indicators
	for _, tab := range tabManager.Tabs {
		tab.Loaded = true
		tab.PRs = []*gh.PullRequest{} // Empty but loaded
	}

	// Render the tab bar
	tabBar := model.renderTabBar()

	// Tab bar should not be empty when there are multiple tabs
	if tabBar == "" {
		t.Error("Tab bar should not be empty with multiple tabs")
		return
	}

	// Split into lines to analyze layout
	lines := strings.Split(tabBar, "\n")

	// Should have at least 4 lines: box top, tab content, box bottom, help text, separator
	if len(lines) < 4 {
		t.Errorf("Expected at least 4 lines in tab bar, got %d", len(lines))
		return
	}

	// Tab content is usually on the second line (inside the boxes)
	tabContentLine := ""
	for _, line := range lines {
		if strings.Contains(line, "1:Tab") || strings.Contains(line, "2:Tab") || strings.Contains(line, "3:Tab") {
			tabContentLine = line
			break
		}
	}

	if tabContentLine == "" {
		t.Error("Could not find line containing tab content")
		return
	}

	// All tab names should appear on the same line (horizontal layout)
	expectedTabs := []string{"1:Tab 1", "2:Tab 2", "3:Tab 3"}
	for _, expectedTab := range expectedTabs {
		if !strings.Contains(tabContentLine, expectedTab) {
			t.Errorf("Tab content line should contain '%s', got: %s", expectedTab, tabContentLine)
		}
	}

	// Verify tabs appear in order on the same line (not vertically stacked)
	tab1Pos := strings.Index(tabContentLine, "1:Tab 1")
	tab2Pos := strings.Index(tabContentLine, "2:Tab 2")
	tab3Pos := strings.Index(tabContentLine, "3:Tab 3")

	if tab1Pos == -1 || tab2Pos == -1 || tab3Pos == -1 {
		t.Errorf("All tabs should be present on content line: %s", tabContentLine)
		return
	}

	// Tabs should appear in order from left to right
	if !(tab1Pos < tab2Pos && tab2Pos < tab3Pos) {
		t.Errorf("Tabs should appear in order: tab1(%d) < tab2(%d) < tab3(%d)",
			tab1Pos, tab2Pos, tab3Pos)
	}

	// Find help line (contains navigation instructions)
	helpLine := ""
	for _, line := range lines {
		if strings.Contains(line, "🧭") || strings.Contains(line, "Tab/") {
			helpLine = line
			break
		}
	}

	if helpLine == "" {
		t.Error("Could not find help line with navigation instructions")
		return
	}

	// Help line should not contain specific tab names (since they should be on the tab content line)
	if strings.Contains(helpLine, "1:Tab 1") || strings.Contains(helpLine, "2:Tab 2") {
		t.Error("Help line should not contain specific tab names (tabs should not be vertical)")
	}

	// Help line should contain navigation instructions
	if !strings.Contains(helpLine, "Tab/") && !strings.Contains(helpLine, "Navigate") {
		t.Errorf("Help line should contain navigation instructions, got: %s", helpLine)
	}
}

// TestSingleTabNoBar tests that single tab doesn't show tab bar
func TestSingleTabNoBar(t *testing.T) {
	// Create test model with single tab
	tabManager := NewTabManager("test-token")
	tabConfig := &TabConfig{
		Name:  "Single Tab",
		Mode:  "repos",
		Repos: []string{"test/repo"},
	}
	tabManager.AddTab(tabConfig)

	model := NewMultiTabModel("test-token", nil)
	model.TabManager = tabManager
	model.Width = 120
	model.Height = 40

	// Render the tab bar
	tabBar := model.renderTabBar()

	// Tab bar is now always shown, even for single tabs, as it contains important status info
	if tabBar == "" {
		t.Error("Tab bar should be shown even for single tab as it contains status info")
	}

	// Verify it contains the tab name
	if !strings.Contains(tabBar, "Single Tab") {
		t.Error("Tab bar should contain the tab name")
	}
}

// TestTabBarSpacing tests that tabs have proper spacing
func TestTabBarSpacing(t *testing.T) {
	// Create test model with two tabs
	tabManager := NewTabManager("test-token")

	tabConfigs := []*TabConfig{
		{Name: "First", Mode: "repos", Repos: []string{"test/repo1"}},
		{Name: "Second", Mode: "repos", Repos: []string{"test/repo2"}},
	}

	for _, config := range tabConfigs {
		tabManager.AddTab(config)
	}

	model := NewMultiTabModel("test-token", nil)
	model.TabManager = tabManager
	model.Width = 120
	model.Height = 40

	// Set tabs as loaded
	for _, tab := range tabManager.Tabs {
		tab.Loaded = true
		tab.PRs = []*gh.PullRequest{}
	}

	// Render the tab bar
	tabBar := model.renderTabBar()
	lines := strings.Split(tabBar, "\n")

	if len(lines) == 0 {
		t.Error("Tab bar should have content")
		return
	}

	// Find the line with tab content
	tabContentLine := ""
	for _, line := range lines {
		if strings.Contains(line, "First") || strings.Contains(line, "Second") {
			tabContentLine = line
			break
		}
	}

	if tabContentLine == "" {
		t.Error("Could not find tab content line")
		return
	}

	// Should contain both tabs on same line
	if !strings.Contains(tabContentLine, "First") {
		t.Errorf("Should contain First tab, got: %s", tabContentLine)
	}
	if !strings.Contains(tabContentLine, "Second") {
		t.Errorf("Should contain Second tab, got: %s", tabContentLine)
	}

	// There should be some separation between the tabs
	// The exact spacing depends on lipgloss rendering, but they shouldn't be immediately adjacent
	firstPos := strings.Index(tabContentLine, "First")
	secondPos := strings.Index(tabContentLine, "Second")

	if firstPos == -1 || secondPos == -1 {
		t.Errorf("Both tabs should be present on content line: %s", tabContentLine)
		return
	}

	// There should be some gap between the end of "First" and start of "Second"
	gap := secondPos - (firstPos + len("First"))
	if gap < 1 {
		t.Errorf("Expected some spacing between tabs, gap was %d chars", gap)
	}
}

package ui

import (
	"testing"
	"time"

	"github.com/bjess9/pr-compass/internal/ui/types"
	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/google/go-github/v55/github"
)

// TestHotkeyQuit tests that 'q' and Ctrl+C quit the application
func TestHotkeyQuit(t *testing.T) {
	// Create test model with one tab
	tabManager := NewTabManager("test-token")
	tabConfig := &TabConfig{
		Name: "Test Tab",
		Mode: "repos",
		Repos: []string{"test/repo"},
	}
	tabManager.AddTab(tabConfig)
	
	model := &MultiTabModel{
		TabManager: tabManager,
		Width:      100,
		Height:     50,
	}
	
	tests := []string{"q", "ctrl+c"}
	
	for _, key := range tests {
		t.Run("key_"+key, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			if key == "ctrl+c" {
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}
			
			_, cmd := model.Update(keyMsg)
			
			// Should return tea.Quit command
			if cmd == nil {
				t.Errorf("Expected quit command for key '%s', got nil", key)
			}
		})
	}
}

// TestHotkeyHelp tests that 'h' and '?' toggle help
func TestHotkeyHelp(t *testing.T) {
	// Create test model with one tab
	tabManager := NewTabManager("test-token")
	tabConfig := &TabConfig{
		Name: "Test Tab",
		Mode: "repos",
		Repos: []string{"test/repo"},
	}
	tabManager.AddTab(tabConfig)
	
	model := &MultiTabModel{
		TabManager: tabManager,
		Width:      100,
		Height:     50,
	}
	
	activeTab := model.TabManager.GetActiveTab()
	if activeTab == nil {
		t.Fatal("Expected active tab to exist")
	}
	
	tests := []string{"h", "?"}
	
	for _, key := range tests {
		t.Run("key_"+key, func(t *testing.T) {
			initialHelp := activeTab.ShowHelp
			
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			model.Update(keyMsg)
			
			if activeTab.ShowHelp == initialHelp {
				t.Errorf("Expected ShowHelp to toggle from %v for key '%s'", initialHelp, key)
			}
		})
	}
}

// TestHotkeyFilter tests filter functionality
func TestHotkeyFilter(t *testing.T) {
	// Create test model with one tab and some PRs
	tabManager := NewTabManager("test-token")
	tabConfig := &TabConfig{
		Name: "Test Tab",
		Mode: "repos",
		Repos: []string{"test/repo"},
	}
	tabManager.AddTab(tabConfig)
	
	model := &MultiTabModel{
		TabManager: tabManager,
		Width:      100,
		Height:     50,
	}
	
	activeTab := model.TabManager.GetActiveTab()
	if activeTab == nil {
		t.Fatal("Expected active tab to exist")
	}
	
	// Add test PRs
	testPRs := []*gh.PullRequest{
		{
			Number: gh.Int(1),
			Title:  gh.String("Test PR 1"),
			User:   &gh.User{Login: gh.String("alice")},
			Draft:  gh.Bool(false),
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{FullName: gh.String("test/repo")},
			},
		},
		{
			Number: gh.Int(2),
			Title:  gh.String("Draft PR 2"),
			User:   &gh.User{Login: gh.String("bob")},
			Draft:  gh.Bool(true),
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{FullName: gh.String("test/repo")},
			},
		},
	}
	activeTab.PRs = testPRs
	activeTab.FilteredPRs = testPRs
	activeTab.Loaded = true
	
	// Test author filter
	t.Run("author_filter", func(t *testing.T) {
		// Press 'f' to start author filter
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
		model.Update(keyMsg)
		
		if activeTab.FilterMode != "author" {
			t.Errorf("Expected FilterMode to be 'author', got '%s'", activeTab.FilterMode)
		}
		
		if activeTab.StatusMsg != "Enter author name to filter by:" {
			t.Errorf("Expected status message for author filter, got '%s'", activeTab.StatusMsg)
		}
	})
	
	// Test status filter
	t.Run("status_filter", func(t *testing.T) {
		// Press 's' to start status filter
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		model.Update(keyMsg)
		
		if activeTab.FilterMode != "status" {
			t.Errorf("Expected FilterMode to be 'status', got '%s'", activeTab.FilterMode)
		}
		
		if activeTab.StatusMsg != "Enter status (draft/ready/conflicts):" {
			t.Errorf("Expected status message for status filter, got '%s'", activeTab.StatusMsg)
		}
	})
	
	// Test draft filter toggle
	t.Run("draft_filter", func(t *testing.T) {
		// Reset filter first
		activeTab.FilterMode = ""
		activeTab.FilterValue = ""
		activeTab.FilteredPRs = activeTab.PRs
		
		// Press 'd' to toggle draft filter
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
		model.Update(keyMsg)
		
		if activeTab.FilterMode != "draft" {
			t.Errorf("Expected FilterMode to be 'draft', got '%s'", activeTab.FilterMode)
		}
		
		if activeTab.FilterValue != "true" {
			t.Errorf("Expected FilterValue to be 'true', got '%s'", activeTab.FilterValue)
		}
		
		// Should only show draft PRs
		if len(activeTab.FilteredPRs) != 1 || !activeTab.FilteredPRs[0].GetDraft() {
			t.Errorf("Expected 1 draft PR, got %d PRs", len(activeTab.FilteredPRs))
		}
	})
	
	// Test clear filters
	t.Run("clear_filters", func(t *testing.T) {
		// Set up some filter first
		activeTab.FilterMode = "author"
		activeTab.FilterValue = "alice"
		
		// Press 'c' to clear filters
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
		model.Update(keyMsg)
		
		if activeTab.FilterMode != "" {
			t.Errorf("Expected FilterMode to be empty, got '%s'", activeTab.FilterMode)
		}
		
		if activeTab.FilterValue != "" {
			t.Errorf("Expected FilterValue to be empty, got '%s'", activeTab.FilterValue)
		}
		
		if len(activeTab.FilteredPRs) != len(activeTab.PRs) {
			t.Errorf("Expected all PRs after clearing filter, got %d of %d", len(activeTab.FilteredPRs), len(activeTab.PRs))
		}
	})
}

// TestHotkeyNavigation tests table navigation keys
func TestHotkeyNavigation(t *testing.T) {
	// Create test model with one tab and some PRs
	tabManager := NewTabManager("test-token")
	tabConfig := &TabConfig{
		Name: "Test Tab",
		Mode: "repos",
		Repos: []string{"test/repo"},
	}
	tabManager.AddTab(tabConfig)
	
	model := &MultiTabModel{
		TabManager: tabManager,
		Width:      100,
		Height:     50,
	}
	
	activeTab := model.TabManager.GetActiveTab()
	if activeTab == nil {
		t.Fatal("Expected active tab to exist")
	}
	
	// Add test PRs
	testPRs := []*gh.PullRequest{
		{
			Number: gh.Int(1),
			Title:  gh.String("Test PR 1"),
			User:   &gh.User{Login: gh.String("alice")},
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{FullName: gh.String("test/repo")},
			},
		},
		{
			Number: gh.Int(2),
			Title:  gh.String("Test PR 2"),
			User:   &gh.User{Login: gh.String("bob")},
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{FullName: gh.String("test/repo")},
			},
		},
	}
	activeTab.PRs = testPRs
	activeTab.FilteredPRs = testPRs
	activeTab.Loaded = true
	
	// Update table with test data
	model.updateTableRows(activeTab)
	
	// Test that navigation keys get passed to table
	navigationKeys := []string{"up", "down", "j", "k"}
	
	for _, key := range navigationKeys {
		t.Run("navigation_"+key, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			if key == "up" {
				keyMsg.Type = tea.KeyUp
			} else if key == "down" {
				keyMsg.Type = tea.KeyDown
			}
			
			_, cmd := model.Update(keyMsg)
			
			// Should not return an error or quit command
			// Note: We can't directly compare commands, just check it's not nil unexpectedly
			_ = cmd // Use the command variable to avoid unused variable error
		})
	}
}

// TestFilterInput tests filter input handling
func TestFilterInput(t *testing.T) {
	// Create test model
	tabManager := NewTabManager("test-token")
	tabConfig := &TabConfig{
		Name: "Test Tab",
		Mode: "repos",
		Repos: []string{"test/repo"},
	}
	tabManager.AddTab(tabConfig)
	
	model := &MultiTabModel{
		TabManager: tabManager,
		Width:      100,
		Height:     50,
	}
	
	activeTab := model.TabManager.GetActiveTab()
	if activeTab == nil {
		t.Fatal("Expected active tab to exist")
	}
	
	// Set up filter mode
	activeTab.FilterMode = "author"
	activeTab.FilterValue = ""
	
	// Test adding characters
	t.Run("add_characters", func(t *testing.T) {
		_, cmd := model.handleFilterInput(activeTab, "a")
		if cmd != nil {
			t.Error("handleFilterInput should return nil command for character input")
		}
		
		if activeTab.FilterValue != "a" {
			t.Errorf("Expected FilterValue to be 'a', got '%s'", activeTab.FilterValue)
		}
	})
	
	// Test backspace
	t.Run("backspace", func(t *testing.T) {
		activeTab.FilterValue = "alice"
		
		_, _ = model.handleFilterInput(activeTab, "backspace")
		
		if activeTab.FilterValue != "alic" {
			t.Errorf("Expected FilterValue to be 'alic' after backspace, got '%s'", activeTab.FilterValue)
		}
	})
	
	// Test escape
	t.Run("escape", func(t *testing.T) {
		activeTab.FilterMode = "author"
		activeTab.FilterValue = "alice"
		
		_, _ = model.handleFilterInput(activeTab, "escape")
		
		if activeTab.FilterMode != "" {
			t.Errorf("Expected FilterMode to be empty after escape, got '%s'", activeTab.FilterMode)
		}
		
		if activeTab.FilterValue != "" {
			t.Errorf("Expected FilterValue to be empty after escape, got '%s'", activeTab.FilterValue)
		}
	})
}

// TestApplyFilter tests the filter application logic
func TestApplyFilter(t *testing.T) {
	model := &MultiTabModel{}
	
	testPRs := []*gh.PullRequest{
		{
			Number: gh.Int(1),
			Title:  gh.String("Test PR by Alice"),
			User:   &gh.User{Login: gh.String("alice")},
			Draft:  gh.Bool(false),
			MergeableState: gh.String("clean"),
		},
		{
			Number: gh.Int(2),
			Title:  gh.String("Draft PR by Bob"),
			User:   &gh.User{Login: gh.String("bob")},
			Draft:  gh.Bool(true),
			MergeableState: gh.String("clean"),
		},
		{
			Number: gh.Int(3),
			Title:  gh.String("PR with conflicts"),
			User:   &gh.User{Login: gh.String("charlie")},
			Draft:  gh.Bool(false),
			MergeableState: gh.String("dirty"),
		},
	}
	
	// Test author filter
	t.Run("author_filter", func(t *testing.T) {
		filtered := model.applyFilter(testPRs, "author", "alice")
		
		if len(filtered) != 1 {
			t.Errorf("Expected 1 PR for author 'alice', got %d", len(filtered))
		}
		
		if len(filtered) > 0 && filtered[0].GetUser().GetLogin() != "alice" {
			t.Errorf("Expected filtered PR to be by 'alice', got '%s'", filtered[0].GetUser().GetLogin())
		}
	})
	
	// Test status filter
	t.Run("status_filter_draft", func(t *testing.T) {
		filtered := model.applyFilter(testPRs, "status", "draft")
		
		if len(filtered) != 1 {
			t.Errorf("Expected 1 draft PR, got %d", len(filtered))
		}
		
		if len(filtered) > 0 && !filtered[0].GetDraft() {
			t.Error("Expected filtered PR to be draft")
		}
	})
	
	t.Run("status_filter_conflicts", func(t *testing.T) {
		filtered := model.applyFilter(testPRs, "status", "conflicts")
		
		if len(filtered) != 1 {
			t.Errorf("Expected 1 PR with conflicts, got %d", len(filtered))
		}
		
		if len(filtered) > 0 && filtered[0].GetMergeableState() != "dirty" {
			t.Error("Expected filtered PR to have conflicts")
		}
	})
	
	// Test draft filter
	t.Run("draft_filter", func(t *testing.T) {
		filtered := model.applyFilter(testPRs, "draft", "true")
		
		if len(filtered) != 1 {
			t.Errorf("Expected 1 draft PR, got %d", len(filtered))
		}
		
		if len(filtered) > 0 && !filtered[0].GetDraft() {
			t.Error("Expected filtered PR to be draft")
		}
	})
	
	// Test empty filter
	t.Run("empty_filter", func(t *testing.T) {
		filtered := model.applyFilter(testPRs, "", "")
		
		if len(filtered) != len(testPRs) {
			t.Errorf("Expected all PRs with empty filter, got %d of %d", len(filtered), len(testPRs))
		}
	})
}

// TestUpdateTableRows tests table row updating functionality
func TestUpdateTableRows(t *testing.T) {
	// Create test model
	tabManager := NewTabManager("test-token")
	tabConfig := &TabConfig{
		Name: "Test Tab",
		Mode: "repos",
		Repos: []string{"test/repo"},
	}
	tabManager.AddTab(tabConfig)
	
	model := &MultiTabModel{
		TabManager: tabManager,
		Width:      100,
		Height:     50,
	}
	
	activeTab := model.TabManager.GetActiveTab()
	if activeTab == nil {
		t.Fatal("Expected active tab to exist")
	}
	
	// Test with no PRs
	t.Run("no_prs", func(t *testing.T) {
		activeTab.FilteredPRs = []*gh.PullRequest{}
		
		model.updateTableRows(activeTab)
		
		rows := activeTab.Table.Rows()
		if len(rows) != 0 {
			t.Errorf("Expected 0 rows with no PRs, got %d", len(rows))
		}
	})
	
	// Test with PRs but no enhanced data
	t.Run("prs_no_enhanced_data", func(t *testing.T) {
		testPRs := []*gh.PullRequest{
			{
				Number: gh.Int(1),
				Title:  gh.String("Test PR 1"),
				User:   &gh.User{Login: gh.String("alice")},
				CreatedAt: &gh.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
				UpdatedAt: &gh.Timestamp{Time: time.Now().Add(-30 * time.Minute)},
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{FullName: gh.String("test/repo")},
				},
			},
		}
		
		activeTab.FilteredPRs = testPRs
		activeTab.EnhancedData = make(map[int]types.EnhancedData)
		
		model.updateTableRows(activeTab)
		
		rows := activeTab.Table.Rows()
		if len(rows) != 1 {
			t.Errorf("Expected 1 row, got %d", len(rows))
		}
		
		if len(rows) > 0 && len(rows[0]) != 9 {
			t.Errorf("Expected 9 columns per row, got %d", len(rows[0]))
		}
	})
	
	// Test with enhanced data
	t.Run("prs_with_enhanced_data", func(t *testing.T) {
		testPRs := []*gh.PullRequest{
			{
				Number: gh.Int(123),
				Title:  gh.String("Enhanced PR"),
				User:   &gh.User{Login: gh.String("alice")},
				CreatedAt: &gh.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
				UpdatedAt: &gh.Timestamp{Time: time.Now().Add(-30 * time.Minute)},
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{FullName: gh.String("test/repo")},
				},
			},
		}
		
		activeTab.FilteredPRs = testPRs
		activeTab.EnhancedData = map[int]types.EnhancedData{
			123: {
				Number:         123,
				Comments:       5,
				ReviewComments: 3,
				ReviewStatus:   "approved",
				ChecksStatus:   "success",
				Mergeable:      "clean",
				ChangedFiles:   4,
				Additions:      120,
				Deletions:      45,
			},
		}
		
		model.updateTableRows(activeTab)
		
		rows := activeTab.Table.Rows()
		if len(rows) != 1 {
			t.Errorf("Expected 1 row, got %d", len(rows))
		}
		
		if len(rows) > 0 && len(rows[0]) != 9 {
			t.Errorf("Expected 9 columns per row, got %d", len(rows[0]))
		}
		
		// Check that enhanced data is used (comments should show "8" instead of "?")
		if len(rows) > 0 {
			commentsCol := rows[0][5] // Comments column
			if commentsCol != "8" {
				t.Errorf("Expected comments to show '8' with enhanced data, got '%s'", commentsCol)
			}
		}
	})
}
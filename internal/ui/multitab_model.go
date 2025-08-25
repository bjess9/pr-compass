package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bjess9/pr-compass/internal/github"
	gh "github.com/google/go-github/v55/github"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MultiTabModel is the main model that manages multiple tabs
type MultiTabModel struct {
	TabManager *TabManager
	
	// UI State
	ShowTabNumbers bool // Show numbers when in tab switching mode
	LastKeyTime    time.Time
	HelpMode       bool
	SpinnerIndex   int  // For animating loading spinner
	
	// Global state
	Width  int
	Height int
}

// Tab-specific message types
type tabSwitchMsg struct {
	tabIndex int
}

type tabCloseMsg struct {
	tabIndex int
}

type tabAddMsg struct {
	config *TabConfig
}

type initializeTabsMsg struct{}

type spinnerTickMsg struct{}

type tabRefreshMsg struct {
	tabName string
}

type tabPrsMsg struct {
	tabName string
	prs     []*gh.PullRequest
	err     error
}

// NewMultiTabModel creates a new multi-tab model
func NewMultiTabModel(token string) *MultiTabModel {
	manager := NewTabManager(token)
	
	return &MultiTabModel{
		TabManager: manager,
		ShowTabNumbers: false,
		Width: 120,  // More reasonable default width for modern terminals
		Height: 30,  // More reasonable default height
	}
}

// Init initializes the multi-tab model (called when model is created but before tabs are added)
func (m *MultiTabModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the multi-tab model
func (m *MultiTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		
		// Update all tab table heights
		for _, tab := range m.TabManager.Tabs {
			tableHeight := m.calculateTableHeight(tab)
			tab.Table.SetHeight(tableHeight)
		}
		
		return m, nil
		
	case tea.KeyMsg:
		// Handle global tab switching keys first
		switch msg.String() {
		case "tab":
			m.TabManager.NextTab()
			// Check if newly active tab needs data
			if activeTab := m.TabManager.GetActiveTab(); activeTab != nil && !activeTab.Loaded {
				return m, tea.Batch(
					m.fetchPRsForTab(activeTab),
					m.spinnerTickCmd(), // Start spinner for this tab
				)
			}
			return m, nil
			
		case "shift+tab":
			m.TabManager.PrevTab()
			// Check if newly active tab needs data
			if activeTab := m.TabManager.GetActiveTab(); activeTab != nil && !activeTab.Loaded {
				return m, tea.Batch(
					m.fetchPRsForTab(activeTab),
					m.spinnerTickCmd(), // Start spinner for this tab
				)
			}
			return m, nil
			
		case "ctrl+t":
			// TODO: Add new tab (will implement later)
			return m, nil
			
		case "ctrl+w":
			// Close current tab
			if m.TabManager.GetTabCount() > 1 {
				m.TabManager.CloseTab(m.TabManager.ActiveTabIdx)
			}
			return m, nil
			
		case "ctrl+1", "ctrl+2", "ctrl+3", "ctrl+4", "ctrl+5", "ctrl+6", "ctrl+7", "ctrl+8", "ctrl+9":
			// Switch to specific tab (Ctrl+1 = tab 0, etc.)
			tabNum := int(msg.String()[4] - '1') // Convert '1'-'9' to 0-8
			m.TabManager.SwitchToTab(tabNum)
			// Check if newly active tab needs data
			if activeTab := m.TabManager.GetActiveTab(); activeTab != nil && !activeTab.Loaded {
				return m, tea.Batch(
					m.fetchPRsForTab(activeTab),
					m.spinnerTickCmd(), // Start spinner for this tab
				)
			}
			return m, nil
		}
		
		// Pass other keys to the active tab
		return m.updateActiveTab(msg)
		
	case tabSwitchMsg:
		m.TabManager.SwitchToTab(msg.tabIndex)
		// Check if newly active tab needs data
		if activeTab := m.TabManager.GetActiveTab(); activeTab != nil && !activeTab.Loaded {
			return m, tea.Batch(
				m.fetchPRsForTab(activeTab),
				m.spinnerTickCmd(), // Start spinner for this tab
			)
		}
		return m, nil
		
	case tabAddMsg:
		m.TabManager.AddTab(msg.config)
		return m, nil
		
	case tabCloseMsg:
		m.TabManager.CloseTab(msg.tabIndex)
		return m, nil
		
	case initializeTabsMsg:
		// Initialize tabs after they've been added
		var cmds []tea.Cmd
		
		// Start refresh timers for ALL tabs (they'll only refresh when loaded)
		for _, tab := range m.TabManager.Tabs {
			cmds = append(cmds, m.refreshCmdForTab(tab))
		}
		
		// Fetch data for the active tab immediately
		if activeTab := m.TabManager.GetActiveTab(); activeTab != nil {
			cmds = append(cmds, m.fetchPRsForTab(activeTab))
			cmds = append(cmds, m.spinnerTickCmd()) // Start spinner animation
		}
		
		return m, tea.Batch(cmds...)
		
	case spinnerTickMsg:
		// Update spinner animation
		m.SpinnerIndex = (m.SpinnerIndex + 1) % 10
		
		// Continue spinner animation if any tab is loading
		anyLoading := false
		for _, tab := range m.TabManager.Tabs {
			if !tab.Loaded {
				anyLoading = true
				break
			}
		}
		
		if anyLoading {
			return m, m.spinnerTickCmd()
		}
		return m, nil
		
	case tabRefreshMsg:
		// Handle background refresh for a specific tab
		for _, tab := range m.TabManager.Tabs {
			if tab.Config.Name == msg.tabName {
				// Schedule next refresh for this tab
				nextRefreshCmd := m.refreshCmdForTab(tab)
				
				// Only fetch data if tab is loaded (don't refresh unloaded tabs)
				if tab.Loaded {
					fetchCmd := m.fetchPRsForTab(tab)
					return m, tea.Batch(fetchCmd, nextRefreshCmd)
				} else {
					// Just schedule next refresh, don't fetch data for unloaded tabs
					return m, nextRefreshCmd
				}
			}
		}
		return m, nil
		
	case tabPrsMsg:
		// Handle PR data for a specific tab
		return m.handleTabPRsMessage(msg)
		
	default:
		// Pass other messages to the active tab
		return m.updateActiveTab(msg)
	}
}

// updateActiveTab updates the currently active tab with the given message
func (m *MultiTabModel) updateActiveTab(msg tea.Msg) (tea.Model, tea.Cmd) {
	activeTab := m.TabManager.GetActiveTab()
	if activeTab == nil {
		return m, nil
	}
	
	// Create a temporary single-tab model to handle the message
	singleModel := &model{
		table:               activeTab.Table,
		prs:                 activeTab.PRs,
		filteredPRs:         activeTab.FilteredPRs,
		loaded:              activeTab.Loaded,
		err:                 activeTab.Error,
		token:               m.TabManager.Token,
		refreshIntervalMins: activeTab.Config.RefreshIntervalMinutes,
		showHelp:            activeTab.ShowHelp,
		filterMode:          activeTab.FilterMode,
		filterValue:         activeTab.FilterValue,
		statusMsg:           activeTab.StatusMsg,
		enhancedData:        activeTab.EnhancedData,
		enhancing:           activeTab.Enhancing,
		enhancedCount:       activeTab.EnhancedCount,
		batchManager:        activeTab.BatchManager,
		activeBatchChan:     activeTab.ActiveBatchChan,
		ctx:                 activeTab.Ctx,
		cancel:              activeTab.Cancel,
		prCache:             activeTab.PRCache,
		backgroundRefreshing: activeTab.BackgroundRefreshing,
		lastSelectedPRIndex: activeTab.LastSelectedPRIndex,
		enhancementQueue:    activeTab.EnhancementQueue,
	}
	
	// Update the single model
	updatedModel, cmd := singleModel.Update(msg)
	updatedSingleModel := updatedModel.(*model)
	
	// Copy the state back to the active tab
	activeTab.Table = updatedSingleModel.table
	activeTab.PRs = updatedSingleModel.prs
	activeTab.FilteredPRs = updatedSingleModel.filteredPRs
	activeTab.Loaded = updatedSingleModel.loaded
	activeTab.Error = updatedSingleModel.err
	activeTab.ShowHelp = updatedSingleModel.showHelp
	activeTab.FilterMode = updatedSingleModel.filterMode
	activeTab.FilterValue = updatedSingleModel.filterValue
	activeTab.StatusMsg = updatedSingleModel.statusMsg
	activeTab.EnhancedData = updatedSingleModel.enhancedData
	activeTab.Enhancing = updatedSingleModel.enhancing
	activeTab.EnhancedCount = updatedSingleModel.enhancedCount
	activeTab.ActiveBatchChan = updatedSingleModel.activeBatchChan
	activeTab.BackgroundRefreshing = updatedSingleModel.backgroundRefreshing
	activeTab.LastSelectedPRIndex = updatedSingleModel.lastSelectedPRIndex
	activeTab.EnhancementQueue = updatedSingleModel.enhancementQueue
	
	return m, cmd
}

// View renders the multi-tab interface
func (m *MultiTabModel) View() string {
	if len(m.TabManager.Tabs) == 0 {
		return m.renderNoTabs()
	}
	
	// Render tab bar
	tabBar := m.renderTabBar()
	
	// Render active tab content
	activeTab := m.TabManager.GetActiveTab()
	if activeTab == nil {
		return tabBar + "\n" + "No active tab"
	}
	
	// Render the active tab's content using existing single-tab view logic
	tabContent := m.renderActiveTabContent(activeTab)
	
	return tabBar + "\n" + tabContent
}

// renderTabBar renders the tab bar at the top with rate limiting info
func (m *MultiTabModel) renderTabBar() string {
	if len(m.TabManager.Tabs) <= 1 {
		return "" // Don't show tab bar for single tab
	}
	
	// Rate limit status
	rateLimitInfo := ""
	if m.TabManager.refreshScheduler != nil {
		summary := m.TabManager.refreshScheduler.GetRateLimitSummary()
		rateLimitInfo = fmt.Sprintf(" | Rate Limit: %d remaining | Active: %d", 
			summary.RequestsRemaining, summary.ActiveRequests)
	}
	
	var tabButtons []string
	
	for i, tab := range m.TabManager.Tabs {
		tabName := tab.Config.Name
		if len(tabName) > 15 {
			tabName = tabName[:12] + "..."
		}
		
		// Create tab button with indicators
		var indicator string
		if tab.Loaded {
			prCount := len(tab.PRs)
			if prCount > 0 {
				indicator = fmt.Sprintf(" (%d)", prCount)
			}
		} else if tab.Error != nil {
			indicator = " ❌"
		} else {
			indicator = " ⏳"
		}
		
		tabText := fmt.Sprintf("%d:%s%s", i+1, tabName, indicator)
		
		// Style the tab based on whether it's active
		if i == m.TabManager.ActiveTabIdx {
			// Active tab styling
			tabButton := lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextBright)).
				Background(lipgloss.Color(SelectedBgColor)).
				Bold(true).
				Padding(0, 2).
				Render(tabText)
			tabButtons = append(tabButtons, tabButton)
		} else {
			// Inactive tab styling
			tabButton := lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextSecondary)).
				Background(lipgloss.Color(SurfaceColor)).
				Padding(0, 2).
				Render(tabText)
			tabButtons = append(tabButtons, tabButton)
		}
	}
	
	// Join tabs and add border
	tabBarContent := strings.Join(tabButtons, "")
	
	// Add help text for tab navigation with rate limit info
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextMuted)).
		Render("Tab: Next • Shift+Tab: Prev • Ctrl+1-9: Switch • Ctrl+T: New • Ctrl+W: Close" + rateLimitInfo)
	
	return tabBarContent + "\n" + helpText
}

// renderActiveTabContent renders the content of the active tab
func (m *MultiTabModel) renderActiveTabContent(activeTab *TabState) string {
	// Handle error state
	if activeTab.Error != nil {
		return errorView(activeTab.Error)
	}
	
	// Handle loading state with animated spinner
	if !activeTab.Loaded {
		return loadingViewWithSpinner(m.SpinnerIndex)
	}
	
	// Create a temporary single model for rendering
	singleModel := &model{
		table:               activeTab.Table,
		prs:                 activeTab.PRs,
		filteredPRs:         activeTab.FilteredPRs,
		loaded:              activeTab.Loaded,
		err:                 activeTab.Error,
		token:               m.TabManager.Token,
		refreshIntervalMins: activeTab.Config.RefreshIntervalMinutes,
		showHelp:            activeTab.ShowHelp,
		filterMode:          activeTab.FilterMode,
		filterValue:         activeTab.FilterValue,
		statusMsg:           activeTab.StatusMsg,
		enhancedData:        activeTab.EnhancedData,
		enhancing:           activeTab.Enhancing,
		enhancedCount:       activeTab.EnhancedCount,
		backgroundRefreshing: activeTab.BackgroundRefreshing,
	}
	
	// Use the existing view logic but without the title (since we have tabs)
	singleModel.table.SetStyles(tableStyles())
	
	// Table
	tableView := singleModel.table.View()
	
	// Status message
	statusLine := ""
	if singleModel.statusMsg != "" {
		statusLine = "\n" + statusStyle.Render(singleModel.statusMsg)
	}
	
	// Help text (modified for multi-tab)
	helpText := "🔼🔽 Navigate  •  ⏎ Open PR  •  🔄 Refresh  •  ❓ Help  •  🚪 Quit"
	if singleModel.filterMode != "" {
		helpText = fmt.Sprintf("🔍 Filter: %s=%s  •  🧹 Clear  •  %s", singleModel.filterMode, singleModel.filterValue, helpText)
	}
	
	// Extended help (if active)
	if singleModel.showHelp {
		extendedHelp := "\n" + helpStyle.Render(`
╭─ 🧭 PR Compass Multi-Tab Commands & Visual Guide ──────╮
│                                                         │
│ 🎯 Navigation:                                          │
│   ↑/↓ or j/k     Navigate through PR list              │
│   Enter          🔗 Open PR in browser                  │
│   r              🔄 Manual refresh                      │
│                                                         │
│ 📑 Tab Management:                                      │
│   Tab            ➡️ Next tab                             │
│   Shift+Tab      ⬅️ Previous tab                        │
│   Ctrl+1-9       🔢 Switch to specific tab             │
│   Ctrl+T         ➕ New tab (coming soon)              │
│   Ctrl+W         ❌ Close current tab                   │
│                                                         │
│ 🔍 Filters:                                             │
│   f              👤 Filter by author                    │
│   s              ⚡ Filter by status                     │
│   d              📝 Show drafts only                    │
│   c              🧹 Clear all filters                   │
│                                                         │
│ 📊 Column Symbols:                                      │
│   ✅ Ready        PR ready to merge                     │
│   ⚠️ Conflicts    Merge conflicts                       │
│   🔄 Checks       CI/CD running                         │
│   💬 8c           8 comments                            │
│   📁 5F +120/-45  5 files, +120/-45 lines              │
│   ⏳ Loading...   Fetching enhanced data                │
│                                                         │
│ ❓ Help:                                                 │
│   h/?            Toggle this help                       │
│   q/Ctrl+C       🚪 Exit application                    │
│                                                         │
╰─────────────────────────────────────────────────────────╯`)
		return baseStyle.Render(tableView+statusLine+extendedHelp)
	}
	
	return baseStyle.Render(tableView+statusLine) + "\n" + helpStyle.Render(helpText)
}

// renderNoTabs renders the no tabs state
func (m *MultiTabModel) renderNoTabs() string {
	title := titleStyle.Render("🧭 PR Compass - Multi-Tab Mode")
	message := errorStyle.Render("No tabs configured. Please add tabs to your configuration.")
	help := helpStyle.Render("Press 'q' to quit")
	
	return "\n" + title + "\n\n" + message + "\n\n" + help + "\n"
}

// calculateTableHeight calculates the appropriate table height based on available space
// The table height is FIXED and adapts to terminal size, NOT to number of PRs
func (m *MultiTabModel) calculateTableHeight(tab *TabState) int {
	if m.Height <= 0 {
		return 10 // Conservative default height if window size not yet known
	}
	
	usedLines := 0
	
	// Tab bar (if multiple tabs) - 1 line for tabs, 1 line for border
	if len(m.TabManager.Tabs) > 1 {
		usedLines += 2
	}
	
	// Status line (always reserve space for status)
	usedLines += 1
	
	// Help text (always 1 line for basic help)
	usedLines += 1
	
	// Extended help (if active)
	if tab.ShowHelp {
		usedLines += 25 // Extended help is about 25 lines
	}
	
	// Reserve space for margins, table header, and borders
	usedLines += 4 // Top margin, table header, bottom margin, plus buffer
	
	// Calculate remaining height for table content area
	tableHeight := m.Height - usedLines
	
	// Enforce minimum and maximum bounds
	const minTableHeight = 5  // Always show at least 5 PR rows
	const maxTableHeight = 40 // Cap at reasonable size even on very large terminals
	
	if tableHeight < minTableHeight {
		tableHeight = minTableHeight
	}
	
	if tableHeight > maxTableHeight {
		tableHeight = maxTableHeight
	}
	
	return tableHeight
}

// Helper methods for tab operations

func (m *MultiTabModel) fetchPRsForTab(tab *TabState) tea.Cmd {
	return func() tea.Msg {
		// For initial fetch (when tab is not loaded), bypass rate limiting
		if tab.Loaded {
			// Check if this tab should refresh based on rate limiting (only for subsequent refreshes)
			if m.TabManager.refreshScheduler != nil && !m.TabManager.refreshScheduler.ShouldRefreshTab(tab.Config.Name) {
				// Skip refresh due to rate limiting
				return nil
			}
		}
		
		// Mark refresh as started for rate limiting coordination
		if m.TabManager.refreshScheduler != nil {
			m.TabManager.refreshScheduler.MarkRefreshStarted(tab.Config.Name)
		}

		// Convert tab config to standard config
		cfg := tab.Config.ConvertToConfig()

		var prs []*gh.PullRequest
		var err error
		
		// Create rate-limited request
		if m.TabManager.RateLimiter != nil {
			// Use rate limiter for coordinated API calls
			req := &RateLimitedRequest{
				TabName:  tab.Config.Name,
				Priority: PriorityNormal,
				Timeout:  30 * time.Second,
				ResultChan: make(chan error, 1),
				RequestFunc: func(ctx context.Context) error {
					var fetchErr error
					
					// Use optimized fetching with GraphQL + Caching
					if tab.PRCache != nil {
						prs, fetchErr = github.FetchPRsFromConfigOptimized(ctx, cfg, m.TabManager.Token, tab.PRCache)
					} else {
						prs, fetchErr = github.FetchPRsFromConfig(ctx, cfg, m.TabManager.Token)
					}
					
					return fetchErr
				},
			}
			
			err = m.TabManager.RateLimiter.RequestWithRateLimit(req)
		} else {
			// Fallback to direct fetching
			ctx, cancel := context.WithTimeout(tab.Ctx, 30*time.Second)
			defer cancel()
			
			if tab.PRCache != nil {
				prs, err = github.FetchPRsFromConfigOptimized(ctx, cfg, m.TabManager.Token, tab.PRCache)
			} else {
				prs, err = github.FetchPRsFromConfig(ctx, cfg, m.TabManager.Token)
			}
		}
		
		return tabPrsMsg{
			tabName: tab.Config.Name,
			prs:     prs,
			err:     err,
		}
	}
}

// handleTabPRsMessage handles PR data received for a specific tab
func (m *MultiTabModel) handleTabPRsMessage(msg tabPrsMsg) (tea.Model, tea.Cmd) {
	// Find the tab that this message belongs to
	var targetTab *TabState
	for _, tab := range m.TabManager.Tabs {
		if tab.Config.Name == msg.tabName {
			targetTab = tab
			break
		}
	}
	
	if targetTab == nil {
		// Tab not found - might have been closed
		return m, nil
	}
	
	// Update the tab state based on the message
	if msg.err != nil {
		targetTab.Error = msg.err
		targetTab.Loaded = true
	} else {
		targetTab.PRs = msg.prs
		targetTab.FilteredPRs = msg.prs // TODO: Apply filtering if needed
		targetTab.Loaded = true
		targetTab.Error = nil
		
		// Update table data
		if len(msg.prs) > 0 {
			rows := createTableRows(msg.prs)
			targetTab.Table.SetRows(rows)
		}
		
		// ALWAYS enforce fixed table height regardless of number of rows
		// This ensures the table viewport stays within terminal bounds
		tableHeight := m.calculateTableHeight(targetTab)
		targetTab.Table.SetHeight(tableHeight)
		
		// Ensure table stays focused and scrollable within fixed bounds
		targetTab.Table.Focus()
	}
	
	// If this is the active tab, we might need to return additional commands
	// (like starting enhancement process)
	if targetTab == m.TabManager.GetActiveTab() {
		// TODO: Start background enhancement for the active tab
		return m, nil
	}
	
	return m, nil
}

func (m *MultiTabModel) refreshCmdForTab(tab *TabState) tea.Cmd {
	refreshInterval := tab.Config.RefreshIntervalMinutes
	if refreshInterval == 0 {
		refreshInterval = 5
	}
	
	tabName := tab.Config.Name
	return func() tea.Msg {
		duration := time.Duration(refreshInterval) * time.Minute
		time.Sleep(duration)
		return tabRefreshMsg{tabName: tabName}
	}
}

// spinnerTickCmd creates a command that sends spinner tick messages
func (m *MultiTabModel) spinnerTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}


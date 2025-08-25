package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bjess9/pr-compass/internal/github"
	gh "github.com/google/go-github/v55/github"

	"github.com/charmbracelet/bubbles/table"
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			// Quit the application
			return m, tea.Quit

		case "h", "?":
			// Toggle help
			activeTab.ShowHelp = !activeTab.ShowHelp
			return m, nil

		case "r":
			// Refresh current tab
			if activeTab != nil {
				return m, m.fetchPRsForTab(activeTab)
			}
			return m, nil

		case "f":
			// Start author filter
			activeTab.FilterMode = "author"
			activeTab.FilterValue = ""
			activeTab.StatusMsg = "Enter author name to filter by:"
			return m, nil

		case "s":
			// Start status filter
			activeTab.FilterMode = "status"
			activeTab.FilterValue = ""
			activeTab.StatusMsg = "Enter status (draft/ready/conflicts):"
			return m, nil

		case "d":
			// Toggle draft filter
			if activeTab.FilterMode == "draft" {
				activeTab.FilterMode = ""
				activeTab.FilterValue = ""
				activeTab.FilteredPRs = activeTab.PRs
				activeTab.StatusMsg = "Draft filter cleared"
			} else {
				activeTab.FilterMode = "draft"
				activeTab.FilterValue = "true"
				activeTab.FilteredPRs = m.filterPRsByDraft(activeTab.PRs)
				activeTab.StatusMsg = "Showing draft PRs only"
			}
			m.updateTableRows(activeTab)
			return m, nil

		case "c":
			// Clear all filters
			activeTab.FilterMode = ""
			activeTab.FilterValue = ""
			activeTab.FilteredPRs = activeTab.PRs
			activeTab.StatusMsg = "All filters cleared"
			m.updateTableRows(activeTab)
			return m, nil

		case "enter":
			// Open selected PR in browser
			if len(activeTab.FilteredPRs) > 0 {
				selectedIndex := activeTab.Table.Cursor()
				if selectedIndex < len(activeTab.FilteredPRs) {
					pr := activeTab.FilteredPRs[selectedIndex]
					url := pr.GetHTMLURL()
					if url != "" {
						return m, openURLCmd(url)
					}
				}
			}
			return m, nil

		case "up", "k":
			// Move table cursor up
			activeTab.Table, _ = activeTab.Table.Update(msg)
			return m, nil

		case "down", "j":
			// Move table cursor down
			activeTab.Table, _ = activeTab.Table.Update(msg)
			return m, nil

		case "escape":
			// Cancel current filter input
			if activeTab.FilterMode != "" {
				activeTab.FilterMode = ""
				activeTab.FilterValue = ""
				activeTab.StatusMsg = "Filter cancelled"
			}
			return m, nil

		default:
			// Handle filter input
			if activeTab.FilterMode != "" {
				return m.handleFilterInput(activeTab, msg.String())
			}

			// Pass other keys to table for navigation
			activeTab.Table, _ = activeTab.Table.Update(msg)
			return m, nil
		}
	}

	return m, nil
}

// handleFilterInput processes filter input from the user
func (m *MultiTabModel) handleFilterInput(tab *TabState, input string) (tea.Model, tea.Cmd) {
	switch input {
	case "backspace":
		if len(tab.FilterValue) > 0 {
			tab.FilterValue = tab.FilterValue[:len(tab.FilterValue)-1]
		}
	case "enter":
		// Apply the filter
		tab.FilteredPRs = m.applyFilter(tab.PRs, tab.FilterMode, tab.FilterValue)
		m.updateTableRows(tab)
		tab.StatusMsg = fmt.Sprintf("Filter applied: %s=%s (%d results)", tab.FilterMode, tab.FilterValue, len(tab.FilteredPRs))
		tab.FilterMode = "" // Exit filter input mode
		return m, nil
	case "escape":
		// Cancel filter
		tab.FilterMode = ""
		tab.FilterValue = ""
		tab.StatusMsg = "Filter cancelled"
		return m, nil
	default:
		// Add character to filter
		if len(input) == 1 && input >= " " {
			tab.FilterValue += input
		}
	}
	
	// Update status message with current filter input
	tab.StatusMsg = fmt.Sprintf("Filter %s: %s_", tab.FilterMode, tab.FilterValue)
	return m, nil
}

// applyFilter applies a filter to the PRs list
func (m *MultiTabModel) applyFilter(prs []*gh.PullRequest, mode, value string) []*gh.PullRequest {
	if mode == "" || value == "" {
		return prs
	}
	
	var filtered []*gh.PullRequest
	valueLower := strings.ToLower(value)
	
	for _, pr := range prs {
		include := false
		
		switch mode {
		case "author":
			author := ""
			if pr.GetUser() != nil {
				author = strings.ToLower(pr.GetUser().GetLogin())
			}
			include = strings.Contains(author, valueLower)
			
		case "status":
			status := "ready"
			if pr.GetDraft() {
				status = "draft"
			} else if pr.GetMergeableState() == "dirty" {
				status = "conflicts"
			}
			include = strings.Contains(status, valueLower)
			
		case "draft":
			include = pr.GetDraft() == (value == "true")
		}
		
		if include {
			filtered = append(filtered, pr)
		}
	}
	
	return filtered
}

// filterPRsByDraft returns only draft PRs
func (m *MultiTabModel) filterPRsByDraft(prs []*gh.PullRequest) []*gh.PullRequest {
	var drafts []*gh.PullRequest
	for _, pr := range prs {
		if pr.GetDraft() {
			drafts = append(drafts, pr)
		}
	}
	return drafts
}

// updateTableRows updates the table with current filtered PRs
func (m *MultiTabModel) updateTableRows(tab *TabState) {
	if len(tab.FilteredPRs) == 0 {
		tab.Table.SetRows([]table.Row{})
		return
	}
	
	// Use enhanced table rows if we have enhanced data
	if len(tab.EnhancedData) > 0 {
		rows := createTableRowsWithEnhancement(tab.FilteredPRs, tab.EnhancedData)
		tab.Table.SetRows(rows)
	} else {
		rows := createTableRows(tab.FilteredPRs)
		tab.Table.SetRows(rows)
	}
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

// renderTabBar renders the enhanced tab bar at the top with rate limiting info
func (m *MultiTabModel) renderTabBar() string {
	if len(m.TabManager.Tabs) <= 1 {
		return "" // Don't show tab bar for single tab
	}
	
	// Rate limit status with better formatting
	rateLimitInfo := ""
	if m.TabManager.refreshScheduler != nil {
		summary := m.TabManager.refreshScheduler.GetRateLimitSummary()
		rateLimitColor := TextMuted
		if summary.RequestsRemaining < 100 {
			rateLimitColor = ErrorColor // Red when low
		} else if summary.RequestsRemaining < 500 {
			rateLimitColor = WarningColor // Yellow when getting low
		}
		
		rateLimitInfo = fmt.Sprintf(" │ %s %d/5000 │ Active: %d", 
			lipgloss.NewStyle().Foreground(lipgloss.Color(rateLimitColor)).Render("API:"),
			summary.RequestsRemaining, 
			summary.ActiveRequests)
	}
	
	var tabButtons []string
	
	for i, tab := range m.TabManager.Tabs {
		tabName := tab.Config.Name
		if len(tabName) > 15 {
			tabName = tabName[:12] + "..."
		}
		
		// Enhanced status indicators with better visuals
		var indicator, statusColor string
		if tab.Error != nil {
			indicator = " 🚨"
			statusColor = ErrorColor
		} else if !tab.Loaded {
			// Show loading progress with animated dots
			dots := []string{"⏳", "⏳ .", "⏳ ..", "⏳ ..."}
			dotIndex := (m.SpinnerIndex / 5) % len(dots)
			indicator = " " + dots[dotIndex]
			statusColor = WarningColor
		} else {
			prCount := len(tab.PRs)
			if prCount > 0 {
				indicator = fmt.Sprintf(" 📋%d", prCount)
				statusColor = SuccessColor
			} else {
				indicator = " ✅"
				statusColor = TextSecondary
			}
		}
		
		tabText := fmt.Sprintf("%d:%s%s", i+1, tabName, indicator)
		
		// Enhanced tab styling with borders and rounded corners
		if i == m.TabManager.ActiveTabIdx {
			// Active tab - prominent with border
			tabButton := lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextBright)).
				Background(lipgloss.Color(SelectedBgColor)).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(statusColor)).
				Padding(0, 1).
				Render(tabText)
			tabButtons = append(tabButtons, tabButton)
		} else {
			// Inactive tab - subtle with rounded corners
			tabButton := lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextSecondary)).
				Background(lipgloss.Color(SurfaceColor)).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(BorderColor)).
				Padding(0, 1).
				Render(tabText)
			tabButtons = append(tabButtons, tabButton)
		}
	}
	
	// Add spacing between tabs and join horizontally  
	var spacedButtons []string
	for i, button := range tabButtons {
		if i > 0 {
			spacedButtons = append(spacedButtons, " ") // Add space between tabs
		}
		spacedButtons = append(spacedButtons, button)
	}
	tabBarContent := lipgloss.JoinHorizontal(lipgloss.Top, spacedButtons...)
	
	// Compact help text with compass theme
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextMuted)).
		Italic(true).
		Render("🧭 Tab/⇧Tab Navigate • ^1-9 Switch • h Help" + rateLimitInfo)
	
	// Compact separator line
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(BorderColor)).
		Render(strings.Repeat("─", 60)) // Shorter line
	
	return tabBarContent + "\n" + helpText + "\n" + separator
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
	
	// Render the table directly without creating the old model
	
	// Use the existing view logic but without the title (since we have tabs)
	activeTab.Table.SetStyles(tableStyles())
	
	// Table
	tableView := activeTab.Table.View()
	
	// Status message
	statusLine := ""
	if activeTab.StatusMsg != "" {
		statusLine = "\n" + statusStyle.Render(activeTab.StatusMsg)
	}
	
	// Help text (compact with compass emoji)
	helpText := "🧭 Navigate: ↑↓  •  ⏎ Open  •  r Refresh  •  h Help  •  q Quit"
	if activeTab.FilterMode != "" {
		helpText = fmt.Sprintf("🔍 %s: %s  •  c Clear  •  %s", activeTab.FilterMode, activeTab.FilterValue, helpText)
	}
	
	// Extended help (compact with compass theme)
	if activeTab.ShowHelp {
		extendedHelp := "\n" + helpStyle.Render(`
╭─ 🧭 PR Compass - Navigation Guide ─╮
│ 🎯 Navigate: ↑↓/jk  ⏎ Open PR      │
│ 📑 Tabs: Tab/⇧Tab  ^1-9 Switch     │
│ 🔍 Filter: f Author s Status d Draft │
│ 🧹 Clear: c  🔄 Refresh: r  ❓ Help: h │
│                                     │
│ 🏷️  Status: ✅Ready ⚠️Conflict 🔄CI  │
│ 💬 Comments 📁 Files ⏳ Loading     │
╰─────────────────────────────────────╯`)
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
		return 8 // More compact default for small terminals
	}
	
	usedLines := 0
	
	// Account for terminal window chrome (title bar, tabs, etc.) - be more aggressive
	usedLines += 6 // Terminal app chrome, status bars, etc.
	
	// Tab bar (compact) - tabs are now more compact boxes
	if len(m.TabManager.Tabs) > 1 {
		usedLines += 3 // Box border + content + separator
	}
	
	// Status line (compact)
	usedLines += 1
	
	// Help text (compact, single line)
	usedLines += 1
	
	// Extended help (if active) - make more compact
	if tab.ShowHelp {
		usedLines += 20 // Reduced from 25 lines
	}
	
	// Reserve minimal space for table header and borders (more compact)
	usedLines += 2 // Just table header + minimal buffer
	
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


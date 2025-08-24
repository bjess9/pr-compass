package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/ui/components"
	"github.com/bjess9/pr-compass/internal/ui/services"
	"github.com/bjess9/pr-compass/internal/ui/types"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// Message types for the refactored model
type (
	prsLoadedMsg struct {
		prs []*types.PRData
		cfg *config.Config
	}
	
	prEnhancedMsg struct {
		prNumber int
		enhanced *types.EnhancedData
		err      error
	}
	
	refreshCompleteMsg struct {
		prs []*types.PRData
		cfg *config.Config
	}
	
	errorMsg struct {
		err error
	}
	
	refreshTickMsg struct{}
)

// Model represents the refactored application model
type Model struct {
	// Core dependencies
	services   *services.Registry
	components struct {
		table *components.TableComponent
	}
	
	// UI state
	table         table.Model
	token         string
	refreshInterval time.Duration
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// InitialModelNew creates a new refactored model
func InitialModelNew(token string) Model {
	// Initialize cache
	prCache, _ := cache.NewPRCache() // Ignore error, continue without cache
	
	// Create services
	serviceRegistry := services.NewRegistry(token, prCache)
	
	// Create UI components
	tableComponent := components.NewTableComponent()
	
	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create the table
	tableModel := tableComponent.CreateTable()
	
	return Model{
		services: serviceRegistry,
		components: struct {
			table *components.TableComponent
		}{
			table: tableComponent,
		},
		table:           tableModel,
		token:           token,
		refreshInterval: 5 * time.Minute,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadPRsCmd(),
		m.refreshTickCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	switch msg := msg.(type) {
	case prsLoadedMsg:
		return m.handlePRsLoaded(msg, cmd)
		
	case prEnhancedMsg:
		return m.handlePREnhanced(msg, cmd)
		
	case refreshCompleteMsg:
		return m.handleRefreshComplete(msg, cmd)
		
	case refreshTickMsg:
		return m.handleRefreshTick(cmd)
		
	case errorMsg:
		m.services.State.SetError(msg.err)
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg, cmd)
	}

	return m, cmd
}

// View renders the model
func (m Model) View() string {
	state := m.services.State.GetState()
	
	if state.Error != nil {
		return m.renderError(state.Error)
	}
	
	if !state.Loaded {
		return m.renderLoading()
	}
	
	return m.renderMain(state)
}

// Command functions
func (m Model) loadPRsCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadConfig()
		if err != nil {
			return errorMsg{err}
		}

		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		prs, err := m.services.PR.FetchPRs(ctx, cfg)
		if err != nil {
			return errorMsg{err}
		}

		return prsLoadedMsg{prs: prs, cfg: cfg}
	}
}

func (m Model) refreshTickCmd() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func (m Model) enhancePRCmd(prNumber int) tea.Cmd {
	return func() tea.Msg {
		state := m.services.State.GetState()
		
		// Find the PR
		var targetPR *types.PRData
		for _, pr := range state.PRs {
			if pr.GetNumber() == prNumber {
				targetPR = pr
				break
			}
		}
		
		if targetPR == nil {
			return prEnhancedMsg{prNumber: prNumber, err: fmt.Errorf("PR #%d not found", prNumber)}
		}
		
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()
		
		enhanced, err := m.services.Enhancement.EnhancePR(ctx, targetPR.PullRequest)
		return prEnhancedMsg{prNumber: prNumber, enhanced: enhanced, err: err}
	}
}

// Message handlers
func (m Model) handlePRsLoaded(msg prsLoadedMsg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	// Update state
	m.services.State.UpdatePRs(msg.prs)
	m.services.State.SetLoaded(true)
	m.services.State.ClearError()
	
	// Update table
	state := m.services.State.GetState()
	rows := m.components.table.CreateRows(state.FilteredPRs, state.EnhancementQueue)
	m.table.SetRows(rows)
	
	// Update status
	statusMsg := fmt.Sprintf("Loaded %d PRs", len(msg.prs))
	if msg.cfg != nil {
		statusMsg += m.formatConfigInfo(msg.cfg)
	}
	m.services.State.UpdateStatusMessage(statusMsg)
	
	// Start enhancing the first PR
	var enhanceCmd tea.Cmd
	if len(msg.prs) > 0 {
		firstPR := msg.prs[0]
		m.services.State.AddToEnhancementQueue(firstPR.GetNumber())
		enhanceCmd = m.enhancePRCmd(firstPR.GetNumber())
	}
	
	return m, tea.Batch(cmd, enhanceCmd, m.refreshTickCmd())
}

func (m Model) handlePREnhanced(msg prEnhancedMsg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.services.State.RemoveFromEnhancementQueue(msg.prNumber)
	
	if msg.err == nil && msg.enhanced != nil {
		m.services.State.UpdatePREnhancement(msg.prNumber, msg.enhanced)
		
		// Update table with new enhanced data
		state := m.services.State.GetState()
		rows := m.components.table.CreateRows(state.FilteredPRs, state.EnhancementQueue)
		m.table.SetRows(rows)
		
		// Update enhanced count
		count := state.EnhancedCount + 1
		m.services.State.UpdateEnhancedCount(count)
		
		// Update status message
		statusMsg := fmt.Sprintf("Loaded %d PRs - enhanced %d", len(state.PRs), count)
		m.services.State.UpdateStatusMessage(statusMsg)
	}
	
	return m, cmd
}

func (m Model) handleRefreshComplete(msg refreshCompleteMsg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	// Update PRs
	m.services.State.UpdatePRs(msg.prs)
	m.services.State.SetBackgroundRefreshing(false)
	
	// Update table
	state := m.services.State.GetState()
	rows := m.components.table.CreateRows(state.FilteredPRs, state.EnhancementQueue)
	m.table.SetRows(rows)
	
	// Update status
	statusMsg := fmt.Sprintf("Refreshed %d PRs", len(msg.prs))
	m.services.State.UpdateStatusMessage(statusMsg)
	
	return m, tea.Batch(cmd, m.refreshTickCmd())
}

func (m Model) handleRefreshTick(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	state := m.services.State.GetState()
	
	if state.Loaded && len(state.PRs) > 0 {
		m.services.State.SetBackgroundRefreshing(true)
		return m, tea.Batch(cmd, m.refreshCmd())
	}
	
	return m, tea.Batch(cmd, m.refreshTickCmd())
}

func (m Model) handleKeyPress(msg tea.KeyMsg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
		
	case "r":
		return m, tea.Batch(cmd, m.loadPRsCmd())
		
	case "h", "?":
		state := m.services.State.GetState()
		m.services.State.SetShowHelp(!state.UI.ShowHelp)
		
	case "enter":
		return m.handleEnterPress(cmd)
		
	case "f":
		return m.handleFilterPress(cmd)
		
	case "c":
		return m.handleClearFilter(cmd)
	}
	
	// Check for cursor movement to trigger lazy enhancement
	currentIndex := m.table.Cursor()
	state := m.services.State.GetState()
	
	if currentIndex != state.UI.TableCursor && currentIndex < len(state.FilteredPRs) {
		m.services.State.UpdateTableCursor(currentIndex)
		
		// Trigger lazy enhancement for selected PR
		selectedPR := state.FilteredPRs[currentIndex]
		prNumber := selectedPR.GetNumber()
		
		if !m.services.Enhancement.IsEnhanced(prNumber) && !m.services.State.IsInEnhancementQueue(prNumber) {
			m.services.State.AddToEnhancementQueue(prNumber)
			return m, tea.Batch(cmd, m.enhancePRCmd(prNumber))
		}
	}
	
	return m, cmd
}

// Helper methods
func (m Model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadConfig()
		if err != nil {
			return errorMsg{err}
		}

		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		prs, err := m.services.PR.RefreshPRs(ctx, cfg)
		if err != nil {
			// On error, don't fail completely - just continue with current data
			return refreshTickMsg{}
		}

		return refreshCompleteMsg{prs: prs, cfg: cfg}
	}
}

func (m Model) handleEnterPress(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	state := m.services.State.GetState()
	if !state.Loaded || len(state.FilteredPRs) == 0 {
		return m, cmd
	}
	
	idx := m.table.Cursor()
	if idx >= 0 && idx < len(state.FilteredPRs) {
		pr := state.FilteredPRs[idx]
		prURL := pr.GetHTMLURL()
		if prURL != "" {
			return m, tea.Batch(cmd, openURLCmd(prURL))
		}
	}
	
	return m, cmd
}

func (m Model) handleFilterPress(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	state := m.services.State.GetState()
	if !state.Loaded || len(state.PRs) == 0 {
		return m, cmd
	}
	
	// Simple author filter for now
	idx := m.table.Cursor()
	if idx >= 0 && idx < len(state.FilteredPRs) {
		selectedPR := state.FilteredPRs[idx]
		author := selectedPR.GetUser().GetLogin()
		
		filter := types.FilterOptions{
			Mode:   "author",
			Value:  author,
			Active: true,
		}
		
		filteredPRs := m.services.Filter.ApplyFilter(state.PRs, filter)
		m.services.State.UpdateFilter(filter)
		m.services.State.UpdateFilteredPRs(filteredPRs)
		
		// Update table
		rows := m.components.table.CreateRows(filteredPRs, state.EnhancementQueue)
		m.table.SetRows(rows)
		
		statusMsg := fmt.Sprintf("Showing %d PRs by %s", len(filteredPRs), author)
		m.services.State.UpdateStatusMessage(statusMsg)
	}
	
	return m, cmd
}

func (m Model) handleClearFilter(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	state := m.services.State.GetState()
	
	// Clear filter
	filter := types.FilterOptions{Active: false}
	m.services.State.UpdateFilter(filter)
	m.services.State.UpdateFilteredPRs(state.PRs)
	
	// Update table
	rows := m.components.table.CreateRows(state.PRs, state.EnhancementQueue)
	m.table.SetRows(rows)
	
	statusMsg := fmt.Sprintf("Showing all %d PRs", len(state.PRs))
	m.services.State.UpdateStatusMessage(statusMsg)
	
	return m, cmd
}

// Rendering methods
func (m Model) renderError(err error) string {
	return errorView(err)
}

func (m Model) renderLoading() string {
	return loadingView()
}


func (m Model) renderMain(state *types.AppState) string {
	m.table.SetStyles(tableStyles())
	
	// Title
	title := titleStyle.Render("PR Compass - Pull Request Monitor")
	
	// Table
	tableView := m.table.View()
	
	// Status
	statusLine := ""
	if state.UI.StatusMsg != "" {
		statusLine = "\n" + statusStyle.Render(state.UI.StatusMsg)
	}
	
	// Help
	helpText := "↑/↓: Navigate  •  Enter: Open PR  •  r: Refresh  •  f: Filter  •  c: Clear  •  h: Help  •  q: Quit"
	if state.UI.Filter.Active {
		filterDesc := m.services.Filter.GetActiveFiltersDescription(state.UI.Filter)
		helpText = fmt.Sprintf("Filter: %s  •  %s", filterDesc, helpText)
	}
	
	if state.UI.ShowHelp {
		// Extended help would go here
		return title + "\n" + baseStyle.Render(tableView+statusLine) + "\n" + helpStyle.Render(helpText)
	}
	
	return title + "\n" + baseStyle.Render(tableView+statusLine) + "\n" + helpStyle.Render(helpText)
}

func (m Model) formatConfigInfo(cfg *config.Config) string {
	switch cfg.Mode {
	case "topics":
		if cfg.TopicOrg != "" && len(cfg.Topics) > 0 {
			return fmt.Sprintf(" from %s (topics: %v)", cfg.TopicOrg, cfg.Topics)
		}
	case "organization":
		if cfg.Organization != "" {
			return fmt.Sprintf(" from %s organization", cfg.Organization)
		}
	case "teams":
		if cfg.Organization != "" && len(cfg.Teams) > 0 {
			return fmt.Sprintf(" from %s teams: %v", cfg.Organization, cfg.Teams)
		}
	}
	return ""
}
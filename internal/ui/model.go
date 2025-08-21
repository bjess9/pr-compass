package ui

import (
	"fmt"
	"time"

	"github.com/bjess9/pr-pilot/internal/config"
	"github.com/bjess9/pr-pilot/internal/github"
	gh "github.com/google/go-github/v55/github"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type refreshMsg struct{}

type model struct {
	table       table.Model
	prs         []*gh.PullRequest
	filteredPRs []*gh.PullRequest
	loaded      bool
	err         error
	token       string
	showHelp    bool
	filterMode  string // "", "author", "repo", "status"
	filterValue string
	statusMsg   string
}

func InitialModel(token string) model {
	columns := createTableColumns()
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	t.Focus()

	return model{table: t, token: token}
}

func refreshCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(60 * time.Second)
		return refreshMsg{}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.fetchPRs, refreshCmd())
}

func (m model) fetchPRs() tea.Msg {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errMsg{err}
	}
	prs, err := github.FetchPRsFromConfig(cfg, m.token)
	if err != nil {
		return errMsg{err}
	}
	return prs
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.table, cmd = m.table.Update(msg)

	switch msg := msg.(type) {
	case []*gh.PullRequest:
		m.loaded = true
		m.prs = msg
		m.filteredPRs = msg // Initially show all PRs
		rows := createTableRows(m.filteredPRs)
		m.table.SetRows(rows)
		m.statusMsg = fmt.Sprintf("Loaded %d PRs", len(msg))
		return m, tea.Batch(cmd, refreshCmd())

	case refreshMsg:
		return m, tea.Batch(m.fetchPRs, refreshCmd())

	case errMsg:
		m.err = msg.err

	case error:
		m.err = msg

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "h", "?":
			m.showHelp = !m.showHelp

		case "r":
			// Manual refresh
			m.statusMsg = "Refreshing..."
			return m, m.fetchPRs

		case "c":
			// Clear filters
			m.filterMode = ""
			m.filterValue = ""
			m.filteredPRs = m.prs
			rows := createTableRows(m.filteredPRs)
			m.table.SetRows(rows)
			m.statusMsg = fmt.Sprintf("Showing all %d PRs", len(m.prs))

		case "f":
			// Filter by author
			if m.loaded && len(m.prs) > 0 {
				return m.filterByAuthor()
			}

		case "s":
			// Filter by status
			if m.loaded && len(m.prs) > 0 {
				return m.filterByStatus()
			}

		case "d":
			// Show only draft PRs
			if m.loaded {
				return m.filterByDraft()
			}

		case "enter":
			if m.loaded && len(m.filteredPRs) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.filteredPRs) {
					pr := m.filteredPRs[idx]
					prURL := pr.GetHTMLURL()
					if prURL != "" {
						return m, openURLCmd(prURL)
					}
				}
			}
		}
	}

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return errorView(m.err)
	}
	if !m.loaded {
		return loadingView()
	}

	m.table.SetStyles(tableStyles())
	
	tableView := m.table.View()
	
	// Status message
	statusLine := ""
	if m.statusMsg != "" {
		statusLine = "\n" + m.statusMsg
	}
	
	// Help text
	helpText := "↑/↓: Navigate  •  Enter: Open PR  •  r: Refresh  •  h: Help  •  q: Quit"
	if m.filterMode != "" {
		helpText = fmt.Sprintf("Filtered by %s: %s  •  c: Clear filter  •  %s", m.filterMode, m.filterValue, helpText)
	}

	// Extended help
	if m.showHelp {
		extendedHelp := "\n" + helpStyle.Render(`
Additional Commands:
  f: Filter by author    s: Filter by status    d: Show drafts only
  c: Clear filters       r: Manual refresh      h/?: Toggle help
  ↑/↓ or j/k: Navigate   Enter: Open PR in browser   q: Quit`)
		return baseStyle.Render(tableView + statusLine + extendedHelp)
	}

	return baseStyle.Render(tableView + statusLine + "\n" + helpStyle.Render(helpText))
}

// Filter methods
func (m model) filterByAuthor() (model, tea.Cmd) {
	// Get unique authors from current PRs
	authorMap := make(map[string]int)
	for _, pr := range m.prs {
		author := pr.GetUser().GetLogin()
		authorMap[author]++
	}
	
	// If only one author, filter by them
	if len(authorMap) == 1 {
		for author := range authorMap {
			return m.applyAuthorFilter(author)
		}
	}
	
	// For now, filter by the current selected PR's author
	if len(m.filteredPRs) > 0 {
		idx := m.table.Cursor()
		if idx >= 0 && idx < len(m.filteredPRs) {
			author := m.filteredPRs[idx].GetUser().GetLogin()
			return m.applyAuthorFilter(author)
		}
	}
	
	return m, nil
}

func (m model) applyAuthorFilter(author string) (model, tea.Cmd) {
	filtered := []*gh.PullRequest{}
	for _, pr := range m.prs {
		if pr.GetUser().GetLogin() == author {
			filtered = append(filtered, pr)
		}
	}
	
	m.filteredPRs = filtered
	m.filterMode = "author"
	m.filterValue = author
	rows := createTableRows(m.filteredPRs)
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d PRs by %s", len(filtered), author)
	
	return m, nil
}

func (m model) filterByStatus() (model, tea.Cmd) {
	// Show only PRs that are ready (not drafts, and mergeable)
	filtered := []*gh.PullRequest{}
	for _, pr := range m.prs {
		if !pr.GetDraft() && pr.GetMergeable() {
			filtered = append(filtered, pr)
		}
	}
	
	m.filteredPRs = filtered
	m.filterMode = "status"
	m.filterValue = "ready"
	rows := createTableRows(m.filteredPRs)
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d ready PRs", len(filtered))
	
	return m, nil
}

func (m model) filterByDraft() (model, tea.Cmd) {
	filtered := []*gh.PullRequest{}
	for _, pr := range m.prs {
		if pr.GetDraft() {
			filtered = append(filtered, pr)
		}
	}
	
	m.filteredPRs = filtered
	m.filterMode = "status"
	m.filterValue = "draft"
	rows := createTableRows(m.filteredPRs)
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d draft PRs", len(filtered))
	
	return m, nil
}

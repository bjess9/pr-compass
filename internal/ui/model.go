package ui

import (
	"context"
	"fmt"
	"sync"
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

// PrsWithConfigMsg contains both PRs and config info for status display
type PrsWithConfigMsg struct {
	Prs []*gh.PullRequest
	Cfg *config.Config
}

// Enhanced PR data from individual API calls
type enhancedPRData struct {
	Number         int
	Comments       int
	ReviewComments int
	ReviewStatus   string // "approved", "changes_requested", "pending", "unknown"
	ChecksStatus   string // "success", "failure", "pending", "unknown"
	Mergeable      string // "clean", "conflicts", "unknown"
	Additions      int
	Deletions      int
	ChangedFiles   int
	EnhancedAt     time.Time
}

// Message types for background enhancement
type prEnhancementStartMsg struct{}
type prEnhancementUpdateMsg struct {
	prData enhancedPRData
	error  error
}
type prEnhancementCompleteMsg struct{}

type model struct {
	table               table.Model
	prs                 []*gh.PullRequest
	filteredPRs         []*gh.PullRequest
	loaded              bool
	err                 error
	token               string
	refreshIntervalMins int // Refresh interval in minutes
	showHelp            bool
	filterMode          string // "", "author", "repo", "status"
	filterValue         string
	statusMsg           string

	// Enhanced data tracking
	enhancedData     map[int]enhancedPRData // PR number -> enhanced data
	enhancementMutex sync.RWMutex
	enhancing        bool
	enhancedCount    int
}

func InitialModel(token string) model {
	columns := createTableColumns()
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	t.Focus()

	return model{
		table:               t,
		token:               token,
		refreshIntervalMins: 5, // default, will be updated from config
		enhancedData:        make(map[int]enhancedPRData),
	}
}

func refreshCmd(intervalMinutes int) tea.Cmd {
	return func() tea.Msg {
		duration := time.Duration(intervalMinutes) * time.Minute
		time.Sleep(duration)
		return refreshMsg{}
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.fetchPRs, refreshCmd(m.refreshIntervalMins))
}

func (m *model) fetchPRs() tea.Msg {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errMsg{err}
	}
	prs, err := github.FetchPRsFromConfig(cfg, m.token)
	if err != nil {
		return errMsg{err}
	}
	return PrsWithConfigMsg{Prs: prs, Cfg: cfg}
}

// Background enhancement command that fetches individual PR details
func (m *model) enhancePRs() tea.Cmd {
	cmds := m.enhancePRsIndividually()
	return tea.Batch(cmds...)
}

// enhancePRsIndividually creates commands for each PR to be enhanced individually
func (m *model) enhancePRsIndividually() []tea.Cmd {
	// Limit to top 20 PRs for better performance (since refresh is now 5 min)
	prsToEnhance := m.prs
	if len(prsToEnhance) > 20 {
		prsToEnhance = prsToEnhance[:20]
	}

	var cmds []tea.Cmd

	// Create a command for each PR with staggered delays
	for i, pr := range prsToEnhance {
		cmds = append(cmds, m.enhanceSinglePR(pr, time.Duration(i)*time.Second))
	}

	return cmds
}

// enhanceSinglePR creates a command to enhance a single PR with a delay
func (m *model) enhanceSinglePR(pr *gh.PullRequest, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		// Staggered delay for rate limiting
		time.Sleep(delay)

		// Get GitHub client
		client, err := github.NewClient(m.token)
		if err != nil {
			return errMsg{err}
		}

		ctx := context.Background()

		// Fetch enhanced data for this PR
		enhanced, err := m.fetchEnhancedPRData(ctx, client, pr)
		if err != nil {
			// Return an error for this specific PR, don't crash the whole thing
			return prEnhancementUpdateMsg{
				prData: enhancedPRData{Number: pr.GetNumber()}, // Empty data indicates error
				error:  err,
			}
		}

		return prEnhancementUpdateMsg{prData: enhanced}
	}
}

// fetchEnhancedPRData gets detailed PR information from individual API call
func (m *model) fetchEnhancedPRData(ctx context.Context, client *gh.Client, pr *gh.PullRequest) (enhancedPRData, error) {
	owner := pr.GetBase().GetRepo().GetOwner().GetLogin()
	repo := pr.GetBase().GetRepo().GetName()
	number := pr.GetNumber()

	// Get detailed PR data
	detailedPR, _, err := client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return enhancedPRData{}, err
	}

	// Get review status
	reviews, _, err := client.PullRequests.ListReviews(ctx, owner, repo, number, nil)
	reviewStatus := "unknown"
	if err == nil {
		reviewStatus = determineReviewStatus(reviews)
	}

	// Get checks status
	checksStatus := "unknown"
	if sha := pr.GetHead().GetSHA(); sha != "" {
		checks, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
		if err == nil && checks != nil {
			checksStatus = determineChecksStatus(checks.CheckRuns)
		}
	}

	// Determine mergeable status
	mergeableStatus := "unknown"
	if detailedPR.Mergeable != nil {
		if *detailedPR.Mergeable {
			mergeableStatus = "clean"
		} else {
			mergeableStatus = "conflicts"
		}
	}

	return enhancedPRData{
		Number:         number,
		Comments:       detailedPR.GetComments(),
		ReviewComments: detailedPR.GetReviewComments(),
		ReviewStatus:   reviewStatus,
		ChecksStatus:   checksStatus,
		Mergeable:      mergeableStatus,
		Additions:      detailedPR.GetAdditions(),
		Deletions:      detailedPR.GetDeletions(),
		ChangedFiles:   detailedPR.GetChangedFiles(),
		EnhancedAt:     time.Now(),
	}, nil
}

// determineReviewStatus analyzes review data to determine overall status
func determineReviewStatus(reviews []*gh.PullRequestReview) string {
	if len(reviews) == 0 {
		return "no_review"
	}

	// Get latest review by each reviewer
	latestReviews := make(map[string]string)
	for _, review := range reviews {
		user := review.GetUser().GetLogin()
		state := review.GetState()
		latestReviews[user] = state
	}

	// Check for blocking states
	for _, state := range latestReviews {
		if state == "CHANGES_REQUESTED" {
			return "changes_requested"
		}
	}

	// Check if all reviews are approved
	approvedCount := 0
	for _, state := range latestReviews {
		if state == "APPROVED" {
			approvedCount++
		}
	}

	if approvedCount > 0 && approvedCount == len(latestReviews) {
		return "approved"
	}

	return "pending"
}

// determineChecksStatus analyzes check runs to determine overall status
func determineChecksStatus(checkRuns []*gh.CheckRun) string {
	if len(checkRuns) == 0 {
		return "none"
	}

	hasFailure := false
	hasPending := false

	for _, check := range checkRuns {
		switch check.GetStatus() {
		case "completed":
			if check.GetConclusion() == "failure" || check.GetConclusion() == "cancelled" {
				hasFailure = true
			}
		case "in_progress", "queued":
			hasPending = true
		}
	}

	if hasFailure {
		return "failure"
	}
	if hasPending {
		return "pending"
	}
	return "success"
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.table, cmd = m.table.Update(msg)

	switch msg := msg.(type) {
	case PrsWithConfigMsg:
		m.loaded = true
		m.prs = sortPRsByNewest(msg.Prs) // Sort by newest first
		m.filteredPRs = m.prs            // Initially show all PRs (already filtered by config)

		// Update refresh interval from config
		if msg.Cfg != nil {
			m.refreshIntervalMins = msg.Cfg.RefreshIntervalMinutes
		}

		// Clear previous enhanced data
		m.enhancementMutex.Lock()
		m.enhancedData = make(map[int]enhancedPRData)
		m.enhancing = false
		m.enhancedCount = 0
		m.enhancementMutex.Unlock()

		rows := createTableRowsWithEnhancement(m.filteredPRs, m.enhancedData)
		m.table.SetRows(rows)

		// Create informative status message with org/topic info
		statusInfo := fmt.Sprintf("Loaded %d PRs", len(msg.Prs))
		if msg.Cfg != nil {
			switch msg.Cfg.Mode {
			case "topics":
				if msg.Cfg.TopicOrg != "" && len(msg.Cfg.Topics) > 0 {
					statusInfo += fmt.Sprintf(" from %s (topics: %v)", msg.Cfg.TopicOrg, msg.Cfg.Topics)
				}
			case "organization":
				if msg.Cfg.Organization != "" {
					statusInfo += fmt.Sprintf(" from %s organization", msg.Cfg.Organization)
				}
			case "teams":
				if msg.Cfg.Organization != "" && len(msg.Cfg.Teams) > 0 {
					statusInfo += fmt.Sprintf(" from %s teams: %v", msg.Cfg.Organization, msg.Cfg.Teams)
				}
			}
		}
		m.statusMsg = statusInfo

		// Start background enhancement
		enhanceCmd := func() tea.Msg {
			// Small delay to let UI render first
			time.Sleep(100 * time.Millisecond)
			return prEnhancementStartMsg{}
		}

		return m, tea.Batch(cmd, refreshCmd(m.refreshIntervalMins), enhanceCmd)

	case refreshMsg:
		return m, tea.Batch(m.fetchPRs, refreshCmd(m.refreshIntervalMins))

	case errMsg:
		m.err = msg.err

	case error:
		m.err = msg

	case prEnhancementStartMsg:
		m.enhancementMutex.Lock()
		m.enhancing = true
		m.enhancedCount = 0 // Reset count when starting enhancement
		m.enhancementMutex.Unlock()

		m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhancing with detailed data...", len(m.prs))
		return m, m.enhancePRs()

	case prEnhancementUpdateMsg:
		if msg.error != nil {
			// Skip this PR on error, don't crash the whole thing
			return m, nil
		}

		// Update enhanced data
		m.enhancementMutex.Lock()
		m.enhancedData[msg.prData.Number] = msg.prData
		m.enhancedCount++
		totalPRsToEnhance := len(m.prs)
		if totalPRsToEnhance > 20 {
			totalPRsToEnhance = 20 // We only enhance top 20
		}
		currentCount := m.enhancedCount
		m.enhancementMutex.Unlock()

		// Update table rows immediately with this new enhanced data
		rows := createTableRowsWithEnhancement(m.filteredPRs, m.enhancedData)
		m.table.SetRows(rows)
		m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhancing %d/%d ⏳", len(m.prs), currentCount, totalPRsToEnhance)

		// Check if we're done enhancing
		if currentCount >= totalPRsToEnhance {
			m.enhancing = false
			m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhanced top %d with detailed data ✅", len(m.prs), currentCount)
		}

	case prEnhancementCompleteMsg:
		// This is now just a fallback/cleanup message
		if m.enhancing {
			m.enhancing = false
			m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhancement complete", len(m.prs))
		}

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
			rows := createTableRowsWithEnhancement(m.filteredPRs, m.enhancedData)
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

func (m *model) View() string {
	if m.err != nil {
		return errorView(m.err)
	}
	if !m.loaded {
		return loadingView()
	}

	m.table.SetStyles(tableStyles())

	// Title
	title := titleStyle.Render("PR Pilot - Pull Request Monitor")

	// Table
	tableView := m.table.View()

	// Status message
	statusLine := ""
	if m.statusMsg != "" {
		statusLine = "\n" + statusStyle.Render(m.statusMsg)
	}

	// Help text
	helpText := "↑/↓: Navigate  •  Enter: Open PR  •  r: Refresh  •  h: Help  •  q: Quit"
	if m.filterMode != "" {
		helpText = fmt.Sprintf("Filter: %s=%s  •  c: Clear  •  %s", m.filterMode, m.filterValue, helpText)
	}

	// Extended help
	if m.showHelp {
		extendedHelp := "\n" + helpStyle.Render(`
┌─ Commands & Column Guide ────────────────────────────────┐
│ Navigation:  ↑/↓ or j/k    Navigate through PR list     │
│ Actions:     Enter         Open PR in browser           │
│              r             Manual refresh               │
│                                                         │
│ Filters:     f             Filter by author             │
│              s             Filter by status             │
│              d             Show drafts only             │
│              c             Clear all filters            │
│                                                         │
│ Activity Column Shows:                                   │
│   8c              8 comments                            │
│   5F +120/-45     5 files changed (+120, -45 lines)    │
│   8c • 5F +120/-45  8 comments AND 5 files changed    │
│   ?               Loading enhanced data...              │
│   -               No activity                           │
│                                                         │
│ Help/Exit:   h/?           Toggle this help             │
│              q/Ctrl+C      Quit application             │
└─────────────────────────────────────────────────────────┘`)
		return title + "\n" + baseStyle.Render(tableView+statusLine+extendedHelp)
	}

	return title + "\n" + baseStyle.Render(tableView+statusLine) + "\n" + helpStyle.Render(helpText)
}

// Filter methods
func (m *model) filterByAuthor() (*model, tea.Cmd) {
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

func (m *model) applyAuthorFilter(author string) (*model, tea.Cmd) {
	filtered := []*gh.PullRequest{}
	for _, pr := range m.prs {
		if pr.GetUser().GetLogin() == author {
			filtered = append(filtered, pr)
		}
	}

	m.filteredPRs = filtered
	m.filterMode = "author"
	m.filterValue = author
	rows := createTableRowsWithEnhancement(m.filteredPRs, m.enhancedData)
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d PRs by %s", len(filtered), author)

	return m, nil
}

func (m *model) filterByStatus() (*model, tea.Cmd) {
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
	rows := createTableRowsWithEnhancement(m.filteredPRs, m.enhancedData)
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d ready PRs", len(filtered))

	return m, nil
}

func (m *model) filterByDraft() (*model, tea.Cmd) {
	filtered := []*gh.PullRequest{}
	for _, pr := range m.prs {
		if pr.GetDraft() {
			filtered = append(filtered, pr)
		}
	}

	m.filteredPRs = filtered
	m.filterMode = "status"
	m.filterValue = "draft"
	rows := createTableRowsWithEnhancement(m.filteredPRs, m.enhancedData)
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d draft PRs", len(filtered))

	return m, nil
}

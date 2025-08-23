package ui

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bjess9/pr-compass/internal/batch"
	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/github"
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
type prBatchEnhancementMsg struct {
	enhancedData   map[int]enhancedPRData
	processedCount int
	successCount   int
}
type prBatchStartedMsg struct {
	resultChan chan prEnhancementUpdateMsg
}

type backgroundRefreshStartMsg struct{}
type backgroundRefreshCompleteMsg struct {
	prs []*gh.PullRequest
	cfg *config.Config
}

type lazyEnhancementMsg struct {
	prNumber int
}

type singlePREnhancementCompleteMsg struct {
	prData enhancedPRData
	error  error
}

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

	// Batch processing for concurrent PR enhancement
	batchManager *batch.Manager[*gh.PullRequest, enhancedPRData]
	activeBatchChan chan prEnhancementUpdateMsg // Channel for streaming batch results

	// Context for cancellation support
	ctx    context.Context
	cancel context.CancelFunc

	// Cache for improved performance
	prCache *cache.PRCache
	
	// API optimization tracking  
	rateLimitInfo    *github.RateLimitInfo
	
	// Background refresh state
	backgroundRefreshing bool
	
	// Lazy loading state
	lastSelectedPRIndex int
	enhancementQueue    map[int]bool // Track PRs being enhanced
}

func InitialModel(token string) model {
	columns := createTableColumns()
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	t.Focus()

	// Create context with cancellation support for proper resource cleanup
	ctx, cancel := context.WithCancel(context.Background())

	// Create worker function for batch PR enhancement
	enhancePRWorker := func(batchCtx context.Context, pr *gh.PullRequest) (enhancedPRData, error) {
		// Create timeout context for this specific PR (10 seconds)
		prCtx, prCancel := context.WithTimeout(batchCtx, 10*time.Second)
		defer prCancel()

		// Get GitHub client
		client, err := github.NewClient(token)
		if err != nil {
			return enhancedPRData{Number: pr.GetNumber()}, err
		}

		// Fetch enhanced data for this PR (we'll need to extract this method)
		return fetchEnhancedPRDataStatic(prCtx, client, pr)
	}

	// Create batch manager with 5 concurrent workers for optimal performance
	batchManager := batch.NewManager(5, enhancePRWorker)

	// Initialize PR cache
	prCache, err := cache.NewPRCache()
	if err != nil {
		log.Printf("Warning: Failed to initialize PR cache: %v", err)
		prCache = nil // Continue without caching
	}

	return model{
		table:               t,
		token:               token,
		refreshIntervalMins: 5, // default, will be updated from config
		enhancedData:        make(map[int]enhancedPRData),
		batchManager:        batchManager,
		ctx:                 ctx,
		cancel:              cancel,
		prCache:             prCache,
		lastSelectedPRIndex: -1,
		enhancementQueue:    make(map[int]bool),
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
	// Create a timeout context for fetching PRs (30 seconds should be enough)
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		return errMsg{err}
	}

	var prs []*gh.PullRequest
	// Use optimized fetching with GraphQL + Caching + Rate Limiting
	if m.prCache != nil {
		prs, err = github.FetchPRsFromConfigOptimized(ctx, cfg, m.token, m.prCache)
	} else {
		prs, err = github.FetchPRsFromConfig(ctx, cfg, m.token)
	}
	
	if err != nil {
		return errMsg{err}
	}
	return PrsWithConfigMsg{Prs: prs, Cfg: cfg}
}

// backgroundFetchPRs performs a background refresh without disrupting the UI
func (m *model) backgroundFetchPRs() tea.Cmd {
	return func() tea.Msg {
		// Create a timeout context for fetching PRs
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		cfg, err := config.LoadConfig()
		if err != nil {
			// Silently fail background refresh on config errors
			return backgroundRefreshCompleteMsg{prs: m.prs, cfg: nil}
		}

		var prs []*gh.PullRequest
		// Use optimized fetching for background refresh
		if m.prCache != nil {
			prs, err = github.FetchPRsFromConfigOptimized(ctx, cfg, m.token, m.prCache)
		} else {
			prs, err = github.FetchPRsFromConfig(ctx, cfg, m.token)
		}
		
		if err != nil {
			// On error, return current data (graceful degradation)
			return backgroundRefreshCompleteMsg{prs: m.prs, cfg: cfg}
		}
		
		return backgroundRefreshCompleteMsg{prs: prs, cfg: cfg}
	}
}

// Background enhancement command that fetches individual PR details
func (m *model) enhancePRs() tea.Cmd {
	return m.enhancePRsWithBatch()
}

// listenForBatchResults creates a command that listens for streaming batch results
func (m *model) listenForBatchResults(resultChan chan prEnhancementUpdateMsg) tea.Cmd {
	return func() tea.Msg {
		// Listen for the next result from the channel
		result, ok := <-resultChan
		if !ok {
			// Channel closed, all results processed
			return prEnhancementCompleteMsg{}
		}
		
		// Return the individual result message
		return result
	}
}

// enhanceSinglePR enhances a single PR on-demand for lazy loading
func (m *model) enhanceSinglePR(prNumber int) tea.Cmd {
	return func() tea.Msg {
		// Find the PR by number
		var targetPR *gh.PullRequest
		for _, pr := range m.prs {
			if pr.GetNumber() == prNumber {
				targetPR = pr
				break
			}
		}

		if targetPR == nil {
			return singlePREnhancementCompleteMsg{
				prData: enhancedPRData{Number: prNumber},
				error:  fmt.Errorf("PR #%d not found", prNumber),
			}
		}

		// Create timeout context for this specific PR
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		// Get GitHub client
		client, err := github.NewClient(m.token)
		if err != nil {
			return singlePREnhancementCompleteMsg{
				prData: enhancedPRData{Number: prNumber},
				error:  err,
			}
		}

		// Fetch enhanced data for this PR
		enhancedData, err := fetchEnhancedPRDataStatic(ctx, client, targetPR)
		return singlePREnhancementCompleteMsg{
			prData: enhancedData,
			error:  err,
		}
	}
}

// checkAndEnhanceSelected checks if the currently selected PR needs enhancement
func (m *model) checkAndEnhanceSelected() tea.Cmd {
	if !m.loaded || len(m.filteredPRs) == 0 {
		return nil
	}

	currentIndex := m.table.Cursor()
	if currentIndex < 0 || currentIndex >= len(m.filteredPRs) {
		return nil
	}

	// If selection hasn't changed, no need to do anything
	if currentIndex == m.lastSelectedPRIndex {
		return nil
	}

	m.lastSelectedPRIndex = currentIndex
	
	// Get the selected PR and nearby PRs for buffer enhancement
	selectedPR := m.filteredPRs[currentIndex]
	prNumber := selectedPR.GetNumber()

	// Check if this PR is already enhanced or being enhanced
	m.enhancementMutex.RLock()
	_, alreadyEnhanced := m.enhancedData[prNumber]
	_, beingEnhanced := m.enhancementQueue[prNumber]
	m.enhancementMutex.RUnlock()

	if alreadyEnhanced || beingEnhanced {
		return nil
	}

	// Mark as being enhanced and start enhancement
	m.enhancementMutex.Lock()
	m.enhancementQueue[prNumber] = true
	m.enhancementMutex.Unlock()

	return tea.Batch(
		func() tea.Msg { return lazyEnhancementMsg{prNumber: prNumber} },
		m.enhanceSinglePR(prNumber),
	)
}

// enhancePRsWithBatch - kept for compatibility but now does minimal enhancement
func (m *model) enhancePRsWithBatch() tea.Cmd {
	// For lazy loading, we don't enhance all PRs upfront
	// Instead, we enhance just the first visible PR to get started
	if len(m.prs) > 0 {
		firstPR := m.prs[0]
		prNumber := firstPR.GetNumber()
		
		// Check if already enhanced
		m.enhancementMutex.RLock()
		_, alreadyEnhanced := m.enhancedData[prNumber]
		m.enhancementMutex.RUnlock()
		
		if !alreadyEnhanced {
			m.enhancementMutex.Lock()
			m.enhancementQueue[prNumber] = true
			m.enhancementMutex.Unlock()
			
			return m.enhanceSinglePR(prNumber)
		}
	}
	
	return func() tea.Msg { return prEnhancementCompleteMsg{} }
}

// fetchEnhancedPRDataStatic is a static version of fetchEnhancedPRData for batch processing
func fetchEnhancedPRDataStatic(ctx context.Context, client *gh.Client, pr *gh.PullRequest) (enhancedPRData, error) {
	// Validate PR structure to avoid nil pointer panics
	if pr == nil {
		return enhancedPRData{}, fmt.Errorf("PR is nil")
	}
	if pr.GetBase() == nil || pr.GetBase().GetRepo() == nil {
		return enhancedPRData{}, fmt.Errorf("PR base or repository is nil for PR #%d", pr.GetNumber())
	}
	if pr.GetBase().GetRepo().GetOwner() == nil {
		return enhancedPRData{}, fmt.Errorf("PR repository owner is nil for PR #%d", pr.GetNumber())
	}

	owner := pr.GetBase().GetRepo().GetOwner().GetLogin()
	repo := pr.GetBase().GetRepo().GetName()
	number := pr.GetNumber()

	// Additional validation for required fields
	if owner == "" {
		return enhancedPRData{}, fmt.Errorf("PR owner is empty for PR #%d", number)
	}
	if repo == "" {
		return enhancedPRData{}, fmt.Errorf("PR repository name is empty for PR #%d", number)
	}

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
	if pr.GetHead() != nil {
		if sha := pr.GetHead().GetSHA(); sha != "" {
			checks, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
			if err == nil && checks != nil {
				checksStatus = determineChecksStatus(checks.CheckRuns)
			}
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

		// Clear previous enhanced data and cancel ongoing batch processing
		m.enhancementMutex.Lock()
		m.enhancedData = make(map[int]enhancedPRData)
		m.enhancing = false
		m.enhancedCount = 0
		m.activeBatchChan = nil // Cancel any ongoing batch processing
		m.enhancementQueue = make(map[int]bool) // Clear enhancement queue
		m.enhancementMutex.Unlock()

		rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
		m.table.SetRows(rows)

		// Create informative status message with org/topic info and rate limit
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
		
		// Add rate limit info to status if available
		if m.rateLimitInfo != nil {
			resetTime := m.rateLimitInfo.ResetAt.Format("15:04")
			statusInfo += fmt.Sprintf(" | API: %d/%d (resets %s)", 
				m.rateLimitInfo.Remaining, m.rateLimitInfo.Limit, resetTime)
		}
		
		// Add background refresh indicator
		if m.backgroundRefreshing {
			statusInfo += " 🔄"
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
		// Start background refresh if we already have data
		if m.loaded && len(m.prs) > 0 {
			m.backgroundRefreshing = true
			return m, tea.Batch(
				func() tea.Msg { return backgroundRefreshStartMsg{} },
				refreshCmd(m.refreshIntervalMins),
			)
		}
		// First load - use regular refresh
		return m, tea.Batch(m.fetchPRs, refreshCmd(m.refreshIntervalMins))

	case backgroundRefreshStartMsg:
		return m, m.backgroundFetchPRs()

	case backgroundRefreshCompleteMsg:
		// Update data seamlessly without clearing enhanced data
		oldPRCount := len(m.prs)
		m.prs = sortPRsByNewest(msg.prs)
		m.filteredPRs = m.prs
		m.backgroundRefreshing = false
		
		// Update config if provided
		if msg.cfg != nil {
			m.refreshIntervalMins = msg.cfg.RefreshIntervalMinutes
		}

		// Update table rows with progressive loading indicators
		rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
		m.table.SetRows(rows)

		// Update status message with change indication
		newPRCount := len(m.prs)
		var changeIndicator string
		if newPRCount > oldPRCount {
			changeIndicator = fmt.Sprintf(" (+%d new)", newPRCount-oldPRCount)
		} else if newPRCount < oldPRCount {
			changeIndicator = fmt.Sprintf(" (-%d closed)", oldPRCount-newPRCount)
		} else {
			changeIndicator = " (updated)"
		}

		statusInfo := fmt.Sprintf("Loaded %d PRs%s", newPRCount, changeIndicator)
		if msg.cfg != nil {
			switch msg.cfg.Mode {
			case "topics":
				if msg.cfg.TopicOrg != "" && len(msg.cfg.Topics) > 0 {
					statusInfo += fmt.Sprintf(" from %s (topics: %v)", msg.cfg.TopicOrg, msg.cfg.Topics)
				}
			case "organization":
				if msg.cfg.Organization != "" {
					statusInfo += fmt.Sprintf(" from %s organization", msg.cfg.Organization)
				}
			case "teams":
				if msg.cfg.Organization != "" && len(msg.cfg.Teams) > 0 {
					statusInfo += fmt.Sprintf(" from %s teams: %v", msg.cfg.Organization, msg.cfg.Teams)
				}
			}
		}
		
		// Add rate limit info if available
		if m.rateLimitInfo != nil {
			resetTime := m.rateLimitInfo.ResetAt.Format("15:04")
			statusInfo += fmt.Sprintf(" | API: %d/%d (resets %s)", 
				m.rateLimitInfo.Remaining, m.rateLimitInfo.Limit, resetTime)
		}
		
		m.statusMsg = statusInfo

		// Start enhancement for any new PRs
		if newPRCount > 0 {
			enhanceCmd := func() tea.Msg {
				time.Sleep(100 * time.Millisecond)
				return prEnhancementStartMsg{}
			}
			return m, enhanceCmd
		}

		return m, nil

	case lazyEnhancementMsg:
		// Visual indicator that a PR is being enhanced
		return m, nil

	case singlePREnhancementCompleteMsg:
		// Update enhanced data for single PR
		m.enhancementMutex.Lock()
		if msg.error == nil {
			m.enhancedData[msg.prData.Number] = msg.prData
		}
		delete(m.enhancementQueue, msg.prData.Number) // Remove from enhancement queue
		m.enhancementMutex.Unlock()

		// Update table rows with new enhanced data
		rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
		m.table.SetRows(rows)

		return m, nil

	case errMsg:
		m.err = msg.err

	case error:
		m.err = msg

	case prEnhancementStartMsg:
		m.enhancementMutex.Lock()
		m.enhancing = true
		m.enhancedCount = 0 // Reset count when starting enhancement
		m.enhancementMutex.Unlock()

		m.statusMsg = fmt.Sprintf("Loaded %d PRs - lazy loading enabled", len(m.prs))
		return m, m.enhancePRs()

	case prEnhancementUpdateMsg:
		// Handle individual PR enhancement updates (works for both old and new batch approaches)
		var continueListening tea.Cmd
		
		if msg.error == nil {
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
			rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
			m.table.SetRows(rows)
			m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhancing %d/%d ⏳", len(m.prs), currentCount, totalPRsToEnhance)

			// Check if we're done enhancing
			if currentCount >= totalPRsToEnhance {
				m.enhancing = false
				m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhanced top %d with detailed data ✅", len(m.prs), currentCount)
			}
		}
		
		// Continue listening for more results if we have an active batch channel
		if m.activeBatchChan != nil {
			continueListening = m.listenForBatchResults(m.activeBatchChan)
		}
		
		return m, continueListening

	case prBatchStartedMsg:
		// Store the channel and start listening for streaming results
		m.activeBatchChan = msg.resultChan
		return m, m.listenForBatchResults(msg.resultChan)

	case prBatchEnhancementMsg:
		// Handle batch enhancement results (legacy - kept for compatibility)
		m.enhancementMutex.Lock()
		// Update all enhanced data at once
		for prNumber, prData := range msg.enhancedData {
			m.enhancedData[prNumber] = prData
		}
		m.enhancedCount = msg.successCount
		m.enhancing = false // Mark enhancement as complete
		m.enhancementMutex.Unlock()

		// Update table rows with all enhanced data
		rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
		m.table.SetRows(rows)
		
		// Update status message
		if msg.successCount == msg.processedCount {
			m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhanced top %d with detailed data ✅", len(m.prs), msg.successCount)
		} else {
			m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhanced %d/%d with detailed data ⚠️", len(m.prs), msg.successCount, msg.processedCount)
		}

	case prEnhancementCompleteMsg:
		// Enhancement complete - cleanup batch channel
		m.activeBatchChan = nil
		if m.enhancing {
			m.enhancing = false
			m.statusMsg = fmt.Sprintf("Loaded %d PRs - enhancement complete", len(m.prs))
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
					// Cancel all ongoing operations before quitting
		if m.cancel != nil {
			m.cancel()
		}
		// Stop batch manager gracefully
		if m.batchManager != nil {
			m.batchManager.Stop()
		}
		// Clear active batch channel
		m.activeBatchChan = nil
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
			rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
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

	// Check if cursor position changed for lazy loading
	if m.loaded && len(m.filteredPRs) > 0 {
		enhanceCmd := m.checkAndEnhanceSelected()
		if enhanceCmd != nil {
			cmd = tea.Batch(cmd, enhanceCmd)
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
	title := titleStyle.Render("PR Compass - Pull Request Monitor")

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
	rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
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
	rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
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
	rows := createProgressiveTableRows(m.filteredPRs, m.enhancedData, m.enhancementQueue)
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d draft PRs", len(filtered))

	return m, nil
}

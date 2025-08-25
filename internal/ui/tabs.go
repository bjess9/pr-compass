package ui

import (
	"context"
	"sync"
	"time"

	"github.com/bjess9/pr-compass/internal/batch"
	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/github"
	gh "github.com/google/go-github/v55/github"

	"github.com/charmbracelet/bubbles/table"
)



// TabConfig represents the configuration for a single tab
type TabConfig struct {
	Name         string `mapstructure:"name" yaml:"name"`                 // Display name for the tab
	Mode         string `mapstructure:"mode" yaml:"mode"`                 // "repos", "organization", "teams", "search", "topics"
	
	// Mode-specific configurations
	Repos        []string `mapstructure:"repos" yaml:"repos,omitempty"`
	Organization string   `mapstructure:"organization" yaml:"organization,omitempty"`
	Teams        []string `mapstructure:"teams" yaml:"teams,omitempty"`
	SearchQuery  string   `mapstructure:"search_query" yaml:"search_query,omitempty"`
	Topics       []string `mapstructure:"topics" yaml:"topics,omitempty"`
	TopicOrg     string   `mapstructure:"topic_org" yaml:"topic_org,omitempty"`
	
	// Filtering options (can be different per tab)
	ExcludeBots    bool     `mapstructure:"exclude_bots" yaml:"exclude_bots,omitempty"`
	ExcludeAuthors []string `mapstructure:"exclude_authors" yaml:"exclude_authors,omitempty"`
	ExcludeTitles  []string `mapstructure:"exclude_titles" yaml:"exclude_titles,omitempty"`
	IncludeDrafts  bool     `mapstructure:"include_drafts" yaml:"include_drafts,omitempty"`
	
	// Tab-specific refresh interval
	RefreshIntervalMinutes int `mapstructure:"refresh_interval_minutes" yaml:"refresh_interval_minutes,omitempty"`
	
	// Performance options
	MaxPRs int `mapstructure:"max_prs" yaml:"max_prs,omitempty"` // Maximum PRs to fetch for this tab
}

// ConvertToConfig converts a TabConfig to the standard Config format
func (tc *TabConfig) ConvertToConfig() *config.Config {
	maxPRs := tc.MaxPRs
	if maxPRs == 0 {
		maxPRs = 50 // Default limit
	}
	
	return &config.Config{
		Mode:                   tc.Mode,
		Repos:                  tc.Repos,
		Organization:          tc.Organization,
		Teams:                 tc.Teams,
		SearchQuery:           tc.SearchQuery,
		Topics:                tc.Topics,
		TopicOrg:              tc.TopicOrg,
		ExcludeBots:           tc.ExcludeBots,
		ExcludeAuthors:        tc.ExcludeAuthors,
		ExcludeTitles:         tc.ExcludeTitles,
		IncludeDrafts:         tc.IncludeDrafts,
		RefreshIntervalMinutes: tc.RefreshIntervalMinutes,
		MaxPRs:                 maxPRs,
	}
}

// TabState represents the state of a single tab (similar to current model)
type TabState struct {
	Config   *TabConfig
	
	// UI State
	Table       table.Model
	ShowHelp    bool
	FilterMode  string // "", "author", "repo", "status"
	FilterValue string
	StatusMsg   string
	
	// Data State
	PRs         []*gh.PullRequest
	FilteredPRs []*gh.PullRequest
	Loaded      bool
	Error       error
	
	// Enhanced data tracking
	EnhancedData     map[int]enhancedPRData // PR number -> enhanced data
	EnhancementMutex sync.RWMutex
	Enhancing        bool
	EnhancedCount    int
	
	// Background processing
	BatchManager    *batch.Manager[*gh.PullRequest, enhancedPRData]
	ActiveBatchChan chan prEnhancementUpdateMsg
	
	// Context for cancellation
	Ctx    context.Context
	Cancel context.CancelFunc
	
	// Cache
	PRCache *cache.PRCache
	
	// State management
	BackgroundRefreshing bool
	LastSelectedPRIndex  int
	EnhancementQueue     map[int]bool
	
	// Tab metadata
	LastRefreshTime time.Time
	LoadTime        time.Time
}

// TabManager manages multiple tabs and their states
type TabManager struct {
	Tabs          []*TabState
	ActiveTabIdx  int
	Token         string
	
	// Global settings
	GlobalRefreshInterval int
	
	// Tab switching state
	TabSwitchMode bool // When true, show tab numbers for quick switching
	
	// Rate limiting and coordination
	RateLimiter   *GlobalRateLimiter
	SharedCache   *SharedCache
	
	// Request coordination
	mu                sync.RWMutex
	refreshScheduler  *RefreshScheduler
}

// NewTabState creates a new tab state with the given configuration
func NewTabState(tabConfig *TabConfig, token string) *TabState {
	// Create table columns
	columns := createTableColumns()
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10), // Will be dynamically adjusted based on terminal size
	)
	t.Focus()
	
	// Apply table styles to ensure proper sizing behavior
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Bold(false)
	t.SetStyles(s)
	
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create worker function for batch PR enhancement
	enhancePRWorker := func(batchCtx context.Context, pr *gh.PullRequest) (enhancedPRData, error) {
		prCtx, prCancel := context.WithTimeout(batchCtx, 10*time.Second)
		defer prCancel()
		
		// Get GitHub client
		client, err := github.NewClient(token)
		if err != nil {
			return enhancedPRData{Number: pr.GetNumber()}, err
		}
		
		return fetchEnhancedPRDataStatic(prCtx, client, pr)
	}
	
	// Create batch manager
	batchManager := batch.NewManager(5, enhancePRWorker)
	
	// Initialize PR cache
	prCache, err := cache.NewPRCache()
	if err != nil {
		// Continue without caching if cache initialization fails
		prCache = nil
	}
	
	// Set default refresh interval if not configured
	refreshInterval := tabConfig.RefreshIntervalMinutes
	if refreshInterval == 0 {
		refreshInterval = 5 // default: 5 minutes
	}
	
	return &TabState{
		Config:               tabConfig,
		Table:                t,
		EnhancedData:         make(map[int]enhancedPRData),
		BatchManager:         batchManager,
		Ctx:                  ctx,
		Cancel:               cancel,
		PRCache:              prCache,
		LastSelectedPRIndex:  -1,
		EnhancementQueue:     make(map[int]bool),
		LoadTime:             time.Now(),
	}
}

// NewTabManager creates a new tab manager with rate limiting
func NewTabManager(token string) *TabManager {
	// Initialize global rate limiter if not already done
	if GlobalLimiter == nil {
		InitGlobalRateLimiter()
	}
	
	manager := &TabManager{
		Tabs:                  make([]*TabState, 0),
		ActiveTabIdx:          0,
		Token:                 token,
		GlobalRefreshInterval: 5,
		TabSwitchMode:         false,
		RateLimiter:          GlobalLimiter,
		SharedCache:          GlobalLimiter.sharedCache,
		refreshScheduler:     NewRefreshScheduler(),
	}
	
	return manager
}

// AddTab adds a new tab with the given configuration
func (tm *TabManager) AddTab(tabConfig *TabConfig) *TabState {
	tabState := NewTabState(tabConfig, tm.Token)
	tm.Tabs = append(tm.Tabs, tabState)
	
	// Register with refresh scheduler for rate limiting coordination
	if tm.refreshScheduler != nil {
		refreshInterval := time.Duration(tabConfig.RefreshIntervalMinutes) * time.Minute
		if refreshInterval == 0 {
			refreshInterval = time.Duration(tm.GlobalRefreshInterval) * time.Minute
		}
		
		// Determine priority based on refresh interval (faster = higher priority)
		var priority RefreshPriority = RefreshPriorityNormal
		if refreshInterval < 3*time.Minute {
			priority = RefreshPriorityHigh
		} else if refreshInterval > 10*time.Minute {
			priority = RefreshPriorityLow
		}
		
		tm.refreshScheduler.AddTab(tabConfig.Name, refreshInterval, priority)
	}
	
	return tabState
}

// GetActiveTab returns the currently active tab
func (tm *TabManager) GetActiveTab() *TabState {
	if tm.ActiveTabIdx >= 0 && tm.ActiveTabIdx < len(tm.Tabs) {
		return tm.Tabs[tm.ActiveTabIdx]
	}
	return nil
}

// SwitchToTab switches to the tab at the given index
func (tm *TabManager) SwitchToTab(index int) bool {
	if index >= 0 && index < len(tm.Tabs) {
		tm.ActiveTabIdx = index
		return true
	}
	return false
}

// NextTab switches to the next tab (wraps around)
func (tm *TabManager) NextTab() {
	if len(tm.Tabs) > 1 {
		tm.ActiveTabIdx = (tm.ActiveTabIdx + 1) % len(tm.Tabs)
	}
}

// PrevTab switches to the previous tab (wraps around)
func (tm *TabManager) PrevTab() {
	if len(tm.Tabs) > 1 {
		tm.ActiveTabIdx = (tm.ActiveTabIdx - 1 + len(tm.Tabs)) % len(tm.Tabs)
	}
}

// CloseTab closes the tab at the given index
func (tm *TabManager) CloseTab(index int) bool {
	if index < 0 || index >= len(tm.Tabs) || len(tm.Tabs) <= 1 {
		return false // Can't close the last tab
	}
	
	// Cancel the tab's context to stop background operations
	if tm.Tabs[index].Cancel != nil {
		tm.Tabs[index].Cancel()
	}
	
	// Remove the tab
	tm.Tabs = append(tm.Tabs[:index], tm.Tabs[index+1:]...)
	
	// Adjust active tab index
	if tm.ActiveTabIdx >= len(tm.Tabs) {
		tm.ActiveTabIdx = len(tm.Tabs) - 1
	} else if tm.ActiveTabIdx > index {
		tm.ActiveTabIdx--
	}
	
	return true
}

// GetTabCount returns the number of tabs
func (tm *TabManager) GetTabCount() int {
	return len(tm.Tabs)
}

// GetTabNames returns the names of all tabs for display
func (tm *TabManager) GetTabNames() []string {
	names := make([]string, len(tm.Tabs))
	for i, tab := range tm.Tabs {
		names[i] = tab.Config.Name
	}
	return names
}

// Cleanup properly closes all tabs and their resources
func (tm *TabManager) Cleanup() {
	for _, tab := range tm.Tabs {
		if tab.Cancel != nil {
			tab.Cancel()
		}
		if tab.BatchManager != nil {
			tab.BatchManager.Stop()
		}
	}
}
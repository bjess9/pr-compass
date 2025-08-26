package types

import (
	"time"

	"github.com/bjess9/pr-compass/internal/config"
	gh "github.com/google/go-github/v55/github"
)

// PRData represents a pull request with enhanced information
type PRData struct {
	*gh.PullRequest
	Enhanced *EnhancedData `json:"enhanced,omitempty"`
}

// EnhancedData contains additional PR information from detailed API calls
type EnhancedData struct {
	Number         int       `json:"number"`
	Comments       int       `json:"comments"`
	ReviewComments int       `json:"review_comments"`
	ReviewStatus   string    `json:"review_status"` // "approved", "changes_requested", "pending", "unknown"
	ChecksStatus   string    `json:"checks_status"` // "success", "failure", "pending", "unknown"
	Mergeable      string    `json:"mergeable"`     // "clean", "conflicts", "unknown"
	Additions      int       `json:"additions"`
	Deletions      int       `json:"deletions"`
	ChangedFiles   int       `json:"changed_files"`
	EnhancedAt     time.Time `json:"enhanced_at"`
}

// FilterOptions represents filtering criteria for PRs
type FilterOptions struct {
	Mode   string `json:"mode"`   // "", "author", "repo", "status", "draft"
	Value  string `json:"value"`  // The filter value
	Active bool   `json:"active"` // Whether filter is currently applied
}

// UIState represents the current state of the user interface
type UIState struct {
	ShowHelp    bool          `json:"show_help"`
	Filter      FilterOptions `json:"filter"`
	StatusMsg   string        `json:"status_msg"`
	SelectedPR  int           `json:"selected_pr"` // Index of currently selected PR
	TableCursor int           `json:"table_cursor"`
}

// AppState represents the overall application state
type AppState struct {
	// Data state
	PRs         []*PRData      `json:"prs"`
	FilteredPRs []*PRData      `json:"filtered_prs"`
	Config      *config.Config `json:"config"`

	// UI state
	UI UIState `json:"ui"`

	// Loading states
	Loaded               bool `json:"loaded"`
	BackgroundRefreshing bool `json:"background_refreshing"`
	Enhancing            bool `json:"enhancing"`

	// Enhancement tracking
	EnhancedCount    int          `json:"enhanced_count"`
	EnhancementQueue map[int]bool `json:"enhancement_queue"`

	// Error state
	Error error `json:"error,omitempty"`
}

// PRDisplayInfo contains formatted information for displaying a PR in the table
type PRDisplayInfo struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Repo        string `json:"repo"`
	Status      string `json:"status"`
	Reviews     string `json:"reviews"`
	Comments    string `json:"comments"`
	Files       string `json:"files"`
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	IsEnhancing bool   `json:"is_enhancing"`
	IsEnhanced  bool   `json:"is_enhanced"`
}

// Legacy type aliases for backward compatibility (used in utils.go and tabs.go)
type enhancedPRData = EnhancedData

// Message types for background enhancement
type PrEnhancementUpdateMsg struct {
	PrData EnhancedData
	Error  error
}

// Error message type (exported for use in commands.go)
type ErrorMsg struct {
	Error error
}

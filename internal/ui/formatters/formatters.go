package formatters

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bjess9/pr-compass/internal/ui/types"
	"github.com/charmbracelet/bubbles/table"
	gh "github.com/google/go-github/v55/github"
)

// PRFormatter handles formatting of PR data for display
type PRFormatter struct{}

// NewPRFormatter creates a new PR formatter
func NewPRFormatter() *PRFormatter {
	return &PRFormatter{}
}

// FormatNumber formats a number for display
func (f *PRFormatter) FormatNumber(n int) string {
	if n == 0 {
		return "-"
	}
	return strconv.Itoa(n)
}

// FormatChanges formats additions and deletions
func (f *PRFormatter) FormatChanges(additions, deletions int) string {
	if additions == 0 && deletions == 0 {
		return ""
	}
	return fmt.Sprintf("+%d/-%d", additions, deletions)
}

// HumanizeTimeSince formats duration since a given time in a human-readable way
func (f *PRFormatter) HumanizeTimeSince(t time.Time) string {
	duration := time.Since(t)

	// Handle very recent times - show < 1m instead of "now"
	if duration < time.Minute {
		return "< 1m"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		return fmt.Sprintf("%dw", weeks)
	} else {
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%dmo", months)
	}
}

// GetBasicStatus returns basic PR status
func (f *PRFormatter) GetBasicStatus(pr *gh.PullRequest) string {
	if pr.GetDraft() {
		return "Draft"
	}

	mergeableState := pr.GetMergeableState()
	switch mergeableState {
	case "dirty":
		return "Conflicts"
	case "blocked":
		return "Blocked"
	case "behind":
		return "Behind"
	case "clean":
		return "Ready"
	case "unstable":
		return "Checks"
	default:
		return "Ready"
	}
}

// GetEnhancedStatus returns enhanced PR status
func (f *PRFormatter) GetEnhancedStatus(enhanced *types.EnhancedData, baseStatus string) string {
	switch enhanced.Mergeable {
	case "clean":
		if enhanced.ChecksStatus == "failure" {
			return "[!] Checks"
		}
		return "[✓] Ready"
	case "conflicts":
		return "[X] Conflicts"
	default:
		return baseStatus
	}
}

// GetBasicReviewStatus returns basic review status
func (f *PRFormatter) GetBasicReviewStatus(pr *gh.PullRequest) string {
	if pr.GetDraft() {
		return "WIP"
	}

	totalReviewers := len(pr.RequestedReviewers) + len(pr.RequestedTeams)
	if totalReviewers > 0 {
		return fmt.Sprintf("0/%d", totalReviewers)
	}

	daysSinceUpdated := time.Since(pr.GetUpdatedAt().Time).Hours() / 24
	if daysSinceUpdated > 5 {
		return "Stale"
	}

	if daysSinceUpdated < 1 {
		return "Recent"
	}

	return "None"
}

// GetEnhancedReviewStatus returns enhanced review status
func (f *PRFormatter) GetEnhancedReviewStatus(enhanced *types.EnhancedData) string {
	switch enhanced.ReviewStatus {
	case "approved":
		return "Approved"
	case "changes_requested":
		return "Changes"
	case "pending":
		return "Pending"
	case "no_review":
		return "No Review"
	default:
		return "Unknown"
	}
}

// CreateTableColumns creates table columns with proper widths
func (f *PRFormatter) CreateTableColumns() []table.Column {
	totalWidth := f.getTerminalWidth() - 12 // Account for borders and padding

	// Calculate optimal widths
	prNameWidth := max(32, totalWidth*22/100)
	authorWidth := max(10, totalWidth*10/100)
	repoWidth := max(12, totalWidth*12/100)
	statusWidth := max(10, totalWidth*12/100)
	reviewsWidth := max(8, totalWidth*9/100)
	commentsWidth := max(8, totalWidth*7/100)
	filesWidth := max(10, totalWidth*10/100)
	createdWidth := max(8, totalWidth*9/100)
	updatedWidth := max(8, totalWidth*9/100)

	return []table.Column{
		{Title: "📋 Pull Request", Width: prNameWidth},
		{Title: "👤 Author", Width: authorWidth},
		{Title: "📦 Repo", Width: repoWidth},
		{Title: "⚡ Status/CI", Width: statusWidth},
		{Title: "👀 Review", Width: reviewsWidth},
		{Title: "💬 Comments", Width: commentsWidth},
		{Title: "📁 Files", Width: filesWidth},
		{Title: "📅 Created", Width: createdWidth},
		{Title: "🕐 Updated", Width: updatedWidth},
	}
}

// getTerminalWidth returns terminal width with fallback
func (f *PRFormatter) getTerminalWidth() int {
	return 160 // Conservative default
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
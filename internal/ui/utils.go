package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/bjess9/pr-compass/internal/errors"
	"github.com/bjess9/pr-compass/internal/ui/formatters"
	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"

	"github.com/charmbracelet/bubbles/table"
)

func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	if _, err := exec.LookPath("wslpath"); err == nil {
		return true
	}

	if content, err := exec.Command("uname", "-r").Output(); err == nil {
		if strings.Contains(strings.ToLower(string(content)), "microsoft") {
			return true
		}
	}

	if os.Getenv("WSL_INTEROP") != "" || os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	return false
}

// getTerminalWidth returns the current terminal width, with fallback to reasonable default
func getTerminalWidth() int {
	// Try to get actual terminal width
	if width := getActualTerminalWidth(); width > 0 {
		return width
	}
	// Use a larger fallback that fills most modern terminal windows
	return 160
}

// getActualTerminalWidth attempts to get the real terminal width
func getActualTerminalWidth() int {
	// Try to get terminal size from bubbletea's term package
	// This is a more robust approach than trying to implement our own
	// For now, we'll use a conservative default that works well
	return 0 // Use fallback for consistency
}

func createTableColumns() []table.Column {
	totalWidth := getTerminalWidth() - 12 // Account for borders, padding, selection indicators, and scrollbars

	// Minimum widths to ensure readability
	minPRWidth := 32      // PR title only (no number)
	minAuthorWidth := 10  // Author only
	minRepoWidth := 12    // Repo only
	minStatusWidth := 10  // Status + CI
	minReviewsWidth := 8  // Review status
	minCommentsWidth := 8 // Comments only
	minFilesWidth := 10   // Files only
	minCreatedWidth := 8  // Created time
	minUpdatedWidth := 8  // Updated time

	// Calculate optimal widths for 9 columns (split Author/Repo into separate columns)
	prNameWidth := max(minPRWidth, totalWidth*22/100)        // 22% - PR title (no number)
	authorWidth := max(minAuthorWidth, totalWidth*10/100)    // 10% - Author only
	repoWidth := max(minRepoWidth, totalWidth*12/100)        // 12% - Repo only (guaranteed visibility)
	statusWidth := max(minStatusWidth, totalWidth*12/100)    // 12% - Status + CI
	reviewsWidth := max(minReviewsWidth, totalWidth*9/100)   // 9%  - Review status
	commentsWidth := max(minCommentsWidth, totalWidth*7/100) // 7%  - Comments only
	filesWidth := max(minFilesWidth, totalWidth*10/100)      // 10% - Files only
	createdWidth := max(minCreatedWidth, totalWidth*9/100)   // 9%  - Created time
	updatedWidth := max(minUpdatedWidth, totalWidth*9/100)   // 9%  - Updated time

	return []table.Column{
		{Title: "📋 Pull Request", Width: prNameWidth}, // PR title only (no number)
		{Title: "👤 Author", Width: authorWidth},       // Author only
		{Title: "📦 Repo", Width: repoWidth},           // Repo only (guaranteed visible)
		{Title: "⚡ Status/CI", Width: statusWidth},    // Status + CI
		{Title: "👀 Review", Width: reviewsWidth},      // Review process
		{Title: "💬 Comments", Width: commentsWidth},   // Comments only
		{Title: "📁 Files", Width: filesWidth},         // File changes only
		{Title: "📅 Created", Width: createdWidth},     // When PR was created
		{Title: "🕐 Updated", Width: updatedWidth},     // When PR was last updated
	}
}

// max returns the maximum of two integers (helper function for Go < 1.21)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func createTableRows(prs []*gh.PullRequest) []table.Row {
	// Get column widths for dynamic truncation
	columns := createTableColumns()
	prColumnWidth := columns[0].Width

	rows := make([]table.Row, len(prs))
	for i, pr := range prs {
		// PR Name (smart formatting with ticket detection)
		prName := formatPRTitle(pr, prColumnWidth)

		// Author and Repo (now separate columns for better visibility)
		author := "Unknown"
		if pr.GetUser() != nil && pr.GetUser().GetLogin() != "" {
			author = pr.GetUser().GetLogin()
		}

		repoFullName := "Unknown"
		if pr.GetBase() != nil && pr.GetBase().GetRepo() != nil && pr.GetBase().GetRepo().GetFullName() != "" {
			repoFullName = pr.GetBase().GetRepo().GetFullName()
		}
		repoParts := strings.Split(repoFullName, "/")
		repoName := repoFullName
		if len(repoParts) == 2 {
			repoName = repoParts[1] // Just the repo name
		}

		// Status + CI (compact single-line format)
		mergeStatus := getPRStatusIndicator(pr)
		ciStatus := "CI:?" // Default CI indicator
		statusCombined := mergeStatus + " " + ciStatus

		// Review Status
		reviews := getPRReviewIndicator(pr)

		// Comments (split from activity)
		comments := getPRCommentCount(pr)

		// Files (placeholder for now, will be enhanced)
		files := "-"

		// Use the formatter for consistent timestamps
		formatter := formatters.NewPRFormatter()
		timeSinceCreated := formatter.HumanizeTimeSince(pr.GetCreatedAt().Time)
		timeSinceUpdated := formatter.HumanizeTimeSince(pr.GetUpdatedAt().Time)

		row := table.Row{
			prName,
			author,           // Author only
			repoName,         // Repo only (guaranteed visible)
			statusCombined,   // Status + CI
			reviews,          // Review status
			comments,         // Comments only
			files,            // Files only (placeholder)
			timeSinceCreated, // When PR was created
			timeSinceUpdated, // When PR was updated
		}

		rows[i] = row
	}
	return rows
}

// createTableRowsWithEnhancement creates table rows using enhanced data when available
func createTableRowsWithEnhancement(prs []*gh.PullRequest, enhancedData map[int]types.EnhancedData) []table.Row {
	// Get column widths for dynamic truncation
	columns := createTableColumns()
	prColumnWidth := columns[0].Width

	rows := make([]table.Row, len(prs))
	for i, pr := range prs {
		// PR Name (smart formatting with ticket detection)
		prName := formatPRTitle(pr, prColumnWidth)

		// Author and Repo (now separate columns for better visibility)
		author := "Unknown"
		if pr.GetUser() != nil && pr.GetUser().GetLogin() != "" {
			author = pr.GetUser().GetLogin()
		}

		repoFullName := "Unknown"
		if pr.GetBase() != nil && pr.GetBase().GetRepo() != nil && pr.GetBase().GetRepo().GetFullName() != "" {
			repoFullName = pr.GetBase().GetRepo().GetFullName()
		}
		repoParts := strings.Split(repoFullName, "/")
		repoName := repoFullName
		if len(repoParts) == 2 {
			repoName = repoParts[1] // Just the repo name
		}

		// Status + CI (compact single-line format with enhanced CI data)
		mergeStatus := getPRStatusIndicatorEnhanced(pr, enhancedData)
		ciStatus := getCIStatusEnhanced(pr, enhancedData) // New function for CI status
		statusCombined := mergeStatus + " " + ciStatus

		// Review Status - enhanced with detailed review info
		reviews := getPRReviewIndicatorEnhanced(pr, enhancedData)

		// Comments - enhanced with detailed comment counts when available
		comments := getPRCommentCountEnhanced(pr, enhancedData)

		// Files - enhanced with file change info when available
		files := getPRFileChangesEnhanced(pr, enhancedData)

		// Use the formatter for consistent timestamps
		formatter := formatters.NewPRFormatter()
		timeSinceCreated := formatter.HumanizeTimeSince(pr.GetCreatedAt().Time)
		timeSinceUpdated := formatter.HumanizeTimeSince(pr.GetUpdatedAt().Time)

		row := table.Row{
			prName,
			author,           // Author only
			repoName,         // Repo only (guaranteed visible)
			statusCombined,   // Status + CI
			reviews,          // Review status (enhanced)
			comments,         // Comments only (enhanced)
			files,            // Files only (enhanced)
			timeSinceCreated, // When PR was created
			timeSinceUpdated, // When PR was updated
		}

		rows[i] = row
	}
	return rows
}

// getPRStatusIndicator returns merge readiness status
func getPRStatusIndicator(pr *gh.PullRequest) string {
	// Focus on MERGE READINESS with enhanced visual indicators

	if pr.GetDraft() {
		return "📝 Draft"
	}

	// Check merge conflicts and blocking issues
	mergeableState := pr.GetMergeableState()
	switch mergeableState {
	case "dirty":
		return "⚠️ Conflicts"
	case "blocked":
		return "🚫 Blocked"
	case "behind":
		return "📥 Behind"
	case "clean":
		return "✅ Ready"
	case "unstable":
		return "🔄 Checks"
	default:
		// For non-draft PRs without explicit state, assume ready
		return "✅ Ready"
	}
}

// getPRReviewIndicator returns HUMAN REVIEW STATUS with enhanced visual indicators
func getPRReviewIndicator(pr *gh.PullRequest) string {
	// Focus ONLY on human review process - completely separate from merge status

	// For drafts, they're work in progress so reviews don't make sense yet
	if pr.GetDraft() {
		return "🚧 WIP"
	}

	// Count requested reviewers to show review progress
	totalReviewers := len(pr.RequestedReviewers) + len(pr.RequestedTeams)
	if totalReviewers > 0 {
		return fmt.Sprintf("⏳ 0/%d", totalReviewers) // Shows "⏳ 0/3" etc
	}

	// Check if PR is old and might need attention (regardless of merge state)
	daysSinceUpdated := time.Since(pr.GetUpdatedAt().Time).Hours() / 24
	if daysSinceUpdated > 5 { // Older than 5 days
		return "🕰️ Stale"
	}

	// Check if it's recent and might not need formal review
	if daysSinceUpdated < 1 {
		return "🆕 Recent"
	}

	// Default for PRs with no explicit reviewers
	return "❓ None"
}

// getPRCommentCount returns the total number of comments (issue + review comments)
// Uses enhanced data when available, falls back to "?" when not
func getPRCommentCount(pr *gh.PullRequest) string {
	// Check if the API actually returned comment counts (non-zero means populated)
	issueComments := pr.GetComments()        // General discussion comments
	reviewComments := pr.GetReviewComments() // Code review comments on specific lines

	// If both are 0, it's likely the API didn't populate them (very common)
	// Only show count if we actually have non-zero counts
	totalComments := issueComments + reviewComments
	if totalComments > 0 {
		return fmt.Sprintf("%d", totalComments)
	}

	// Show "?" to indicate comment data not available from list API
	return "?"
}

// getPRCommentCountEnhanced returns comment count from enhanced data or falls back to basic logic
func getPRCommentCountEnhanced(pr *gh.PullRequest, enhancedData map[int]types.EnhancedData) string {
	prNumber := pr.GetNumber()

	// Try to get enhanced data first
	if enhanced, exists := enhancedData[prNumber]; exists {
		total := enhanced.Comments + enhanced.ReviewComments
		if total == 0 {
			return "-"
		}
		return fmt.Sprintf("%d", total)
	}

	// Fall back to original logic
	return getPRCommentCount(pr)
}

// getPRStatusIndicatorEnhanced returns enhanced merge readiness status
func getPRStatusIndicatorEnhanced(pr *gh.PullRequest, enhancedData map[int]types.EnhancedData) string {
	prNumber := pr.GetNumber()

	// Try to get enhanced data first
	if enhanced, exists := enhancedData[prNumber]; exists {
		// Use enhanced mergeable status if available
		switch enhanced.Mergeable {
		case "clean":
			if enhanced.ChecksStatus == "failure" {
				return "❌ Failed Checks"
			}
			return "✅ Ready"
		case "conflicts":
			return "⚠️ Conflicts"
		default:
			// Fall through to basic logic
		}
	}

	// Fall back to original logic
	return getPRStatusIndicator(pr)
}

// getPRReviewIndicatorEnhanced returns enhanced review status
func getPRReviewIndicatorEnhanced(pr *gh.PullRequest, enhancedData map[int]types.EnhancedData) string {
	prNumber := pr.GetNumber()

	// Try to get enhanced data first
	if enhanced, exists := enhancedData[prNumber]; exists {
		switch enhanced.ReviewStatus {
		case "approved":
			return "✅ Approved"
		case "changes_requested":
			return "🔄 Changes"
		case "pending":
			return "⏳ Pending"
		case "no_review":
			return "📝 No Review"
		default:
			return "❓ Unknown"
		}
	}

	// Fall back to original logic
	return getPRReviewIndicator(pr)
}

// getCIStatusEnhanced returns CI status from enhanced data
func getCIStatusEnhanced(pr *gh.PullRequest, enhancedData map[int]types.EnhancedData) string {
	prNumber := pr.GetNumber()

	// Try to get enhanced data first
	if enhanced, exists := enhancedData[prNumber]; exists {
		switch enhanced.ChecksStatus {
		case "success":
			return "✅ CI"
		case "failure":
			return "❌ CI"
		case "pending":
			return "🔄 CI"
		case "skipped":
			return "⚪ CI"
		default:
			return "❓ CI"
		}
	}

	// No enhanced data available
	return "CI:?"
}

// formatPRTitle intelligently formats PR title for better readability
func formatPRTitle(pr *gh.PullRequest, maxWidth int) string {
	title := pr.GetTitle()

	// Try to extract ticket prefix (Common Jira/ticket patterns: ABC-123, PROJ-456, etc.)
	if len(title) > 8 { // Only if title is reasonably long
		// Simple pattern matching without regex for now
		for i := 3; i < len(title) && i < 12; i++ { // Look in first 12 chars
			if title[i] == '-' && i >= 2 {
				// Found potential ticket pattern
				prefix := title[:i+1] // Include the dash
				restOfTitle := title[i+1:]

				// Clean up rest of title (remove leading colons, spaces)
				for len(restOfTitle) > 0 && (restOfTitle[0] == ':' || restOfTitle[0] == ' ') {
					restOfTitle = restOfTitle[1:]
				}

				// Create compact version: "A-123: Fix bug with..." instead of "ACC-1234: Fix bug with..."
				if len(prefix) >= 4 { // At least 2 chars + dash + 1 digit
					shortPrefix := string(prefix[0]) + prefix[len(prefix)-4:] // "A" + "-123"
					smartTitle := shortPrefix + ": " + restOfTitle

					// Check if smart version fits better
					if len(smartTitle) <= maxWidth {
						return smartTitle
					}
				}
				break
			}
		}
	}

	// Fallback: Use original title, truncated if necessary
	if len(title) > maxWidth {
		if maxWidth > 3 {
			title = title[:maxWidth-3] + "..."
		} else {
			title = title[:maxWidth]
		}
	}

	return title
}

// getPRActivityEnhanced returns combined activity info (comments + file changes) when enhanced data is available
func getPRActivityEnhanced(pr *gh.PullRequest, enhancedData map[int]types.EnhancedData) string {
	prNumber := pr.GetNumber()

	// Try to get enhanced data first
	if enhanced, exists := enhancedData[prNumber]; exists {
		// Show comments and file changes if available
		totalComments := enhanced.Comments + enhanced.ReviewComments

		if enhanced.ChangedFiles > 0 && totalComments > 0 {
			// Show both: "8c • 5F +120/-45" (compact format)
			return fmt.Sprintf("%dc • %dF +%d/-%d", totalComments, enhanced.ChangedFiles, enhanced.Additions, enhanced.Deletions)
		} else if enhanced.ChangedFiles > 0 {
			// Show just files: "5F +120/-45"
			return fmt.Sprintf("%dF +%d/-%d", enhanced.ChangedFiles, enhanced.Additions, enhanced.Deletions)
		} else if totalComments > 0 {
			// Show just comments: "8c"
			return fmt.Sprintf("%dc", totalComments)
		} else {
			// No activity
			return "-"
		}
	}

	// Fall back to comment count from basic data
	return getPRCommentCount(pr)
}

// getPRFileChangesEnhanced returns file change info when enhanced data is available
func getPRFileChangesEnhanced(pr *gh.PullRequest, enhancedData map[int]types.EnhancedData) string {
	prNumber := pr.GetNumber()

	// Try to get enhanced data first
	if enhanced, exists := enhancedData[prNumber]; exists {
		if enhanced.ChangedFiles > 0 {
			// Show file changes: "5 +120/-45"
			return fmt.Sprintf("%d +%d/-%d", enhanced.ChangedFiles, enhanced.Additions, enhanced.Deletions)
		} else {
			// No file changes
			return "-"
		}
	}

	// Fall back - no enhanced data available
	return "?"
}

// getPRLabelsDisplay returns a smart display of important labels
func getPRLabelsDisplay(pr *gh.PullRequest) string {
	if len(pr.Labels) == 0 {
		return ""
	}

	// Show up to 3-4 labels, prioritize certain important ones
	importantLabels := []string{}
	otherLabels := []string{}

	for _, label := range pr.Labels {
		labelName := label.GetName()
		// Prioritize certain label patterns
		if strings.Contains(strings.ToLower(labelName), "bug") ||
			strings.Contains(strings.ToLower(labelName), "urgent") ||
			strings.Contains(strings.ToLower(labelName), "breaking") ||
			strings.Contains(strings.ToLower(labelName), "security") ||
			strings.Contains(strings.ToLower(labelName), "critical") ||
			strings.Contains(strings.ToLower(labelName), "hotfix") {
			importantLabels = append(importantLabels, labelName)
		} else {
			otherLabels = append(otherLabels, labelName)
		}
	}

	// Combine important first, then others
	allLabels := append(importantLabels, otherLabels...)

	if len(allLabels) == 0 {
		return ""
	}

	// Build result string, fitting as many labels as possible
	result := ""
	maxWidth := 35 // Increased from 15 to show more labels

	for i, label := range allLabels {
		if i == 0 {
			result = label
		} else {
			candidate := result + ", " + label
			if len(candidate) <= maxWidth {
				result = candidate
			} else {
				// Add count of remaining labels if there are more
				remaining := len(allLabels) - i
				if remaining > 0 {
					countStr := fmt.Sprintf(" +%d", remaining)
					if len(result)+len(countStr) <= maxWidth {
						result += countStr
					}
				}
				break
			}
		}
	}

	return result
}

func loadingView() string {
	return loadingViewWithSpinner(0) // Default spinner index for backward compatibility
}

func loadingViewWithSpinner(spinnerIndex int) string {
	title := titleStyle.Render("🧭 PR Compass - Pull Request Monitor")
	
	// Enhanced loading animation with multiple symbols
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	currentSpinner := spinner[spinnerIndex%len(spinner)]
	
	// Create a more informative loading message
	loadingMsg := fmt.Sprintf("%s Fetching pull requests from GitHub...", currentSpinner)
	message := loadingStyle.Render(loadingMsg)
	
	// Compact help text 
	helpText := "🧭 Press q to quit • Fetching data..."
	help := helpStyle.Render(helpText)

	return "\n" + title + "\n\n" + message + "\n\n" + help + "\n"
}

func errorView(err error) string {
	title := titleStyle.Render("🧭 PR Compass - Pull Request Monitor")

	var message string
	var suggestions string

	// Check if this is a domain-specific error
	if prErr, isPRError := errors.IsPRCompassError(err); isPRError {
		// Enhanced error display with better formatting
		errorMsg := fmt.Sprintf("🚫 %s", prErr.UserFriendlyError())
		message = errorStyle.Render(errorMsg)
		
		// Add helpful suggestions based on error type
		suggestions = mutedStyle.Render("💡 Try: Check your config file or run 'gh auth login' to authenticate")
	} else {
		// Enhanced generic error display
		errorMsg := fmt.Sprintf("🚫 Unexpected error: %v", err)
		message = errorStyle.Render(errorMsg)
		
		suggestions = mutedStyle.Render("💡 This might be a network issue or GitHub API problem")
	}

	// Compact help text with compass emoji
	helpText := "🧭 Press q to quit • r to retry • Check connection"
	help := helpStyle.Render(helpText)

	return "\n" + title + "\n\n" + message + "\n\n" + suggestions + "\n\n" + help + "\n"
}

// sortPRsByNewest sorts PRs by most recently updated first (not created)
func sortPRsByNewest(prs []*gh.PullRequest) []*gh.PullRequest {
	// Make a copy to avoid modifying the original slice
	sorted := make([]*gh.PullRequest, len(prs))
	copy(sorted, prs)

	// Sort by updated time (most recent first) for better relevance
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetUpdatedAt().Time.After(sorted[j].GetUpdatedAt().Time)
	})

	return sorted
}


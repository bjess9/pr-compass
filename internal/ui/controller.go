package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/bjess9/pr-compass/internal/github"
	gh "github.com/google/go-github/v55/github"
)

// UIController handles business logic for UI operations
// This is separated from the Bubble Tea model for better testability
type UIController struct {
	token string
}

// NewUIController creates a new UI controller
func NewUIController(token string) *UIController {
	return &UIController{
		token: token,
	}
}

// FilterResult represents the result of filtering operations
type FilterResult struct {
	FilteredPRs []*gh.PullRequest
	StatusMsg   string
}

// ApplyFilter applies filtering logic to PRs
func (c *UIController) ApplyFilter(prs []*gh.PullRequest, mode, value string) FilterResult {
	if mode == "" || value == "" {
		return FilterResult{
			FilteredPRs: prs,
			StatusMsg:   "",
		}
	}

	filtered := c.filterPRs(prs, mode, value)
	statusMsg := c.formatFilterStatus(mode, value, len(filtered))

	return FilterResult{
		FilteredPRs: filtered,
		StatusMsg:   statusMsg,
	}
}

// filterPRs performs the actual filtering logic
func (c *UIController) filterPRs(prs []*gh.PullRequest, mode, value string) []*gh.PullRequest {
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

// formatFilterStatus creates a status message for filtering
func (c *UIController) formatFilterStatus(mode, value string, resultCount int) string {
	return fmt.Sprintf("Filter applied: %s=%s (%d results)", mode, value, resultCount)
}

// FilterDraftPRs returns only draft PRs
func (c *UIController) FilterDraftPRs(prs []*gh.PullRequest) []*gh.PullRequest {
	var drafts []*gh.PullRequest
	for _, pr := range prs {
		if pr.GetDraft() {
			drafts = append(drafts, pr)
		}
	}
	return drafts
}

// FetchPRsForTab fetches PRs for a specific tab configuration
func (c *UIController) FetchPRsForTab(ctx context.Context, tabConfig *TabConfig) ([]*gh.PullRequest, error) {
	config := tabConfig.ConvertToConfig()
	return github.FetchPRsFromConfig(ctx, config, c.token)
}

// CalculateTableHeight calculates the appropriate table height based on terminal dimensions
func (c *UIController) CalculateTableHeight(terminalHeight int) int {
	// Smart viewport management: ensure critical UI elements always stay visible
	
	// Critical elements that must always be visible:
	tabBarLines := 3      // Tab indicators and borders
	statusLineLines := 1  // Status message area
	tableHeaderLines := 1 // Column headers
	marginsLines := 3     // App borders and spacing
	safetyBuffer := 8     // Buffer to ensure tabs don't get pushed off screen
	
	criticalLines := tabBarLines + statusLineLines + tableHeaderLines + marginsLines + safetyBuffer
	
	availableHeight := terminalHeight - criticalLines
	
	// Minimum table height to show some data
	minHeight := 3
	if availableHeight < minHeight {
		return minHeight
	}
	
	// Maximum table height to prevent excessive scrolling and ensure good UX
	maxHeight := 20
	if availableHeight > maxHeight {
		return maxHeight
	}
	
	return availableHeight
}

// ValidateTabSwitch validates if a tab switch operation is valid
func (c *UIController) ValidateTabSwitch(currentIdx, targetIdx, totalTabs int) bool {
	return targetIdx >= 0 && targetIdx < totalTabs && targetIdx != currentIdx
}

// GetSpinnerFrame returns the current spinner frame for loading animations
func (c *UIController) GetSpinnerFrame(index int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return spinners[index%len(spinners)]
}
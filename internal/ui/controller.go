package ui

import (
	"context"
	"fmt"

	"github.com/bjess9/pr-compass/internal/ui/services"
	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

// UIController handles business logic for UI operations
// This is separated from the Bubble Tea model for better testability
type UIController struct {
	services *services.Registry
}

// NewUIController creates a new UI controller
func NewUIController(services *services.Registry) *UIController {
	return &UIController{
		services: services,
	}
}

// FilterResult represents the result of filtering operations
type FilterResult struct {
	FilteredPRs []*gh.PullRequest
	StatusMsg   string
}

// ApplyFilter applies filtering logic to PRs
func (c *UIController) ApplyFilter(prs []*types.PRData, filter types.FilterOptions) FilterResult {
	if filter.Mode == "" {
		converted := make([]*gh.PullRequest, len(prs))
		for i, pr := range prs {
			converted[i] = pr.PullRequest
		}
		return FilterResult{
			FilteredPRs: converted,
			StatusMsg:   "",
		}
	}

	filteredPRs := c.services.Filter.FilterPRs(prs, filter)
	statusMsg := fmt.Sprintf("Filter applied: %s=%s (%d results)", filter.Mode, filter.Value, len(filteredPRs))

	converted := make([]*gh.PullRequest, len(filteredPRs))
	for i, pr := range filteredPRs {
		converted[i] = pr.PullRequest
	}

	return FilterResult{
		FilteredPRs: converted,
		StatusMsg:   statusMsg,
	}
}

// FilterDraftPRs returns only draft PRs
func (c *UIController) FilterDraftPRs(prs []*types.PRData) []*types.PRData {
	filter := types.FilterOptions{
		Mode:  "draft",
		Value: "true",
	}
	return c.services.Filter.FilterPRs(prs, filter)
}

// FetchPRsForTab fetches PRs for a specific tab configuration
func (c *UIController) FetchPRsForTab(ctx context.Context, tabConfig *TabConfig) ([]*types.PRData, error) {
	config := tabConfig.ConvertToConfig()
	return c.services.PR.FetchPRs(ctx, config)
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

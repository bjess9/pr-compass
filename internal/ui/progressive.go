package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	gh "github.com/google/go-github/v55/github"
)

// createProgressiveTableRows creates table rows with progressive loading indicators
func createProgressiveTableRows(prs []*gh.PullRequest, enhancedData map[int]enhancedPRData, enhancementQueue map[int]bool) []table.Row {
	// Get column widths for dynamic truncation
	columns := createTableColumns()
	prColumnWidth := columns[0].Width

	rows := make([]table.Row, len(prs))
	for i, pr := range prs {
		prNumber := pr.GetNumber()
		
		// PR Name (smart formatting with ticket detection)
		prName := formatPRTitle(pr, prColumnWidth)

		// Author and Repo (basic info - always available)
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

		// Progressive enhancement: check if enhanced data is available
		_, hasEnhanced := enhancedData[prNumber]
		_, isEnhancing := enhancementQueue[prNumber]

		var statusCombined, reviews, comments, files string

		if hasEnhanced {
			// Show enhanced data
			mergeStatus := getPRStatusIndicatorEnhanced(pr, enhancedData)
			ciStatus := getCIStatusEnhanced(pr, enhancedData)
			statusCombined = mergeStatus + " " + ciStatus
			reviews = getPRReviewIndicatorEnhanced(pr, enhancedData)
			comments = getPRCommentCountEnhanced(pr, enhancedData)
			files = getPRFileChangesEnhanced(pr, enhancedData)
		} else if isEnhancing {
			// Show enhanced loading indicators
			statusCombined = getPRStatusIndicator(pr) + " 🔄"
			reviews = "🔍 Loading..."
			comments = "⏳"
			files = "📊 Loading..."
		} else {
			// Show basic data with indicators that more is available
			statusCombined = getPRStatusIndicator(pr) + " 💫"
			reviews = getPRReviewIndicator(pr)
			comments = getPRCommentCount(pr) + " 💬"
			files = "📁 More..."
		}

		// Time info is always available
		timeSinceCreated := humanizeTimeSince(pr.GetCreatedAt().Time)
		timeSinceUpdated := humanizeTimeSince(pr.GetUpdatedAt().Time)

		row := table.Row{
			prName,
			author,
			repoName,
			statusCombined,
			reviews,
			comments,
			files,
			timeSinceCreated,
			timeSinceUpdated,
		}

		rows[i] = row
	}
	return rows
}

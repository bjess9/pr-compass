package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	gh "github.com/google/go-github/v55/github"

	"github.com/charmbracelet/bubbles/table"
	"golang.org/x/term"
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

func createTableColumns() []table.Column {
	totalWidth := getTerminalWidth() - 6 // Account for borders and padding
	
	// Allocate width more intelligently based on content
	prNameWidth := totalWidth * 25 / 100     // 25% - PR title and number
	authorWidth := totalWidth * 12 / 100      // 12% - Author name  
	repoWidth := totalWidth * 18 / 100        // 18% - Repository name
	statusWidth := totalWidth * 15 / 100      // 15% - Draft/Ready/Review status
	reviewsWidth := totalWidth * 15 / 100     // 15% - Review status (✅❌🔄)
	timeWidth := totalWidth * 8 / 100         // 8%  - Time open
	labelsWidth := totalWidth * 7 / 100       // 7%  - Important labels
	
	return []table.Column{
		{Title: "PR", Width: prNameWidth},
		{Title: "Author", Width: authorWidth},
		{Title: "Repository", Width: repoWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Reviews", Width: reviewsWidth},
		{Title: "Age", Width: timeWidth},
		{Title: "Labels", Width: labelsWidth},
	}
}

func createTableRows(prs []*gh.PullRequest) []table.Row {
	rows := make([]table.Row, len(prs))
	for i, pr := range prs {
		// PR Name (number + title, truncated if needed)
		prNumber := fmt.Sprintf("#%d", pr.GetNumber())
		title := pr.GetTitle()
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		prName := prNumber + " " + title
		
		// Author
		author := pr.GetUser().GetLogin()
		
		// Repository (show just repo name, not full owner/repo)
		repoFullName := pr.GetBase().GetRepo().GetFullName()
		repoParts := strings.Split(repoFullName, "/")
		repoName := repoFullName
		if len(repoParts) == 2 {
			repoName = repoParts[1] // Just the repo name
		}
		
		// Status (Draft, Ready, Mergeable, etc.)
		status := getPRStatusIndicator(pr)
		
		// Review Status (visual indicators)
		reviews := getPRReviewIndicator(pr)
		
		// Time since opened
		timeSinceOpened := humanizeTimeSince(pr.GetCreatedAt().Time)
		
		// Labels (show important ones, truncated)
		labels := getPRLabelsDisplay(pr)

		row := table.Row{
			prName,
			author,
			repoName,
			status,
			reviews,
			timeSinceOpened,
			labels,
		}

		rows[i] = row
	}
	return rows
}

// getPRStatusIndicator returns a status indicator for the PR
func getPRStatusIndicator(pr *gh.PullRequest) string {
	if pr.GetDraft() {
		return "🚧 Draft"
	}
	
	// Note: GetMergeable() might not always be populated, depends on GitHub API response
	// We'll use available information and default to "Ready" for non-draft PRs
	if pr.GetMergeable() {
		return "✅ Ready"
	} else if pr.GetMergeableState() == "dirty" {
		return "❌ Conflicts"
	}
	return "✅ Ready"
}

// getPRReviewIndicator returns review status indicators
func getPRReviewIndicator(pr *gh.PullRequest) string {
	// Note: This is basic - we'd need to make additional API calls to get full review status
	// For now, we'll use available information from the PR object
	
	// Check if there are requested reviewers (users or teams)
	if pr.RequestedReviewers != nil && len(pr.RequestedReviewers) > 0 {
		return "🔄 Pending"
	}
	
	if pr.RequestedTeams != nil && len(pr.RequestedTeams) > 0 {
		return "🔄 Pending"
	}
	
	// Could enhance this by fetching reviews with additional API calls
	return "❓ Unknown"
}

// getPRLabelsDisplay returns a truncated display of important labels
func getPRLabelsDisplay(pr *gh.PullRequest) string {
	if pr.Labels == nil || len(pr.Labels) == 0 {
		return ""
	}
	
	// Show up to 2 labels, prioritize certain important ones
	importantLabels := []string{}
	otherLabels := []string{}
	
	for _, label := range pr.Labels {
		labelName := label.GetName()
		// Prioritize certain label patterns
		if strings.Contains(strings.ToLower(labelName), "bug") ||
		   strings.Contains(strings.ToLower(labelName), "urgent") ||
		   strings.Contains(strings.ToLower(labelName), "breaking") ||
		   strings.Contains(strings.ToLower(labelName), "security") {
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
	
	if len(allLabels) == 1 {
		return allLabels[0]
	}
	
	// Show first two, or indicate there are more
	result := allLabels[0]
	if len(allLabels) > 2 {
		result += " +" + fmt.Sprintf("%d", len(allLabels)-1)
	} else {
		result += " " + allLabels[1]
	}
	
	// Truncate if too long
	if len(result) > 15 {
		result = result[:12] + "..."
	}
	
	return result
}

func humanizeTimeSince(t time.Time) string {
	duration := time.Since(t)
	if duration.Hours() < 24 {
		if duration.Hours() >= 1 {
			return fmt.Sprintf("%.0fh", duration.Hours())
		}
		return fmt.Sprintf("%.0fm", duration.Minutes())
	}
	return fmt.Sprintf("%.0fd", duration.Hours()/24)
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}
	return width
}

func loadingView() string {
	return "Loading PRs...\nPress 'q' to quit."
}

func errorView(err error) string {
	return fmt.Sprintf("Error: %v\nPress 'q' to quit.", err)
}

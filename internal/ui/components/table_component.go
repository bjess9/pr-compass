package components

import (
	"github.com/bjess9/pr-compass/internal/ui/formatters"
	"github.com/bjess9/pr-compass/internal/ui/types"
	"github.com/charmbracelet/bubbles/table"
)

// TableComponent handles table creation and management
type TableComponent struct {
	columns   []table.Column
	formatter *formatters.PRFormatter
}

// NewTableComponent creates a new table component
func NewTableComponent() *TableComponent {
	formatter := formatters.NewPRFormatter()
	return &TableComponent{
		columns:   formatter.CreateTableColumns(),
		formatter: formatter,
	}
}

// CreateTable creates a new table with proper configuration
func (tc *TableComponent) CreateTable() table.Model {
	t := table.New(
		table.WithColumns(tc.columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	t.Focus()
	return t
}

// CreateRows creates table rows from PR data
func (tc *TableComponent) CreateRows(prs []*types.PRData, enhancementQueue map[int]bool) []table.Row {
	rows := make([]table.Row, len(prs))

	for i, pr := range prs {
		displayInfo := tc.createPRDisplayInfo(pr, enhancementQueue)
		rows[i] = table.Row{
			displayInfo.Title,
			displayInfo.Author,
			displayInfo.Repo,
			displayInfo.Status,
			displayInfo.Reviews,
			displayInfo.Comments,
			displayInfo.Files,
			displayInfo.CreatedTime,
			displayInfo.UpdatedTime,
		}
	}

	return rows
}

// createPRDisplayInfo creates display information for a PR
func (tc *TableComponent) createPRDisplayInfo(pr *types.PRData, enhancementQueue map[int]bool) *types.PRDisplayInfo {
	prNumber := pr.GetNumber()
	_, isEnhancing := enhancementQueue[prNumber]
	isEnhanced := pr.Enhanced != nil

	return &types.PRDisplayInfo{
		Title:       tc.formatTitle(pr),
		Author:      tc.formatAuthor(pr),
		Repo:        tc.formatRepo(pr),
		Status:      tc.formatStatus(pr, isEnhancing, isEnhanced),
		Reviews:     tc.formatReviews(pr, isEnhancing, isEnhanced),
		Comments:    tc.formatComments(pr, isEnhancing, isEnhanced),
		Files:       tc.formatFiles(pr, isEnhancing, isEnhanced),
		CreatedTime: tc.formatCreatedTime(pr),
		UpdatedTime: tc.formatUpdatedTime(pr),
		IsEnhancing: isEnhancing,
		IsEnhanced:  isEnhanced,
	}
}

// Helper methods for formatting different fields
func (tc *TableComponent) formatTitle(pr *types.PRData) string {
	title := pr.GetTitle()
	maxWidth := tc.columns[0].Width

	// Smart title formatting logic
	if len(title) > maxWidth {
		if maxWidth > 3 {
			title = title[:maxWidth-3] + "..."
		} else {
			title = title[:maxWidth]
		}
	}

	return title
}

func (tc *TableComponent) formatAuthor(pr *types.PRData) string {
	if pr.GetUser() != nil && pr.GetUser().GetLogin() != "" {
		return pr.GetUser().GetLogin()
	}
	return "Unknown"
}

func (tc *TableComponent) formatRepo(pr *types.PRData) string {
	if pr.GetBase() != nil && pr.GetBase().GetRepo() != nil {
		repoName := pr.GetBase().GetRepo().GetName()
		if repoName != "" {
			return repoName
		}
	}
	return "Unknown"
}

func (tc *TableComponent) formatStatus(pr *types.PRData, isEnhancing, isEnhanced bool) string {
	baseStatus := tc.formatter.GetBasicStatus(pr.PullRequest)

	if isEnhanced && pr.Enhanced != nil {
		return tc.formatter.GetEnhancedStatus(pr.Enhanced, baseStatus)
	}

	if isEnhancing {
		return baseStatus + " ⏳"
	}

	return baseStatus + " ••"
}

func (tc *TableComponent) formatReviews(pr *types.PRData, isEnhancing, isEnhanced bool) string {
	if isEnhanced && pr.Enhanced != nil {
		return tc.formatter.GetEnhancedReviewStatus(pr.Enhanced)
	}

	if isEnhancing {
		return "Loading..."
	}

	return tc.formatter.GetBasicReviewStatus(pr.PullRequest)
}

func (tc *TableComponent) formatComments(pr *types.PRData, isEnhancing, isEnhanced bool) string {
	if isEnhanced && pr.Enhanced != nil {
		total := pr.Enhanced.Comments + pr.Enhanced.ReviewComments
		if total == 0 {
			return "-"
		}
		return tc.formatter.FormatNumber(total)
	}

	if isEnhancing {
		return "⏳"
	}

	return "?"
}

func (tc *TableComponent) formatFiles(pr *types.PRData, isEnhancing, isEnhanced bool) string {
	if isEnhanced && pr.Enhanced != nil {
		if pr.Enhanced.ChangedFiles == 0 {
			return "-"
		}
		changesStr := tc.formatter.FormatChanges(pr.Enhanced.Additions, pr.Enhanced.Deletions)
		if changesStr != "" {
			return tc.formatter.FormatNumber(pr.Enhanced.ChangedFiles) + " " + changesStr
		}
		return tc.formatter.FormatNumber(pr.Enhanced.ChangedFiles)
	}

	if isEnhancing {
		return "⏳"
	}

	return "••"
}

func (tc *TableComponent) formatCreatedTime(pr *types.PRData) string {
	return tc.formatter.HumanizeTimeSince(pr.GetCreatedAt().Time)
}

func (tc *TableComponent) formatUpdatedTime(pr *types.PRData) string {
	return tc.formatter.HumanizeTimeSince(pr.GetUpdatedAt().Time)
}

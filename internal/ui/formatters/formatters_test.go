package formatters

import (
	"testing"
	"time"

	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

func TestNewPRFormatter(t *testing.T) {
	formatter := NewPRFormatter()

	if formatter == nil {
		t.Fatal("NewPRFormatter returned nil")
	}
}

func TestPRFormatter_FormatNumber(t *testing.T) {
	formatter := NewPRFormatter()

	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"Zero", 0, "-"},
		{"Positive", 5, "5"},
		{"Large number", 999, "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPRFormatter_FormatChanges(t *testing.T) {
	formatter := NewPRFormatter()

	tests := []struct {
		name      string
		additions int
		deletions int
		expected  string
	}{
		{"No changes", 0, 0, ""},
		{"Only additions", 5, 0, "+5/-0"},
		{"Only deletions", 0, 3, "+0/-3"},
		{"Both", 10, 5, "+10/-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatChanges(tt.additions, tt.deletions)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPRFormatter_HumanizeTimeSince(t *testing.T) {
	formatter := NewPRFormatter()
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		contains string
	}{
		{"Just now", now.Add(-30 * time.Second), "< 1m"},
		{"Minutes ago", now.Add(-5 * time.Minute), "5m"},
		{"Hours ago", now.Add(-3 * time.Hour), "3h"},
		{"Days ago", now.Add(-2 * 24 * time.Hour), "2d"},
		{"Weeks ago", now.Add(-2 * 7 * 24 * time.Hour), "2w"},
		{"Months ago", now.Add(-2 * 30 * 24 * time.Hour), "2mo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.HumanizeTimeSince(tt.time)
			if result != tt.contains {
				t.Errorf("Expected '%s', got '%s'", tt.contains, result)
			}
		})
	}
}

func TestPRFormatter_GetBasicStatus(t *testing.T) {
	formatter := NewPRFormatter()

	tests := []struct {
		name     string
		pr       *gh.PullRequest
		expected string
	}{
		{
			name: "Draft PR",
			pr: &gh.PullRequest{
				Draft: gh.Bool(true),
			},
			expected: "Draft",
		},
		{
			name: "Clean PR",
			pr: &gh.PullRequest{
				Draft:          gh.Bool(false),
				MergeableState: gh.String("clean"),
			},
			expected: "Ready",
		},
		{
			name: "Dirty PR",
			pr: &gh.PullRequest{
				Draft:          gh.Bool(false),
				MergeableState: gh.String("dirty"),
			},
			expected: "Conflicts",
		},
		{
			name: "Blocked PR",
			pr: &gh.PullRequest{
				Draft:          gh.Bool(false),
				MergeableState: gh.String("blocked"),
			},
			expected: "Blocked",
		},
		{
			name: "Behind PR",
			pr: &gh.PullRequest{
				Draft:          gh.Bool(false),
				MergeableState: gh.String("behind"),
			},
			expected: "Behind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.GetBasicStatus(tt.pr)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPRFormatter_GetEnhancedStatus(t *testing.T) {
	formatter := NewPRFormatter()

	tests := []struct {
		name       string
		enhanced   *types.EnhancedData
		baseStatus string
		expected   string
	}{
		{
			name: "Clean with passing checks",
			enhanced: &types.EnhancedData{
				Mergeable:    "clean",
				ChecksStatus: "success",
			},
			baseStatus: "Ready",
			expected:   "[✓] Ready",
		},
		{
			name: "Clean with failing checks",
			enhanced: &types.EnhancedData{
				Mergeable:    "clean",
				ChecksStatus: "failure",
			},
			baseStatus: "Ready",
			expected:   "[!] Checks",
		},
		{
			name: "Conflicts",
			enhanced: &types.EnhancedData{
				Mergeable: "conflicts",
			},
			baseStatus: "Ready",
			expected:   "[X] Conflicts",
		},
		{
			name: "Unknown state",
			enhanced: &types.EnhancedData{
				Mergeable: "unknown",
			},
			baseStatus: "Ready",
			expected:   "Ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.GetEnhancedStatus(tt.enhanced, tt.baseStatus)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPRFormatter_GetBasicReviewStatus(t *testing.T) {
	formatter := NewPRFormatter()

	now := time.Now()
	tests := []struct {
		name     string
		pr       *gh.PullRequest
		expected string
	}{
		{
			name: "Draft PR",
			pr: &gh.PullRequest{
				Draft: gh.Bool(true),
			},
			expected: "WIP",
		},
		{
			name: "PR with reviewers",
			pr: &gh.PullRequest{
				Draft: gh.Bool(false),
				RequestedReviewers: []*gh.User{
					{Login: gh.String("reviewer1")},
					{Login: gh.String("reviewer2")},
				},
			},
			expected: "0/2",
		},
		{
			name: "Recent PR",
			pr: &gh.PullRequest{
				Draft:     gh.Bool(false),
				UpdatedAt: &gh.Timestamp{Time: now.Add(-30 * time.Minute)},
			},
			expected: "Recent",
		},
		{
			name: "Stale PR",
			pr: &gh.PullRequest{
				Draft:     gh.Bool(false),
				UpdatedAt: &gh.Timestamp{Time: now.Add(-6 * 24 * time.Hour)},
			},
			expected: "Stale",
		},
		{
			name: "No review PR",
			pr: &gh.PullRequest{
				Draft:     gh.Bool(false),
				UpdatedAt: &gh.Timestamp{Time: now.Add(-2 * 24 * time.Hour)},
			},
			expected: "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.GetBasicReviewStatus(tt.pr)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPRFormatter_GetEnhancedReviewStatus(t *testing.T) {
	formatter := NewPRFormatter()

	tests := []struct {
		name     string
		enhanced *types.EnhancedData
		expected string
	}{
		{
			name: "Approved",
			enhanced: &types.EnhancedData{
				ReviewStatus: "approved",
			},
			expected: "Approved",
		},
		{
			name: "Changes requested",
			enhanced: &types.EnhancedData{
				ReviewStatus: "changes_requested",
			},
			expected: "Changes",
		},
		{
			name: "Pending",
			enhanced: &types.EnhancedData{
				ReviewStatus: "pending",
			},
			expected: "Pending",
		},
		{
			name: "No review",
			enhanced: &types.EnhancedData{
				ReviewStatus: "no_review",
			},
			expected: "No Review",
		},
		{
			name: "Unknown status",
			enhanced: &types.EnhancedData{
				ReviewStatus: "unknown_status",
			},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.GetEnhancedReviewStatus(tt.enhanced)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPRFormatter_CreateTableColumns(t *testing.T) {
	formatter := NewPRFormatter()

	columns := formatter.CreateTableColumns()

	// Should have 9 columns
	if len(columns) != 9 {
		t.Fatalf("Expected 9 columns, got %d", len(columns))
	}

	// Check column titles
	expectedTitles := []string{
		"📋 Pull Request",
		"👤 Author",
		"📦 Repo",
		"⚡ Status/CI",
		"👀 Review",
		"💬 Comments",
		"📁 Files",
		"📅 Created",
		"🕐 Updated",
	}

	for i, expected := range expectedTitles {
		if columns[i].Title != expected {
			t.Errorf("Column %d: expected title '%s', got '%s'", i, expected, columns[i].Title)
		}
		if columns[i].Width <= 0 {
			t.Errorf("Column %d: width should be positive, got %d", i, columns[i].Width)
		}
	}
}

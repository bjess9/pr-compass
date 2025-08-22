package ui

import (
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
)

// TestCreateTableColumns verifies column width allocation
func TestCreateTableColumns(t *testing.T) {
	columns := createTableColumns()
	
	// Should have 8 columns (added Comments column)
	expectedColumns := 8
	if len(columns) != expectedColumns {
		t.Errorf("Expected %d columns, got %d", expectedColumns, len(columns))
	}
	
	// Verify column titles (now with 8 columns: split Activity, added Created, removed Labels)
	expectedTitles := []string{"Pull Request", "Author/Repo", "Status/CI", "Review", "Comments", "Files", "Created", "Updated"}
	for i, col := range columns {
		if col.Title != expectedTitles[i] {
			t.Errorf("Column %d: expected title %q, got %q", i, expectedTitles[i], col.Title)
		}
	}
	
	// Verify all columns have positive width
	for i, col := range columns {
		if col.Width <= 0 {
			t.Errorf("Column %d (%s) should have positive width, got %d", i, col.Title, col.Width)
		}
	}
	
	// PR column should be widest (for readability)
	prColumn := columns[0]
	if prColumn.Title != "Pull Request" {
		t.Error("First column should be Pull Request")
	}
	
	// Updated column should be last (at index 7 for 8 columns)
	updatedColumn := columns[7]  // Updated now at index 7 (8 columns total)
	if updatedColumn.Title != "Updated" {
		t.Error("Last column should be Updated")
	}
	
	// Created column should be at index 6
	createdColumn := columns[6]  // Created at index 6
	if createdColumn.Title != "Created" {
		t.Error("Created column should be at index 6")
	}
	
	// Comments column should be at index 4
	commentsColumn := columns[4] // Comments now at index 4
	if commentsColumn.Title != "Comments" {
		t.Error("Comments column should be at index 4")
	}
	
	// Files column should be at index 5
	filesColumn := columns[5] // Files now at index 5
	if filesColumn.Title != "Files" {
		t.Error("Files column should be at index 5")
	}

}

// TestCreateTableRows verifies PR data is formatted correctly
func TestCreateTableRows(t *testing.T) {
	// Create test PR data
	testPRs := []*github.PullRequest{
		{
			Number: github.Int(123),
			Title:  github.String("feat: add awesome new feature that has a very long title"),
			User:   &github.User{Login: github.String("developer")},
			Base: &github.PullRequestBranch{
				Repo: &github.Repository{
					FullName: github.String("company/awesome-repo"),
				},
			},
			Draft:          github.Bool(false),
			Mergeable:      github.Bool(true),
			CreatedAt:      &github.Timestamp{Time: time.Now().Add(-2 * time.Hour)},
			UpdatedAt:      &github.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
			Comments:       github.Int(5),       // General discussion comments
			ReviewComments: github.Int(12),      // Code review comments
			Labels: []*github.Label{
				{Name: github.String("feature")},
				{Name: github.String("backend")},
				{Name: github.String("urgent")},
			},
		},
		{
			Number:         github.Int(456),
			Title:          github.String("fix: short title"),
			User:           &github.User{Login: github.String("another-dev")},
			Base:           &github.PullRequestBranch{Repo: &github.Repository{FullName: github.String("company/other-repo")}},
			Draft:          github.Bool(true),
			Mergeable:      github.Bool(false),
			CreatedAt:      &github.Timestamp{Time: time.Now().Add(-24 * time.Hour)},
			UpdatedAt:      &github.Timestamp{Time: time.Now().Add(-12 * time.Hour)},
			Comments:       github.Int(0),        // No general comments
			ReviewComments: github.Int(3),        // Few review comments
			Labels:         []*github.Label{},    // No labels
		},
	}
	
	rows := createTableRows(testPRs)
	
	// Should have same number of rows as PRs
	if len(rows) != len(testPRs) {
		t.Errorf("Expected %d rows, got %d", len(testPRs), len(rows))
	}
	
	// Each row should have 8 columns (Comments, Files, Created, Updated separate)
	for i, row := range rows {
		if len(row) != 8 {
			t.Errorf("Row %d should have 8 columns, got %d", i, len(row))
		}
	}
	
	// First row should contain title (without PR number)
	firstRow := rows[0]
	if contains(firstRow[0], "#123") {
		t.Error("First column should NOT contain PR number (we removed PR numbers)")
	}
	
	if !contains(firstRow[0], "feat: add awesome") {
		t.Error("First column should contain truncated title")
	}
	
	// Author+Repo column should be formatted intelligently to show repo name
	authorRepoCol := firstRow[1]
	if !contains(authorRepoCol, "/") {
		t.Error("Author+Repo column should contain '/' separator")
	}
	if !contains(authorRepoCol, "awesome-repo") && !contains(authorRepoCol, "awesome-rep") {
		t.Errorf("Author+Repo column should contain repo name (possibly truncated), got %q", authorRepoCol)
	}
	
	// Status+CI column should have proper merge status indicators (compact format)
	if !contains(firstRow[2], "[") && !contains(firstRow[2], "]") {
		t.Error("Status+CI column should have bracket indicators")
	}
	
	// Comments column should show total comment count (5 + 12 = 17)
	commentsCol := firstRow[4]  // Comments column at index 4
	if commentsCol != "17" {
		t.Errorf("Comments column should show '17' (5+12 comments), got %q", commentsCol)
	}
	
	// Files column should show "-" (placeholder since no enhanced data)
	filesCol := firstRow[5]  // Files column at index 5
	if filesCol != "-" {
		t.Errorf("Files column should show '-' for basic data, got %q", filesCol)
	}
	
	// Created column should have time (index 6)
	createdCol := firstRow[6]  // Created column at index 6
	if createdCol == "" {
		t.Error("Created column should not be empty")
	}
	
	// Updated column should have time (index 7)
	updatedCol := firstRow[7]  // Updated column at index 7
	if updatedCol == "" {
		t.Error("Updated column should not be empty")
	}
}

// TestGetPRLabelsDisplay verifies label formatting works correctly
func TestGetPRLabelsDisplay(t *testing.T) {
	tests := []struct {
		name     string
		labels   []*github.Label
		expected string // partial match - we just check it's not empty
	}{
		{
			name:     "no labels",
			labels:   []*github.Label{},
			expected: "", // empty
		},
		{
			name: "single label",
			labels: []*github.Label{
				{Name: github.String("bug")},
			},
			expected: "bug",
		},
		{
			name: "multiple labels",
			labels: []*github.Label{
				{Name: github.String("feature")},
				{Name: github.String("backend")},
				{Name: github.String("urgent")},
			},
			expected: "feature", // should start with first label
		},
		{
			name: "prioritizes important labels",
			labels: []*github.Label{
				{Name: github.String("documentation")},
				{Name: github.String("urgent")}, // This should be prioritized
			},
			expected: "urgent", // urgent should come first
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &github.PullRequest{Labels: tt.labels}
			result := getPRLabelsDisplay(pr)
			
			if tt.expected == "" {
				if result != "" {
					t.Errorf("Expected empty result, got %q", result)
				}
			} else {
				if !contains(result, tt.expected) {
					t.Errorf("Expected result to contain %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		   (len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestGetPRCommentCount verifies comment count calculation
func TestGetPRCommentCount(t *testing.T) {
	tests := []struct {
		name           string
		pr             *github.PullRequest
		expectedOutput string
	}{
		{
			name: "PR with both comment types",
			pr: &github.PullRequest{
				Comments:       github.Int(5),
				ReviewComments: github.Int(12),
			},
			expectedOutput: "17", // 5 + 12
		},
		{
			name: "PR with only general comments",
			pr: &github.PullRequest{
				Comments:       github.Int(3),
				ReviewComments: github.Int(0),
			},
			expectedOutput: "3",
		},
		{
			name: "PR with only review comments",
			pr: &github.PullRequest{
				Comments:       github.Int(0),
				ReviewComments: github.Int(8),
			},
			expectedOutput: "8",
		},
		{
			name: "PR with no comments",
			pr: &github.PullRequest{
				Comments:       github.Int(0),
				ReviewComments: github.Int(0),
			},
			expectedOutput: "?", // Shows ? when API doesn't populate comment counts
		},
		{
			name: "PR with nil comment counts (should handle gracefully)",
			pr: &github.PullRequest{
				// Comments and ReviewComments are nil
			},
			expectedOutput: "?", // Shows ? when comment data not available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPRCommentCount(tt.pr)
			if result != tt.expectedOutput {
				t.Errorf("getPRCommentCount() = %v, want %v", result, tt.expectedOutput)
			}
		})
	}
}

// TestCreateTableRowsWithEnhancement tests the enhanced table row creation
func TestCreateTableRowsWithEnhancement(t *testing.T) {
	prs := []*github.PullRequest{
		{
			Number: github.Int(123),
			Title:  github.String("Test PR"),
			User:   &github.User{Login: github.String("testuser")},
			Base: &github.PullRequestBranch{
				Repo: &github.Repository{
					FullName: github.String("test/repo"),
				},
			},
			UpdatedAt: &github.Timestamp{Time: time.Now()},
			Draft:     github.Bool(false),
		},
	}
	
	// Test without enhanced data
	enhancedData := make(map[int]enhancedPRData)
	rows := createTableRowsWithEnhancement(prs, enhancedData)
	
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}
	
	if len(rows[0]) != 8 {
		t.Errorf("Expected 8 columns, got %d", len(rows[0]))
	}
	
	// Test with enhanced data
	enhancedData[123] = enhancedPRData{
		Number:         123,
		Comments:       5,
		ReviewComments: 3,
		ReviewStatus:   "approved",
		ChecksStatus:   "success",
		Mergeable:      "clean",
	}
	
	rowsEnhanced := createTableRowsWithEnhancement(prs, enhancedData)
	
	if len(rowsEnhanced) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rowsEnhanced))
	}
	
	// Comments column should show "8" (8 comments)
	commentsCol := rowsEnhanced[0][4] // Comments is 5th column (index 4)
	if commentsCol != "8" {
		t.Errorf("Expected comments column to be '8', got '%s'", commentsCol)
	}
	
	// Files column should show file changes when enhanced data is available
	filesCol := rowsEnhanced[0][5] // Files is 6th column (index 5)
	// This will be empty since we didn't set file change data in the enhanced data
	if filesCol != "-" {
		t.Errorf("Expected files column to be '-' when no file data, got '%s'", filesCol)
	}
	
	// Status column should show enhanced merge status with CI status
	statusCol := rowsEnhanced[0][2] // Status+CI is 3rd column (index 2) 
	// Note: This will have two lines, so we check if it contains the merge status
	if !contains(statusCol, "[✓] Ready") {
		t.Errorf("Expected status column to contain '[✓] Ready', got '%s'", statusCol)
	}
	
	// Review column should show enhanced review status
	reviewCol := rowsEnhanced[0][3] // Review is 4th column (index 3)
	if reviewCol != "Approved" {
		t.Errorf("Expected review column to be 'Approved', got '%s'", reviewCol)
	}
}

// TestGetPRCommentCountEnhanced tests enhanced comment count logic
func TestGetPRCommentCountEnhanced(t *testing.T) {
	pr := &github.PullRequest{
		Number:         github.Int(123),
		Comments:       github.Int(0), // List API returns 0
		ReviewComments: github.Int(0), // List API returns 0
	}
	
	tests := []struct {
		name         string
		enhancedData map[int]enhancedPRData
		expected     string
	}{
		{
			name:         "no enhanced data falls back to original logic",
			enhancedData: make(map[int]enhancedPRData),
			expected:     "?", // Should fall back to original logic returning "?"
		},
		{
			name: "enhanced data with comments shows count",
			enhancedData: map[int]enhancedPRData{
				123: {
					Comments:       5,
					ReviewComments: 3,
				},
			},
			expected: "8", // 5 + 3
		},
		{
			name: "enhanced data with zero comments shows dash",
			enhancedData: map[int]enhancedPRData{
				123: {
					Comments:       0,
					ReviewComments: 0,
				},
			},
			expected: "-",
		},
		{
			name: "enhanced data for different PR number falls back",
			enhancedData: map[int]enhancedPRData{
				999: { // Different PR number
					Comments:       10,
					ReviewComments: 5,
				},
			},
			expected: "?", // Should fall back since PR 123 not found
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPRCommentCountEnhanced(pr, tt.enhancedData)
			if result != tt.expected {
				t.Errorf("getPRCommentCountEnhanced() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetPRStatusIndicatorEnhanced tests enhanced merge status logic
func TestGetPRStatusIndicatorEnhanced(t *testing.T) {
	pr := &github.PullRequest{
		Number: github.Int(123),
		Draft:  github.Bool(false),
	}
	
	tests := []struct {
		name         string
		enhancedData map[int]enhancedPRData
		expected     string
	}{
		{
			name:         "no enhanced data falls back to original logic",
			enhancedData: make(map[int]enhancedPRData),
			expected:     "[✓] Ready", // Default for non-draft PR
		},
		{
			name: "enhanced data shows clean mergeable",
			enhancedData: map[int]enhancedPRData{
				123: {
					Mergeable:    "clean",
					ChecksStatus: "success",
				},
			},
			expected: "[✓] Ready",
		},
		{
			name: "enhanced data shows conflicts",
			enhancedData: map[int]enhancedPRData{
				123: {
					Mergeable:    "conflicts",
					ChecksStatus: "success",
				},
			},
			expected: "[X] Conflicts",
		},
		{
			name: "enhanced data shows failed checks",
			enhancedData: map[int]enhancedPRData{
				123: {
					Mergeable:    "clean",
					ChecksStatus: "failure",
				},
			},
			expected: "[!] Checks",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPRStatusIndicatorEnhanced(pr, tt.enhancedData)
			if result != tt.expected {
				t.Errorf("getPRStatusIndicatorEnhanced() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetPRReviewIndicatorEnhanced tests enhanced review status logic
func TestGetPRReviewIndicatorEnhanced(t *testing.T) {
	pr := &github.PullRequest{
		Number:    github.Int(123),
		UpdatedAt: &github.Timestamp{Time: time.Now()}, // Make it recent so it returns "Recent"
		Draft:     github.Bool(false),
	}
	
	tests := []struct {
		name         string
		enhancedData map[int]enhancedPRData
		expected     string
	}{
		{
			name:         "no enhanced data falls back to original logic",
			enhancedData: make(map[int]enhancedPRData),
			expected:     "Recent", // Recent PR (< 1 day old) with no reviewers
		},
		{
			name: "enhanced data shows approved",
			enhancedData: map[int]enhancedPRData{
				123: {
					ReviewStatus: "approved",
				},
			},
			expected: "Approved",
		},
		{
			name: "enhanced data shows changes requested",
			enhancedData: map[int]enhancedPRData{
				123: {
					ReviewStatus: "changes_requested",
				},
			},
			expected: "Changes",
		},
		{
			name: "enhanced data shows pending",
			enhancedData: map[int]enhancedPRData{
				123: {
					ReviewStatus: "pending",
				},
			},
			expected: "Pending",
		},
		{
			name: "enhanced data shows no review",
			enhancedData: map[int]enhancedPRData{
				123: {
					ReviewStatus: "no_review",
				},
			},
			expected: "No Review",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPRReviewIndicatorEnhanced(pr, tt.enhancedData)
			if result != tt.expected {
				t.Errorf("getPRReviewIndicatorEnhanced() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetPRActivityEnhanced tests the new activity column logic
func TestGetPRActivityEnhanced(t *testing.T) {
	pr := &github.PullRequest{
		Number:         github.Int(123),
		Comments:       github.Int(0), // List API returns 0
		ReviewComments: github.Int(0), // List API returns 0
	}
	
	tests := []struct {
		name         string
		enhancedData map[int]enhancedPRData
		expected     string
	}{
		{
			name:         "no enhanced data falls back to original logic",
			enhancedData: make(map[int]enhancedPRData),
			expected:     "?", // Should fall back to original getPRCommentCount
		},
		{
			name: "enhanced data with comments only shows comment count",
			enhancedData: map[int]enhancedPRData{
				123: {
					Comments:       5,
					ReviewComments: 3,
					ChangedFiles:   0, // No files changed
				},
			},
			expected: "8c", // 8 comments (compact format)
		},
		{
			name: "enhanced data with file changes only shows file stats",
			enhancedData: map[int]enhancedPRData{
				123: {
					Comments:       0,
					ReviewComments: 0,
					ChangedFiles:   5,
					Additions:      120,
					Deletions:      45,
				},
			},
			expected: "5F +120/-45", // 5 files with additions/deletions (compact format)
		},
		{
			name: "enhanced data with both comments and files shows both",
			enhancedData: map[int]enhancedPRData{
				123: {
					Comments:       3,
					ReviewComments: 2,
					ChangedFiles:   5,
					Additions:      120,
					Deletions:      45,
				},
			},
			expected: "5c • 5F +120/-45", // 5 comments, 5 files (compact format)
		},
		{
			name: "enhanced data with no activity shows dash",
			enhancedData: map[int]enhancedPRData{
				123: {
					Comments:       0,
					ReviewComments: 0,
					ChangedFiles:   0,
				},
			},
			expected: "-", // No activity (compact format)
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPRActivityEnhanced(pr, tt.enhancedData)
			if result != tt.expected {
				t.Errorf("getPRActivityEnhanced() = %v, want %v", result, tt.expected)
			}
		})
	}
}

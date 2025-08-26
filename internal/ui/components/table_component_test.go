package components

import (
	"strings"
	"testing"
	"time"

	"github.com/bjess9/pr-compass/internal/ui/types"
	gh "github.com/google/go-github/v55/github"
)

func TestNewTableComponent(t *testing.T) {
	component := NewTableComponent()

	if component == nil {
		t.Fatal("NewTableComponent returned nil")
	}
}

func TestTableComponent_CreateTable(t *testing.T) {
	component := NewTableComponent()

	table := component.CreateTable()

	// Verify table is created (just check it doesn't panic)
	// The table.Model is a struct, so it can't be nil
	_ = table
}

func TestTableComponent_CreateRows_EmptyPRs(t *testing.T) {
	component := NewTableComponent()

	rows := component.CreateRows([]*types.PRData{}, map[int]bool{})

	// Should return empty rows
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows for empty PRs, got %d", len(rows))
	}
}

func TestTableComponent_CreateRows_NilPRs(t *testing.T) {
	component := NewTableComponent()

	rows := component.CreateRows(nil, map[int]bool{})

	// Should return empty rows
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows for nil PRs, got %d", len(rows))
	}
}

func TestTableComponent_CreateRows_Basic(t *testing.T) {
	component := NewTableComponent()

	now := time.Now()
	pr := &types.PRData{
		PullRequest: &gh.PullRequest{
			Number:    gh.Int(456),
			Title:     gh.String("Another Test PR"),
			User:      &gh.User{Login: gh.String("anotheruser")},
			UpdatedAt: &gh.Timestamp{Time: now},
			CreatedAt: &gh.Timestamp{Time: now},
			State:     gh.String("open"),
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{Name: gh.String("test-repo")},
			},
		},
	}

	rows := component.CreateRows([]*types.PRData{pr}, map[int]bool{})

	// Verify we get one row
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	// Verify row has correct number of columns (9 columns as per CreateTableColumns)
	if len(row) != 9 {
		t.Fatalf("Expected 9 columns, got %d", len(row))
	}

	// Check some column content
	if !strings.Contains(row[0], "Another Test PR") {
		t.Errorf("Expected title in first column, got '%s'", row[0])
	}
	if row[1] != "anotheruser" {
		t.Errorf("Expected author 'anotheruser', got '%s'", row[1])
	}
	if row[2] != "test-repo" {
		t.Errorf("Expected repo 'test-repo', got '%s'", row[2])
	}
}

func TestTableComponent_CreateRows_WithEnhancement(t *testing.T) {
	component := NewTableComponent()

	now := time.Now()
	pr := &types.PRData{
		PullRequest: &gh.PullRequest{
			Number:    gh.Int(789),
			Title:     gh.String("Enhanced PR"),
			User:      &gh.User{Login: gh.String("enhanceduser")},
			UpdatedAt: &gh.Timestamp{Time: now},
			CreatedAt: &gh.Timestamp{Time: now},
			State:     gh.String("open"),
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{Name: gh.String("enhanced-repo")},
			},
		},
		Enhanced: &types.EnhancedData{
			Number:   789,
			Comments: 5,
		},
	}

	rows := component.CreateRows([]*types.PRData{pr}, map[int]bool{})

	// Should have one row
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if len(row) != 9 {
		t.Fatalf("Expected 9 columns, got %d", len(row))
	}
}

func TestTableComponent_CreateRows_NoEnhancement(t *testing.T) {
	component := NewTableComponent()

	now := time.Now()
	pr := &types.PRData{
		PullRequest: &gh.PullRequest{
			Number:    gh.Int(101),
			Title:     gh.String("Unenhanced PR"),
			User:      &gh.User{Login: gh.String("basicuser")},
			UpdatedAt: &gh.Timestamp{Time: now},
			CreatedAt: &gh.Timestamp{Time: now},
			State:     gh.String("closed"),
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{Name: gh.String("basic-repo")},
			},
		},
		// No Enhanced data
	}

	rows := component.CreateRows([]*types.PRData{pr}, map[int]bool{})

	// Should have one row
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if len(row) != 9 {
		t.Fatalf("Expected 9 columns, got %d", len(row))
	}
}

func TestTableComponent_CreateRows_SpecialCharacters(t *testing.T) {
	component := NewTableComponent()

	now := time.Now()
	pr := &types.PRData{
		PullRequest: &gh.PullRequest{
			Number:    gh.Int(555),
			Title:     gh.String("PR with émojis 🚀 and spëcial chars"),
			User:      &gh.User{Login: gh.String("spëcial-usér")},
			UpdatedAt: &gh.Timestamp{Time: now},
			CreatedAt: &gh.Timestamp{Time: now},
			State:     gh.String("open"),
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{Name: gh.String("special-repo")},
			},
		},
	}

	rows := component.CreateRows([]*types.PRData{pr}, map[int]bool{})

	// Should handle special characters without panicking
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if len(row) != 9 {
		t.Fatalf("Expected 9 columns, got %d", len(row))
	}

	// Should contain the special characters
	if !strings.Contains(row[0], "émojis") {
		t.Error("Title should preserve special characters")
	}
	if row[1] != "spëcial-usér" {
		t.Error("Author should preserve special characters")
	}
}

func TestTableComponent_CreateRows_MultiplePRs(t *testing.T) {
	component := NewTableComponent()

	now := time.Now()
	prs := []*types.PRData{
		{
			PullRequest: &gh.PullRequest{
				Number:    gh.Int(1),
				Title:     gh.String("First PR"),
				User:      &gh.User{Login: gh.String("user1")},
				UpdatedAt: &gh.Timestamp{Time: now},
				CreatedAt: &gh.Timestamp{Time: now},
				State:     gh.String("open"),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{Name: gh.String("repo1")},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number:    gh.Int(2),
				Title:     gh.String("Second PR"),
				User:      &gh.User{Login: gh.String("user2")},
				UpdatedAt: &gh.Timestamp{Time: now},
				CreatedAt: &gh.Timestamp{Time: now},
				State:     gh.String("closed"),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{Name: gh.String("repo2")},
				},
			},
		},
		{
			PullRequest: &gh.PullRequest{
				Number:    gh.Int(3),
				Title:     gh.String("Third PR"),
				User:      &gh.User{Login: gh.String("user3")},
				UpdatedAt: &gh.Timestamp{Time: now},
				CreatedAt: &gh.Timestamp{Time: now},
				State:     gh.String("open"),
				Base: &gh.PullRequestBranch{
					Repo: &gh.Repository{Name: gh.String("repo3")},
				},
			},
		},
	}

	rows := component.CreateRows(prs, map[int]bool{})

	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	// Check that all PRs are represented
	titles := []string{rows[0][0], rows[1][0], rows[2][0]}
	for i, expectedTitle := range []string{"First PR", "Second PR", "Third PR"} {
		if !strings.Contains(titles[i], expectedTitle) {
			t.Errorf("Expected row %d to contain '%s', got '%s'", i, expectedTitle, titles[i])
		}
	}
}

func TestTableComponent_CreateRows_WithEnhancementQueue(t *testing.T) {
	component := NewTableComponent()

	now := time.Now()
	pr := &types.PRData{
		PullRequest: &gh.PullRequest{
			Number:    gh.Int(999),
			Title:     gh.String("Enhancing PR"),
			User:      &gh.User{Login: gh.String("enhancinguser")},
			UpdatedAt: &gh.Timestamp{Time: now},
			CreatedAt: &gh.Timestamp{Time: now},
			State:     gh.String("open"),
			Base: &gh.PullRequestBranch{
				Repo: &gh.Repository{Name: gh.String("enhancing-repo")},
			},
		},
	}

	enhancementQueue := map[int]bool{999: true}
	rows := component.CreateRows([]*types.PRData{pr}, enhancementQueue)

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	// Check that enhancement queue status affects the display
	// (specific formatting depends on implementation)
	if len(row) != 9 {
		t.Fatalf("Expected 9 columns, got %d", len(row))
	}
}

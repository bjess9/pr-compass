package github

import (
	"testing"

	"github.com/google/go-github/v55/github"
)

// TestShouldExcludePR verifies the bot and title filtering logic works correctly
func TestShouldExcludePR(t *testing.T) {
	tests := []struct {
		name     string
		pr       *github.PullRequest
		filter   *PRFilter
		expected bool
	}{
		{
			name: "excludes renovate bot",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("renovate[bot]")},
				Title: github.String("Update dependency react"),
			},
			filter: DefaultFilter(),
			expected: true,
		},
		{
			name: "excludes dependabot",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("dependabot[bot]")},
				Title: github.String("Bump lodash from 4.17.20 to 4.17.21"),
			},
			filter: DefaultFilter(),
			expected: true,
		},
		{
			name: "excludes github-actions bot",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("github-actions[bot]")},
				Title: github.String("Auto-update dependencies"),
			},
			filter: DefaultFilter(),
			expected: true,
		},
		{
			name: "excludes by title pattern",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("human-dev")},
				Title: github.String("chore(deps): update all dependencies"),
			},
			filter: DefaultFilter(),
			expected: true,
		},
		{
			name: "includes human PR",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("human-dev")},
				Title: github.String("feat: add new authentication system"),
			},
			filter: DefaultFilter(),
			expected: false,
		},
		{
			name: "includes draft when drafts enabled",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("human-dev")},
				Title: github.String("draft: work in progress"),
				Draft: github.Bool(true),
			},
			filter: &PRFilter{IncludeDrafts: true},
			expected: false,
		},
		{
			name: "excludes draft when drafts disabled",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("human-dev")},
				Title: github.String("draft: work in progress"),
				Draft: github.Bool(true),
			},
			filter: &PRFilter{IncludeDrafts: false},
			expected: true,
		},
		{
			name: "partial bot name matching",
			pr: &github.PullRequest{
				User: &github.User{Login: github.String("renovate-enterprise-corp")},
				Title: github.String("Update all dependencies"),
			},
			filter: DefaultFilter(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldExcludePR(tt.pr, tt.filter)
			if result != tt.expected {
				t.Errorf("shouldExcludePR() = %v, want %v for PR: %s by %s", 
					result, tt.expected, tt.pr.GetTitle(), tt.pr.GetUser().GetLogin())
			}
		})
	}
}

// TestDefaultFilter verifies our default filter has sensible exclusions
func TestDefaultFilter(t *testing.T) {
	filter := DefaultFilter()
	
	if len(filter.ExcludeAuthors) == 0 {
		t.Error("Default filter should exclude bot authors")
	}
	
	if len(filter.ExcludeTitles) == 0 {
		t.Error("Default filter should exclude dependency update titles")
	}
	
	if !filter.IncludeDrafts {
		t.Error("Default filter should include drafts")
	}
	
	// Test that it includes common bot names
	botNames := []string{"renovate[bot]", "dependabot[bot]", "github-actions[bot]"}
	for _, botName := range botNames {
		found := false
		for _, excludeName := range filter.ExcludeAuthors {
			if excludeName == botName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Default filter should exclude %s", botName)
		}
	}
}

// TestPRFilterWithConfig verifies configuration-based filtering
func TestPRFilterWithConfig(t *testing.T) {
	customFilter := &PRFilter{
		ExcludeAuthors: []string{"custom-bot", "ci-user"},
		ExcludeTitles:  []string{"chore:", "docs:"},
		IncludeDrafts:  false,
	}
	
	tests := []struct {
		name     string
		author   string
		title    string
		isDraft  bool
		expected bool
	}{
		{"excludes custom bot", "custom-bot", "Some PR", false, true},
		{"excludes ci user", "ci-user", "Deploy fix", false, true},
		{"excludes chore title", "human-dev", "chore: update config", false, true},
		{"excludes docs title", "human-dev", "docs: update README", false, true},
		{"excludes draft", "human-dev", "WIP: new feature", true, true},
		{"includes normal PR", "human-dev", "feat: awesome feature", false, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &github.PullRequest{
				User:  &github.User{Login: github.String(tt.author)},
				Title: github.String(tt.title),
				Draft: github.Bool(tt.isDraft),
			}
			
			result := shouldExcludePR(pr, customFilter)
			if result != tt.expected {
				t.Errorf("shouldExcludePR() = %v, want %v for %s", result, tt.expected, tt.name)
			}
		})
	}
}

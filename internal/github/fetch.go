package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bjess9/pr-pilot/internal/config"
	"github.com/bjess9/pr-pilot/internal/errors"
	"github.com/google/go-github/v55/github"
)

// PRFilter holds filtering options for PRs
type PRFilter struct {
	ExcludeAuthors []string // Authors to exclude (e.g., "renovate[bot]", "dependabot[bot]")
	ExcludeTitles  []string // Title patterns to exclude (e.g., "chore(deps)", "Update")
	IncludeDrafts  bool     // Whether to include draft PRs
}

// DefaultFilter returns a sensible default filter that excludes common bots
func DefaultFilter() *PRFilter {
	return &PRFilter{
		ExcludeAuthors: []string{
			"renovate[bot]",
			"dependabot[bot]",
			"dependabot-preview[bot]",
			"github-actions[bot]",
			"greenkeeper[bot]",
		},
		ExcludeTitles: []string{
			"chore(deps):",
			"Update dependency",
			"Bump ",
			"Auto-update",
		},
		IncludeDrafts: true,
	}
}

// shouldExcludePR determines if a PR should be filtered out
func shouldExcludePR(pr *github.PullRequest, filter *PRFilter) bool {
	if filter == nil {
		return false
	}

	// Check if it's a draft and we're excluding drafts
	if pr.GetDraft() && !filter.IncludeDrafts {
		return true
	}

	// Check author exclusions
	author := pr.GetUser().GetLogin()
	for _, excludeAuthor := range filter.ExcludeAuthors {
		if strings.EqualFold(author, excludeAuthor) {
			return true
		}
		// Also check for partial matches for bot accounts
		if strings.Contains(excludeAuthor, "[bot]") {
			botName := strings.Replace(excludeAuthor, "[bot]", "", 1)
			if strings.Contains(strings.ToLower(author), strings.ToLower(botName)) {
				return true
			}
		}
	}

	// Check title exclusions
	title := pr.GetTitle()
	for _, excludePattern := range filter.ExcludeTitles {
		if strings.Contains(strings.ToLower(title), strings.ToLower(excludePattern)) {
			return true
		}
	}

	return false
}

// FetchOpenPRs fetches PRs from multiple repositories concurrently for better performance
func FetchOpenPRs(ctx context.Context, repos []string, token string) ([]*github.PullRequest, error) {
	return FetchOpenPRsWithFilter(ctx, repos, token, DefaultFilter())
}

// FetchOpenPRsWithFilter fetches PRs with filtering options
func FetchOpenPRsWithFilter(ctx context.Context, repos []string, token string, filter *PRFilter) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}
	return fetchOpenPRsWithFilter(ctx, client, repos, filter)
}

// fetchOpenPRsWithFilter fetches PRs with filtering options (private implementation)
func fetchOpenPRsWithFilter(ctx context.Context, client *github.Client, repos []string, filter *PRFilter) ([]*github.PullRequest, error) {

	// Use buffered channel and worker pool for better performance
	type repoResult struct {
		prs []*github.PullRequest
		err error
	}

	// Limit concurrent requests to avoid rate limiting
	const maxConcurrent = 15
	semaphore := make(chan struct{}, maxConcurrent)
	results := make(chan repoResult, len(repos))

	var wg sync.WaitGroup

	for _, repoFullName := range repos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				results <- repoResult{err: ctx.Err()}
				return
			case semaphore <- struct{}{}: // Acquire semaphore
			}
			defer func() { <-semaphore }() // Release semaphore

			parts := strings.Split(repo, "/")
			if len(parts) != 2 {
				results <- repoResult{err: errors.NewRepositoryInvalidError(repo, nil)}
				return
			}
			owner, repoName := parts[0], parts[1]

			opts := &github.PullRequestListOptions{
				State:     "open",
				Sort:      "updated", // Use 'updated' instead of 'created' for more relevant sorting
				Direction: "desc",
				ListOptions: github.ListOptions{
					PerPage: 50, // Increased from 15 for fewer API calls
				},
			}

			var repoPRs []*github.PullRequest
			for {
				// Check context before each API call
				select {
				case <-ctx.Done():
					results <- repoResult{err: ctx.Err()}
					return
				default:
				}

				prs, resp, err := client.PullRequests.List(ctx, owner, repoName, opts)
				if err != nil {
					results <- repoResult{err: fmt.Errorf("failed to fetch PRs from %s: %w", repo, err)}
					return
				}

				// Filter PRs before adding them
				for _, pr := range prs {
					if !shouldExcludePR(pr, filter) {
						repoPRs = append(repoPRs, pr)
					}
				}

				if resp.NextPage == 0 || len(repoPRs) >= 10 { // Limit to 10 PRs per repo for better performance
					break
				}
				opts.Page = resp.NextPage
			}

			results <- repoResult{prs: repoPRs}
		}(repoFullName)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allPRs []*github.PullRequest

	for result := range results {
		if result.err != nil {
			// Silently skip errors - don't interfere with TUI
			continue
		}
		allPRs = append(allPRs, result.prs...)
	}

	// Silently handle repository errors - don't interfere with TUI

	// Sort by updated time (most recently updated first)
	sort.Slice(allPRs, func(i, j int) bool {
		return allPRs[i].GetUpdatedAt().Time.After(allPRs[j].GetUpdatedAt().Time)
	})

	// Completed loading PRs
	return allPRs, nil
}

// FetchPRsFromConfig fetches PRs based on the configuration mode using the fetcher interface
func FetchPRsFromConfig(ctx context.Context, cfg *config.Config, token string) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}

	// Create filter based on config
	filter := createFilterFromConfig(cfg)

	// Create appropriate fetcher based on configuration
	fetcher := NewFetcher(cfg)

	return fetcher.FetchPRs(ctx, client, filter)
}

// createFilterFromConfig creates a PRFilter from configuration settings
func createFilterFromConfig(cfg *config.Config) *PRFilter {
	filter := &PRFilter{
		IncludeDrafts: true, // Default to including drafts
	}

	if cfg.ExcludeBots {
		filter.ExcludeAuthors = []string{
			"renovate[bot]",
			"renovate-bot",
			"renovate-enterprise",
			"dependabot[bot]",
			"dependabot-preview[bot]",
			"github-actions[bot]",
			"greenkeeper[bot]",
		}
		filter.ExcludeTitles = []string{
			"chore(deps):",
			"Update dependency",
			"Update all ",
			"Bump ",
			"Auto-update",
			"renovate-enterprise",
		}
	}

	// Add custom exclusions from config
	filter.ExcludeAuthors = append(filter.ExcludeAuthors, cfg.ExcludeAuthors...)
	filter.ExcludeTitles = append(filter.ExcludeTitles, cfg.ExcludeTitles...)
	filter.IncludeDrafts = cfg.IncludeDrafts

	return filter
}

// FetchPRsFromOrganization fetches all open PRs from all repositories in an organization
func FetchPRsFromOrganization(ctx context.Context, org string, token string) ([]*github.PullRequest, error) {
	return FetchPRsFromOrganizationWithFilter(ctx, org, token, DefaultFilter())
}

// FetchPRsFromOrganizationWithFilter fetches all open PRs from all repositories in an organization with filtering
func FetchPRsFromOrganizationWithFilter(ctx context.Context, org string, token string, filter *PRFilter) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}
	return fetchPRsFromOrganizationWithFilter(ctx, client, org, filter)
}

// fetchPRsFromOrganizationWithFilter fetches all open PRs from all repositories in an organization with filtering (private implementation)
func fetchPRsFromOrganizationWithFilter(ctx context.Context, client *github.Client, org string, filter *PRFilter) ([]*github.PullRequest, error) {

	// Loading repositories from organization...

	// Get all repositories in the organization
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Type:        "all",
		Sort:        "updated", // Get most recently updated repos first
	}

	var allRepos []string
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for org %s: %w", org, err)
		}

		for _, repo := range repos {
			if repo.GetArchived() || repo.GetDisabled() {
				continue // skip archived/disabled repos
			}
			// Only include repos that have been active recently and have recent activity
			if time.Since(repo.GetUpdatedAt().Time) < 60*24*time.Hour { // 2 months for better performance
				allRepos = append(allRepos, fmt.Sprintf("%s/%s", org, repo.GetName()))
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Found repositories, loading PRs...
	return fetchOpenPRsWithFilter(ctx, client, allRepos, filter)
}

// FetchPRsFromTeams fetches PRs from repositories belonging to specific teams
func FetchPRsFromTeams(ctx context.Context, org string, teams []string, token string) ([]*github.PullRequest, error) {
	return FetchPRsFromTeamsWithFilter(ctx, org, teams, token, DefaultFilter())
}

// FetchPRsFromTeamsWithFilter fetches PRs from repositories belonging to specific teams with filtering
func FetchPRsFromTeamsWithFilter(ctx context.Context, org string, teams []string, token string, filter *PRFilter) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}
	return fetchPRsFromTeamsWithFilter(ctx, client, org, teams, filter)
}

// fetchPRsFromTeamsWithFilter fetches PRs from repositories belonging to specific teams with filtering (private implementation)
func fetchPRsFromTeamsWithFilter(ctx context.Context, client *github.Client, org string, teams []string, filter *PRFilter) ([]*github.PullRequest, error) {

	// Loading repositories for teams...

	repoSet := make(map[string]bool)

	// Get repositories for each team
	for _, teamSlug := range teams {
		opts := &github.ListOptions{PerPage: 100}

		for {
			repos, resp, err := client.Teams.ListTeamReposBySlug(ctx, org, teamSlug, opts)
			if err != nil {
				fmt.Printf("Warning: Could not access team %s in org %s: %v\n", teamSlug, org, err)
				break
			}

			for _, repo := range repos {
				if repo.GetArchived() || repo.GetDisabled() {
					continue
				}
				repoName := fmt.Sprintf("%s/%s", org, repo.GetName())
				repoSet[repoName] = true
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}

	// Convert set to slice
	var allRepos []string
	for repo := range repoSet {
		allRepos = append(allRepos, repo)
	}

	if len(allRepos) == 0 {
		return []*github.PullRequest{}, nil
	}

	// Found repositories from teams, loading PRs...
	return fetchOpenPRsWithFilter(ctx, client, allRepos, filter)
}

// FetchPRsFromSearch uses GitHub's search API to find PRs based on a custom query
func FetchPRsFromSearch(ctx context.Context, query string, token string) ([]*github.PullRequest, error) {
	return FetchPRsFromSearchWithFilter(ctx, query, token, DefaultFilter())
}

// FetchPRsFromSearchWithFilter uses GitHub's search API to find PRs based on a custom query with filtering
func FetchPRsFromSearchWithFilter(ctx context.Context, query string, token string, filter *PRFilter) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}
	return fetchPRsFromSearchWithFilter(ctx, client, query, filter)
}

// fetchPRsFromSearchWithFilter uses GitHub's search API to find PRs based on a custom query with filtering (private implementation)
func fetchPRsFromSearchWithFilter(ctx context.Context, client *github.Client, query string, filter *PRFilter) ([]*github.PullRequest, error) {

	// Ensure the query includes PR and open filters
	if !strings.Contains(query, "is:pr") {
		query += " is:pr"
	}
	if !strings.Contains(query, "is:open") {
		query += " is:open"
	}

	// Add bot exclusions to search query for better performance if filter excludes them
	if filter != nil && len(filter.ExcludeAuthors) > 0 {
		for _, author := range filter.ExcludeAuthors {
			if strings.Contains(author, "[bot]") { // Only add bot exclusions to search query
				query += " -author:" + author
			}
		}
	}

	// Searching GitHub...

	opts := &github.SearchOptions{
		Sort:        "updated",
		Order:       "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allPRs []*github.PullRequest

	for {
		result, resp, err := client.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, errors.NewGitHubUnknownError(0, fmt.Errorf("search query failed: %w", err))
		}

		// Processing search results...

		// Process results concurrently for better performance
		type prResult struct {
			pr  *github.PullRequest
			err error
		}

		const maxConcurrent = 10
		semaphore := make(chan struct{}, maxConcurrent)
		prResults := make(chan prResult, len(result.Issues))
		var wg sync.WaitGroup

		// Convert Issues to PullRequests (GitHub's search returns Issues for PRs)
		for _, issue := range result.Issues {
			if !issue.IsPullRequest() {
				continue
			}

			wg.Add(1)
			go func(issue *github.Issue) {
				defer wg.Done()

				select {
				case <-ctx.Done():
					prResults <- prResult{err: ctx.Err()}
					return
				case semaphore <- struct{}{}:
				}
				defer func() { <-semaphore }()

				parts := strings.Split(issue.GetRepositoryURL(), "/")
				if len(parts) >= 2 {
					owner := parts[len(parts)-2]
					repo := parts[len(parts)-1]

					pr, _, err := client.PullRequests.Get(ctx, owner, repo, issue.GetNumber())
					prResults <- prResult{pr: pr, err: err}
				} else {
					prResults <- prResult{err: fmt.Errorf("invalid repository URL")}
				}
			}(issue)
		}

		go func() {
			wg.Wait()
			close(prResults)
		}()

		// Collect PR results
		for result := range prResults {
			if result.err != nil {
				continue // skip failed PR fetches
			}
			if result.pr != nil && !shouldExcludePR(result.pr, filter) {
				allPRs = append(allPRs, result.pr)
			}
		}

		if resp.NextPage == 0 || len(allPRs) >= 200 { // Reasonable limit
			break
		}
		opts.Page = resp.NextPage
	}

	// Sort by updated date (newest first)
	sort.Slice(allPRs, func(i, j int) bool {
		return allPRs[i].GetUpdatedAt().Time.After(allPRs[j].GetUpdatedAt().Time)
	})

	// Search completed
	return allPRs, nil
}

// FetchPRsFromTopics fetches PRs from repositories that have specific topics/labels
func FetchPRsFromTopics(ctx context.Context, org string, topics []string, token string) ([]*github.PullRequest, error) {
	return FetchPRsFromTopicsWithFilter(ctx, org, topics, token, DefaultFilter())
}

// FetchPRsFromTopicsWithFilter fetches PRs from repositories that have specific topics/labels with filtering
func FetchPRsFromTopicsWithFilter(ctx context.Context, org string, topics []string, token string, filter *PRFilter) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}
	return fetchPRsFromTopicsWithFilter(ctx, client, org, topics, filter)
}

// fetchPRsFromTopicsWithFilter fetches PRs from repositories that have specific topics/labels with filtering (private implementation)
func fetchPRsFromTopicsWithFilter(ctx context.Context, client *github.Client, org string, topics []string, filter *PRFilter) ([]*github.PullRequest, error) {

	repoSet := make(map[string]bool)

	// Search for repositories with each topic
	for _, topic := range topics {
		query := fmt.Sprintf("org:%s topic:%s", org, topic)

		opts := &github.SearchOptions{
			Sort:        "updated",
			Order:       "desc",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		for {
			result, resp, err := client.Search.Repositories(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search repositories with topic %s: %w", topic, err)
			}

			for _, repo := range result.Repositories {
				if repo.GetArchived() || repo.GetDisabled() {
					continue
				}
				repoName := repo.GetFullName()
				repoSet[repoName] = true
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}

	// Convert set to slice (limit for performance)
	var allRepos []string
	for repo := range repoSet {
		allRepos = append(allRepos, repo)
		if len(allRepos) >= 30 { // Limit to 30 repos for better performance
			break
		}
	}

	if len(allRepos) == 0 {
		return []*github.PullRequest{}, nil
	}

	// Loading PRs from repositories...
	return fetchOpenPRsWithFilter(ctx, client, allRepos, filter)
}

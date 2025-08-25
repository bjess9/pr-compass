package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/errors"
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

	prs, err := fetcher.FetchPRs(ctx, client, filter)
	if err != nil {
		return nil, err
	}
	
	// Apply global PR limit
	maxPRs := cfg.MaxPRs
	if maxPRs == 0 {
		maxPRs = 50 // Default limit
	}
	
	if len(prs) > maxPRs {
		// Sort by updated time (most recent first) and take the top N
		sort.Slice(prs, func(i, j int) bool {
			return prs[i].GetUpdatedAt().Time.After(prs[j].GetUpdatedAt().Time)
		})
		prs = prs[:maxPRs]
	}
	
	return prs, nil
}

// FetchPRsFromConfigWithCache fetches PRs using caching for improved performance
func FetchPRsFromConfigWithCache(ctx context.Context, cfg *config.Config, token string, prCache *cache.PRCache) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}

	// Create filter based on config
	filter := createFilterFromConfig(cfg)

	// Create appropriate fetcher based on configuration
	baseFetcher := NewFetcher(cfg)

	// If cache is provided, wrap fetcher with caching (5 minute TTL)
	var fetcher PRFetcher
	if prCache != nil {
		cacheTTL := 5 * time.Minute
		fetcher = NewCachedFetcher(baseFetcher, prCache, cacheTTL)
	} else {
		fetcher = baseFetcher
	}

	prs, err := fetcher.FetchPRs(ctx, client, filter)
	if err != nil {
		return nil, err
	}
	
	// Apply global PR limit
	maxPRs := cfg.MaxPRs
	if maxPRs == 0 {
		maxPRs = 50 // Default limit
	}
	
	if len(prs) > maxPRs {
		// Sort by updated time (most recent first) and take the top N
		sort.Slice(prs, func(i, j int) bool {
			return prs[i].GetUpdatedAt().Time.After(prs[j].GetUpdatedAt().Time)
		})
		prs = prs[:maxPRs]
	}
	
	return prs, nil
}

// FetchPRsFromConfigOptimized fetches PRs using all optimizations (GraphQL + Cache + Rate Limiting)
func FetchPRsFromConfigOptimized(ctx context.Context, cfg *config.Config, token string, prCache *cache.PRCache) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}

	// Create filter based on config
	filter := createFilterFromConfig(cfg)

	// Create appropriate base fetcher
	baseFetcher := NewFetcher(cfg)

	// Create fully optimized fetcher with GraphQL + Caching + Rate Limiting
	optimizedFetcher := NewOptimizedFetcher(baseFetcher, prCache, token)

	prs, err := optimizedFetcher.FetchPRs(ctx, client, filter)
	if err != nil {
		return nil, err
	}
	
	// Apply global PR limit
	maxPRs := cfg.MaxPRs
	if maxPRs == 0 {
		maxPRs = 50 // Default limit
	}
	
	if len(prs) > maxPRs {
		// Sort by updated time (most recent first) and take the top N
		sort.Slice(prs, func(i, j int) bool {
			return prs[i].GetUpdatedAt().Time.After(prs[j].GetUpdatedAt().Time)
		})
		prs = prs[:maxPRs]
	}
	
	return prs, nil
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

// FetchOpenPRsWithFilter fetches PRs with filtering options (kept for backward compatibility)
func FetchOpenPRsWithFilter(ctx context.Context, repos []string, token string, filter *PRFilter) ([]*github.PullRequest, error) {
	client, err := NewClient(token)
	if err != nil {
		return nil, err
	}
	return fetchOpenPRsWithFilter(ctx, client, repos, filter)
}

// fetchOpenPRsWithFilter fetches PRs with filtering options (used by ReposFetcher)
func fetchOpenPRsWithFilter(ctx context.Context, client *github.Client, repos []string, filter *PRFilter) ([]*github.PullRequest, error) {
	type repoResult struct {
		prs []*github.PullRequest
		err error
	}

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
			case semaphore <- struct{}{}:
			}
			defer func() { <-semaphore }()

			parts := strings.Split(repo, "/")
			if len(parts) != 2 {
				results <- repoResult{err: errors.NewRepositoryInvalidError(repo, nil)}
				return
			}
			owner, repoName := parts[0], parts[1]

			opts := &github.PullRequestListOptions{
				State:       "open",
				Sort:        "updated",
				Direction:   "desc",
				ListOptions: github.ListOptions{PerPage: 50},
			}

			var repoPRs []*github.PullRequest
			for {
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

				for _, pr := range prs {
					if !shouldExcludePR(pr, filter) {
						repoPRs = append(repoPRs, pr)
					}
				}

				if resp.NextPage == 0 || len(repoPRs) >= 10 {
					break
				}
				opts.Page = resp.NextPage
			}

			results <- repoResult{prs: repoPRs}
		}(repoFullName)
	}

	go func() {
		defer close(results)
		// Wait for all goroutines to complete or context to be cancelled
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// All goroutines completed normally
		case <-ctx.Done():
			// Context was cancelled, but we still need to wait for goroutines
			// to avoid resource leaks. They should exit quickly due to context cancellation.
			wg.Wait()
		}
	}()

	var allPRs []*github.PullRequest
	for result := range results {
		if result.err != nil {
			// Skip failed repos - could log the error if needed
			continue
		}
		allPRs = append(allPRs, result.prs...)
	}

	// Note: Could log summary if repos were skipped, but skipping for now
	// Note: Could check if we successfully fetched from repositories, but allowing empty results

	sort.Slice(allPRs, func(i, j int) bool {
		return allPRs[i].GetUpdatedAt().Time.After(allPRs[j].GetUpdatedAt().Time)
	})

	return allPRs, nil
}

// fetchPRsFromOrganizationWithFilter fetches PRs from an organization (used by OrganizationFetcher)
func fetchPRsFromOrganizationWithFilter(ctx context.Context, client *github.Client, org string, filter *PRFilter) ([]*github.PullRequest, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Type:        "all",
		Sort:        "updated",
	}

	var allRepos []string
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for org %s: %w", org, err)
		}

		for _, repo := range repos {
			if repo.GetArchived() || repo.GetDisabled() {
				continue
			}
			if time.Since(repo.GetUpdatedAt().Time) < 60*24*time.Hour {
				allRepos = append(allRepos, fmt.Sprintf("%s/%s", org, repo.GetName()))
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return fetchOpenPRsWithFilter(ctx, client, allRepos, filter)
}

// fetchPRsFromTeamsWithFilter fetches PRs from team repositories (used by TeamsFetcher)
func fetchPRsFromTeamsWithFilter(ctx context.Context, client *github.Client, org string, teams []string, filter *PRFilter) ([]*github.PullRequest, error) {
	repoSet := make(map[string]bool)

	for _, teamSlug := range teams {
		opts := &github.ListOptions{PerPage: 100}

		for {
			repos, resp, err := client.Teams.ListTeamReposBySlug(ctx, org, teamSlug, opts)
			if err != nil {
				// Break out of pagination loop, but continue with next team
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

	var allRepos []string
	for repo := range repoSet {
		allRepos = append(allRepos, repo)
	}

	if len(allRepos) == 0 {
		return []*github.PullRequest{}, nil
	}

	return fetchOpenPRsWithFilter(ctx, client, allRepos, filter)
}

// fetchPRsFromSearchWithFilter uses GitHub search API (used by SearchFetcher)
func fetchPRsFromSearchWithFilter(ctx context.Context, client *github.Client, query string, filter *PRFilter) ([]*github.PullRequest, error) {
	if !strings.Contains(query, "is:pr") {
		query += " is:pr"
	}
	if !strings.Contains(query, "is:open") {
		query += " is:open"
	}

	if filter != nil && len(filter.ExcludeAuthors) > 0 {
		for _, author := range filter.ExcludeAuthors {
			if strings.Contains(author, "[bot]") {
				query += " -author:" + author
			}
		}
	}

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

		type prResult struct {
			pr  *github.PullRequest
			err error
		}

		const maxConcurrent = 10
		semaphore := make(chan struct{}, maxConcurrent)
		prResults := make(chan prResult, len(result.Issues))
		var wg sync.WaitGroup

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
			defer close(prResults)
			// Wait for all goroutines to complete or context to be cancelled
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// All goroutines completed normally
			case <-ctx.Done():
				// Context was cancelled, but we still need to wait for goroutines
				// to avoid resource leaks. They should exit quickly due to context cancellation.
				wg.Wait()
			}
		}()

		var failedPRs int
		for result := range prResults {
			if result.err != nil {
				failedPRs++
				continue
			}
			if result.pr != nil && !shouldExcludePR(result.pr, filter) {
				allPRs = append(allPRs, result.pr)
			}
		}


		if resp.NextPage == 0 || len(allPRs) >= 200 {
			break
		}
		opts.Page = resp.NextPage
	}

	sort.Slice(allPRs, func(i, j int) bool {
		return allPRs[i].GetUpdatedAt().Time.After(allPRs[j].GetUpdatedAt().Time)
	})

	return allPRs, nil
}

// fetchPRsFromTopicsWithFilter fetches PRs from repositories with topics (used by TopicsFetcher)
func fetchPRsFromTopicsWithFilter(ctx context.Context, client *github.Client, org string, topics []string, filter *PRFilter) ([]*github.PullRequest, error) {
	repoSet := make(map[string]bool)

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

	var allRepos []string
	for repo := range repoSet {
		allRepos = append(allRepos, repo)
		if len(allRepos) >= 30 {
			break
		}
	}

	if len(allRepos) == 0 {
		return []*github.PullRequest{}, nil
	}

	return fetchOpenPRsWithFilter(ctx, client, allRepos, filter)
}

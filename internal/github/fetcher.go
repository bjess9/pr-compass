package github

import (
	"context"

	"github.com/bjess9/pr-pilot/internal/config"
	"github.com/google/go-github/v55/github"
)

// PRFetcher defines the interface for fetching pull requests from different sources
type PRFetcher interface {
	FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error)
}

// ReposFetcher fetches PRs from specific repositories
type ReposFetcher struct {
	Repos []string
}

func (f *ReposFetcher) FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error) {
	return fetchOpenPRsWithFilter(ctx, client, f.Repos, filter)
}

// OrganizationFetcher fetches PRs from all repositories in an organization
type OrganizationFetcher struct {
	Organization string
}

func (f *OrganizationFetcher) FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error) {
	return fetchPRsFromOrganizationWithFilter(ctx, client, f.Organization, filter)
}

// TeamsFetcher fetches PRs from repositories belonging to specific teams
type TeamsFetcher struct {
	Organization string
	Teams        []string
}

func (f *TeamsFetcher) FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error) {
	return fetchPRsFromTeamsWithFilter(ctx, client, f.Organization, f.Teams, filter)
}

// SearchFetcher fetches PRs using GitHub's search API
type SearchFetcher struct {
	SearchQuery string
}

func (f *SearchFetcher) FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error) {
	return fetchPRsFromSearchWithFilter(ctx, client, f.SearchQuery, filter)
}

// TopicsFetcher fetches PRs from repositories with specific topics
type TopicsFetcher struct {
	Organization string
	Topics       []string
}

func (f *TopicsFetcher) FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error) {
	return fetchPRsFromTopicsWithFilter(ctx, client, f.Organization, f.Topics, filter)
}

// NewFetcher creates a PRFetcher based on the configuration
func NewFetcher(cfg *config.Config) PRFetcher {
	switch cfg.Mode {
	case "repos":
		return &ReposFetcher{Repos: cfg.Repos}
	case "organization":
		return &OrganizationFetcher{Organization: cfg.Organization}
	case "teams":
		return &TeamsFetcher{Organization: cfg.Organization, Teams: cfg.Teams}
	case "search":
		return &SearchFetcher{SearchQuery: cfg.SearchQuery}
	case "topics":
		return &TopicsFetcher{Organization: cfg.TopicOrg, Topics: cfg.Topics}
	default:
		// fallback to repo mode
		return &ReposFetcher{Repos: cfg.Repos}
	}
}

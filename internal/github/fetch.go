package github

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/bjess9/pr-pilot/internal/config"
	"github.com/google/go-github/v55/github"
)

func FetchOpenPRs(repos []string, token string) ([]*github.PullRequest, error) {
    client, err := NewClient(token)
    if err != nil {
        return nil, err
    }
    ctx := context.Background()
    var allPRs []*github.PullRequest

    for _, repoFullName := range repos {
        parts := strings.Split(repoFullName, "/")
        if len(parts) != 2 {
            continue
        }
        owner, repo := parts[0], parts[1]

        opts := &github.PullRequestListOptions{
            State:     "open",
            Sort:      "created",
            Direction: "desc",
            ListOptions: github.ListOptions{
                PerPage: 15,
            },
        }

        prs, _, err := client.PullRequests.List(ctx, owner, repo, opts)
        if err != nil {
            return nil, err
        }

        allPRs = append(allPRs, prs...)
    }

    sort.Slice(allPRs, func(i, j int) bool {
        return allPRs[i].GetCreatedAt().Time.After(allPRs[j].GetCreatedAt().Time)
    })

    return allPRs, nil
}

// FetchPRsFromConfig fetches PRs based on the configuration mode
func FetchPRsFromConfig(cfg *config.Config, token string) ([]*github.PullRequest, error) {
    switch cfg.Mode {
    case "repos":
        return FetchOpenPRs(cfg.Repos, token)
    case "organization":
        return FetchPRsFromOrganization(cfg.Organization, token)
    case "teams":
        return FetchPRsFromTeams(cfg.Organization, cfg.Teams, token)
    case "search":
        return FetchPRsFromSearch(cfg.SearchQuery, token)
    case "topics":
        return FetchPRsFromTopics(cfg.TopicOrg, cfg.Topics, token)
    default:
        return FetchOpenPRs(cfg.Repos, token) // fallback to repo mode
    }
}

// FetchPRsFromOrganization fetches all open PRs from all repositories in an organization
func FetchPRsFromOrganization(org string, token string) ([]*github.PullRequest, error) {
    client, err := NewClient(token)
    if err != nil {
        return nil, err
    }
    ctx := context.Background()
    
    // Get all repositories in the organization
    opts := &github.RepositoryListByOrgOptions{
        ListOptions: github.ListOptions{PerPage: 100},
    }
    
    var allRepos []string
    for {
        repos, resp, err := client.Repositories.ListByOrg(ctx, org, opts)
        if err != nil {
            return nil, fmt.Errorf("failed to list repositories for org %s: %w", org, err)
        }
        
        for _, repo := range repos {
            if repo.GetArchived() {
                continue // skip archived repos
            }
            allRepos = append(allRepos, fmt.Sprintf("%s/%s", org, repo.GetName()))
        }
        
        if resp.NextPage == 0 {
            break
        }
        opts.Page = resp.NextPage
    }
    
    return FetchOpenPRs(allRepos, token)
}

// FetchPRsFromTeams fetches PRs from repositories belonging to specific teams
func FetchPRsFromTeams(org string, teams []string, token string) ([]*github.PullRequest, error) {
    client, err := NewClient(token)
    if err != nil {
        return nil, err
    }
    ctx := context.Background()
    
    repoSet := make(map[string]bool)
    
    // Get repositories for each team
    for _, teamSlug := range teams {
        opts := &github.ListOptions{PerPage: 100}
        
        for {
            repos, resp, err := client.Teams.ListTeamReposBySlug(ctx, org, teamSlug, opts)
            if err != nil {
                // If team doesn't exist or we don't have access, log and continue
                fmt.Printf("Warning: Could not access team %s in org %s: %v\n", teamSlug, org, err)
                break
            }
            
            for _, repo := range repos {
                if repo.GetArchived() {
                    continue // skip archived repos
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
    
    return FetchOpenPRs(allRepos, token)
}

// FetchPRsFromSearch uses GitHub's search API to find PRs based on a custom query
func FetchPRsFromSearch(query string, token string) ([]*github.PullRequest, error) {
    client, err := NewClient(token)
    if err != nil {
        return nil, err
    }
    ctx := context.Background()
    
    // Ensure the query includes PR and open filters
    if !strings.Contains(query, "is:pr") {
        query += " is:pr"
    }
    if !strings.Contains(query, "is:open") {
        query += " is:open"
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
            return nil, fmt.Errorf("search query failed: %w", err)
        }
        
        // Convert Issues to PullRequests (GitHub's search returns Issues for PRs)
        for _, issue := range result.Issues {
            if issue.IsPullRequest() {
                // Get the full PR details
                parts := strings.Split(issue.GetRepositoryURL(), "/")
                if len(parts) >= 2 {
                    owner := parts[len(parts)-2]
                    repo := parts[len(parts)-1]
                    
                    pr, _, err := client.PullRequests.Get(ctx, owner, repo, issue.GetNumber())
                    if err != nil {
                        continue // skip if we can't get PR details
                    }
                    allPRs = append(allPRs, pr)
                }
            }
        }
        
        if resp.NextPage == 0 {
            break
        }
        opts.Page = resp.NextPage
    }
    
    // Sort by created date (newest first)
    sort.Slice(allPRs, func(i, j int) bool {
        return allPRs[i].GetCreatedAt().Time.After(allPRs[j].GetCreatedAt().Time)
    })
    
    return allPRs, nil
}

// FetchPRsFromTopics fetches PRs from repositories that have specific topics/labels
func FetchPRsFromTopics(org string, topics []string, token string) ([]*github.PullRequest, error) {
    client, err := NewClient(token)
    if err != nil {
        return nil, err
    }
    ctx := context.Background()
    
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
                if repo.GetArchived() {
                    continue // skip archived repos
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
    
    // Convert set to slice
    var allRepos []string
    for repo := range repoSet {
        allRepos = append(allRepos, repo)
    }
    
    if len(allRepos) == 0 {
        return []*github.PullRequest{}, nil
    }
    
    // Fetch PRs from all repositories with the specified topics
    return FetchOpenPRs(allRepos, token)
}

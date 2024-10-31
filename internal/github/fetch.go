package github

import (
	"context"
	"sort"
	"strings"

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

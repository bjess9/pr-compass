package github

import (
	"context"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

func NewClient(token string) (*github.Client, error) {
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(context.Background(), ts)
    client := github.NewClient(tc)

    return client, nil
}

package github

import (
	"context"
	"os"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

// TODO: Implement user friendly auth
func NewClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		panic("GITHUB_TOKEN environment variable is not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return client
}

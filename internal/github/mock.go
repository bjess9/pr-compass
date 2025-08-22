package github

import (
	"fmt"
	"strings"
	"time"

	"github.com/bjess9/pr-pilot/internal/config"
	"github.com/google/go-github/v55/github"
)

// MockClient implements a mock GitHub client for testing
type MockClient struct {
	PRs          []*github.PullRequest
	Repositories []*github.Repository
	Error        error
}

// NewMockClient creates a new mock GitHub client with test data
func NewMockClient() *MockClient {
	return &MockClient{
		PRs:          generateTestPRs(),
		Repositories: generateTestRepos(),
	}
}

// FetchPRsFromConfig mocks the PR fetching functionality
func (m *MockClient) FetchPRsFromConfig(cfg *config.Config) ([]*github.PullRequest, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	switch cfg.Mode {
	case "repos":
		return m.filterPRsByRepos(cfg.Repos), nil
	case "organization":
		return m.filterPRsByOrg(cfg.Organization), nil
	case "teams":
		return m.filterPRsByTeams(cfg.Organization, cfg.Teams), nil
	case "search":
		return m.filterPRsBySearch(cfg.SearchQuery), nil
	case "topics":
		return m.filterPRsByTopics(cfg.TopicOrg, cfg.Topics), nil
	default:
		return m.PRs, nil
	}
}

func (m *MockClient) filterPRsByRepos(repos []string) []*github.PullRequest {
	var filtered []*github.PullRequest
	for _, pr := range m.PRs {
		repoName := pr.GetBase().GetRepo().GetFullName()
		for _, repo := range repos {
			if repoName == strings.TrimSpace(repo) {
				filtered = append(filtered, pr)
				break
			}
		}
	}
	return filtered
}

func (m *MockClient) filterPRsByOrg(org string) []*github.PullRequest {
	var filtered []*github.PullRequest
	for _, pr := range m.PRs {
		if pr.GetBase().GetRepo().GetOwner().GetLogin() == org {
			filtered = append(filtered, pr)
		}
	}
	return filtered
}

func (m *MockClient) filterPRsByTeams(org string, teams []string) []*github.PullRequest {
	// For mock, return PRs that match org and have certain patterns
	var filtered []*github.PullRequest
	for _, pr := range m.PRs {
		if pr.GetBase().GetRepo().GetOwner().GetLogin() == org {
			// Mock team membership by checking repo names
			for _, team := range teams {
				if strings.Contains(pr.GetBase().GetRepo().GetName(), team) {
					filtered = append(filtered, pr)
					break
				}
			}
		}
	}
	return filtered
}

func (m *MockClient) filterPRsBySearch(query string) []*github.PullRequest {
	// Simple mock search - return PRs that match basic patterns
	var filtered []*github.PullRequest
	for _, pr := range m.PRs {
		if strings.Contains(query, "is:pr is:open") {
			if strings.Contains(query, pr.GetBase().GetRepo().GetOwner().GetLogin()) {
				filtered = append(filtered, pr)
			}
		}
	}
	return filtered
}

func (m *MockClient) filterPRsByTopics(org string, topics []string) []*github.PullRequest {
	// Mock topic filtering by checking repo names/descriptions
	var filtered []*github.PullRequest
	for _, pr := range m.PRs {
		if pr.GetBase().GetRepo().GetOwner().GetLogin() == org {
			for _, topic := range topics {
				// Mock: if repo name contains topic, include it
				if strings.Contains(pr.GetBase().GetRepo().GetName(), topic) ||
					strings.Contains(pr.GetBase().GetRepo().GetDescription(), topic) {
					filtered = append(filtered, pr)
					break
				}
			}
		}
	}
	return filtered
}

// SetError allows tests to simulate API errors
func (m *MockClient) SetError(err error) {
	m.Error = err
}

// AddPR adds a PR to the mock data
func (m *MockClient) AddPR(pr *github.PullRequest) {
	m.PRs = append(m.PRs, pr)
}

// generateTestPRs creates realistic test PR data
func generateTestPRs() []*github.PullRequest {
	now := time.Now()

	prs := []*github.PullRequest{
		createTestPR(1, "Fix authentication bug", "johnsmith", "testorg/api-service", false, true, now.Add(-2*time.Hour), []string{"bug", "critical"}),
		createTestPR(2, "Add user dashboard feature", "janedoe", "testorg/frontend", true, false, now.Add(-1*time.Hour), []string{"feature", "ui"}),
		createTestPR(3, "Update database schema", "bobwilson", "testorg/backend-service", false, false, now.Add(-30*time.Minute), []string{"database", "migration"}),
		createTestPR(4, "Refactor payment processing", "alicechen", "testorg/payment-api", false, true, now.Add(-4*time.Hour), []string{"refactor", "payments"}),
		createTestPR(5, "Add integration tests", "davidkim", "testorg/test-automation", true, false, now.Add(-45*time.Minute), []string{"testing", "ci"}),
	}

	return prs
}

// generateTestRepos creates test repository data
func generateTestRepos() []*github.Repository {
	repos := []*github.Repository{
		createTestRepo("api-service", "testorg", "Main API service", []string{"backend", "api"}),
		createTestRepo("frontend", "testorg", "React frontend application", []string{"frontend", "ui"}),
		createTestRepo("backend-service", "testorg", "Core backend services", []string{"backend", "core"}),
		createTestRepo("payment-api", "testorg", "Payment processing API", []string{"backend", "payments"}),
		createTestRepo("test-automation", "testorg", "Automated testing suite", []string{"testing", "ci"}),
	}

	return repos
}

// createTestPR creates a test pull request
func createTestPR(number int, title, author, repoFullName string, isDraft, mergeable bool, createdAt time.Time, labels []string) *github.PullRequest {
	parts := strings.Split(repoFullName, "/")
	owner := parts[0]
	repoName := parts[1]

	pr := &github.PullRequest{
		Number:    &number,
		Title:     &title,
		CreatedAt: &github.Timestamp{Time: createdAt},
		Draft:     &isDraft,
		Mergeable: &mergeable,
		User: &github.User{
			Login: &author,
		},
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{
				Name:     &repoName,
				FullName: &repoFullName,
				Owner: &github.User{
					Login: &owner,
				},
			},
		},
		HTMLURL: github.String(fmt.Sprintf("https://github.com/%s/pull/%d", repoFullName, number)),
	}

	// Add labels
	for _, label := range labels {
		pr.Labels = append(pr.Labels, &github.Label{
			Name: &label,
		})
	}

	// Add mock reviewers for some PRs
	if number%2 == 0 {
		pr.RequestedReviewers = []*github.User{
			{Login: github.String("reviewer1")},
		}
	}

	return pr
}

// createTestRepo creates a test repository
func createTestRepo(name, owner, description string, topics []string) *github.Repository {
	fullName := fmt.Sprintf("%s/%s", owner, name)
	archived := false

	repo := &github.Repository{
		Name:        &name,
		FullName:    &fullName,
		Description: &description,
		Archived:    &archived,
		Owner: &github.User{
			Login: &owner,
		},
		Topics: topics,
	}

	return repo
}

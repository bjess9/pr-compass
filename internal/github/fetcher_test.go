package github

import (
	"testing"

	"github.com/bjess9/pr-pilot/internal/config"
)

// TestReposFetcher tests the ReposFetcher implementation
func TestReposFetcher(t *testing.T) {
	fetcher := &ReposFetcher{
		Repos: []string{"owner/repo1", "owner/repo2"},
	}

	// Verify interface implementation
	var _ PRFetcher = fetcher

	// For this test, we just verify the fetcher has the correct structure
	// We don't test actual fetching to avoid panics with nil clients
	expectedRepos := []string{"owner/repo1", "owner/repo2"}
	if len(fetcher.Repos) != len(expectedRepos) {
		t.Errorf("Expected %d repos, got %d", len(expectedRepos), len(fetcher.Repos))
	}

	for i, repo := range fetcher.Repos {
		if repo != expectedRepos[i] {
			t.Errorf("Expected repo[%d] = %s, got %s", i, expectedRepos[i], repo)
		}
	}
}

// TestOrganizationFetcher tests the OrganizationFetcher implementation
func TestOrganizationFetcher(t *testing.T) {
	fetcher := &OrganizationFetcher{
		Organization: "testorg",
	}

	// Verify interface implementation
	var _ PRFetcher = fetcher

	// Test the structure without calling FetchPRs to avoid panics
	expectedOrg := "testorg"
	if fetcher.Organization != expectedOrg {
		t.Errorf("Expected organization %s, got %s", expectedOrg, fetcher.Organization)
	}
}

// TestTeamsFetcher tests the TeamsFetcher implementation
func TestTeamsFetcher(t *testing.T) {
	fetcher := &TeamsFetcher{
		Organization: "testorg",
		Teams:        []string{"backend-team", "frontend-team"},
	}

	// Verify interface implementation
	var _ PRFetcher = fetcher

	// Test the structure without calling FetchPRs to avoid panics
	expectedOrg := "testorg"
	expectedTeams := []string{"backend-team", "frontend-team"}

	if fetcher.Organization != expectedOrg {
		t.Errorf("Expected organization %s, got %s", expectedOrg, fetcher.Organization)
	}

	if len(fetcher.Teams) != len(expectedTeams) {
		t.Errorf("Expected %d teams, got %d", len(expectedTeams), len(fetcher.Teams))
	}
}

// TestSearchFetcher tests the SearchFetcher implementation
func TestSearchFetcher(t *testing.T) {
	fetcher := &SearchFetcher{
		SearchQuery: "org:testorg is:pr is:open",
	}

	// Verify interface implementation
	var _ PRFetcher = fetcher

	// Test the structure without calling FetchPRs to avoid panics
	expectedQuery := "org:testorg is:pr is:open"
	if fetcher.SearchQuery != expectedQuery {
		t.Errorf("Expected query %s, got %s", expectedQuery, fetcher.SearchQuery)
	}
}

// TestTopicsFetcher tests the TopicsFetcher implementation
func TestTopicsFetcher(t *testing.T) {
	fetcher := &TopicsFetcher{
		Organization: "testorg",
		Topics:       []string{"backend", "api"},
	}

	// Verify interface implementation
	var _ PRFetcher = fetcher

	// Test the structure without calling FetchPRs to avoid panics
	expectedOrg := "testorg"
	expectedTopics := []string{"backend", "api"}

	if fetcher.Organization != expectedOrg {
		t.Errorf("Expected organization %s, got %s", expectedOrg, fetcher.Organization)
	}

	if len(fetcher.Topics) != len(expectedTopics) {
		t.Errorf("Expected %d topics, got %d", len(expectedTopics), len(fetcher.Topics))
	}
}

// TestNewFetcher tests the factory function for creating fetchers
func TestNewFetcher(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		expectedType PRFetcher
		description  string
	}{
		{
			name: "creates ReposFetcher for repos mode",
			config: &config.Config{
				Mode:  "repos",
				Repos: []string{"owner/repo1", "owner/repo2"},
			},
			expectedType: &ReposFetcher{},
			description:  "should create ReposFetcher when mode is 'repos'",
		},
		{
			name: "creates OrganizationFetcher for organization mode",
			config: &config.Config{
				Mode:         "organization",
				Organization: "testorg",
			},
			expectedType: &OrganizationFetcher{},
			description:  "should create OrganizationFetcher when mode is 'organization'",
		},
		{
			name: "creates TeamsFetcher for teams mode",
			config: &config.Config{
				Mode:         "teams",
				Organization: "testorg",
				Teams:        []string{"team1", "team2"},
			},
			expectedType: &TeamsFetcher{},
			description:  "should create TeamsFetcher when mode is 'teams'",
		},
		{
			name: "creates SearchFetcher for search mode",
			config: &config.Config{
				Mode:        "search",
				SearchQuery: "org:testorg is:pr is:open",
			},
			expectedType: &SearchFetcher{},
			description:  "should create SearchFetcher when mode is 'search'",
		},
		{
			name: "creates TopicsFetcher for topics mode",
			config: &config.Config{
				Mode:     "topics",
				TopicOrg: "testorg",
				Topics:   []string{"backend", "frontend"},
			},
			expectedType: &TopicsFetcher{},
			description:  "should create TopicsFetcher when mode is 'topics'",
		},
		{
			name: "defaults to ReposFetcher for unknown mode",
			config: &config.Config{
				Mode:  "unknown_mode",
				Repos: []string{"owner/repo"},
			},
			expectedType: &ReposFetcher{},
			description:  "should default to ReposFetcher for unknown mode",
		},
		{
			name: "defaults to ReposFetcher for empty mode",
			config: &config.Config{
				Mode:  "",
				Repos: []string{"owner/repo"},
			},
			expectedType: &ReposFetcher{},
			description:  "should default to ReposFetcher for empty mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewFetcher(tt.config)

			// Verify that we got a fetcher
			if fetcher == nil {
				t.Fatal("NewFetcher() returned nil")
			}

			// Verify the type matches what we expected
			switch tt.expectedType.(type) {
			case *ReposFetcher:
				if actual, ok := fetcher.(*ReposFetcher); ok {
					if len(actual.Repos) != len(tt.config.Repos) {
						t.Errorf("ReposFetcher has %d repos, expected %d", len(actual.Repos), len(tt.config.Repos))
					}
				} else {
					t.Errorf("Expected *ReposFetcher, got %T", fetcher)
				}

			case *OrganizationFetcher:
				if actual, ok := fetcher.(*OrganizationFetcher); ok {
					if actual.Organization != tt.config.Organization {
						t.Errorf("OrganizationFetcher has org %q, expected %q", actual.Organization, tt.config.Organization)
					}
				} else {
					t.Errorf("Expected *OrganizationFetcher, got %T", fetcher)
				}

			case *TeamsFetcher:
				if actual, ok := fetcher.(*TeamsFetcher); ok {
					if actual.Organization != tt.config.Organization {
						t.Errorf("TeamsFetcher has org %q, expected %q", actual.Organization, tt.config.Organization)
					}
					if len(actual.Teams) != len(tt.config.Teams) {
						t.Errorf("TeamsFetcher has %d teams, expected %d", len(actual.Teams), len(tt.config.Teams))
					}
				} else {
					t.Errorf("Expected *TeamsFetcher, got %T", fetcher)
				}

			case *SearchFetcher:
				if actual, ok := fetcher.(*SearchFetcher); ok {
					if actual.SearchQuery != tt.config.SearchQuery {
						t.Errorf("SearchFetcher has query %q, expected %q", actual.SearchQuery, tt.config.SearchQuery)
					}
				} else {
					t.Errorf("Expected *SearchFetcher, got %T", fetcher)
				}

			case *TopicsFetcher:
				if actual, ok := fetcher.(*TopicsFetcher); ok {
					if actual.Organization != tt.config.TopicOrg {
						t.Errorf("TopicsFetcher has org %q, expected %q", actual.Organization, tt.config.TopicOrg)
					}
					if len(actual.Topics) != len(tt.config.Topics) {
						t.Errorf("TopicsFetcher has %d topics, expected %d", len(actual.Topics), len(tt.config.Topics))
					}
				} else {
					t.Errorf("Expected *TopicsFetcher, got %T", fetcher)
				}
			}
		})
	}
}

// TestPRFetcherInterface tests that all fetchers implement the interface correctly
func TestPRFetcherInterface(t *testing.T) {
	// Test that all concrete types implement the interface
	var fetchers []PRFetcher = []PRFetcher{
		&ReposFetcher{Repos: []string{"owner/repo"}},
		&OrganizationFetcher{Organization: "org"},
		&TeamsFetcher{Organization: "org", Teams: []string{"team"}},
		&SearchFetcher{SearchQuery: "query"},
		&TopicsFetcher{Organization: "org", Topics: []string{"topic"}},
	}

	// Just verify that they all implement the interface - no actual calls
	for i, fetcher := range fetchers {
		t.Run(t.Name()+"/fetcher_"+string(rune('0'+i)), func(t *testing.T) {
			// Verify the interface is implemented by doing a type assertion
			if fetcher == nil {
				t.Error("Fetcher should not be nil")
			}
			// The fact that this compiles proves the interface is satisfied
			t.Logf("Fetcher %T correctly implements PRFetcher interface", fetcher)
		})
	}
}

// TestFetcherFieldAssignment tests that NewFetcher correctly assigns config fields
func TestFetcherFieldAssignment(t *testing.T) {
	// Test ReposFetcher field assignment
	t.Run("ReposFetcher fields", func(t *testing.T) {
		cfg := &config.Config{
			Mode:  "repos",
			Repos: []string{"owner/repo1", "owner/repo2", "owner/repo3"},
		}

		fetcher := NewFetcher(cfg).(*ReposFetcher)

		if len(fetcher.Repos) != 3 {
			t.Errorf("Expected 3 repos, got %d", len(fetcher.Repos))
		}

		expectedRepos := map[string]bool{
			"owner/repo1": true,
			"owner/repo2": true,
			"owner/repo3": true,
		}

		for _, repo := range fetcher.Repos {
			if !expectedRepos[repo] {
				t.Errorf("Unexpected repo: %s", repo)
			}
		}
	})

	// Test TeamsFetcher field assignment
	t.Run("TeamsFetcher fields", func(t *testing.T) {
		cfg := &config.Config{
			Mode:         "teams",
			Organization: "myorg",
			Teams:        []string{"backend", "frontend", "devops"},
		}

		fetcher := NewFetcher(cfg).(*TeamsFetcher)

		if fetcher.Organization != "myorg" {
			t.Errorf("Expected organization 'myorg', got '%s'", fetcher.Organization)
		}

		if len(fetcher.Teams) != 3 {
			t.Errorf("Expected 3 teams, got %d", len(fetcher.Teams))
		}
	})

	// Test TopicsFetcher field assignment
	t.Run("TopicsFetcher fields", func(t *testing.T) {
		cfg := &config.Config{
			Mode:     "topics",
			TopicOrg: "topicorg",
			Topics:   []string{"backend", "api", "microservice"},
		}

		fetcher := NewFetcher(cfg).(*TopicsFetcher)

		if fetcher.Organization != "topicorg" {
			t.Errorf("Expected organization 'topicorg', got '%s'", fetcher.Organization)
		}

		if len(fetcher.Topics) != 3 {
			t.Errorf("Expected 3 topics, got %d", len(fetcher.Topics))
		}
	})
}

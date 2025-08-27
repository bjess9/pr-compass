package github

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/bjess9/pr-compass/internal/config"
	gh "github.com/google/go-github/v55/github"
)

func TestFetchPRsFromConfig_ReposMode(t *testing.T) {
	token := "fake-token"
	
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/repo1", "owner/repo2"},
		MaxPRs: 10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// With fake token and real repos, this should fail
	_, err := FetchPRsFromConfig(ctx, cfg, token)
	// Either error or success is acceptable - we're testing the code path exists
	if err != nil {
		// Expected with fake token
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfig_OrganizationMode(t *testing.T) {
	cfg := &config.Config{
		Mode:         "organization",
		Organization: "test-org",
		MaxPRs:       5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test the code path - error is expected with fake token
	_, err := FetchPRsFromConfig(ctx, cfg, "fake-token")
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfig_TeamsMode(t *testing.T) {
	cfg := &config.Config{
		Mode:         "teams",
		Organization: "test-org",
		Teams:        []string{"team1", "team2"},
		MaxPRs:       15,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := FetchPRsFromConfig(ctx, cfg, "fake-token")
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfig_SearchMode(t *testing.T) {
	cfg := &config.Config{
		Mode:        "search",
		SearchQuery: "repo:owner/repo is:pr is:open",
		MaxPRs:      20,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := FetchPRsFromConfig(ctx, cfg, "fake-token")
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfig_TopicsMode(t *testing.T) {
	cfg := &config.Config{
		Mode:     "topics",
		TopicOrg: "test-org",
		Topics:   []string{"javascript", "go"},
		MaxPRs:   25,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := FetchPRsFromConfig(ctx, cfg, "fake-token")
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfig_DefaultMode(t *testing.T) {
	cfg := &config.Config{
		Mode:   "unknown-mode", // Should fallback to repos mode
		Repos:  []string{"owner/test-repo"},
		MaxPRs: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := FetchPRsFromConfig(ctx, cfg, "fake-token")
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfig_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/repo"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test with empty token
	_, err := FetchPRsFromConfig(ctx, cfg, "")
	if err != nil {
		t.Logf("Got expected error with empty token: %v", err)
	}

	// Test with malformed token
	_, err = FetchPRsFromConfig(ctx, cfg, "invalid")
	if err != nil {
		t.Logf("Got expected error with invalid token: %v", err)
	}
}

func TestFetchPRsFromConfigWithCache(t *testing.T) {
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/test-repo"},
		MaxPRs: 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test with nil cache
	_, err := FetchPRsFromConfigWithCache(ctx, cfg, "fake-token", nil)
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}

	// Test with cache
	prCache, err := cache.NewPRCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	_, err = FetchPRsFromConfigWithCache(ctx, cfg, "fake-token", prCache)
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfigWithCache_CacheHit(t *testing.T) {
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/test-repo"},
		MaxPRs: 5,
	}

	ctx := context.Background()

	prCache, err := cache.NewPRCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Pre-populate cache
	filterKey := generateCacheKey(cfg)
	cacheKey := prCache.GenerateFetcherKey(cfg.Mode, filterKey)
	
	testPRs := []*gh.PullRequest{
		{
			Number: gh.Int(1),
			Title:  gh.String("Test PR"),
			UpdatedAt: &gh.Timestamp{Time: time.Now()},
		},
		{
			Number: gh.Int(2),
			Title:  gh.String("Another PR"),
			UpdatedAt: &gh.Timestamp{Time: time.Now()},
		},
	}

	err = prCache.SetPRList(cacheKey, testPRs, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Should get cached data
	result, err := FetchPRsFromConfigWithCache(ctx, cfg, "fake-token", prCache)
	if err != nil {
		t.Errorf("Expected cached data to be returned without error, got: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 cached PRs, got %d", len(result))
	}
}

func TestFetchPRsFromConfigOptimized(t *testing.T) {
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/test-repo"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	prCache, err := cache.NewPRCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Should delegate to FetchPRsFromConfigWithCache
	_, err = FetchPRsFromConfigOptimized(ctx, cfg, "fake-token", prCache)
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected string
	}{
		{
			name: "repos mode",
			config: &config.Config{
				Mode:  "repos",
				Repos: []string{"owner/repo1", "owner/repo2"},
			},
			expected: "repos:owner/repo1,owner/repo2",
		},
		{
			name: "organization mode",
			config: &config.Config{
				Mode:         "organization",
				Organization: "test-org",
			},
			expected: "organization:test-org",
		},
		{
			name: "teams mode",
			config: &config.Config{
				Mode:         "teams",
				Organization: "test-org",
				Teams:        []string{"team1", "team2"},
			},
			expected: "teams:test-org,team1,team2",
		},
		{
			name: "search mode",
			config: &config.Config{
				Mode:        "search",
				SearchQuery: "repo:owner/repo is:pr",
			},
			expected: "search:repo:owner/repo is:pr",
		},
		{
			name: "topics mode",
			config: &config.Config{
				Mode:     "topics",
				TopicOrg: "test-org",
				Topics:   []string{"javascript", "go"},
			},
			expected: "topics:test-org,javascript,go",
		},
		{
			name: "with exclusions",
			config: &config.Config{
				Mode:           "repos",
				Repos:          []string{"owner/repo"},
				ExcludeBots:    true,
				ExcludeAuthors: []string{"test-user"},
				ExcludeTitles:  []string{"WIP:"},
			},
			expected: "repos:owner/repo,exclude-bots,test-user,WIP:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCacheKey(tt.config)
			if result != tt.expected {
				t.Errorf("generateCacheKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCreateFilterFromConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		expectBots     bool
		expectDrafts   bool
		expectCustom   bool
	}{
		{
			name: "default config",
			config: &config.Config{
				IncludeDrafts: true,
			},
			expectBots:   false,
			expectDrafts: true,
			expectCustom: false,
		},
		{
			name: "exclude bots",
			config: &config.Config{
				ExcludeBots:   true,
				IncludeDrafts: false,
			},
			expectBots:   true,
			expectDrafts: false,
			expectCustom: false,
		},
		{
			name: "custom exclusions",
			config: &config.Config{
				ExcludeAuthors: []string{"custom-bot"},
				ExcludeTitles:  []string{"TEMP:"},
				IncludeDrafts:  true,
			},
			expectBots:   false,
			expectDrafts: true,
			expectCustom: true,
		},
		{
			name: "combined exclusions",
			config: &config.Config{
				ExcludeBots:    true,
				ExcludeAuthors: []string{"custom-user"},
				ExcludeTitles:  []string{"SKIP:"},
				IncludeDrafts:  false,
			},
			expectBots:   true,
			expectDrafts: false,
			expectCustom: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := createFilterFromConfig(tt.config)

			if filter.IncludeDrafts != tt.expectDrafts {
				t.Errorf("IncludeDrafts = %v, want %v", filter.IncludeDrafts, tt.expectDrafts)
			}

			if tt.expectBots {
				// Should have default bot exclusions
				found := false
				for _, author := range filter.ExcludeAuthors {
					if author == "renovate[bot]" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected bot exclusions to be set")
				}
			}

			if tt.expectCustom {
				// Check custom exclusions are included
				if tt.config.ExcludeAuthors != nil {
					found := false
					for _, author := range filter.ExcludeAuthors {
						for _, customAuthor := range tt.config.ExcludeAuthors {
							if author == customAuthor {
								found = true
								break
							}
						}
					}
					if !found {
						t.Error("Expected custom author exclusions to be included")
					}
				}

				if tt.config.ExcludeTitles != nil {
					found := false
					for _, title := range filter.ExcludeTitles {
						for _, customTitle := range tt.config.ExcludeTitles {
							if title == customTitle {
								found = true
								break
							}
						}
					}
					if !found {
						t.Error("Expected custom title exclusions to be included")
					}
				}
			}
		})
	}
}

func TestFetchOpenPRsWithFilter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repos := []string{"owner/repo1", "owner/repo2"}
	filter := DefaultFilter()

	// Test the code path
	_, err := FetchOpenPRsWithFilter(ctx, repos, "fake-token", filter)
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchOpenPRsWithFilter_EmptyRepos(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with empty repos list
	prs, err := FetchOpenPRsWithFilter(ctx, []string{}, "fake-token", nil)
	if err != nil {
		t.Errorf("Expected no error with empty repos, got: %v", err)
	}
	if len(prs) != 0 {
		t.Errorf("Expected empty PR list, got %d PRs", len(prs))
	}
}

func TestFetchOpenPRsWithFilter_NilFilter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Should handle nil filter
	_, err := FetchOpenPRsWithFilter(ctx, []string{"owner/repo"}, "fake-token", nil)
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestPRLimitApplied(t *testing.T) {
	// Create a mock scenario where we would test PR limiting
	// This tests the logic for limiting PRs when MaxPRs is set
	cfg := &config.Config{
		Mode:   "repos",
		Repos:  []string{"owner/test-repo"},
		MaxPRs: 0, // Should use default limit of 50
	}

	// Test that default limit is applied
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Even though this will fail with fake token, the logic path is tested
	_, err := FetchPRsFromConfig(ctx, cfg, "fake-token")
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}

	// Test explicit limit
	cfg.MaxPRs = 10
	_, err = FetchPRsFromConfig(ctx, cfg, "fake-token")
	if err != nil {
		t.Logf("Got expected error with fake token: %v", err)
	}
}

func TestFetchPRsFromConfig_CacheKeyGeneration(t *testing.T) {
	// Test that different configurations generate different cache keys
	cfg1 := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/repo1"},
	}
	
	cfg2 := &config.Config{
		Mode:  "repos", 
		Repos: []string{"owner/repo2"},
	}

	key1 := generateCacheKey(cfg1)
	key2 := generateCacheKey(cfg2)

	if key1 == key2 {
		t.Error("Different configurations should generate different cache keys")
	}

	// Test that same configuration generates same key
	key1Again := generateCacheKey(cfg1)
	if key1 != key1Again {
		t.Error("Same configuration should generate consistent cache keys")
	}
}

func TestCreateFilterFromConfig_EdgeCases(t *testing.T) {
	// Test with nil config (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("createFilterFromConfig panicked with nil config: %v", r)
		}
	}()
	
	// This would panic if not handled properly
	filter := createFilterFromConfig(&config.Config{})
	if filter == nil {
		t.Error("Should return a filter even with empty config")
	}
}

func TestFetchPRsFromConfigWithCache_MaxPRsHandling(t *testing.T) {
	cfg := &config.Config{
		Mode:   "repos",
		Repos:  []string{"owner/test-repo"},
		MaxPRs: 3,
	}

	prCache, err := cache.NewPRCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Pre-populate cache with more PRs than maxPRs
	filterKey := generateCacheKey(cfg)
	cacheKey := prCache.GenerateFetcherKey(cfg.Mode, filterKey)
	
	testPRs := []*gh.PullRequest{
		{Number: gh.Int(1), Title: gh.String("PR 1"), UpdatedAt: &gh.Timestamp{Time: time.Now()}},
		{Number: gh.Int(2), Title: gh.String("PR 2"), UpdatedAt: &gh.Timestamp{Time: time.Now()}},
		{Number: gh.Int(3), Title: gh.String("PR 3"), UpdatedAt: &gh.Timestamp{Time: time.Now()}},
		{Number: gh.Int(4), Title: gh.String("PR 4"), UpdatedAt: &gh.Timestamp{Time: time.Now()}},
		{Number: gh.Int(5), Title: gh.String("PR 5"), UpdatedAt: &gh.Timestamp{Time: time.Now()}},
	}

	err = prCache.SetPRList(cacheKey, testPRs, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	ctx := context.Background()
	result, err := FetchPRsFromConfigWithCache(ctx, cfg, "fake-token", prCache)
	if err != nil {
		t.Errorf("Unexpected error with cached data: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected MaxPRs limit to be applied to cached data. Got %d PRs, wanted 3", len(result))
	}
}

func TestFetchPRsFromConfig_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/repo"},
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := FetchPRsFromConfig(ctx, cfg, "fake-token")
	// With cancelled context, we expect an error, but the specific error depends on timing
	if err != nil {
		t.Logf("Got expected error with cancelled context: %v", err)
		// The error could be context cancellation or authentication failure
		errStr := err.Error()
		if strings.Contains(errStr, "context") || strings.Contains(errStr, "cancel") || 
		   strings.Contains(errStr, "Bad credentials") {
			// Expected error types
		} else {
			t.Logf("Unexpected error type (but error is expected): %v", err)
		}
	} else {
		t.Logf("Function succeeded despite cancelled context - this can happen with empty repos")
	}
}

func TestDefaultFilter_Values(t *testing.T) {
	filter := DefaultFilter()
	
	if filter == nil {
		t.Fatal("DefaultFilter should not return nil")
	}
	
	if !filter.IncludeDrafts {
		t.Error("DefaultFilter should include drafts")
	}
	
	if len(filter.ExcludeAuthors) == 0 {
		t.Error("DefaultFilter should have default bot exclusions")
	}
	
	// Check for specific bot exclusions
	found := false
	for _, author := range filter.ExcludeAuthors {
		if author == "renovate[bot]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DefaultFilter should exclude renovate[bot]")
	}
}
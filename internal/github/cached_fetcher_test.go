package github

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/google/go-github/v55/github"
)

func TestNewCachedFetcher(t *testing.T) {
	prCache, err := cache.NewPRCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	baseFetcher := &ReposFetcher{Repos: []string{"owner/repo1"}}
	cacheTTL := 5 * time.Minute

	cachedFetcher := NewCachedFetcher(baseFetcher, prCache, cacheTTL)

	if cachedFetcher == nil {
		t.Fatal("NewCachedFetcher() returned nil")
	}

	if cachedFetcher.fetcher != baseFetcher {
		t.Error("CachedFetcher.fetcher not set correctly")
	}

	if cachedFetcher.cache != prCache {
		t.Error("CachedFetcher.cache not set correctly")
	}

	if cachedFetcher.cacheTTL != cacheTTL {
		t.Error("CachedFetcher.cacheTTL not set correctly")
	}
}

func TestGenerateFetcherID(t *testing.T) {
	tests := []struct {
		name     string
		fetcher  PRFetcher
		expected string
	}{
		{
			"ReposFetcher",
			&ReposFetcher{Repos: []string{"owner/repo1", "owner/repo2"}},
			"repos:owner/repo1,owner/repo2",
		},
		{
			"OrganizationFetcher",
			&OrganizationFetcher{Organization: "myorg"},
			"org:myorg",
		},
		{
			"TeamsFetcher",
			&TeamsFetcher{Organization: "myorg", Teams: []string{"team1", "team2"}},
			"teams:myorg:team1,team2",
		},
		{
			"SearchFetcher",
			&SearchFetcher{SearchQuery: "is:pr is:open org:myorg"},
			"search:is:pr is:open org:myorg",
		},
		{
			"TopicsFetcher",
			&TopicsFetcher{Organization: "myorg", Topics: []string{"backend", "api"}},
			"topics:myorg:backend,api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFetcherID(tt.fetcher)
			if got != tt.expected {
				t.Errorf("generateFetcherID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateFilterKey(t *testing.T) {
	tests := []struct {
		name     string
		filter   *PRFilter
		expected string
	}{
		{
			"nil filter",
			nil,
			"nofilter",
		},
		{
			"empty filter",
			&PRFilter{},
			"nofilter",
		},
		{
			"filter with drafts",
			&PRFilter{IncludeDrafts: true},
			"drafts",
		},
		{
			"filter with exclude authors",
			&PRFilter{ExcludeAuthors: []string{"bot1", "bot2"}},
			"noauthors:bot1,bot2",
		},
		{
			"filter with exclude titles",
			&PRFilter{ExcludeTitles: []string{"chore:", "docs:"}},
			"notitles:chore:,docs:",
		},
		{
			"filter with multiple options",
			&PRFilter{
				IncludeDrafts:  true,
				ExcludeAuthors: []string{"renovate[bot]"},
				ExcludeTitles:  []string{"chore(deps):"},
			},
			"drafts_noauthors:renovate[bot]_notitles:chore(deps):",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFilterKey(tt.filter)
			if got != tt.expected {
				t.Errorf("generateFilterKey() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// MockFetcher implements PRFetcher for testing
type mockFetcher struct {
	prs   []*github.PullRequest
	err   error
	calls int
}

func (m *mockFetcher) FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error) {
	m.calls++
	return m.prs, m.err
}

// createTestCache creates a PRCache instance in a temporary directory for testing
func createTestCache(t *testing.T) *cache.PRCache {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")

	prCache, err := cache.NewPRCacheWithDir(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create test cache: %v", err)
	}

	return prCache
}

func TestCachedFetcher_CacheHitAndMiss(t *testing.T) {
	prCache := createTestCache(t)

	// Create test PRs
	testPRs := []*github.PullRequest{
		{Number: github.Int(1), Title: github.String("Test PR 1")},
		{Number: github.Int(2), Title: github.String("Test PR 2")},
	}

	mockFetcher := &mockFetcher{prs: testPRs}
	cachedFetcher := NewCachedFetcher(mockFetcher, prCache, 5*time.Minute)

	ctx := context.Background()
	filter := &PRFilter{IncludeDrafts: true}

	// First call should be cache miss and fetch from mock
	prs1, err := cachedFetcher.FetchPRs(ctx, nil, filter)
	if err != nil {
		t.Fatalf("FetchPRs() error = %v", err)
	}
	if len(prs1) != len(testPRs) {
		t.Errorf("Expected %d PRs, got %d", len(testPRs), len(prs1))
	}
	if mockFetcher.calls != 1 {
		t.Errorf("Expected 1 call to mock fetcher, got %d", mockFetcher.calls)
	}

	// Second call should be cache hit and not call mock
	prs2, err := cachedFetcher.FetchPRs(ctx, nil, filter)
	if err != nil {
		t.Fatalf("FetchPRs() error = %v", err)
	}
	if len(prs2) != len(testPRs) {
		t.Errorf("Expected %d PRs from cache, got %d", len(testPRs), len(prs2))
	}
	if mockFetcher.calls != 1 {
		t.Errorf("Expected still only 1 call to mock fetcher (cache hit), got %d", mockFetcher.calls)
	}
}

func TestCachedFetcher_DifferentFiltersGenerateDifferentCacheKeys(t *testing.T) {
	prCache := createTestCache(t)

	testPRs := []*github.PullRequest{
		{Number: github.Int(1), Title: github.String("Test PR")},
	}

	// Create fresh mock fetcher for this test
	mockFetcher := &mockFetcher{prs: testPRs}
	cachedFetcher := NewCachedFetcher(mockFetcher, prCache, 5*time.Minute)

	ctx := context.Background()
	filter1 := &PRFilter{IncludeDrafts: true}
	filter2 := &PRFilter{IncludeDrafts: false}

	// First call with filter1
	if _, err := cachedFetcher.FetchPRs(ctx, nil, filter1); err != nil {
		t.Fatalf("FetchPRs() error = %v", err)
	}

	// Second call with filter2 should be cache miss (different filter)
	if _, err := cachedFetcher.FetchPRs(ctx, nil, filter2); err != nil {
		t.Fatalf("FetchPRs() error = %v", err)
	}

	// Should have been called twice since filters are different
	if mockFetcher.calls != 2 {
		t.Errorf("Expected 2 calls to mock fetcher (different filters), got %d", mockFetcher.calls)
	}
}

func TestCachedFetcher_InvalidateCache(t *testing.T) {
	prCache := createTestCache(t)

	testPRs := []*github.PullRequest{
		{Number: github.Int(1), Title: github.String("Test PR")},
	}

	// Create fresh mock fetcher for this test
	mockFetcher := &mockFetcher{prs: testPRs}
	cachedFetcher := NewCachedFetcher(mockFetcher, prCache, 5*time.Minute)

	ctx := context.Background()
	filter := &PRFilter{IncludeDrafts: false} // Use different filter to avoid cache collision

	// First call to populate cache
	if _, err := cachedFetcher.FetchPRs(ctx, nil, filter); err != nil {
		t.Fatalf("FetchPRs() error = %v", err)
	}

	// Invalidate cache
	if err := cachedFetcher.InvalidateCache(filter); err != nil {
		t.Fatalf("InvalidateCache() error = %v", err)
	}

	// Third call should fetch again since cache was invalidated
	if _, err := cachedFetcher.FetchPRs(ctx, nil, filter); err != nil {
		t.Fatalf("FetchPRs() error = %v", err)
	}

	// Should have been called twice (initial + after invalidation)
	if mockFetcher.calls != 2 {
		t.Errorf("Expected 2 calls to mock fetcher (after invalidation), got %d", mockFetcher.calls)
	}
}

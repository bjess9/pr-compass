package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bjess9/pr-compass/internal/cache"
	ghApi "github.com/google/go-github/v55/github"
)

// CachedFetcher wraps any PRFetcher with caching capabilities
type CachedFetcher struct {
	fetcher   PRFetcher
	cache     *cache.PRCache
	fetcherID string
	cacheTTL  time.Duration
}

// NewCachedFetcher creates a new cached fetcher wrapper
func NewCachedFetcher(fetcher PRFetcher, cache *cache.PRCache, cacheTTL time.Duration) *CachedFetcher {
	return &CachedFetcher{
		fetcher:   fetcher,
		cache:     cache,
		fetcherID: generateFetcherID(fetcher),
		cacheTTL:  cacheTTL,
	}
}

// generateFetcherID creates a unique identifier for the fetcher type and config
func generateFetcherID(fetcher PRFetcher) string {
	switch f := fetcher.(type) {
	case *ReposFetcher:
		return fmt.Sprintf("repos:%s", strings.Join(f.Repos, ","))
	case *OrganizationFetcher:
		return fmt.Sprintf("org:%s", f.Organization)
	case *TeamsFetcher:
		return fmt.Sprintf("teams:%s:%s", f.Organization, strings.Join(f.Teams, ","))
	case *SearchFetcher:
		return fmt.Sprintf("search:%s", f.SearchQuery)
	case *TopicsFetcher:
		return fmt.Sprintf("topics:%s:%s", f.Organization, strings.Join(f.Topics, ","))
	default:
		return fmt.Sprintf("unknown:%T", fetcher)
	}
}

// FetchPRs implements the PRFetcher interface with caching
func (cf *CachedFetcher) FetchPRs(ctx context.Context, client *ghApi.Client, filter *PRFilter) ([]*ghApi.PullRequest, error) {
	// Generate cache key including filter parameters
	filterKey := generateFilterKey(filter)
	cacheKey := cf.cache.GenerateFetcherKey(cf.fetcherID, filterKey)

	// Try to get from cache first
	if cachedPRs, found := cf.cache.GetPRList(cacheKey); found {
		return cachedPRs, nil
	}

	// Cache miss - fetch from GitHub
	prs, err := cf.fetcher.FetchPRs(ctx, client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PRs: %w", err)
	}

	// Cache the results - ignore errors to not fail the request
	_ = cf.cache.SetPRList(cacheKey, prs, cf.cacheTTL)

	return prs, nil
}

// generateFilterKey creates a cache key component from filter parameters
func generateFilterKey(filter *PRFilter) string {
	if filter == nil {
		return "nofilter"
	}

	var parts []string
	
	if filter.IncludeDrafts {
		parts = append(parts, "drafts")
	}
	
	if len(filter.ExcludeAuthors) > 0 {
		parts = append(parts, fmt.Sprintf("noauthors:%s", strings.Join(filter.ExcludeAuthors, ",")))
	}
	
	if len(filter.ExcludeTitles) > 0 {
		parts = append(parts, fmt.Sprintf("notitles:%s", strings.Join(filter.ExcludeTitles, ",")))
	}

	if len(parts) == 0 {
		return "nofilter"
	}
	
	return strings.Join(parts, "_")
}

// BackgroundRefresh performs background refresh of cached data
func (cf *CachedFetcher) BackgroundRefresh(ctx context.Context, client *ghApi.Client, filter *PRFilter) {
	ticker := time.NewTicker(cf.cacheTTL / 2) // Refresh at half TTL
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			
			// Fetch fresh data in background
			filterKey := generateFilterKey(filter)
			cacheKey := cf.cache.GenerateFetcherKey(cf.fetcherID, filterKey)
			
			prs, err := cf.fetcher.FetchPRs(ctx, client, filter)
			if err != nil {
				continue
			}

			// Update cache
			_ = cf.cache.SetPRList(cacheKey, prs, cf.cacheTTL)
		}
	}
}

// InvalidateCache removes cached data for this fetcher
func (cf *CachedFetcher) InvalidateCache(filter *PRFilter) error {
	filterKey := generateFilterKey(filter)
	cacheKey := cf.cache.GenerateFetcherKey(cf.fetcherID, filterKey)
	
	// Try to remove both PR list and enhanced data
	prListPath := cf.cache.GetCachePath(cacheKey, "prlist")
	enhancedPath := cf.cache.GetCachePath(cacheKey, "enhanced")
	
	// Remove files (ignore errors as files might not exist)
	_ = cf.cache.RemoveCacheFile(prListPath)
	_ = cf.cache.RemoveCacheFile(enhancedPath)
	
	return nil
}

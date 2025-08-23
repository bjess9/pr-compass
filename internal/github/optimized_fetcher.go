package github

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/google/go-github/v55/github"
)

// OptimizedFetcher provides GraphQL + Caching + Rate Limiting for maximum performance
type OptimizedFetcher struct {
	graphqlFetcher *GraphQLPRFetcher
	cachedFetcher  *CachedFetcher
	rateLimiter    *RateLimiter
	token          string
}

// RateLimiter handles exponential backoff and rate limit management
type RateLimiter struct {
	lastRequest   time.Time
	backoffDelay  time.Duration
	maxBackoff    time.Duration
	minDelay      time.Duration
}

// NewRateLimiter creates a new rate limiter with exponential backoff
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		minDelay:   100 * time.Millisecond,
		maxBackoff: 60 * time.Second,
	}
}

// Wait implements intelligent rate limiting with exponential backoff
func (r *RateLimiter) Wait(ctx context.Context, rateLimit *RateLimitInfo) error {
	now := time.Now()
	
	// Check if we're approaching rate limits
	if rateLimit != nil && rateLimit.Remaining < 100 {
		// Calculate time until rate limit resets
		resetDelay := time.Until(rateLimit.ResetAt)
		
		if rateLimit.Remaining < 10 {
			// Very close to limit, wait for reset
			log.Printf("Rate limit critical (remaining: %d), waiting %v for reset", 
				rateLimit.Remaining, resetDelay)
			
			select {
			case <-time.After(resetDelay):
				r.backoffDelay = r.minDelay // Reset backoff after waiting
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		} else if rateLimit.Remaining < 100 {
			// Approaching limit, increase delays
			r.backoffDelay = min(r.backoffDelay*2, r.maxBackoff)
			if r.backoffDelay == 0 {
				r.backoffDelay = 1 * time.Second
			}
			
			log.Printf("Rate limit warning (remaining: %d), backing off %v", 
				rateLimit.Remaining, r.backoffDelay)
		}
	} else {
		// Plenty of rate limit, use minimal delay
		r.backoffDelay = r.minDelay
	}

	// Ensure minimum delay between requests
	timeSinceLastRequest := now.Sub(r.lastRequest)
	if timeSinceLastRequest < r.backoffDelay {
		delay := r.backoffDelay - timeSinceLastRequest
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	r.lastRequest = time.Now()
	return nil
}

// min returns the smaller of two durations
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// NewOptimizedFetcher creates a high-performance fetcher with all optimizations
func NewOptimizedFetcher(baseFetcher PRFetcher, prCache *cache.PRCache, token string) *OptimizedFetcher {
	// Create GraphQL fetcher with fallback to base fetcher
	graphqlFetcher := NewGraphQLPRFetcher(token, baseFetcher)
	
	// Wrap with caching for instant subsequent requests
	var cachedFetcher *CachedFetcher
	if prCache != nil {
		cachedFetcher = NewCachedFetcher(graphqlFetcher, prCache, 5*time.Minute)
	}
	
	return &OptimizedFetcher{
		graphqlFetcher: graphqlFetcher,
		cachedFetcher:  cachedFetcher,
		rateLimiter:    NewRateLimiter(),
		token:          token,
	}
}

// FetchPRs implements PRFetcher with full optimization stack
func (o *OptimizedFetcher) FetchPRs(ctx context.Context, client *github.Client, filter *PRFilter) ([]*github.PullRequest, error) {
	// Apply rate limiting before making any requests
	rateLimit := o.graphqlFetcher.GetRateLimit()
	if err := o.rateLimiter.Wait(ctx, rateLimit); err != nil {
		return nil, fmt.Errorf("rate limiter cancelled: %w", err)
	}

	// Use cached fetcher if available (GraphQL + Cache)
	if o.cachedFetcher != nil {
		log.Printf("Using optimized fetcher: GraphQL + Caching + Rate Limiting")
		return o.cachedFetcher.FetchPRs(ctx, client, filter)
	}

	// Fall back to GraphQL only
	log.Printf("Using GraphQL fetcher with rate limiting")
	return o.graphqlFetcher.FetchPRs(ctx, client, filter)
}

// GetRateLimit returns current rate limit info for status display
func (o *OptimizedFetcher) GetRateLimit() *RateLimitInfo {
	return o.graphqlFetcher.GetRateLimit()
}

// GetCacheStats returns cache statistics if caching is enabled
func (o *OptimizedFetcher) GetCacheStats() (int, int64, error) {
	if o.cachedFetcher != nil && o.cachedFetcher.cache != nil {
		return o.cachedFetcher.cache.GetCacheStats()
	}
	return 0, 0, nil
}


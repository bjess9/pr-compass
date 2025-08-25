package ui

import (
	"context"
	"sync"
	"time"

	"github.com/bjess9/pr-compass/internal/errors"
)

// GlobalRateLimiter manages API rate limits across all tabs
type GlobalRateLimiter struct {
	mu                sync.RWMutex
	requestsPerHour   int
	requestsRemaining int
	resetTime         time.Time
	lastUpdate        time.Time
	
	// Request queue for prioritization
	requestQueue    chan *RateLimitedRequest
	priorityQueue   chan *RateLimitedRequest
	
	// Tab coordination
	activeRequests  map[string]int // tab name -> active request count
	maxConcurrent   int
	
	// Shared resources
	sharedCache     *SharedCache
}

// RateLimitedRequest represents a request waiting for rate limit approval
type RateLimitedRequest struct {
	TabName     string
	Priority    RequestPriority
	RequestFunc func(context.Context) error
	ResultChan  chan error
	Timeout     time.Duration
}

// RequestPriority defines the priority levels for requests
type RequestPriority int

const (
	PriorityLow RequestPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent // User-initiated actions
)

// SharedCache manages cached data across all tabs to reduce API calls
type SharedCache struct {
	mu                sync.RWMutex
	prCache          map[string]*CachedPRData   // repo/number -> cached PR data
	repoCache        map[string]*CachedRepoData // org/repo -> cached repo metadata
	userCache        map[string]*CachedUserData // username -> cached user data
	enhancementCache map[string]*CachedEnhancementData // pr_key -> enhancement data
	
	// TTL management
	prTTL          time.Duration
	repoTTL        time.Duration
	userTTL        time.Duration
	enhancementTTL time.Duration
}

type CachedPRData struct {
	Data      interface{}
	ExpiresAt time.Time
	TabsUsing []string // Track which tabs are using this data
}

type CachedRepoData struct {
	Data      interface{}
	ExpiresAt time.Time
}

type CachedUserData struct {
	Data      interface{}
	ExpiresAt time.Time
}

type CachedEnhancementData struct {
	Data      interface{}
	ExpiresAt time.Time
	TabsUsing []string
}

// NewGlobalRateLimiter creates a new global rate limiter
func NewGlobalRateLimiter() *GlobalRateLimiter {
	limiter := &GlobalRateLimiter{
		requestsPerHour:   5000, // GitHub's limit for authenticated users
		requestsRemaining: 5000,
		resetTime:         time.Now().Add(time.Hour),
		lastUpdate:        time.Now(),
		requestQueue:      make(chan *RateLimitedRequest, 100),
		priorityQueue:     make(chan *RateLimitedRequest, 50),
		activeRequests:    make(map[string]int),
		maxConcurrent:     10, // Conservative limit for concurrent requests
		sharedCache:       NewSharedCache(),
	}
	
	// Start the request processor
	go limiter.processRequests()
	
	return limiter
}

// NewSharedCache creates a new shared cache
func NewSharedCache() *SharedCache {
	return &SharedCache{
		prCache:          make(map[string]*CachedPRData),
		repoCache:        make(map[string]*CachedRepoData),
		userCache:        make(map[string]*CachedUserData),
		enhancementCache: make(map[string]*CachedEnhancementData),
		prTTL:            5 * time.Minute,
		repoTTL:          30 * time.Minute,
		userTTL:          15 * time.Minute,
		enhancementTTL:   10 * time.Minute,
	}
}

// RequestWithRateLimit queues a request with rate limiting
func (rl *GlobalRateLimiter) RequestWithRateLimit(req *RateLimitedRequest) error {
	// Set default timeout if not specified
	if req.Timeout == 0 {
		req.Timeout = 30 * time.Second
	}
	
	// Choose the appropriate queue based on priority
	select {
	case rl.priorityQueue <- req:
		// High priority request queued
	case rl.requestQueue <- req:
		// Normal priority request queued
	case <-time.After(req.Timeout):
		return errors.NewGitHubRateLimitError("", nil)
	}
	
	// Wait for the result
	select {
	case err := <-req.ResultChan:
		return err
	case <-time.After(req.Timeout):
		return errors.NewGitHubRateLimitError("", nil)
	}
}

// processRequests handles the request queue with intelligent throttling
func (rl *GlobalRateLimiter) processRequests() {
	ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			rl.updateRateLimitInfo()
			rl.processNextRequest()
			
		case <-time.After(1 * time.Second):
			// Cleanup expired cache entries
			rl.sharedCache.cleanup()
		}
	}
}

// processNextRequest processes the next available request if rate limits allow
func (rl *GlobalRateLimiter) processNextRequest() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Check if we can make more requests
	totalActive := 0
	for _, count := range rl.activeRequests {
		totalActive += count
	}
	
	if totalActive >= rl.maxConcurrent {
		return // Too many concurrent requests
	}
	
	if rl.requestsRemaining <= 10 {
		return // Too close to rate limit
	}
	
	// Try priority queue first, then normal queue
	var req *RateLimitedRequest
	select {
	case req = <-rl.priorityQueue:
		// Got priority request
	case req = <-rl.requestQueue:
		// Got normal request
	default:
		return // No requests waiting
	}
	
	// Safety check
	if req == nil {
		return
	}
	
	// Track active request
	rl.activeRequests[req.TabName]++
	
	// Process request in goroutine
	go rl.executeRequest(req)
}

// executeRequest executes a rate-limited request
func (rl *GlobalRateLimiter) executeRequest(req *RateLimitedRequest) {
	defer func() {
		// Clean up active request tracking
		rl.mu.Lock()
		rl.activeRequests[req.TabName]--
		if rl.activeRequests[req.TabName] <= 0 {
			delete(rl.activeRequests, req.TabName)
		}
		rl.mu.Unlock()
	}()
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
	defer cancel()
	
	// Execute the request
	err := req.RequestFunc(ctx)
	
	// Update rate limit counters
	rl.mu.Lock()
	rl.requestsRemaining--
	rl.mu.Unlock()
	
	// Send result
	select {
	case req.ResultChan <- err:
		// Result sent successfully
	default:
		// Channel might be closed or full, ignore
	}
}

// updateRateLimitInfo updates rate limit information from GitHub headers
func (rl *GlobalRateLimiter) updateRateLimitInfo() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	
	// Reset counters if we've passed the reset time
	if now.After(rl.resetTime) {
		rl.requestsRemaining = rl.requestsPerHour
		rl.resetTime = now.Add(time.Hour)
	}
}

// GetRateLimitStatus returns current rate limit status
func (rl *GlobalRateLimiter) GetRateLimitStatus() (remaining int, resetTime time.Time) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.requestsRemaining, rl.resetTime
}

// UpdateFromGitHubHeaders updates rate limit info from actual GitHub response headers
func (rl *GlobalRateLimiter) UpdateFromGitHubHeaders(remaining int, resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.requestsRemaining = remaining
	rl.resetTime = resetTime
	rl.lastUpdate = time.Now()
}

// Shared cache methods

// cleanup removes expired entries from the shared cache
func (sc *SharedCache) cleanup() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	now := time.Now()
	
	// Clean PR cache
	for key, cached := range sc.prCache {
		if now.After(cached.ExpiresAt) {
			delete(sc.prCache, key)
		}
	}
	
	// Clean repo cache
	for key, cached := range sc.repoCache {
		if now.After(cached.ExpiresAt) {
			delete(sc.repoCache, key)
		}
	}
	
	// Clean user cache
	for key, cached := range sc.userCache {
		if now.After(cached.ExpiresAt) {
			delete(sc.userCache, key)
		}
	}
	
	// Clean enhancement cache
	for key, cached := range sc.enhancementCache {
		if now.After(cached.ExpiresAt) {
			delete(sc.enhancementCache, key)
		}
	}
}

// GetCachedPR retrieves cached PR data if available and fresh
func (sc *SharedCache) GetCachedPR(key string, tabName string) (interface{}, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	cached, exists := sc.prCache[key]
	if !exists || time.Now().After(cached.ExpiresAt) {
		return nil, false
	}
	
	// Track that this tab is using this data
	for _, tab := range cached.TabsUsing {
		if tab == tabName {
			return cached.Data, true
		}
	}
	
	// Add this tab to the users list
	sc.mu.RUnlock()
	sc.mu.Lock()
	cached.TabsUsing = append(cached.TabsUsing, tabName)
	sc.mu.Unlock()
	sc.mu.RLock()
	
	return cached.Data, true
}

// SetCachedPR stores PR data in the shared cache
func (sc *SharedCache) SetCachedPR(key string, data interface{}, tabName string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.prCache[key] = &CachedPRData{
		Data:      data,
		ExpiresAt: time.Now().Add(sc.prTTL),
		TabsUsing: []string{tabName},
	}
}

// Smart batching functions

// BatchRequestsByRepo groups requests by repository to optimize API calls
func (rl *GlobalRateLimiter) BatchRequestsByRepo(requests []*RateLimitedRequest) [][]*RateLimitedRequest {
	repoMap := make(map[string][]*RateLimitedRequest)
	
	for _, req := range requests {
		// Extract repo from tab name or request context (simplified)
		repo := extractRepoFromRequest(req)
		repoMap[repo] = append(repoMap[repo], req)
	}
	
	var batches [][]*RateLimitedRequest
	for _, repoRequests := range repoMap {
		// Split large batches to avoid overwhelming single repos
		for i := 0; i < len(repoRequests); i += 10 {
			end := i + 10
			if end > len(repoRequests) {
				end = len(repoRequests)
			}
			batches = append(batches, repoRequests[i:end])
		}
	}
	
	return batches
}

// extractRepoFromRequest extracts repository information from request for batching
func extractRepoFromRequest(req *RateLimitedRequest) string {
	// Simplified extraction - in real implementation, this would parse the request
	// to determine which repository it's targeting
	return req.TabName // Fallback to tab name
}

// Global instance
var GlobalLimiter *GlobalRateLimiter

// InitGlobalRateLimiter initializes the global rate limiter (call once at startup)
func InitGlobalRateLimiter() {
	GlobalLimiter = NewGlobalRateLimiter()
}
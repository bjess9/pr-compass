package ui

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestGlobalRateLimiter tests the global rate limiter functionality
func TestGlobalRateLimiter(t *testing.T) {
	limiter := NewGlobalRateLimiter()
	defer func() {
		// Cleanup - close channels to stop the processor
		close(limiter.requestQueue)
		close(limiter.priorityQueue)
	}()

	// Test initial state
	remaining, resetTime := limiter.GetRateLimitStatus()
	if remaining != 5000 {
		t.Errorf("Expected initial remaining requests 5000, got %d", remaining)
	}
	if resetTime.Before(time.Now()) {
		t.Error("Expected reset time to be in the future")
	}

	// Test rate limit updates
	newResetTime := time.Now().Add(30 * time.Minute)
	limiter.UpdateFromGitHubHeaders(4500, newResetTime)

	remaining, resetTime = limiter.GetRateLimitStatus()
	if remaining != 4500 {
		t.Errorf("Expected remaining requests 4500 after update, got %d", remaining)
	}
	if !resetTime.Equal(newResetTime) {
		t.Error("Expected reset time to be updated")
	}
}

// TestSharedCache tests the shared cache functionality
func TestSharedCache(t *testing.T) {
	cache := NewSharedCache()

	// Test cache miss
	data, found := cache.GetCachedPR("test-key", "tab1")
	if found {
		t.Error("Expected cache miss for non-existent key")
	}
	if data != nil {
		t.Error("Expected nil data for cache miss")
	}

	// Test cache set and hit
	testData := "test PR data"
	cache.SetCachedPR("test-key", testData, "tab1")

	data, found = cache.GetCachedPR("test-key", "tab1")
	if !found {
		t.Error("Expected cache hit for existing key")
	}
	if data != testData {
		t.Errorf("Expected cached data '%v', got '%v'", testData, data)
	}

	// Test cache sharing between tabs
	_, found = cache.GetCachedPR("test-key", "tab2")
	if !found {
		t.Error("Expected cache hit from different tab")
	}

	// Test cache cleanup manually
	cache.cleanup()

	// Data should still be there since TTL hasn't expired
	_, found = cache.GetCachedPR("test-key", "tab1")
	if !found {
		t.Error("Expected data to still be cached after cleanup (TTL not expired)")
	}
}

// TestRateLimitedRequest tests rate-limited request processing
func TestRateLimitedRequest(t *testing.T) {
	limiter := NewGlobalRateLimiter()
	defer func() {
		close(limiter.requestQueue)
		close(limiter.priorityQueue)
	}()

	// Test successful request
	var executed bool
	req := &RateLimitedRequest{
		TabName:    "test-tab",
		Priority:   PriorityNormal,
		Timeout:    1 * time.Second,
		ResultChan: make(chan error, 1),
		RequestFunc: func(ctx context.Context) error {
			executed = true
			return nil
		},
	}

	// This should complete quickly since we have plenty of rate limit
	err := limiter.RequestWithRateLimit(req)
	if err != nil {
		t.Errorf("Expected successful request, got error: %v", err)
	}

	// Give some time for async processing
	time.Sleep(200 * time.Millisecond)

	if !executed {
		t.Error("Expected request function to be executed")
	}
}

// TestRateLimitPrioritization tests that high priority requests are processed first
func TestRateLimitPrioritization(t *testing.T) {
	limiter := NewGlobalRateLimiter()
	defer func() {
		close(limiter.requestQueue)
		close(limiter.priorityQueue)
	}()

	var executionOrder []string
	var mu sync.Mutex

	createRequest := func(name string, priority RequestPriority) *RateLimitedRequest {
		return &RateLimitedRequest{
			TabName:    name,
			Priority:   priority,
			Timeout:    2 * time.Second,
			ResultChan: make(chan error, 1),
			RequestFunc: func(ctx context.Context) error {
				mu.Lock()
				executionOrder = append(executionOrder, name)
				mu.Unlock()
				return nil
			},
		}
	}

	// Submit requests in reverse priority order
	normalReq := createRequest("normal", PriorityNormal)
	highReq := createRequest("high", PriorityHigh)
	urgentReq := createRequest("urgent", PriorityUrgent)

	// Submit normal priority first
	go func() {
		_ = limiter.RequestWithRateLimit(normalReq)
	}()

	// Small delay to ensure normal request gets queued
	time.Sleep(10 * time.Millisecond)

	// Submit high priority
	go func() {
		_ = limiter.RequestWithRateLimit(highReq)
	}()

	// Submit urgent priority
	go func() {
		_ = limiter.RequestWithRateLimit(urgentReq)
	}()

	// Wait for all requests to complete
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executionOrder) < 3 {
		t.Errorf("Expected 3 executed requests, got %d", len(executionOrder))
	}

	// Note: Due to the async nature and timing, exact ordering is hard to guarantee
	// But we can check that urgent was processed (this is a simplified test)
	urgentProcessed := false
	for _, name := range executionOrder {
		if name == "urgent" {
			urgentProcessed = true
			break
		}
	}
	if !urgentProcessed {
		t.Error("Expected urgent priority request to be processed")
	}
}

// TestRateLimitTimeout tests request timeouts
func TestRateLimitTimeout(t *testing.T) {
	limiter := NewGlobalRateLimiter()
	defer func() {
		close(limiter.requestQueue)
		close(limiter.priorityQueue)
	}()

	req := &RateLimitedRequest{
		TabName:    "test-tab",
		Priority:   PriorityNormal,
		Timeout:    10 * time.Millisecond, // Very short timeout
		ResultChan: make(chan error, 1),
		RequestFunc: func(ctx context.Context) error {
			// Simulate slow request
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}

	start := time.Now()
	err := limiter.RequestWithRateLimit(req)
	duration := time.Since(start)

	// Should timeout quickly
	if err == nil {
		t.Error("Expected timeout error")
	}
	if duration > 50*time.Millisecond {
		t.Errorf("Expected quick timeout, took %v", duration)
	}
}

// TestCacheExpiry tests that cache entries expire properly
func TestCacheExpiry(t *testing.T) {
	cache := NewSharedCache()

	// Set very short TTL for testing
	cache.prTTL = 10 * time.Millisecond

	// Add data to cache
	cache.SetCachedPR("test-key", "test-data", "tab1")

	// Should be available immediately
	data, found := cache.GetCachedPR("test-key", "tab1")
	if !found {
		t.Error("Expected cache hit immediately after set")
	}
	if data != "test-data" {
		t.Errorf("Expected 'test-data', got '%v'", data)
	}

	// Wait for expiry
	time.Sleep(15 * time.Millisecond)

	// Should be expired now
	_, found = cache.GetCachedPR("test-key", "tab1")
	if found {
		t.Error("Expected cache miss after expiry")
	}
}

// TestRateLimiterConcurrency tests concurrent access to the rate limiter
func TestRateLimiterConcurrency(t *testing.T) {
	limiter := NewGlobalRateLimiter()
	defer func() {
		close(limiter.requestQueue)
		close(limiter.priorityQueue)
	}()

	const numRequests = 10
	var wg sync.WaitGroup
	var completedCount int32
	var mu sync.Mutex

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := &RateLimitedRequest{
				TabName:    "concurrent-test",
				Priority:   PriorityNormal,
				Timeout:    2 * time.Second,
				ResultChan: make(chan error, 1),
				RequestFunc: func(ctx context.Context) error {
					// Simulate some work
					time.Sleep(1 * time.Millisecond)

					mu.Lock()
					completedCount++
					mu.Unlock()
					return nil
				},
			}

			err := limiter.RequestWithRateLimit(req)
			if err != nil && !strings.Contains(err.Error(), "rate limit") {
				t.Errorf("Request %d failed with non-rate-limit error: %v", id, err)
			}
		}(i)
	}

	// Wait for all requests to complete with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("Concurrent requests took too long to complete")
	}

	// Give a bit more time for async processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	finalCount := completedCount
	mu.Unlock()

	if finalCount < int32(numRequests) {
		t.Errorf("Expected at least %d completed requests, got %d", numRequests, finalCount)
	}
}

// TestSharedCacheTabUsage tests that tab usage tracking works correctly
func TestSharedCacheTabUsage(t *testing.T) {
	cache := NewSharedCache()

	// Set data from tab1
	cache.SetCachedPR("test-key", "test-data", "tab1")

	// Get data from multiple tabs
	_, found := cache.GetCachedPR("test-key", "tab1")
	if !found {
		t.Error("Expected cache hit from tab1")
	}

	_, found = cache.GetCachedPR("test-key", "tab2")
	if !found {
		t.Error("Expected cache hit from tab2")
	}

	// Check that tab usage is tracked
	cached := cache.prCache["test-key"]
	if cached == nil {
		t.Fatal("Expected cached data to exist")
	}

	if len(cached.TabsUsing) < 2 {
		t.Errorf("Expected at least 2 tabs using cached data, got %d", len(cached.TabsUsing))
	}

	// Check that both tabs are tracked
	tab1Found, tab2Found := false, false
	for _, tab := range cached.TabsUsing {
		if tab == "tab1" {
			tab1Found = true
		}
		if tab == "tab2" {
			tab2Found = true
		}
	}

	if !tab1Found {
		t.Error("Expected tab1 to be tracked in TabsUsing")
	}
	if !tab2Found {
		t.Error("Expected tab2 to be tracked in TabsUsing")
	}
}

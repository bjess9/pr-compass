package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
)

// createTestCache creates a PRCache instance in a temporary directory for testing
func createTestCache(t *testing.T) *PRCache {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")

	cache, err := NewPRCacheWithDir(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create test cache: %v", err)
	}

	return cache
}

func TestNewPRCache(t *testing.T) {
	cache, err := NewPRCache()
	if err != nil {
		t.Fatalf("NewPRCache() error = %v", err)
	}

	if cache == nil {
		t.Fatal("NewPRCache() returned nil cache")
	}

	// Verify cache directory was created
	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".cache", "pr-compass")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created: %s", expectedDir)
	}
}

func TestCacheEntryExpiration(t *testing.T) {
	tests := []struct {
		name      string
		ttl       time.Duration
		sleepTime time.Duration
		expected  bool
	}{
		{"not expired", 100 * time.Millisecond, 50 * time.Millisecond, false},
		{"expired", 50 * time.Millisecond, 100 * time.Millisecond, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := CacheEntry[string]{
				Data:      "test",
				Timestamp: time.Now(),
				TTL:       tt.ttl,
			}

			time.Sleep(tt.sleepTime)

			if got := entry.IsExpired(); got != tt.expected {
				t.Errorf("CacheEntry.IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPRListCaching(t *testing.T) {
	cache := createTestCache(t)

	// Create test PRs
	testPRs := []*github.PullRequest{
		{
			Number: github.Int(1),
			Title:  github.String("Test PR 1"),
		},
		{
			Number: github.Int(2),
			Title:  github.String("Test PR 2"),
		},
	}

	cacheKey := "test-key"
	ttl := 5 * time.Minute

	// Cache miss initially
	if cachedPRs, found := cache.GetPRList(cacheKey); found {
		t.Errorf("Expected cache miss but got %d PRs", len(cachedPRs))
	}

	// Set PRs in cache
	if err := cache.SetPRList(cacheKey, testPRs, ttl); err != nil {
		t.Fatalf("SetPRList() error = %v", err)
	}

	// Cache hit after setting
	if cachedPRs, found := cache.GetPRList(cacheKey); !found {
		t.Error("Expected cache hit but got cache miss")
	} else if len(cachedPRs) != len(testPRs) {
		t.Errorf("Expected %d cached PRs, got %d", len(testPRs), len(cachedPRs))
	}
}

func TestEnhancedPRDataCaching(t *testing.T) {
	cache := createTestCache(t)

	// Create test enhanced PR data
	testData := map[string]EnhancedPRData{
		"1": {
			Number:       1,
			ReviewStatus: "approved",
			ChecksStatus: "success",
			Author:       "test-user",
		},
		"2": {
			Number:       2,
			ReviewStatus: "pending",
			ChecksStatus: "pending",
			Author:       "another-user",
		},
	}

	prKey := "test-pr-key"
	ttl := 5 * time.Minute

	// Cache miss initially
	if cachedData, found := cache.GetEnhancedPRData(prKey); found {
		t.Errorf("Expected cache miss but got %d enhanced entries", len(cachedData))
	}

	// Set enhanced data in cache
	if err := cache.SetEnhancedPRData(prKey, testData, ttl); err != nil {
		t.Fatalf("SetEnhancedPRData() error = %v", err)
	}

	// Cache hit after setting
	if cachedData, found := cache.GetEnhancedPRData(prKey); !found {
		t.Error("Expected cache hit but got cache miss")
	} else if len(cachedData) != len(testData) {
		t.Errorf("Expected %d cached enhanced entries, got %d", len(testData), len(cachedData))
	}
}

func TestExpiredCacheCleanup(t *testing.T) {
	cache := createTestCache(t)

	// Create test PRs with very short TTL
	testPRs := []*github.PullRequest{
		{Number: github.Int(1), Title: github.String("Short-lived PR")},
	}

	cacheKey := "expiring-key"
	shortTTL := 10 * time.Millisecond

	// Set PRs in cache with short TTL
	if err := cache.SetPRList(cacheKey, testPRs, shortTTL); err != nil {
		t.Fatalf("SetPRList() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should be cache miss due to expiration
	if _, found := cache.GetPRList(cacheKey); found {
		t.Error("Expected cache miss due to expiration but got cache hit")
	}
}

func TestGenerateFetcherKey(t *testing.T) {
	cache := createTestCache(t)

	// Test that keys are deterministic
	key1 := cache.GenerateFetcherKey("repos", "owner/repo1")
	key2 := cache.GenerateFetcherKey("repos", "owner/repo1")
	if key1 != key2 {
		t.Error("Same params should generate same key")
	}

	// Test that different params generate different keys
	key3 := cache.GenerateFetcherKey("repos", "owner/repo2")
	if key1 == key3 {
		t.Error("Different params should generate different key")
	}

	// Test key length
	if len(key1) != 16 {
		t.Errorf("GenerateFetcherKey() key length = %d, want 16", len(key1))
	}

	// Test different fetcher types generate different keys
	key4 := cache.GenerateFetcherKey("teams", "owner/repo1")
	if key1 == key4 {
		t.Error("Different fetcher types should generate different keys")
	}
}

func TestCleanExpiredEntries(t *testing.T) {
	cache := createTestCache(t)

	// Create some test cache entries with different TTLs
	testPRs := []*github.PullRequest{
		{Number: github.Int(1), Title: github.String("Test PR")},
	}

	// Set one entry that will expire quickly
	shortKey := "short-lived"
	longKey := "long-lived"

	if err := cache.SetPRList(shortKey, testPRs, 10*time.Millisecond); err != nil {
		t.Fatalf("SetPRList() error = %v", err)
	}

	if err := cache.SetPRList(longKey, testPRs, 10*time.Minute); err != nil {
		t.Fatalf("SetPRList() error = %v", err)
	}

	// Wait for short entry to expire
	time.Sleep(20 * time.Millisecond)

	// Clean expired entries
	ctx := context.Background()
	if err := cache.CleanExpiredEntries(ctx); err != nil {
		t.Fatalf("CleanExpiredEntries() error = %v", err)
	}

	// Verify expired entry is gone and non-expired entry remains
	if _, found := cache.GetPRList(shortKey); found {
		t.Error("Expected expired entry to be cleaned up")
	}

	if _, found := cache.GetPRList(longKey); !found {
		t.Error("Expected non-expired entry to remain")
	}
}

func TestGetCacheStats(t *testing.T) {
	cache := createTestCache(t)

	// Initially should have no cache files
	initialCount, initialSize, err := cache.GetCacheStats()
	if err != nil {
		t.Fatalf("GetCacheStats() error = %v", err)
	}

	// Create some cache entries
	testPRs := []*github.PullRequest{
		{Number: github.Int(1), Title: github.String("Test PR 1")},
		{Number: github.Int(2), Title: github.String("Test PR 2")},
	}

	if err := cache.SetPRList("test-key-1", testPRs, 5*time.Minute); err != nil {
		t.Fatalf("SetPRList() error = %v", err)
	}

	if err := cache.SetPRList("test-key-2", testPRs, 5*time.Minute); err != nil {
		t.Fatalf("SetPRList() error = %v", err)
	}

	// Check stats after adding cache entries
	count, size, err := cache.GetCacheStats()
	if err != nil {
		t.Fatalf("GetCacheStats() error = %v", err)
	}

	if count <= initialCount {
		t.Errorf("Expected cache file count to increase from %d, got %d", initialCount, count)
	}

	if size <= initialSize {
		t.Errorf("Expected cache size to increase from %d, got %d", initialSize, size)
	}
}

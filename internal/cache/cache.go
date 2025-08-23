package cache

import (
	"context"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/v55/github"
)

// CacheEntry represents a cached item with TTL
type CacheEntry[T any] struct {
	Data      T         `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry[T]) IsExpired() bool {
	return time.Since(e.Timestamp) > e.TTL
}

// PRCache handles caching of PR data
type PRCache struct {
	cacheDir string
}

// NewPRCache creates a new PR cache instance
func NewPRCache() (*PRCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cache", "pr-compass")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &PRCache{cacheDir: cacheDir}, nil
}

// NewPRCacheWithDir creates a new PR cache instance with custom directory (for testing)
func NewPRCacheWithDir(cacheDir string) (*PRCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &PRCache{cacheDir: cacheDir}, nil
}

// generateCacheKey creates a cache key from configuration parameters
func (c *PRCache) generateCacheKey(params ...string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%v", params)))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars for shorter filenames
}

// getCachePath returns the full path for a cache file
func (c *PRCache) getCachePath(key string, suffix string) string {
	return filepath.Join(c.cacheDir, fmt.Sprintf("%s_%s.cache", key, suffix))
}

// GetCachePath returns the full path for a cache file (exported for external use)
func (c *PRCache) GetCachePath(key string, suffix string) string {
	return c.getCachePath(key, suffix)
}

// removeCacheFile removes a cache file (used for invalidation)
func (c *PRCache) removeCacheFile(path string) error {
	return os.Remove(path)
}

// RemoveCacheFile removes a cache file (exported for external use)
func (c *PRCache) RemoveCacheFile(path string) error {
	return c.removeCacheFile(path)
}

// saveCacheEntry saves a cache entry to disk
func (c *PRCache) saveCacheEntry(path string, entry interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(entry)
}

// loadCacheEntry loads a cache entry from disk
func (c *PRCache) loadCacheEntry(path string, entry interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err // File doesn't exist or can't be opened
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	return decoder.Decode(entry)
}

// GetPRList retrieves cached PR list
func (c *PRCache) GetPRList(cacheKey string) ([]*github.PullRequest, bool) {
	path := c.getCachePath(cacheKey, "prlist")
	
	var entry CacheEntry[[]*github.PullRequest]
	if err := c.loadCacheEntry(path, &entry); err != nil {
		return nil, false
	}

	if entry.IsExpired() {
		// Clean up expired cache file
		os.Remove(path)
		return nil, false
	}

	return entry.Data, true
}

// SetPRList caches PR list with TTL
func (c *PRCache) SetPRList(cacheKey string, prs []*github.PullRequest, ttl time.Duration) error {
	path := c.getCachePath(cacheKey, "prlist")
	
	entry := CacheEntry[[]*github.PullRequest]{
		Data:      prs,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	return c.saveCacheEntry(path, &entry)
}

// EnhancedPRData represents the enhanced PR information we cache
type EnhancedPRData struct {
	Number          int    `json:"number"`
	ReviewStatus    string `json:"review_status"`
	ChecksStatus    string `json:"checks_status"`
	MergeableStatus string `json:"mergeable_status"`
	Author          string `json:"author"`
	Title           string `json:"title"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GetEnhancedPRData retrieves cached enhanced PR data
func (c *PRCache) GetEnhancedPRData(prKey string) (map[string]EnhancedPRData, bool) {
	path := c.getCachePath(prKey, "enhanced")
	
	var entry CacheEntry[map[string]EnhancedPRData]
	if err := c.loadCacheEntry(path, &entry); err != nil {
		return nil, false
	}

	if entry.IsExpired() {
		// Clean up expired cache file
		os.Remove(path)
		return nil, false
	}

	return entry.Data, true
}

// SetEnhancedPRData caches enhanced PR data with TTL
func (c *PRCache) SetEnhancedPRData(prKey string, data map[string]EnhancedPRData, ttl time.Duration) error {
	path := c.getCachePath(prKey, "enhanced")
	
	entry := CacheEntry[map[string]EnhancedPRData]{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	return c.saveCacheEntry(path, &entry)
}

// GenerateFetcherKey creates a cache key for a specific fetcher configuration
func (c *PRCache) GenerateFetcherKey(fetcherType string, params ...string) string {
	allParams := append([]string{fetcherType}, params...)
	return c.generateCacheKey(allParams...)
}

// CleanExpiredEntries removes expired cache files
func (c *PRCache) CleanExpiredEntries(ctx context.Context) error {
	files, err := filepath.Glob(filepath.Join(c.cacheDir, "*.cache"))
	if err != nil {
		return err
	}

	for _, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to load and check if expired
		var entry CacheEntry[interface{}]
		if err := c.loadCacheEntry(file, &entry); err == nil {
			if entry.IsExpired() {
				os.Remove(file)
			}
		}
	}

	return nil
}

// GetCacheStats returns cache statistics
func (c *PRCache) GetCacheStats() (int, int64, error) {
	files, err := filepath.Glob(filepath.Join(c.cacheDir, "*.cache"))
	if err != nil {
		return 0, 0, err
	}

	var totalSize int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
		}
	}

	return len(files), totalSize, nil
}

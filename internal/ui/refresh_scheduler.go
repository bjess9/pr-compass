package ui

import (
	"math"
	"sort"
	"sync"
	"time"
)

// RefreshScheduler coordinates refresh timing across all tabs to minimize rate limit impact
type RefreshScheduler struct {
	mu sync.RWMutex
	
	// Tab refresh timing
	tabSchedules map[string]*TabSchedule
	
	// Global coordination
	lastGlobalRefresh time.Time
	minInterval       time.Duration // Minimum time between any refreshes
	
	// Rate limit awareness
	rateLimitBuffer   int           // Buffer of requests to keep available
	maxSimultaneous   int           // Max tabs that can refresh simultaneously
}

// TabSchedule tracks refresh timing for a single tab
type TabSchedule struct {
	TabName         string
	RefreshInterval time.Duration
	LastRefresh     time.Time
	NextRefresh     time.Time
	Priority        RefreshPriority
	RequestCount    int // Estimated API requests per refresh
}

// RefreshPriority determines the order of refresh operations
type RefreshPriority int

const (
	RefreshPriorityLow    RefreshPriority = iota
	RefreshPriorityNormal
	RefreshPriorityHigh
	RefreshPriorityUrgent
)

// NewRefreshScheduler creates a new refresh scheduler
func NewRefreshScheduler() *RefreshScheduler {
	return &RefreshScheduler{
		tabSchedules:     make(map[string]*TabSchedule),
		minInterval:      30 * time.Second, // Min 30 seconds between any refreshes
		rateLimitBuffer:  100,               // Keep 100 requests in reserve
		maxSimultaneous:  2,                 // Max 2 tabs refreshing at once
	}
}

// AddTab registers a tab with the scheduler
func (rs *RefreshScheduler) AddTab(tabName string, refreshInterval time.Duration, priority RefreshPriority) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	
	// Estimate request count based on tab configuration
	requestCount := rs.estimateRequestCount(tabName)
	
	schedule := &TabSchedule{
		TabName:         tabName,
		RefreshInterval: refreshInterval,
		LastRefresh:     time.Time{}, // Never refreshed
		NextRefresh:     time.Now().Add(rs.staggerInitialRefresh(tabName)),
		Priority:        priority,
		RequestCount:    requestCount,
	}
	
	rs.tabSchedules[tabName] = schedule
}

// RemoveTab removes a tab from the scheduler
func (rs *RefreshScheduler) RemoveTab(tabName string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	delete(rs.tabSchedules, tabName)
}

// ShouldRefreshTab returns whether a tab should refresh now
func (rs *RefreshScheduler) ShouldRefreshTab(tabName string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	schedule, exists := rs.tabSchedules[tabName]
	if !exists {
		return false
	}
	
	now := time.Now()
	
	// Check if it's time for this tab
	if now.Before(schedule.NextRefresh) {
		return false
	}
	
	// Check global rate limiting constraints
	if rs.wouldExceedRateLimit(schedule) {
		return false
	}
	
	// Check if too many tabs are refreshing simultaneously
	if rs.countActiveRefreshes() >= rs.maxSimultaneous {
		return false
	}
	
	// Check minimum interval since last global refresh
	if now.Sub(rs.lastGlobalRefresh) < rs.minInterval {
		return false
	}
	
	return true
}

// MarkRefreshStarted records that a tab has started refreshing
func (rs *RefreshScheduler) MarkRefreshStarted(tabName string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	
	schedule, exists := rs.tabSchedules[tabName]
	if !exists {
		return
	}
	
	now := time.Now()
	schedule.LastRefresh = now
	schedule.NextRefresh = now.Add(schedule.RefreshInterval)
	rs.lastGlobalRefresh = now
}

// GetNextRefreshTimes returns when each tab should refresh next
func (rs *RefreshScheduler) GetNextRefreshTimes() map[string]time.Time {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	result := make(map[string]time.Time)
	for name, schedule := range rs.tabSchedules {
		result[name] = schedule.NextRefresh
	}
	return result
}

// GetOptimalRefreshOrder returns tabs in the order they should refresh to minimize rate limit impact
func (rs *RefreshScheduler) GetOptimalRefreshOrder() []string {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	now := time.Now()
	var candidates []string
	
	// Find tabs that need refreshing
	for name, schedule := range rs.tabSchedules {
		if now.After(schedule.NextRefresh) {
			candidates = append(candidates, name)
		}
	}
	
	// Sort by priority and request count (lower request count first)
	sort.Slice(candidates, func(i, j int) bool {
		scheduleI := rs.tabSchedules[candidates[i]]
		scheduleJ := rs.tabSchedules[candidates[j]]
		
		// Higher priority first
		if scheduleI.Priority != scheduleJ.Priority {
			return scheduleI.Priority > scheduleJ.Priority
		}
		
		// Lower request count first (to spread out heavy requests)
		return scheduleI.RequestCount < scheduleJ.RequestCount
	})
	
	return candidates
}

// staggerInitialRefresh spreads out initial refresh times to avoid thundering herd
func (rs *RefreshScheduler) staggerInitialRefresh(tabName string) time.Duration {
	// Use tab name hash to determine stagger time
	hash := 0
	for _, r := range tabName {
		hash = hash*31 + int(r)
	}
	
	// Stagger between 0-60 seconds based on hash
	stagger := time.Duration(hash%60) * time.Second
	return stagger
}

// estimateRequestCount estimates how many API requests a tab refresh will make
func (rs *RefreshScheduler) estimateRequestCount(tabName string) int {
	// This is a simplified estimate - in practice, you'd analyze the tab configuration
	// to determine the expected number of requests
	
	baseRequests := 5 // Basic PR list fetch
	
	// Add estimates for enhancement requests
	enhancementRequests := 20 // Conservative estimate for PR details, reviews, checks
	
	return baseRequests + enhancementRequests
}

// wouldExceedRateLimit checks if refreshing this tab would risk hitting rate limits
func (rs *RefreshScheduler) wouldExceedRateLimit(schedule *TabSchedule) bool {
	if GlobalLimiter == nil {
		return false
	}
	
	remaining, _ := GlobalLimiter.GetRateLimitStatus()
	
	// Don't refresh if we don't have enough requests + buffer
	requiredRequests := schedule.RequestCount + rs.rateLimitBuffer
	
	return remaining < requiredRequests
}

// countActiveRefreshes counts how many tabs are currently refreshing
func (rs *RefreshScheduler) countActiveRefreshes() int {
	if GlobalLimiter == nil {
		return 0
	}
	
	GlobalLimiter.mu.RLock()
	defer GlobalLimiter.mu.RUnlock()
	
	total := 0
	for _, count := range GlobalLimiter.activeRequests {
		total += count
	}
	
	// Estimate how many tabs this represents (rough approximation)
	return int(math.Ceil(float64(total) / 10.0)) // Assume ~10 requests per tab refresh
}

// AdjustRefreshIntervals dynamically adjusts refresh intervals based on rate limit status
func (rs *RefreshScheduler) AdjustRefreshIntervals() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	
	if GlobalLimiter == nil {
		return
	}
	
	remaining, resetTime := GlobalLimiter.GetRateLimitStatus()
	timeUntilReset := time.Until(resetTime)
	
	// If we're running low on rate limit, slow down refreshes
	if timeUntilReset > 0 && remaining > 0 {
		if remaining < 1000 { // Less than 1000 requests remaining
			for _, schedule := range rs.tabSchedules {
				// Increase refresh interval by 50%
				newInterval := time.Duration(float64(schedule.RefreshInterval) * 1.5)
				maxInterval := 30 * time.Minute // Don't go too slow
				if newInterval > maxInterval {
					newInterval = maxInterval
				}
				schedule.RefreshInterval = newInterval
			}
		} else if remaining > 3000 { // Plenty of requests available
			for _, schedule := range rs.tabSchedules {
				// Decrease refresh interval slightly (speed up)
				newInterval := time.Duration(float64(schedule.RefreshInterval) * 0.9)
				minInterval := 1 * time.Minute // Don't go too fast
				if newInterval < minInterval {
					newInterval = minInterval
				}
				schedule.RefreshInterval = newInterval
			}
		}
	}
}

// GetRateLimitSummary returns a summary of current rate limit status
func (rs *RefreshScheduler) GetRateLimitSummary() RateLimitSummary {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	summary := RateLimitSummary{
		TabCount: len(rs.tabSchedules),
	}
	
	if GlobalLimiter != nil {
		summary.RequestsRemaining, summary.ResetTime = GlobalLimiter.GetRateLimitStatus()
		
		GlobalLimiter.mu.RLock()
		summary.ActiveRequests = 0
		for _, count := range GlobalLimiter.activeRequests {
			summary.ActiveRequests += count
		}
		GlobalLimiter.mu.RUnlock()
	}
	
	return summary
}

// RateLimitSummary provides a summary of current rate limit status
type RateLimitSummary struct {
	TabCount          int
	RequestsRemaining int
	ResetTime         time.Time
	ActiveRequests    int
}
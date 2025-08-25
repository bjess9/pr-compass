package ui

import (
	"testing"
	"time"
)

// TestRefreshScheduler tests basic refresh scheduler functionality
func TestRefreshScheduler(t *testing.T) {
	scheduler := NewRefreshScheduler()

	// Test initial state
	if len(scheduler.tabSchedules) != 0 {
		t.Error("Expected no tabs initially")
	}

	// Add a tab
	scheduler.AddTab("test-tab", 5*time.Minute, RefreshPriorityNormal)

	if len(scheduler.tabSchedules) != 1 {
		t.Errorf("Expected 1 tab after adding, got %d", len(scheduler.tabSchedules))
	}

	schedule, exists := scheduler.tabSchedules["test-tab"]
	if !exists {
		t.Fatal("Expected tab schedule to exist")
	}

	if schedule.TabName != "test-tab" {
		t.Errorf("Expected tab name 'test-tab', got '%s'", schedule.TabName)
	}
	if schedule.RefreshInterval != 5*time.Minute {
		t.Errorf("Expected refresh interval 5m, got %v", schedule.RefreshInterval)
	}
	if schedule.Priority != RefreshPriorityNormal {
		t.Errorf("Expected normal priority, got %v", schedule.Priority)
	}

	// Test removal
	scheduler.RemoveTab("test-tab")
	if len(scheduler.tabSchedules) != 0 {
		t.Errorf("Expected no tabs after removal, got %d", len(scheduler.tabSchedules))
	}
}

// TestRefreshScheduling tests the scheduling logic
func TestRefreshScheduling(t *testing.T) {
	scheduler := NewRefreshScheduler()
	scheduler.minInterval = 5 * time.Millisecond // Very short for testing

	// Add tab with very short interval for testing
	scheduler.AddTab("quick-tab", 10*time.Millisecond, RefreshPriorityHigh)

	// Manually set the next refresh time to be very soon to override staggering
	scheduler.mu.Lock()
	schedule := scheduler.tabSchedules["quick-tab"]
	schedule.NextRefresh = time.Now().Add(5 * time.Millisecond)
	scheduler.mu.Unlock()

	// Initially should not need refresh (just added)
	if scheduler.ShouldRefreshTab("quick-tab") {
		t.Error("Expected tab to not need refresh immediately after adding")
	}

	// Wait for the refresh time to pass
	time.Sleep(50 * time.Millisecond)

	// Now should need refresh
	if !scheduler.ShouldRefreshTab("quick-tab") {
		t.Error("Expected tab to need refresh after interval")
	}

	// Mark refresh as started
	scheduler.MarkRefreshStarted("quick-tab")

	// Should not need refresh immediately after marking as started
	if scheduler.ShouldRefreshTab("quick-tab") {
		t.Error("Expected tab to not need refresh immediately after marking started")
	}

	// Test non-existent tab
	if scheduler.ShouldRefreshTab("non-existent") {
		t.Error("Expected non-existent tab to not need refresh")
	}
}

// TestRefreshOrder tests the optimal refresh ordering
func TestRefreshOrder(t *testing.T) {
	scheduler := NewRefreshScheduler()

	// Add tabs with different priorities and request counts
	scheduler.AddTab("urgent-tab", 1*time.Minute, RefreshPriorityUrgent)
	scheduler.AddTab("normal-tab", 5*time.Minute, RefreshPriorityNormal)
	scheduler.AddTab("low-tab", 10*time.Minute, RefreshPriorityLow)

	// Set all tabs to need refresh (simulate time passing)
	now := time.Now()
	for _, schedule := range scheduler.tabSchedules {
		schedule.NextRefresh = now.Add(-1 * time.Minute) // In the past
	}

	order := scheduler.GetOptimalRefreshOrder()

	if len(order) != 3 {
		t.Errorf("Expected 3 tabs in refresh order, got %d", len(order))
	}

	// Urgent should be first
	if order[0] != "urgent-tab" {
		t.Errorf("Expected urgent-tab first in order, got '%s'", order[0])
	}

	// Should contain all tabs
	tabsFound := make(map[string]bool)
	for _, tabName := range order {
		tabsFound[tabName] = true
	}

	expectedTabs := []string{"urgent-tab", "normal-tab", "low-tab"}
	for _, expectedTab := range expectedTabs {
		if !tabsFound[expectedTab] {
			t.Errorf("Expected tab '%s' in refresh order", expectedTab)
		}
	}
}

// TestStaggeredInitialRefresh tests that initial refreshes are staggered
func TestStaggeredInitialRefresh(t *testing.T) {
	scheduler := NewRefreshScheduler()

	// Add multiple tabs quickly
	baseTime := time.Now()
	scheduler.AddTab("tab1", 5*time.Minute, RefreshPriorityNormal)
	scheduler.AddTab("tab2", 5*time.Minute, RefreshPriorityNormal)
	scheduler.AddTab("tab3", 5*time.Minute, RefreshPriorityNormal)

	// Get next refresh times
	refreshTimes := scheduler.GetNextRefreshTimes()

	if len(refreshTimes) != 3 {
		t.Fatalf("Expected 3 refresh times, got %d", len(refreshTimes))
	}

	// All refresh times should be different (staggered)
	times := []time.Time{refreshTimes["tab1"], refreshTimes["tab2"], refreshTimes["tab3"]}
	
	for i := 0; i < len(times); i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i].Equal(times[j]) {
				t.Error("Expected staggered refresh times, but found identical times")
			}
		}
	}

	// All should be in the near future (within 60 seconds of now)
	for tabName, refreshTime := range refreshTimes {
		if refreshTime.Before(baseTime) {
			t.Errorf("Tab %s refresh time is in the past", tabName)
		}
		if refreshTime.After(baseTime.Add(60 * time.Second)) {
			t.Errorf("Tab %s refresh time is too far in the future", tabName)
		}
	}
}

// TestRefreshIntervalAdjustment tests dynamic interval adjustment
func TestRefreshIntervalAdjustment(t *testing.T) {
	// Initialize global limiter for this test
	if GlobalLimiter == nil {
		InitGlobalRateLimiter()
	}
	
	scheduler := NewRefreshScheduler()
	scheduler.AddTab("test-tab", 2*time.Minute, RefreshPriorityNormal)

	originalInterval := scheduler.tabSchedules["test-tab"].RefreshInterval

	// Simulate low rate limit remaining
	GlobalLimiter.UpdateFromGitHubHeaders(500, time.Now().Add(time.Hour)) // Low remaining

	scheduler.AdjustRefreshIntervals()

	newInterval := scheduler.tabSchedules["test-tab"].RefreshInterval

	// Should have increased the interval
	if newInterval <= originalInterval {
		t.Errorf("Expected interval to increase from %v, got %v", originalInterval, newInterval)
	}

	// Simulate high rate limit remaining
	GlobalLimiter.UpdateFromGitHubHeaders(4000, time.Now().Add(time.Hour)) // High remaining

	scheduler.AdjustRefreshIntervals()

	adjustedInterval := scheduler.tabSchedules["test-tab"].RefreshInterval

	// Should have decreased the interval (but not below minimum)
	if adjustedInterval > newInterval {
		t.Errorf("Expected interval to decrease from %v, got %v", newInterval, adjustedInterval)
	}
}

// TestRateLimitSummary tests the rate limit summary functionality
func TestRateLimitSummary(t *testing.T) {
	// Initialize global limiter for this test
	if GlobalLimiter == nil {
		InitGlobalRateLimiter()
	}
	
	scheduler := NewRefreshScheduler()
	scheduler.AddTab("tab1", 5*time.Minute, RefreshPriorityNormal)
	scheduler.AddTab("tab2", 3*time.Minute, RefreshPriorityHigh)

	summary := scheduler.GetRateLimitSummary()

	if summary.TabCount != 2 {
		t.Errorf("Expected tab count 2, got %d", summary.TabCount)
	}

	// Should have some requests remaining (default 5000)
	if summary.RequestsRemaining <= 0 {
		t.Errorf("Expected positive requests remaining, got %d", summary.RequestsRemaining)
	}

	// Reset time should be in the future
	if summary.ResetTime.Before(time.Now()) {
		t.Error("Expected reset time to be in the future")
	}

	// Active requests should be 0 initially
	if summary.ActiveRequests < 0 {
		t.Errorf("Expected non-negative active requests, got %d", summary.ActiveRequests)
	}
}

// TestMinimumInterval tests minimum interval enforcement
func TestMinimumInterval(t *testing.T) {
	scheduler := NewRefreshScheduler()
	scheduler.minInterval = 100 * time.Millisecond // Short for testing

	scheduler.AddTab("test-tab", 10*time.Millisecond, RefreshPriorityNormal)

	// Force next refresh to be ready
	schedule := scheduler.tabSchedules["test-tab"]
	schedule.NextRefresh = time.Now().Add(-1 * time.Minute)

	// Should need refresh
	if !scheduler.ShouldRefreshTab("test-tab") {
		t.Error("Expected tab to need refresh")
	}

	// Mark as started
	scheduler.MarkRefreshStarted("test-tab")

	// Wait less than minimum interval
	time.Sleep(50 * time.Millisecond)

	// Force next refresh to be ready again
	schedule.NextRefresh = time.Now().Add(-1 * time.Minute)

	// Should NOT refresh due to minimum interval
	if scheduler.ShouldRefreshTab("test-tab") {
		t.Error("Expected tab to be blocked by minimum interval")
	}

	// Wait for minimum interval to pass
	time.Sleep(60 * time.Millisecond)

	// Now should be allowed to refresh
	if !scheduler.ShouldRefreshTab("test-tab") {
		t.Error("Expected tab to be allowed to refresh after minimum interval")
	}
}

// TestRequestEstimation tests request count estimation
func TestRequestEstimation(t *testing.T) {
	scheduler := NewRefreshScheduler()

	// Add tabs and check their estimated request counts
	scheduler.AddTab("small-tab", 5*time.Minute, RefreshPriorityNormal)
	scheduler.AddTab("large-tab", 2*time.Minute, RefreshPriorityHigh)

	smallSchedule := scheduler.tabSchedules["small-tab"]
	largeSchedule := scheduler.tabSchedules["large-tab"]

	// Both should have reasonable request estimates
	if smallSchedule.RequestCount <= 0 {
		t.Errorf("Expected positive request count for small tab, got %d", smallSchedule.RequestCount)
	}
	if largeSchedule.RequestCount <= 0 {
		t.Errorf("Expected positive request count for large tab, got %d", largeSchedule.RequestCount)
	}

	// Request counts should be reasonable (not too high or too low)
	if smallSchedule.RequestCount > 100 {
		t.Errorf("Expected reasonable request count for small tab, got %d", smallSchedule.RequestCount)
	}
	if largeSchedule.RequestCount > 100 {
		t.Errorf("Expected reasonable request count for large tab, got %d", largeSchedule.RequestCount)
	}
}

// TestSchedulerEdgeCases tests edge cases and error conditions
func TestSchedulerEdgeCases(t *testing.T) {
	scheduler := NewRefreshScheduler()

	// Test operations on non-existent tabs
	if scheduler.ShouldRefreshTab("non-existent") {
		t.Error("Expected false for non-existent tab")
	}

	scheduler.MarkRefreshStarted("non-existent") // Should not panic

	scheduler.RemoveTab("non-existent") // Should not panic

	// Test with empty scheduler
	order := scheduler.GetOptimalRefreshOrder()
	if len(order) != 0 {
		t.Errorf("Expected empty refresh order for empty scheduler, got %d items", len(order))
	}

	times := scheduler.GetNextRefreshTimes()
	if len(times) != 0 {
		t.Errorf("Expected empty refresh times for empty scheduler, got %d items", len(times))
	}

	summary := scheduler.GetRateLimitSummary()
	if summary.TabCount != 0 {
		t.Errorf("Expected tab count 0 for empty scheduler, got %d", summary.TabCount)
	}
}

// TestSchedulerConcurrency tests concurrent access to scheduler
func TestSchedulerConcurrency(t *testing.T) {
	scheduler := NewRefreshScheduler()

	// Add initial tab
	scheduler.AddTab("concurrent-test", 1*time.Minute, RefreshPriorityNormal)

	// Start multiple goroutines doing different operations
	done := make(chan bool)
	
	// Goroutine 1: Adding and removing tabs
	go func() {
		for i := 0; i < 10; i++ {
			tabName := "temp-tab-" + string(rune('0'+i))
			scheduler.AddTab(tabName, time.Duration(i+1)*time.Minute, RefreshPriorityNormal)
			time.Sleep(1 * time.Millisecond)
			scheduler.RemoveTab(tabName)
		}
		done <- true
	}()

	// Goroutine 2: Checking refresh status
	go func() {
		for i := 0; i < 50; i++ {
			scheduler.ShouldRefreshTab("concurrent-test")
			scheduler.GetOptimalRefreshOrder()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: Marking refreshes
	go func() {
		for i := 0; i < 20; i++ {
			scheduler.MarkRefreshStarted("concurrent-test")
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent test took too long")
		}
	}

	// Scheduler should still be functional - just verify it doesn't panic
	_ = scheduler.ShouldRefreshTab("concurrent-test")
}
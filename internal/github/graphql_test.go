package github

import (
	"context"
	"testing"
	"time"
)

func TestNewGraphQLClient(t *testing.T) {
	token := "test-token"
	client := NewGraphQLClient(token)

	if client == nil {
		t.Fatal("NewGraphQLClient() returned nil")
	}

	if client.token != token {
		t.Errorf("Expected token %s, got %s", token, client.token)
	}

	if client.endpoint != "https://api.github.com/graphql" {
		t.Errorf("Expected endpoint https://api.github.com/graphql, got %s", client.endpoint)
	}

	rateLimit := client.GetRateLimit()
	if rateLimit.Limit != 5000 {
		t.Errorf("Expected default rate limit 5000, got %d", rateLimit.Limit)
	}
}

func TestRateLimiter(t *testing.T) {
	rateLimiter := NewRateLimiter()

	if rateLimiter == nil {
		t.Fatal("NewRateLimiter() returned nil")
	}

	// Test basic rate limiting
	ctx := context.Background()
	
	// Normal rate limit should have minimal delay
	rateLimit := &RateLimitInfo{
		Limit:     5000,
		Remaining: 4000,
		ResetAt:   time.Now().Add(time.Hour),
	}

	start := time.Now()
	err := rateLimiter.Wait(ctx, rateLimit)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Wait() failed: %v", err)
	}

	// Should be quick for normal rate limits
	if duration > 200*time.Millisecond {
		t.Errorf("Expected quick response for normal rate limit, took %v", duration)
	}
}

func TestRateLimiterBackoff(t *testing.T) {
	rateLimiter := NewRateLimiter()
	ctx := context.Background()

	// Make a first request to set lastRequest time
	normalRateLimit := &RateLimitInfo{
		Limit:     5000,
		Remaining: 4000,
		ResetAt:   time.Now().Add(time.Hour),
	}
	rateLimiter.Wait(ctx, normalRateLimit)

	// Wait a bit so the backoff logic kicks in properly
	time.Sleep(10 * time.Millisecond)

	// Test backoff with low remaining rate limit
	lowRateLimit := &RateLimitInfo{
		Limit:     5000,
		Remaining: 50, // Should trigger backoff
		ResetAt:   time.Now().Add(time.Hour),
	}

	start := time.Now()
	err := rateLimiter.Wait(ctx, lowRateLimit)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Wait() with low rate limit failed: %v", err)
	}

	// Should have some backoff delay (starts at minDelay, grows to 1s)  
	if duration < 150*time.Millisecond {
		t.Errorf("Expected some backoff for low rate limit (>150ms), took %v", duration)
	}
}

func TestRateLimiterCritical(t *testing.T) {
	rateLimiter := NewRateLimiter()
	
	// Use context with timeout to avoid long waits in tests
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test critical rate limit (should wait for reset, but we'll cancel via context)
	criticalRateLimit := &RateLimitInfo{
		Limit:     5000,
		Remaining: 5, // Critical level
		ResetAt:   time.Now().Add(10 * time.Second), // Long reset time
	}

	start := time.Now()
	err := rateLimiter.Wait(ctx, criticalRateLimit)
	duration := time.Since(start)

	// Should get context cancelled error due to timeout
	if err == nil {
		t.Error("Expected context cancelled error for critical rate limit")
	}

	// Should wait approximately the context timeout duration
	if duration < 80*time.Millisecond || duration > 150*time.Millisecond {
		t.Errorf("Expected ~100ms wait (context timeout), took %v", duration)
	}
}

func TestConvertPRDataToGitHubPR(t *testing.T) {
	// Test data conversion
	prData := PRData{
		Number:    123,
		Title:     "Test PR",
		Body:      "Test body",
		State:     "OPEN",
		IsDraft:   false,
		Mergeable: "MERGEABLE",
		URL:       "https://github.com/owner/repo/pull/123",
		Author: &AuthorData{
			Login:     "test-user",
			AvatarUrl: "https://github.com/test-user.avatar",
		},
		Repository: &RepositoryData{
			Name: "test-repo",
			Owner: &AuthorData{
				Login: "test-owner",
			},
			NameWithOwner: "test-owner/test-repo",
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
		Comments: &CommentConnection{
			TotalCount: 5,
		},
		LatestReviews: &ReviewConnection{
			TotalCount: 2,
		},
	}

	pr := convertPRDataToGitHubPR(prData)

	// Verify conversion
	if pr.GetNumber() != 123 {
		t.Errorf("Expected number 123, got %d", pr.GetNumber())
	}

	if pr.GetTitle() != "Test PR" {
		t.Errorf("Expected title 'Test PR', got %s", pr.GetTitle())
	}

	if pr.GetState() != "open" {
		t.Errorf("Expected state 'open', got %s", pr.GetState())
	}

	if pr.GetDraft() != false {
		t.Errorf("Expected draft false, got %v", pr.GetDraft())
	}

	if pr.GetMergeable() != true {
		t.Errorf("Expected mergeable true, got %v", pr.GetMergeable())
	}

	if pr.GetUser().GetLogin() != "test-user" {
		t.Errorf("Expected author 'test-user', got %s", pr.GetUser().GetLogin())
	}

	if pr.GetBase().GetRepo().GetName() != "test-repo" {
		t.Errorf("Expected repo name 'test-repo', got %s", pr.GetBase().GetRepo().GetName())
	}

	if pr.GetComments() != 5 {
		t.Errorf("Expected 5 comments, got %d", pr.GetComments())
	}

	if pr.GetReviewComments() != 2 {
		t.Errorf("Expected 2 review comments, got %d", pr.GetReviewComments())
	}
}

package services

import (
	"context"
	"testing"
	"time"

	"github.com/bjess9/pr-compass/internal/config"
	gh "github.com/google/go-github/v55/github"
)

func TestNewPRService(t *testing.T) {
	token := "test-token"
	service := NewPRService(token, nil)
	
	if service == nil {
		t.Fatal("NewPRService returned nil")
	}
	
	prService, ok := service.(*prService)
	if !ok {
		t.Fatal("NewPRService did not return a prService instance")
	}
	
	if prService.token != token {
		t.Errorf("Expected token %s, got %s", token, prService.token)
	}
}

func TestPRService_FetchPRs_RepoMode(t *testing.T) {
	service := NewPRService("fake-token", nil)
	
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/repo"},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// This will fail because we don't have a real token/repos, but we're testing the flow
	prs, err := service.FetchPRs(ctx, cfg)
	
	// With fake token, this should either fail or return empty results
	if err == nil {
		// If no error, check that we got empty results (which is expected when all repos fail)
		if len(prs) > 0 {
			t.Error("Expected empty PRs due to fake token, but got some PRs")
		}
	}
}

func TestPRService_convertAndSort(t *testing.T) {
	service := &prService{
		token: "test-token",
		cache: nil,
	}
	
	now := time.Now()
	older := now.Add(-1 * time.Hour)
	newest := now.Add(1 * time.Hour)
	
	// Create test PRs with different update times
	pr1 := &gh.PullRequest{
		Number:    gh.Int(1),
		UpdatedAt: &gh.Timestamp{Time: older},
		Title:     gh.String("Older PR"),
	}
	pr2 := &gh.PullRequest{
		Number:    gh.Int(2),
		UpdatedAt: &gh.Timestamp{Time: newest},
		Title:     gh.String("Newest PR"),
	}
	pr3 := &gh.PullRequest{
		Number:    gh.Int(3),
		UpdatedAt: &gh.Timestamp{Time: now},
		Title:     gh.String("Middle PR"),
	}
	
	ghPRs := []*gh.PullRequest{pr1, pr2, pr3}
	result := service.convertAndSort(ghPRs)
	
	// Check that we got the right number of results
	if len(result) != 3 {
		t.Fatalf("Expected 3 PRs, got %d", len(result))
	}
	
	// Check that they're sorted by newest first
	if result[0].GetNumber() != 2 {
		t.Errorf("Expected first PR to be #2 (newest), got #%d", result[0].GetNumber())
	}
	if result[1].GetNumber() != 3 {
		t.Errorf("Expected second PR to be #3 (middle), got #%d", result[1].GetNumber())
	}
	if result[2].GetNumber() != 1 {
		t.Errorf("Expected third PR to be #1 (oldest), got #%d", result[2].GetNumber())
	}
	
	// Check that Enhanced field is initialized as nil
	for i, pr := range result {
		if pr.Enhanced != nil {
			t.Errorf("Expected Enhanced to be nil for PR %d, got %+v", i, pr.Enhanced)
		}
	}
}

func TestPRService_RefreshPRs(t *testing.T) {
	service := NewPRService("fake-token", nil)
	
	cfg := &config.Config{
		Mode:  "repos",
		Repos: []string{"owner/repo"},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// RefreshPRs should behave the same as FetchPRs for now
	prs, err := service.RefreshPRs(ctx, cfg)
	
	// With fake token, this should either fail or return empty results
	if err == nil {
		// If no error, check that we got empty results (which is expected when all repos fail)
		if len(prs) > 0 {
			t.Error("Expected empty PRs due to fake token, but got some PRs")
		}
	}
}
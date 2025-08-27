package services

import (
	"testing"

	"github.com/bjess9/pr-compass/internal/cache"
	"github.com/bjess9/pr-compass/internal/ui/types"
)

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		cache   *cache.PRCache
		wantNil bool
	}{
		{
			name:    "with valid token and cache",
			token:   "valid-token",
			cache:   func() *cache.PRCache { c, _ := cache.NewPRCache(); return c }(),
			wantNil: false,
		},
		{
			name:    "with valid token and nil cache",
			token:   "valid-token",
			cache:   nil,
			wantNil: false,
		},
		{
			name:    "with empty token",
			token:   "",
			cache:   func() *cache.PRCache { c, _ := cache.NewPRCache(); return c }(),
			wantNil: false,
		},
		{
			name:    "with nil cache",
			token:   "token",
			cache:   nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry(tt.token, tt.cache)

			if (registry == nil) != tt.wantNil {
				t.Errorf("NewRegistry() = nil: %v, want nil: %v", registry == nil, tt.wantNil)
				return
			}

			if registry != nil {
				// Verify all services are initialized
				if registry.PR == nil {
					t.Error("PR service not initialized")
				}
				if registry.Enhancement == nil {
					t.Error("Enhancement service not initialized")
				}
				if registry.State == nil {
					t.Error("State service not initialized")
				}
				if registry.Filter == nil {
					t.Error("Filter service not initialized")
				}
			}
		})
	}
}

func TestRegistry_ServiceInterfaces(t *testing.T) {
	registry := NewRegistry("test-token", nil)

	// Test that all services implement their interfaces
	var _ PRService = registry.PR
	var _ EnhancementService = registry.Enhancement
	var _ StateService = registry.State
	var _ FilterService = registry.Filter

	// Services should be properly isolated (different interfaces, so we verify they're initialized)
	if registry.PR == nil {
		t.Error("PR service should be initialized")
	}
	if registry.Enhancement == nil {
		t.Error("Enhancement service should be initialized")
	}
	if registry.State == nil {
		t.Error("State service should be initialized")
	}
	if registry.Filter == nil {
		t.Error("Filter service should be initialized")
	}
}

func TestRegistry_ServiceInitialization(t *testing.T) {
	token := "test-token"
	prCache, _ := cache.NewPRCache()
	
	registry := NewRegistry(token, prCache)

	// Test that services are properly initialized and functional
	// These are basic smoke tests to ensure initialization worked

	// Test State service
	state := registry.State.GetState()
	if state == nil {
		t.Error("State service GetState() returned nil")
	}

	// Test Filter service validation
	err := registry.Filter.ValidateFilter(types.FilterOptions{Mode: "author", Value: "test"})
	if err != nil {
		t.Errorf("Filter service ValidateFilter() failed: %v", err)
	}

	// Test that services maintain independence
	// State service should not be affected by other services
	initialState := registry.State.GetState()
	registry.State.UpdatePRs([]*types.PRData{})
	updatedState := registry.State.GetState()
	
	if initialState == updatedState {
		t.Error("State service should return different instances on GetState()")
	}
}

func TestRegistry_MultipleInstances(t *testing.T) {
	// Test that creating multiple registries creates independent instances
	registry1 := NewRegistry("token1", nil)
	registry2 := NewRegistry("token2", nil)

	if registry1 == registry2 {
		t.Error("Multiple registry instances should not be the same")
	}

	// Test that services are independently initialized per registry
	// Since we can't directly compare different interface types, we test functionality
	
	// Modify one state and ensure it doesn't affect the other
	registry1.State.UpdateStatusMessage("test1")
	registry2.State.UpdateStatusMessage("test2")
	
	state1After := registry1.State.GetState()
	state2After := registry2.State.GetState()
	
	if state1After.UI.StatusMsg == state2After.UI.StatusMsg {
		t.Error("Different registries should maintain independent state")
	}
}
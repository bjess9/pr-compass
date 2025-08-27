package services

import (
	"github.com/bjess9/pr-compass/internal/cache"
)

// Registry provides centralized access to all services
type Registry struct {
	PR          PRService
	Enhancement EnhancementService
	State       StateService
	Filter      FilterService
}

// NewRegistry creates a new service registry with all services initialized
func NewRegistry(token string, cache *cache.PRCache) *Registry {
	// Initialize state service first
	stateService := NewStateService()

	// Initialize other services
	prService := NewPRService(token, cache)
	enhancementService := NewEnhancementService(token)
	filterService := NewFilterService()

	return &Registry{
		PR:          prService,
		Enhancement: enhancementService,
		State:       stateService,
		Filter:      filterService,
	}
}

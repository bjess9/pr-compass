package services

import (
	"github.com/bjess9/pr-compass/internal/cache"
)

// Registry holds all services for the application
type Registry struct {
	PR          PRService
	Enhancement EnhancementService
	Filter      FilterService
	State       StateService
}

// NewRegistry creates a new service registry
func NewRegistry(token string, prCache *cache.PRCache) *Registry {
	return &Registry{
		PR:          NewPRService(token, prCache),
		Enhancement: NewEnhancementService(token),
		Filter:      NewFilterService(),
		State:       NewStateService(),
	}
}
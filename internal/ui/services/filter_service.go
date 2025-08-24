package services

import (
	"errors"
	"sort"
	"strings"

	"github.com/bjess9/pr-compass/internal/ui/types"
)

// filterService implements the FilterService interface
type filterService struct{}

// NewFilterService creates a new filter service
func NewFilterService() FilterService {
	return &filterService{}
}

// ApplyFilter filters PRs based on the given options
func (s *filterService) ApplyFilter(prs []*types.PRData, filter types.FilterOptions) []*types.PRData {
	if !filter.Active || filter.Mode == "" {
		return prs
	}

	filtered := []*types.PRData{}

	switch filter.Mode {
	case "author":
		for _, pr := range prs {
			if pr.GetUser() != nil && pr.GetUser().GetLogin() == filter.Value {
				filtered = append(filtered, pr)
			}
		}

	case "repo":
		for _, pr := range prs {
			if pr.GetBase() != nil && pr.GetBase().GetRepo() != nil {
				repoName := pr.GetBase().GetRepo().GetName()
				repoFullName := pr.GetBase().GetRepo().GetFullName()
				if repoName == filter.Value || repoFullName == filter.Value {
					filtered = append(filtered, pr)
				}
			}
		}

	case "status":
		switch filter.Value {
		case "ready":
			// Show only PRs that are ready (not drafts, and mergeable)
			for _, pr := range prs {
				if !pr.GetDraft() && pr.GetMergeable() {
					filtered = append(filtered, pr)
				}
			}
		case "draft":
			// Show only draft PRs
			for _, pr := range prs {
				if pr.GetDraft() {
					filtered = append(filtered, pr)
				}
			}
		case "conflicts":
			// Show only PRs with merge conflicts
			for _, pr := range prs {
				if pr.GetMergeableState() == "dirty" {
					filtered = append(filtered, pr)
				}
			}
		case "blocked":
			// Show only blocked PRs
			for _, pr := range prs {
				if pr.GetMergeableState() == "blocked" {
					filtered = append(filtered, pr)
				}
			}
		}

	case "draft":
		// Special case for draft filter (boolean)
		for _, pr := range prs {
			if pr.GetDraft() {
				filtered = append(filtered, pr)
			}
		}

	case "label":
		// Filter by label
		for _, pr := range prs {
			for _, label := range pr.Labels {
				if label.GetName() == filter.Value {
					filtered = append(filtered, pr)
					break
				}
			}
		}

	case "search":
		// Search in title and body
		searchTerm := strings.ToLower(filter.Value)
		for _, pr := range prs {
			title := strings.ToLower(pr.GetTitle())
			body := strings.ToLower(pr.GetBody())
			if strings.Contains(title, searchTerm) || strings.Contains(body, searchTerm) {
				filtered = append(filtered, pr)
			}
		}

	default:
		// Unknown filter mode, return all PRs
		return prs
	}

	return filtered
}

// GetFilterSuggestions returns suggested filter values for a given mode
func (s *filterService) GetFilterSuggestions(prs []*types.PRData, mode string) []string {
	suggestions := []string{}
	seen := make(map[string]bool)

	switch mode {
	case "author":
		for _, pr := range prs {
			if pr.GetUser() != nil {
				author := pr.GetUser().GetLogin()
				if author != "" && !seen[author] {
					suggestions = append(suggestions, author)
					seen[author] = true
				}
			}
		}

	case "repo":
		for _, pr := range prs {
			if pr.GetBase() != nil && pr.GetBase().GetRepo() != nil {
				repoName := pr.GetBase().GetRepo().GetName()
				repoFullName := pr.GetBase().GetRepo().GetFullName()
				if repoName != "" && !seen[repoName] {
					suggestions = append(suggestions, repoName)
					seen[repoName] = true
				}
				if repoFullName != "" && !seen[repoFullName] {
					suggestions = append(suggestions, repoFullName)
					seen[repoFullName] = true
				}
			}
		}

	case "status":
		// Predefined status options
		suggestions = []string{"ready", "draft", "conflicts", "blocked"}

	case "label":
		for _, pr := range prs {
			for _, label := range pr.Labels {
				labelName := label.GetName()
				if labelName != "" && !seen[labelName] {
					suggestions = append(suggestions, labelName)
					seen[labelName] = true
				}
			}
		}

	default:
		// No suggestions for unknown modes
		return []string{}
	}

	// Sort suggestions alphabetically
	sort.Strings(suggestions)
	return suggestions
}

// ValidateFilter checks if a filter configuration is valid
func (s *filterService) ValidateFilter(filter types.FilterOptions) error {
	if !filter.Active {
		return nil // Inactive filters are always valid
	}

	if filter.Mode == "" {
		return errors.New("filter mode cannot be empty when active")
	}

	// Validate known filter modes
	validModes := map[string]bool{
		"author": true,
		"repo":   true,
		"status": true,
		"draft":  true,
		"label":  true,
		"search": true,
	}

	if !validModes[filter.Mode] {
		return errors.New("unknown filter mode: " + filter.Mode)
	}

	// Validate status filter values
	if filter.Mode == "status" {
		validStatuses := map[string]bool{
			"ready":     true,
			"draft":     true,
			"conflicts": true,
			"blocked":   true,
		}
		if !validStatuses[filter.Value] {
			return errors.New("invalid status filter value: " + filter.Value)
		}
	}

	// For other modes, any non-empty value is valid when active
	if filter.Mode != "draft" && filter.Value == "" {
		return errors.New("filter value cannot be empty for mode: " + filter.Mode)
	}

	return nil
}

// GetActiveFiltersDescription returns a human-readable description of active filters
func (s *filterService) GetActiveFiltersDescription(filter types.FilterOptions) string {
	if !filter.Active || filter.Mode == "" {
		return ""
	}

	switch filter.Mode {
	case "author":
		return "by author: " + filter.Value
	case "repo":
		return "by repository: " + filter.Value
	case "status":
		return "by status: " + filter.Value
	case "draft":
		return "drafts only"
	case "label":
		return "by label: " + filter.Value
	case "search":
		return "search: " + filter.Value
	default:
		return "filter: " + filter.Mode + "=" + filter.Value
	}
}
package services

import (
	"fmt"
	"strings"

	"github.com/bjess9/pr-compass/internal/ui/types"
)

// filterService implements the FilterService interface
type filterService struct{}

// NewFilterService creates a new filter service
func NewFilterService() FilterService {
	return &filterService{}
}

// FilterPRs applies filtering to a list of PRs
func (s *filterService) FilterPRs(prs []*types.PRData, filter types.FilterOptions) []*types.PRData {
	if filter.Mode == "" {
		return prs
	}

	var filtered []*types.PRData
	valueLower := strings.ToLower(filter.Value)

	for _, pr := range prs {
		include := false

		switch filter.Mode {
		case "author":
			author := ""
			if pr.GetUser() != nil {
				author = strings.ToLower(pr.GetUser().GetLogin())
			}
			include = strings.Contains(author, valueLower)

		case "status":
			status := "ready"
			if pr.GetDraft() {
				status = "draft"
			} else if pr.GetMergeableState() == "dirty" {
				status = "conflicts"
			}
			include = strings.Contains(status, valueLower)

		case "draft":
			include = pr.GetDraft() == (filter.Value == "true")

		case "title":
			title := strings.ToLower(pr.GetTitle())
			include = strings.Contains(title, valueLower)

		case "repo":
			repo := ""
			if pr.GetBase() != nil && pr.GetBase().GetRepo() != nil {
				repo = strings.ToLower(pr.GetBase().GetRepo().GetName())
			}
			include = strings.Contains(repo, valueLower)
		}

		if include {
			filtered = append(filtered, pr)
		}
	}

	return filtered
}

// ValidateFilter validates filter options
func (s *filterService) ValidateFilter(filter types.FilterOptions) error {
	validModes := map[string]bool{
		"author": true,
		"status": true,
		"draft":  true,
		"title":  true,
		"repo":   true,
	}

	if filter.Mode != "" && !validModes[filter.Mode] {
		return fmt.Errorf("invalid filter mode: %s", filter.Mode)
	}

	return nil
}

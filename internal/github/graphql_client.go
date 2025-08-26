package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GraphQLClient handles GitHub GraphQL API requests
type GraphQLClient struct {
	client     *http.Client
	token      string
	endpoint   string
	rateLimits *RateLimitInfo
}

// RateLimitInfo tracks GitHub API rate limits
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"resetAt"`
	Cost      int       `json:"cost"`
}

// NewGraphQLClient creates a new GraphQL client
func NewGraphQLClient(token string) *GraphQLClient {
	return &GraphQLClient{
		client:   &http.Client{Timeout: 30 * time.Second},
		token:    token,
		endpoint: "https://api.github.com/graphql",
		rateLimits: &RateLimitInfo{
			Limit:     5000, // Default GitHub rate limit
			Remaining: 5000,
			ResetAt:   time.Now().Add(time.Hour),
		},
	}
}

// GetRateLimit returns current rate limit information
func (gc *GraphQLClient) GetRateLimit() *RateLimitInfo {
	return gc.rateLimits
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data       json.RawMessage `json:"data"`
	Errors     []GraphQLError  `json:"errors,omitempty"`
	Extensions struct {
		Cost struct {
			RequestedQueryCost  int `json:"requestedQueryCost"`
			ActualQueryCost     int `json:"actualQueryCost"`
			MaximumAvailable    int `json:"maximumAvailable"`
			RemainingQuotaAfter int `json:"remainingQuotaAfter"`
		} `json:"cost"`
	} `json:"extensions"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Path       []interface{} `json:"path"`
	Extensions struct {
		Code      string `json:"code"`
		TypeName  string `json:"typeName"`
		FieldName string `json:"fieldName"`
	} `json:"extensions"`
}

// Execute performs a GraphQL request
func (gc *GraphQLClient) Execute(ctx context.Context, req GraphQLRequest) (*GraphQLResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", gc.endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+gc.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/vnd.github.v4+json")

	resp, err := gc.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request failed: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit info from headers
	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		_, _ = fmt.Sscanf(limit, "%d", &gc.rateLimits.Limit)
	}
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		_, _ = fmt.Sscanf(remaining, "%d", &gc.rateLimits.Remaining)
	}
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		var resetTimestamp int64
		if n, _ := fmt.Sscanf(reset, "%d", &resetTimestamp); n == 1 {
			gc.rateLimits.ResetAt = time.Unix(resetTimestamp, 0)
		}
	}

	var gqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode GraphQL response: %w", err)
	}

	// Update rate limit from GraphQL cost info if available
	if gqlResp.Extensions.Cost.RemainingQuotaAfter > 0 {
		gc.rateLimits.Remaining = gqlResp.Extensions.Cost.RemainingQuotaAfter
		gc.rateLimits.Cost = gqlResp.Extensions.Cost.ActualQueryCost
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp, fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	return &gqlResp, nil
}

// PRData represents comprehensive PR data from GraphQL
type PRData struct {
	Number            int                      `json:"number"`
	Title             string                   `json:"title"`
	Body              string                   `json:"body"`
	State             string                   `json:"state"`
	IsDraft           bool                     `json:"isDraft"`
	Mergeable         string                   `json:"mergeable"`
	Author            *AuthorData              `json:"author"`
	Repository        *RepositoryData          `json:"repository"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`
	URL               string                   `json:"url"`
	Labels            *LabelConnection         `json:"labels"`
	Comments          *CommentConnection       `json:"comments"`
	ReviewRequests    *ReviewRequestConnection `json:"reviewRequests"`
	LatestReviews     *ReviewConnection        `json:"latestReviews"`
	StatusCheckRollup *StatusCheckRollupData   `json:"statusCheckRollup"`
	Commits           *CommitConnection        `json:"commits"`
}

// AuthorData represents PR author information
type AuthorData struct {
	Login     string `json:"login"`
	AvatarUrl string `json:"avatarUrl"`
}

// RepositoryData represents repository information
type RepositoryData struct {
	Name          string      `json:"name"`
	Owner         *AuthorData `json:"owner"`
	NameWithOwner string      `json:"nameWithOwner"`
}

// LabelConnection represents the labels connection
type LabelConnection struct {
	Nodes      []LabelData `json:"nodes"`
	TotalCount int         `json:"totalCount"`
}

// LabelData represents a label
type LabelData struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CommentConnection represents comments connection
type CommentConnection struct {
	TotalCount int `json:"totalCount"`
}

// ReviewRequestConnection represents review requests
type ReviewRequestConnection struct {
	TotalCount int `json:"totalCount"`
}

// ReviewConnection represents reviews connection
type ReviewConnection struct {
	Nodes      []ReviewData `json:"nodes"`
	TotalCount int          `json:"totalCount"`
}

// ReviewData represents review information
type ReviewData struct {
	State       string      `json:"state"`
	Author      *AuthorData `json:"author"`
	SubmittedAt time.Time   `json:"submittedAt"`
}

// StatusCheckRollupData represents CI/check status
type StatusCheckRollupData struct {
	State    string                   `json:"state"`
	Contexts *StatusContextConnection `json:"contexts"`
}

// StatusContextConnection represents status contexts
type StatusContextConnection struct {
	Nodes      []StatusContextData `json:"nodes"`
	TotalCount int                 `json:"totalCount"`
}

// StatusContextData represents a status check context
type StatusContextData struct {
	State       string `json:"state"`
	Context     string `json:"context"`
	Description string `json:"description"`
}

// CommitConnection represents commits connection
type CommitConnection struct {
	Nodes      []CommitData `json:"nodes"`
	TotalCount int          `json:"totalCount"`
}

// CommitData represents commit information
type CommitData struct {
	Commit struct {
		Additions    int `json:"additions"`
		Deletions    int `json:"deletions"`
		ChangedFiles int `json:"changedFiles"`
	} `json:"commit"`
}

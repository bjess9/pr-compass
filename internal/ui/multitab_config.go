package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bjess9/pr-compass/internal/config"
	"github.com/bjess9/pr-compass/internal/errors"
	"github.com/spf13/viper"
)

// MultiTabConfig represents a multi-tab configuration
type MultiTabConfig struct {
	// Global settings
	RefreshIntervalMinutes int `mapstructure:"refresh_interval_minutes" yaml:"refresh_interval_minutes,omitempty"`

	// Tab definitions
	Tabs []TabConfig `mapstructure:"tabs" yaml:"tabs"`
}

// LoadMultiTabConfig loads multi-tab configuration from file
func LoadMultiTabConfig() (*MultiTabConfig, error) {
	return LoadMultiTabConfigFromPath(getConfigFilePath())
}

// LoadMultiTabConfigFromPath loads multi-tab configuration from a specific path
func LoadMultiTabConfigFromPath(configPath string) (*MultiTabConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, errors.NewConfigNotFoundError(configPath)
	}

	// First, try to load as multi-tab configuration
	var multiConfig MultiTabConfig
	if err := v.Unmarshal(&multiConfig); err == nil && len(multiConfig.Tabs) > 0 {
		// Validate and set defaults for multi-tab config
		for i := range multiConfig.Tabs {
			tab := &multiConfig.Tabs[i]

			// Set default name if not provided
			if tab.Name == "" {
				tab.Name = fmt.Sprintf("Tab %d", i+1)
			}

			// Set default refresh interval if not provided
			if tab.RefreshIntervalMinutes == 0 {
				if multiConfig.RefreshIntervalMinutes > 0 {
					tab.RefreshIntervalMinutes = multiConfig.RefreshIntervalMinutes
				} else {
					tab.RefreshIntervalMinutes = 5
				}
			}

			// Set default include_drafts if not specified
			if !v.IsSet(fmt.Sprintf("tabs.%d.include_drafts", i)) {
				tab.IncludeDrafts = true
			}

			// Set default max_prs if not specified
			if tab.MaxPRs == 0 {
				tab.MaxPRs = 50 // Conservative default for multi-tab
			}
		}

		return &multiConfig, nil
	}

	// Fall back to single-tab (legacy) configuration
	var legacyConfig config.Config
	if err := v.Unmarshal(&legacyConfig); err != nil {
		return nil, errors.NewConfigInvalidError(err)
	}

	// Convert legacy config to multi-tab format
	tabConfig := TabConfig{
		Name:                   "Main", // Default name for legacy tab
		Mode:                   legacyConfig.Mode,
		Repos:                  legacyConfig.Repos,
		Organization:           legacyConfig.Organization,
		Teams:                  legacyConfig.Teams,
		SearchQuery:            legacyConfig.SearchQuery,
		Topics:                 legacyConfig.Topics,
		TopicOrg:               legacyConfig.TopicOrg,
		ExcludeBots:            legacyConfig.ExcludeBots,
		ExcludeAuthors:         legacyConfig.ExcludeAuthors,
		ExcludeTitles:          legacyConfig.ExcludeTitles,
		IncludeDrafts:          legacyConfig.IncludeDrafts,
		RefreshIntervalMinutes: legacyConfig.RefreshIntervalMinutes,
	}

	// Auto-detect mode if not set (for backward compatibility)
	if tabConfig.Mode == "" {
		if len(tabConfig.Repos) > 0 {
			tabConfig.Mode = "repos"
		} else if len(tabConfig.Topics) > 0 {
			tabConfig.Mode = "topics"
		} else if tabConfig.Organization != "" {
			if len(tabConfig.Teams) > 0 {
				tabConfig.Mode = "teams"
			} else {
				tabConfig.Mode = "organization"
			}
		} else if tabConfig.SearchQuery != "" {
			tabConfig.Mode = "search"
		} else {
			tabConfig.Mode = "repos" // default
		}
	}

	// Set default refresh interval if not configured
	if tabConfig.RefreshIntervalMinutes == 0 {
		tabConfig.RefreshIntervalMinutes = 5 // default: 5 minutes
	}

	// Set default max PRs if not configured
	if tabConfig.MaxPRs == 0 {
		tabConfig.MaxPRs = 50 // default: 50 PRs
	}

	multiConfig = MultiTabConfig{
		RefreshIntervalMinutes: tabConfig.RefreshIntervalMinutes,
		Tabs:                   []TabConfig{tabConfig},
	}

	return &multiConfig, nil
}

// IsMultiTab checks if the configuration file contains multiple tabs
func IsMultiTab(configPath string) bool {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return false
	}

	// Check if "tabs" key exists and has multiple entries
	if v.IsSet("tabs") {
		tabs := v.Get("tabs")
		if tabSlice, ok := tabs.([]interface{}); ok {
			return len(tabSlice) > 1
		}
	}

	return false
}

// getConfigFilePath returns the path to the configuration file
func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Unable to determine user home directory: %v\n", err)
		os.Exit(1)
	}
	return fmt.Sprintf("%s/.prcompass_config.yaml", homeDir)
}

// CreateExampleMultiTabConfig creates an example multi-tab configuration file
func CreateExampleMultiTabConfig(path string) error {
	// Validate path to prevent file inclusion vulnerabilities
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Basic path validation - ensure it ends with .yaml extension and doesn't contain suspicious patterns
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		return fmt.Errorf("config file must have .yaml or .yml extension")
	}

	// Prevent directory traversal attacks
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path cannot contain directory traversal sequences")
	}
	exampleConfig := `# PR Compass Multi-Tab Configuration
# Save this as ~/.prcompass_config.yaml

# Global settings (applied to all tabs unless overridden)
refresh_interval_minutes: 5

# Tab definitions - each tab can monitor different repositories/organizations
tabs:
  # Tab 1: Monitor your team's repositories by topic
  - name: "My Team"
    mode: "topics"
    topics:
      - "my-team"
      - "backend"
    topic_org: "your-org-name"
    exclude_bots: true
    include_drafts: true
    refresh_interval_minutes: 3  # Faster refresh for your team

  # Tab 2: Monitor specific important repositories
  - name: "Critical Repos"
    mode: "repos"
    repos:
      - "your-org/critical-service"
      - "your-org/api-gateway"
      - "your-org/database-service"
    exclude_bots: true
    include_drafts: false  # Only production-ready PRs

  # Tab 3: Monitor entire organization activity
  - name: "Org Overview"
    mode: "organization"
    organization: "your-org-name"
    exclude_bots: true
    exclude_authors:
      - "renovate[bot]"
      - "dependabot[bot]"
    refresh_interval_minutes: 10  # Slower refresh for overview

  # Tab 4: Monitor PRs assigned to your teams
  - name: "Team PRs"
    mode: "teams"
    organization: "your-org-name"
    teams:
      - "backend-team"
      - "frontend-team"
    exclude_bots: true

  # Tab 5: Custom search - urgent PRs across the org
  - name: "Urgent"
    mode: "search"
    search_query: "org:your-org-name label:urgent is:pr is:open"
    refresh_interval_minutes: 2  # Very fast refresh for urgent items

# Alternative: Single tab configuration (legacy format still supported)
# If you don't specify 'tabs', it will create a single tab with these settings:
# mode: "topics"
# topics: ["backend"]
# topic_org: "your-org"
# exclude_bots: true
`

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(exampleConfig)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

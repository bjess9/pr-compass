package config

import (
	"fmt"
	"os"

	"github.com/bjess9/pr-compass/internal/errors"
	"github.com/spf13/viper"
)

type Config struct {
	// Legacy repo list support
	Repos []string `mapstructure:"repos"`

	// New team-based configuration
	Organization string   `mapstructure:"organization"`
	Teams        []string `mapstructure:"teams"`
	SearchQuery  string   `mapstructure:"search_query"`

	// Topic-based configuration
	Topics   []string `mapstructure:"topics"`
	TopicOrg string   `mapstructure:"topic_org"`

	// Configuration mode
	Mode string `mapstructure:"mode"` // "repos", "organization", "teams", "search", "topics"

	// Filtering options
	ExcludeBots    bool     `mapstructure:"exclude_bots"`    // Exclude renovate, dependabot, etc.
	ExcludeAuthors []string `mapstructure:"exclude_authors"` // Custom authors to exclude
	ExcludeTitles  []string `mapstructure:"exclude_titles"`  // Title patterns to exclude
	IncludeDrafts  bool     `mapstructure:"include_drafts"`  // Include draft PRs (default: true)

	// UI/Performance options
	RefreshIntervalMinutes int `mapstructure:"refresh_interval_minutes"` // Auto-refresh interval (default: 5)
}

func LoadConfig() (*Config, error) {
	return LoadConfigFromPath(getConfigFilePath())
}

func LoadConfigFromPath(configPath string) (*Config, error) {
	// Create a new viper instance to avoid global state issues
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, errors.NewConfigNotFoundError(configPath, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.NewConfigInvalidError(err)
	}

	// Auto-detect mode if not set (for backward compatibility)
	if cfg.Mode == "" {
		if len(cfg.Repos) > 0 {
			cfg.Mode = "repos"
		} else if len(cfg.Topics) > 0 {
			cfg.Mode = "topics"
		} else if cfg.Organization != "" {
			if len(cfg.Teams) > 0 {
				cfg.Mode = "teams"
			} else {
				cfg.Mode = "organization"
			}
		} else if cfg.SearchQuery != "" {
			cfg.Mode = "search"
		} else {
			cfg.Mode = "repos" // default
		}
	}

	// Set default refresh interval if not configured
	if cfg.RefreshIntervalMinutes == 0 {
		cfg.RefreshIntervalMinutes = 5 // default: 5 minutes
	}

	return &cfg, nil
}

// All the interactive configuration functions were removed
// Devs can just create their own YAML files - much simpler!

func ConfigExists() bool {
	_, err := LoadConfig()
	return err == nil
}

func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Unable to determine user home directory: %v\n", err)
		os.Exit(1)
	}
	return fmt.Sprintf("%s/.prpilot_config.yaml", homeDir)
}

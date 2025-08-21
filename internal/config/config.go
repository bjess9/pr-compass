package config

import (
	"fmt"
	"os"
	"strings"

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
	Topics       []string `mapstructure:"topics"`
	TopicOrg     string   `mapstructure:"topic_org"`
	
	// Configuration mode
	Mode string `mapstructure:"mode"` // "repos", "organization", "teams", "search", "topics"
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
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
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
	
	return &cfg, nil
}

func ConfigureRepos() error {
	fmt.Println("=== PR Pilot Configuration ===")
	fmt.Println("Choose how you want to track PRs:")
	fmt.Println("1. Specific repositories (legacy mode)")
	fmt.Println("2. Organization-wide")
	fmt.Println("3. Specific teams in an organization")
	fmt.Println("4. Custom search query")
	fmt.Println("5. Repositories with specific topics/labels (RECOMMENDED)")
	
	fmt.Print("Select option (1-5): ")
	var choice string
	if _, err := fmt.Scanln(&choice); err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	viper.SetConfigFile(getConfigFilePath())
	
	switch choice {
	case "1":
		return configureRepoList()
	case "2":
		return configureOrganization()
	case "3":
		return configureTeams()
	case "4":
		return configureSearch()
	case "5":
		return configureTopics()
	default:
		fmt.Println("Invalid choice. Defaulting to repository list mode.")
		return configureRepoList()
	}
}

func configureRepoList() error {
	fmt.Println("\n--- Repository List Configuration ---")
	fmt.Print("Enter repositories (comma-separated, in owner/repo format): ")
	var reposInput string
	if _, err := fmt.Scanln(&reposInput); err != nil {
		return fmt.Errorf("failed to read repositories: %w", err)
	}
	repos := strings.Split(reposInput, ",")

	// Clean up repo names
	for i, repo := range repos {
		repos[i] = strings.TrimSpace(repo)
	}

	viper.Set("mode", "repos")
	viper.Set("repos", repos)
	viper.Set("organization", "")
	viper.Set("teams", []string{})
	viper.Set("search_query", "")
	viper.Set("topics", []string{})
	viper.Set("topic_org", "")

	return saveConfig()
}

func configureOrganization() error {
	fmt.Println("\n--- Organization Configuration ---")
	fmt.Print("Enter GitHub organization name: ")
	var org string
	if _, err := fmt.Scanln(&org); err != nil {
		return fmt.Errorf("failed to read organization: %w", err)
	}

	viper.Set("mode", "organization")
	viper.Set("organization", strings.TrimSpace(org))
	viper.Set("repos", []string{})
	viper.Set("teams", []string{})
	viper.Set("search_query", "")
	viper.Set("topics", []string{})
	viper.Set("topic_org", "")

	return saveConfig()
}

func configureTeams() error {
	fmt.Println("\n--- Team Configuration ---")
	fmt.Print("Enter GitHub organization name: ")
	var org string
	if _, err := fmt.Scanln(&org); err != nil {
		return fmt.Errorf("failed to read organization: %w", err)
	}
	
	fmt.Print("Enter team names (comma-separated): ")
	var teamsInput string
	if _, err := fmt.Scanln(&teamsInput); err != nil {
		return fmt.Errorf("failed to read teams: %w", err)
	}
	teams := strings.Split(teamsInput, ",")
	
	// Clean up team names
	for i, team := range teams {
		teams[i] = strings.TrimSpace(team)
	}

	viper.Set("mode", "teams")
	viper.Set("organization", strings.TrimSpace(org))
	viper.Set("teams", teams)
	viper.Set("repos", []string{})
	viper.Set("search_query", "")
	viper.Set("topics", []string{})
	viper.Set("topic_org", "")

	return saveConfig()
}

func configureSearch() error {
	fmt.Println("\n--- Custom Search Configuration ---")
	fmt.Println("Example queries:")
	fmt.Println("  org:myorg is:pr is:open")
	fmt.Println("  org:myorg team:accounting-core is:pr is:open")
	fmt.Println("  author:username is:pr is:open")
	fmt.Print("Enter GitHub search query: ")
	
	var query string
	if _, err := fmt.Scanln(&query); err != nil {
		return fmt.Errorf("failed to read search query: %w", err)
	}

	viper.Set("mode", "search")
	viper.Set("search_query", strings.TrimSpace(query))
	viper.Set("organization", "")
	viper.Set("teams", []string{})
	viper.Set("repos", []string{})
	viper.Set("topics", []string{})
	viper.Set("topic_org", "")

	return saveConfig()
}

func configureTopics() error {
	fmt.Println("\n--- Repository Topics Configuration ---")
	fmt.Println("This mode tracks PRs from repositories that have specific topics/labels.")
	fmt.Println("Perfect for tracking all PRs in repos tagged with your team's topic")
	fmt.Println("regardless of who opens the PR.")
	fmt.Println()
	
	fmt.Print("Enter GitHub organization name: ")
	var org string
	if _, err := fmt.Scanln(&org); err != nil {
		return fmt.Errorf("failed to read organization: %w", err)
	}
	
	fmt.Print("Enter repository topics/labels (comma-separated, e.g., 'my-team,backend'): ")
	var topicsInput string
	if _, err := fmt.Scanln(&topicsInput); err != nil {
		return fmt.Errorf("failed to read topics: %w", err)
	}
	topics := strings.Split(topicsInput, ",")
	
	// Clean up topic names
	for i, topic := range topics {
		topics[i] = strings.TrimSpace(topic)
	}

	viper.Set("mode", "topics")
	viper.Set("topics", topics)
	viper.Set("topic_org", strings.TrimSpace(org))
	viper.Set("organization", "")
	viper.Set("teams", []string{})
	viper.Set("repos", []string{})
	viper.Set("search_query", "")

	return saveConfig()
}

func saveConfig() error {
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", getConfigFilePath())
	fmt.Println("You can edit this file directly to modify your settings.")
	return nil
}

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

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Repos []string `mapstructure:"repos"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(getConfigFilePath())
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}
	return &cfg, nil
}

func ConfigureRepos() error {
	fmt.Println("Configuring repositories to track.")
	fmt.Print("Enter repositories (comma-separated, in owner/repo format): ")
	var reposInput string
	fmt.Scanln(&reposInput)
	repos := strings.Split(reposInput, ",")

	viper.SetConfigFile(getConfigFilePath())
	viper.Set("repos", repos)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration saved to %s. You can edit this file directly to modify your repository list.\n", getConfigFilePath())
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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bjess9/pr-pilot/internal"
	"github.com/bjess9/pr-pilot/internal/auth"
	"github.com/bjess9/pr-pilot/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "configure" {
		if err := config.ConfigureRepos(); err != nil {
			fmt.Printf("Configuration error: %v\n", err)
		}
		return
	}

	if !config.ConfigExists() {
		fmt.Println("No configuration found. Please run `prpilot configure` to set up your repositories.")
		return
	}

	token, err := auth.Authenticate()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("Authentication successful. Starting PR Pilot...")
	model := internal.InitialModel(token)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}

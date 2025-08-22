package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bjess9/pr-pilot/internal/auth"
	"github.com/bjess9/pr-pilot/internal/config"
	"github.com/bjess9/pr-pilot/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if !config.ConfigExists() {
		fmt.Println("No configuration found. Create ~/.prpilot_config.yaml")
		fmt.Println("See example_config.yaml for reference.")
		return
	}

	token, err := auth.Authenticate()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("Authentication successful. Starting PR Pilot...")
	model := ui.InitialModel(token)

	p := tea.NewProgram(&model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}

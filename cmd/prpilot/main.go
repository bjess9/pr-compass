package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bjess9/pr-pilot/internal"
	"github.com/bjess9/pr-pilot/internal/auth"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
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

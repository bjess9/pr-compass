package main

import (
	"fmt"
	"os"

	"github.com/bjess9/pr-pilot/internal"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(internal.InitialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
}

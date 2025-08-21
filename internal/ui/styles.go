package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

const (
	BorderColor     = "#FFA500" // Orange
	HeaderBgColor   = "#00008B" // Dark Blue
	HeaderFgColor   = "#FFFFFF" // White
	SelectedBgColor = "#FF4500" // OrangeRed
	SelectedFgColor = "#FFFFFF" // White text
	CellFgColor     = "#D9DCCF" // Light Gray
	HelpFgColor     = "#FFFFFF" // White
)

var (
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2, 1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColor))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(HeaderFgColor)).
			Background(lipgloss.Color(HeaderBgColor))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SelectedBgColor)).
			Foreground(lipgloss.Color(SelectedFgColor)).
			Bold(true)

	cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(CellFgColor))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(HelpFgColor)).
			Padding(1, 0, 0, 0)
)

func tableStyles() table.Styles {
	return table.Styles{
		Header:   headerStyle,
		Cell:     cellStyle,
		Selected: selectedStyle,
	}
}

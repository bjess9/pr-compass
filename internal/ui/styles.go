package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Modern, professional color scheme inspired by popular TUI apps
	BorderColor     = "#61AFEF" // Soft Blue
	HeaderBgColor   = "#2E3440" // Dark Gray-Blue
	HeaderFgColor   = "#ECEFF4" // Off White
	SelectedBgColor = "#5E81AC" // Medium Blue
	SelectedFgColor = "#ECEFF4" // Off White
	CellFgColor     = "#D8DEE9" // Light Gray-Blue
	HelpFgColor     = "#88C0D0" // Soft Cyan
	AccentColor     = "#A3BE8C" // Soft Green
	WarnColor       = "#EBCB8B" // Soft Yellow
	ErrorColor      = "#BF616A" // Soft Red
)

var (
	baseStyle = lipgloss.NewStyle().
			Padding(0, 1, 1, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColor))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(HeaderFgColor)).
			Background(lipgloss.Color(HeaderBgColor)).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SelectedBgColor)).
			Foreground(lipgloss.Color(SelectedFgColor)).
			Bold(true).
			Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CellFgColor)).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(HelpFgColor)).
			Margin(1, 0, 0, 0)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentColor)).
			Margin(1, 0, 0, 0)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(HeaderFgColor)).
			Background(lipgloss.Color(HeaderBgColor)).
			Padding(0, 2).
			Margin(0, 0, 1, 0)
)

func tableStyles() table.Styles {
	return table.Styles{
		Header:   headerStyle,
		Cell:     cellStyle,
		Selected: selectedStyle,
	}
}

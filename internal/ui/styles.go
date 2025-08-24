package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Enhanced dark theme with improved contrast and visual hierarchy
	
	// Base colors - darker background for better focus
	BackgroundColor = "#1A1D23" // Deep dark background
	SurfaceColor    = "#252831" // Slightly lighter surface
	BorderColor     = "#3B4048" // Subtle borders
	
	// Primary colors - bright but not overwhelming
	PrimaryColor    = "#7C9CBF" // Calm blue
	SecondaryColor  = "#9CABCA" // Lighter blue-gray
	AccentColor     = "#98C379" // Fresh green
	
	// Status colors with good contrast
	SuccessColor    = "#98C379" // Green
	WarningColor    = "#E5C07B" // Warm yellow
	ErrorColor      = "#E06C75" // Soft red
	InfoColor       = "#61AFEF" // Bright blue
	
	// Text colors - optimized for readability
	TextPrimary     = "#ABB2BF" // Main text
	TextSecondary   = "#828997" // Secondary text
	TextMuted       = "#5C6370" // Muted text
	TextBright      = "#DCDFE4" // Highlighted text
	
	// UI element colors
	HeaderBgColor   = "#2C3038" // Header background
	HeaderFgColor   = "#DCDFE4" // Header text
	SelectedBgColor = "#4B5263" // Selection background
	SelectedFgColor = "#FFFFFF" // Selection text
	HoverBgColor    = "#383C45" // Hover state
)

var (
	// Base container style with improved spacing and borders
	baseStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(BackgroundColor)).
			Foreground(lipgloss.Color(TextPrimary)).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColor))

	// Enhanced header with better contrast and spacing
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(HeaderFgColor)).
			Background(lipgloss.Color(HeaderBgColor)).
			Padding(0, 1).
			Align(lipgloss.Center)

	// Improved selection highlighting
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SelectedBgColor)).
			Foreground(lipgloss.Color(SelectedFgColor)).
			Bold(true).
			Padding(0, 1)

	// Better cell styling with improved readability
	cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextPrimary)).
			Padding(0, 1)

	// Enhanced help text styling
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextSecondary)).
			Background(lipgloss.Color(SurfaceColor)).
			Padding(1, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color(BorderColor)).
			Margin(1, 0, 0, 0)

	// Improved status messaging
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentColor)).
			Bold(false).
			Margin(0, 0, 1, 0)

	// Enhanced title with gradient-like effect
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(TextBright)).
			Background(lipgloss.Color(HeaderBgColor)).
			Padding(1, 3).
			Margin(0, 0, 1, 0).
			Border(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color(PrimaryColor))

	// Additional specialized styles for better UX
	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(InfoColor)).
			Background(lipgloss.Color(SurfaceColor)).
			Padding(2, 4).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderColor)).
			Align(lipgloss.Center)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ErrorColor)).
			Background(lipgloss.Color(SurfaceColor)).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ErrorColor)).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SuccessColor)).
			Bold(false)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(WarningColor)).
			Bold(false)

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextMuted))
)

func tableStyles() table.Styles {
	return table.Styles{
		Header:   headerStyle,
		Cell:     cellStyle,
		Selected: selectedStyle,
	}
}

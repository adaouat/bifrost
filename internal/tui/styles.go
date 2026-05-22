package tui

import "charm.land/lipgloss/v2"

var (
	ColorSuccess = lipgloss.Color("#1db954")
	ColorError   = lipgloss.Color("#e74c3c")
	ColorWarning = lipgloss.Color("#f5a623")
	ColorMuted   = lipgloss.Color("#808080")
	ColorPrimary = lipgloss.Color("#9b59b6")
)

var (
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	PrimaryStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
)

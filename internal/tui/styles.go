package tui

import "charm.land/lipgloss/v2"

var (
	ColorMuted   = lipgloss.Color("#808080")
	ColorPrimary = lipgloss.Color("#9b59b6")
)

var (
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	PrimaryStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
)

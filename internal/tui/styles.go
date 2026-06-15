package tui

import "charm.land/lipgloss/v2"

var (
	ColorMuted   = lipgloss.Color("#808080")
	ColorPrimary = lipgloss.Color("#9b59b6")
	ColorWarning = lipgloss.Color("#e67e22")
)

var (
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	PrimaryStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
)

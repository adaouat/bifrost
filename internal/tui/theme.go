package tui

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	forgeui "github.com/adaouat/forge/ui"
)

// Accent is bifrost's brand — Aurora: teal title/program/flags, violet commands — over forge's
// shared palette (forge ADR-0010).
func Accent() forgeui.Accent {
	return forgeui.Accent{
		Light:          lipgloss.Color("#0E8A8A"),
		Dark:           lipgloss.Color("#2DD4BF"),
		SecondaryLight: lipgloss.Color("#8250DF"),
		SecondaryDark:  lipgloss.Color("#D2A8FF"),
	}
}

// HuhTheme is bifrost's branded interactive-form theme.
func HuhTheme() huh.ThemeFunc { return forgeui.HuhTheme(Accent()) }

package main

import (
	"image/color"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"

	"github.com/adaouat/forge/ui"
)

// colorScheme is bifrost's fang theme: the Aurora accent (teal title/program/flags, violet
// commands) over forge's shared structural palette. See forge ADR-0008.
func colorScheme(c lipgloss.LightDarkFunc) fang.ColorScheme {
	p := ui.NewPalette(c)
	accent := c(lipgloss.Color("#0E8A8A"), lipgloss.Color("#2DD4BF"))    // teal
	secondary := c(lipgloss.Color("#8250DF"), lipgloss.Color("#D2A8FF")) // violet
	return fang.ColorScheme{
		Base:           p.Text,
		Title:          accent,
		Description:    p.Muted,
		Codeblock:      p.Muted,
		Program:        accent,
		DimmedArgument: p.Dim,
		Comment:        p.Muted,
		Flag:           accent,
		FlagDefault:    p.Dim,
		Command:        secondary,
		QuotedString:   p.Success,
		Argument:       p.Argument,
		Help:           p.Muted,
		Dash:           p.Muted,
		ErrorHeader:    [2]color.Color{lipgloss.Color("#FFFFFF"), p.Error},
		ErrorDetails:   p.Error,
	}
}

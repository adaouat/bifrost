package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	forgeui "github.com/adaouat/forge/ui"
)

// serverHeaderWidth is the target visual width of a server header rule line.
const serverHeaderWidth = 64

// DeployHeader returns a bordered panel displaying env › app and release name.
func DeployHeader(mode forgeui.Mode, env, app, release string) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 2)
	if mode.IsHuman() {
		style = style.BorderForeground(ColorPrimary)
	}
	envApp := env + " › " + app
	content := fmt.Sprintf("Environment   %s\nRelease       %s", envApp, release)
	return "\n" + style.Render(content) + "\n"
}

// ServerHeader returns a horizontal rule labeling a server's output block,
// e.g. "── web-01 (192.168.1.10) ────────────". It separates per-server
// sections in human and plain mode for multi-server commands (spec 08).
func ServerHeader(mode forgeui.Mode, name, host string) string {
	label := fmt.Sprintf("── %s (%s) ", name, host)
	pad := max(serverHeaderWidth-len([]rune(label)), 0)
	line := label + strings.Repeat("─", pad)
	if mode.IsHuman() {
		line = MutedStyle.Render(line)
	}
	return "\n" + line + "\n"
}

// PrintDetail writes an indented sub-detail line under a step.
func PrintDetail(mode forgeui.Mode, out io.Writer, label string) {
	dash := "-"
	if mode.IsHuman() {
		dash = MutedStyle.Render("-")
	}
	_, _ = fmt.Fprintf(out, "      %s %s\n", dash, label)
}

// PrintSummary writes the final deploy summary line to out.
func PrintSummary(mode forgeui.Mode, out io.Writer, elapsed time.Duration, release string) {
	rel := release
	if mode.IsHuman() {
		rel = PrimaryStyle.Render(release)
	}
	_, _ = fmt.Fprintf(out, "\n  Deployed in %.1fs  →  %s\n\n", elapsed.Seconds(), rel)
}

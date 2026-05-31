package tui

import (
	"fmt"
	"io"
	"time"

	"charm.land/lipgloss/v2"
)

// DeployHeader returns a bordered panel displaying env › app and release name.
func DeployHeader(env, app, release string) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 2)
	if IsHumanMode() {
		style = style.BorderForeground(ColorPrimary)
	}
	envApp := env + " › " + app
	content := fmt.Sprintf("Environment   %s\nRelease       %s", envApp, release)
	return "\n" + style.Render(content) + "\n"
}

// PrintStep writes a completed step line to out.
// detail is optional; pass "" to omit it.
func PrintStep(out io.Writer, label, detail string) {
	checkmark := "✔"
	if IsHumanMode() {
		checkmark = SuccessStyle.Render("✔")
	}
	if detail != "" {
		d := detail
		if IsHumanMode() {
			d = MutedStyle.Render(detail)
		}
		_, _ = fmt.Fprintf(out, "  %s %s   %s\n", checkmark, label, d)
	} else {
		_, _ = fmt.Fprintf(out, "  %s %s\n", checkmark, label)
	}
}

// PrintDetail writes an indented sub-detail line under a step.
func PrintDetail(out io.Writer, label string) {
	dash := "-"
	if IsHumanMode() {
		dash = MutedStyle.Render("-")
	}
	_, _ = fmt.Fprintf(out, "      %s %s\n", dash, label)
}

// PrintSummary writes the final deploy summary line to out.
func PrintSummary(out io.Writer, elapsed time.Duration, release string) {
	rel := release
	if IsHumanMode() {
		rel = PrimaryStyle.Render(release)
	}
	_, _ = fmt.Fprintf(out, "\n  Deployed in %.1fs  →  %s\n\n", elapsed.Seconds(), rel)
}

package tui

import (
	"fmt"
	"io"

	"charm.land/lipgloss/v2"
)

// DeployHeader returns a bordered panel displaying env › app and release name.
func DeployHeader(env, app, release string) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 2)

	envApp := env + " › " + app
	content := fmt.Sprintf("Environment   %s\nRelease       %s", envApp, release)
	return "\n" + style.Render(content) + "\n"
}

// PrintStep writes a completed step line to out.
// detail is optional; pass "" to omit it.
func PrintStep(out io.Writer, label, detail string) {
	checkmark := SuccessStyle.Render("✔")
	if detail != "" {
		_, _ = fmt.Fprintf(out, "  %s %s   %s\n", checkmark, label, MutedStyle.Render(detail))
	} else {
		_, _ = fmt.Fprintf(out, "  %s %s\n", checkmark, label)
	}
}

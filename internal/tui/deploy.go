package tui

import (
	"fmt"

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

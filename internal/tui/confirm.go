package tui

import (
	"charm.land/huh/v2"

	forgeui "github.com/adaouat/forge/ui"
)

// InteractiveHookConfirm returns a confirm function for interactive hooks: it
// shows a huh prompt in human mode on a TTY, or nil (the hook is skipped with a
// warning) otherwise — including in json/plain mode, where a blocking prompt
// would corrupt the machine-parsed output stream.
func InteractiveHookConfirm(mode forgeui.Mode) func(cmd string) bool {
	if !mode.IsHuman() || !IsTTY() {
		return nil
	}
	return func(hookCmd string) bool {
		var ok bool
		if err := huh.NewConfirm().
			Title("Run interactive hook?").
			Description(hookCmd).
			Value(&ok).
			WithTheme(HuhTheme()).
			Run(); err != nil {
			return false
		}
		return ok
	}
}

// StepConfirm returns a confirm function for `deploy --interactive`: it shows
// a "Continue to next step?" prompt (default yes) on a TTY, or nil when stdout
// is not a terminal.
func StepConfirm() func(step string) bool {
	if !IsTTY() {
		return nil
	}
	return func(step string) bool {
		ok := true
		if err := huh.NewConfirm().
			Title("Continue to next step?").
			Description(step).
			Value(&ok).
			WithTheme(HuhTheme()).
			Run(); err != nil {
			return false
		}
		return ok
	}
}

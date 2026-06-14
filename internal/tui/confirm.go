package tui

import "charm.land/huh/v2"

// InteractiveHookConfirm returns a confirm function for interactive hooks: it
// shows a huh prompt on a TTY, or nil (the hook is skipped with a warning) when
// stdout is not a terminal.
func InteractiveHookConfirm() func(cmd string) bool {
	if !IsTTY() {
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

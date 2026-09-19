package tui_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
)

// TestStepConfirm_NonTTY_ReturnsNil proves StepConfirm degrades to nil (no
// prompt) when stdout is not a real terminal, matching InteractiveHookConfirm.
func TestStepConfirm_NonTTY_ReturnsNil(t *testing.T) {
	assert.Nil(t, tui.StepConfirm())
}

// TestInteractiveHookConfirm_NonTTY_ReturnsNil proves InteractiveHookConfirm
// degrades to nil (no prompt) when stdout is not a real terminal.
func TestInteractiveHookConfirm_NonTTY_ReturnsNil(t *testing.T) {
	assert.Nil(t, tui.InteractiveHookConfirm(forgeui.Human))
}

// TestInteractiveHookConfirm_NonHumanMode_ReturnsNilEvenOnTTY proves a
// non-human output mode (json/plain) skips the prompt even on a real TTY —
// otherwise a blocking huh confirm can render into a machine-parsed stream.
func TestInteractiveHookConfirm_NonHumanMode_ReturnsNilEvenOnTTY(t *testing.T) {
	tui.SetIsTTY(func() bool { return true })
	defer tui.SetIsTTY(nil)

	assert.Nil(t, tui.InteractiveHookConfirm(forgeui.JSON))
	assert.Nil(t, tui.InteractiveHookConfirm(forgeui.Plain))
}

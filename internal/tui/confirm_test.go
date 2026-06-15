package tui_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/tui"
	"github.com/stretchr/testify/assert"
)

// TestStepConfirm_NonTTY_ReturnsNil proves StepConfirm degrades to nil (no
// prompt) when stdout is not a real terminal, matching InteractiveHookConfirm.
func TestStepConfirm_NonTTY_ReturnsNil(t *testing.T) {
	assert.Nil(t, tui.StepConfirm())
}

package tui_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestSetOutputMode(t *testing.T) {
	defer tui.SetOutputMode("human") // restore default after test

	tui.SetOutputMode("human")
	assert.True(t, tui.IsHumanMode(), "human mode should return true")

	tui.SetOutputMode("plain")
	assert.False(t, tui.IsHumanMode(), "plain mode should return false")

	tui.SetOutputMode("json")
	assert.False(t, tui.IsHumanMode(), "json mode should return false")
}

package tui_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestDeployHeader(t *testing.T) {
	result := tui.DeployHeader("prod", "web", "20260520-141500")

	assert.Contains(t, result, "prod › web")
	assert.Contains(t, result, "20260520-141500")
	assert.Contains(t, result, "Environment")
	assert.Contains(t, result, "Release")
	assert.Contains(t, result, "┌")
	assert.Contains(t, result, "└")
}

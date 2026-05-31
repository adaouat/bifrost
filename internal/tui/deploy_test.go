package tui_test

import (
	"bytes"
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

func TestPrintStep_LabelOnly(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintStep(&buf, "Config loaded and validated", "")
	out := buf.String()
	assert.Contains(t, out, "✔")
	assert.Contains(t, out, "Config loaded and validated")
}

func TestPrintStep_WithDetail(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintStep(&buf, "Artifact extracted", "(1.2s)")
	out := buf.String()
	assert.Contains(t, out, "✔")
	assert.Contains(t, out, "Artifact extracted")
	assert.Contains(t, out, "(1.2s)")
}

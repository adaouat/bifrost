package tui_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
)

func TestDeployHeader(t *testing.T) {
	result := tui.DeployHeader(forgeui.Human, "prod", "web", "20260520-141500")

	assert.Contains(t, result, "prod › web")
	assert.Contains(t, result, "20260520-141500")
	assert.Contains(t, result, "Environment")
	assert.Contains(t, result, "Release")
	assert.Contains(t, result, "┌")
	assert.Contains(t, result, "└")
}

func TestPrintStep_LabelOnly(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintStep(forgeui.Human, &buf, "Config loaded and validated", "")
	out := buf.String()
	assert.Contains(t, out, "✔")
	assert.Contains(t, out, "Config loaded and validated")
}

func TestPrintStep_WithDetail(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintStep(forgeui.Human, &buf, "Artifact extracted", "(1.2s)")
	out := buf.String()
	assert.Contains(t, out, "✔")
	assert.Contains(t, out, "Artifact extracted")
	assert.Contains(t, out, "(1.2s)")
}

func TestPrintSummary(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintSummary(forgeui.Human, &buf, 4300*time.Millisecond, "20260520-141500")
	out := buf.String()
	assert.Contains(t, out, "4.3s")
	assert.Contains(t, out, "20260520-141500")
	assert.Contains(t, out, "→")
}

func TestPrintDetail(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintDetail(forgeui.Human, &buf, "Root path: /var/releases")
	out := buf.String()
	assert.Contains(t, out, "Root path: /var/releases")
	assert.Contains(t, out, "-")
	// Must be indented more than a step line.
	assert.True(t, len(out) > 0 && out[0] == ' ', "detail line should be indented")
}

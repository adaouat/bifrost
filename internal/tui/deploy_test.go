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

func TestServerHeader(t *testing.T) {
	result := tui.ServerHeader(forgeui.Human, "web-01", "192.168.1.10")

	assert.Contains(t, result, "web-01")
	assert.Contains(t, result, "192.168.1.10")
	assert.Contains(t, result, "──")
}

func TestServerHeader_Plain(t *testing.T) {
	result := tui.ServerHeader(forgeui.Plain, "web-01", "192.168.1.10")

	assert.Contains(t, result, "── web-01 (192.168.1.10) ──")
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

func TestPrintWarning(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintWarning(forgeui.Human, &buf, "Deploy in progress — cannot be cancelled, continuing...")
	out := buf.String()
	assert.Contains(t, out, "Deploy in progress — cannot be cancelled, continuing...")
	assert.Contains(t, out, "⚠")
}

func TestPrintWarning_Plain(t *testing.T) {
	var buf bytes.Buffer
	tui.PrintWarning(forgeui.Plain, &buf, "Deploy in progress — cannot be cancelled, continuing...")
	out := buf.String()
	assert.Contains(t, out, "Deploy in progress — cannot be cancelled, continuing...")
	assert.NotContains(t, out, "\x1b", "plain mode must not emit ANSI escapes")
}

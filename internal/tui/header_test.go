package tui_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/bchatard/bifrost/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelpLong(t *testing.T) {
	long := tui.HelpLong()

	assert.NotEmpty(t, long)
	assert.True(t, strings.Contains(long, "██"), "ASCII block art expected")
	assert.Contains(t, long, tui.CatchPhrase)
}

func TestVersionTemplate(t *testing.T) {
	tmpl := tui.VersionTemplate()
	require.NotEmpty(t, tmpl)

	assert.True(t, strings.Contains(tmpl, "██"), "ASCII block art expected")

	_, err := template.New("version").Parse(tmpl)
	require.NoError(t, err)
}

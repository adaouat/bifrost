package release

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/cmderr"
)

func TestListReleases_MissingReleasesRoot_ExitCode3(t *testing.T) {
	_, _, err := listReleases("/nonexistent/releases/root")
	require.Error(t, err)

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, cmderr.Runtime, exitErr.Code, "a missing releases root is a runtime error (3), not the generic default (1)")
}

func TestRenderReleaseList_Header(t *testing.T) {
	var buf bytes.Buffer
	renderReleaseList(&buf, "prod", "web", []string{"r2", "r1"}, "r2")
	out := buf.String()
	assert.Contains(t, out, "Releases for prod › web")
	assert.Contains(t, out, "(2 total)")
}

func TestRenderReleaseList_CurrentSuffix(t *testing.T) {
	var buf bytes.Buffer
	renderReleaseList(&buf, "prod", "web", []string{"r2", "r1"}, "r2")
	out := buf.String()
	assert.Contains(t, out, "← current")
	assert.NotContains(t, out, "* r2")
}

func TestRenderReleaseList_NoCurrent(t *testing.T) {
	var buf bytes.Buffer
	renderReleaseList(&buf, "prod", "web", []string{"r1"}, "")
	out := buf.String()
	assert.NotContains(t, out, "← current")
	assert.Contains(t, out, "r1")
}

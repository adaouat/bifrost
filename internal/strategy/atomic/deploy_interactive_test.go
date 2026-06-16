package atomic_test

import (
	"context"
	"errors"
	"testing"

	"github.com/adaouat/forge/exec/exectest"
	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

// TestDeploy_StepConfirm_DeclineAborts proves that when WithStepConfirm's
// function returns false, Deploy stops right after that step and never runs
// later steps (here: extraction, which would create files in the release dir).
func TestDeploy_StepConfirm_DeclineAborts(t *testing.T) {
	merged := &config.MergedConfig{
		Strategy:     "atomic",
		ReleasesRoot: t.TempDir(),
		SharedRoot:   t.TempDir(),
		Settings:     config.Settings{ReleasesToKeep: 3},
	}
	opts := strategy.DeployOptions{
		Config:      merged,
		Artifact:    "../../../testdata/release.tar.gz",
		ReleaseName: "20260101000000",
	}

	var seen []string
	stepConfirm := func(step string) bool {
		seen = append(seen, step)
		return step != "Release directory created"
	}

	d := atomic.New(nopWriter{}, forgeui.Plain, nil).WithStepConfirm(stepConfirm)

	err := d.Deploy(context.Background(), opts)
	require.Error(t, err)

	var exitErr *cmderr.ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, cmderr.Runtime, exitErr.Code)
	assert.Contains(t, err.Error(), "Release directory created")

	assert.Equal(t, []string{"Config loaded and validated", "Release directory created"}, seen,
		"deploy must stop right after the declined step")
}

// TestDeploy_StepConfirm_FiresAfterHookGroupStep proves the step-continue
// prompt also fires after a configured hook-group step (here: pre_extract),
// not just the seven base steps.
func TestDeploy_StepConfirm_FiresAfterHookGroupStep(t *testing.T) {
	merged := &config.MergedConfig{
		Strategy:     "atomic",
		ReleasesRoot: t.TempDir(),
		SharedRoot:   t.TempDir(),
		Settings:     config.Settings{ReleasesToKeep: 3},
		Hooks: config.Hooks{
			PreExtract: []config.HookEntry{{Cmd: "true"}},
		},
	}
	opts := strategy.DeployOptions{
		Config:      merged,
		Artifact:    "../../../testdata/release.tar.gz",
		ReleaseName: "20260101000000",
	}

	mr := exectest.NewMockRunner()
	mr.QueueResponse("", "", nil) // pre_extract hook

	var seen []string
	stepConfirm := func(step string) bool {
		seen = append(seen, step)
		return step != "pre_extract hooks"
	}

	d := atomic.New(nopWriter{}, forgeui.Plain, nil).WithRunner(mr).WithStepConfirm(stepConfirm)

	err := d.Deploy(context.Background(), opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pre_extract hooks")

	assert.Equal(t, []string{"Config loaded and validated", "Release directory created", "pre_extract hooks"}, seen)
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

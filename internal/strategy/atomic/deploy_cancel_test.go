package atomic_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

// cancelingRunner cancels the deploy's context on its first call, simulating
// a Ctrl+C that arrives while a post-extraction hook is running.
type cancelingRunner struct {
	cancel context.CancelFunc
}

func (r *cancelingRunner) Run(name string, args ...string) (string, string, error) {
	return r.RunDir("", nil, name, args...)
}

func (r *cancelingRunner) RunEnv(env []string, name string, args ...string) (string, string, error) {
	return r.RunDir("", env, name, args...)
}

func (r *cancelingRunner) RunDir(_ string, _ []string, _ string, _ ...string) (string, string, error) {
	r.cancel()
	return "", "", nil
}

// TestDeploy_CancelledAfterExtract_WarnsAndCompletes proves that a context
// cancellation arriving after extraction (e.g. Ctrl+C during a post_extract
// hook) does not abort the deploy: it completes successfully and prints a
// one-time warning that cancellation is being ignored.
func TestDeploy_CancelledAfterExtract_WarnsAndCompletes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	merged := &config.MergedConfig{
		Strategy:     "atomic",
		ReleasesRoot: t.TempDir(),
		SharedRoot:   t.TempDir(),
		Settings:     config.Settings{ReleasesToKeep: 3},
		Hooks: config.Hooks{
			PostExtract: []config.HookEntry{{Cmd: "true"}},
		},
	}
	opts := strategy.DeployOptions{
		Config:      merged,
		Artifact:    "../../../testdata/release.tar.gz",
		ReleaseName: "20260101000000",
	}

	var buf bytes.Buffer
	d := atomic.New(&buf, forgeui.Plain, nil).WithRunner(&cancelingRunner{cancel: cancel})

	err := d.Deploy(ctx, opts)
	require.NoError(t, err, "post-extraction cancellation must not abort the deploy")

	out := buf.String()
	assert.Contains(t, out, "cannot be cancelled")
	assert.Equal(t, 1, strings.Count(out, "cannot be cancelled"), "warning must print exactly once")
}

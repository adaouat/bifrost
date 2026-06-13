package atomic_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/adaouat/forge/exec/exectest"
	"github.com/adaouat/forge/exitcode"
	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

// TestDeploy_HookFailure_ExitsRuntime proves a non-zero hook (allow_fail: false)
// makes the deploy resolve to the Runtime exit code (3), matching specs 03/05,
// and that the JSON error event reports the same code.
func TestDeploy_HookFailure_ExitsRuntime(t *testing.T) {
	merged := &config.MergedConfig{
		Strategy:     "atomic",
		ReleasesRoot: t.TempDir(),
		SharedRoot:   t.TempDir(),
		Settings:     config.Settings{ReleasesToKeep: 3},
		Hooks: config.Hooks{
			PostExtract: []config.HookEntry{{Cmd: "false"}},
		},
	}
	opts := strategy.DeployOptions{
		Config:      merged,
		Artifact:    "../../../testdata/release.tar.gz",
		ReleaseName: "20260101000000",
	}

	mr := exectest.NewMockRunner()
	mr.QueueResponse("", "boom\n", errors.New("exit status 1"))

	var buf bytes.Buffer
	d := atomic.New(&buf, forgeui.JSON, nil).WithRunner(mr)

	err := d.Deploy(context.Background(), opts)
	require.Error(t, err)
	assert.Equal(t, exitcode.Runtime, exitcode.Resolve(err), "hook failure must resolve to Runtime (3)")

	sawError := false
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) == nil && ev["event"] == "error" {
			sawError = true
			assert.EqualValues(t, 3, ev["exit_code"], "error event exit_code")
		}
	}
	assert.True(t, sawError, "expected a JSON error event")
}

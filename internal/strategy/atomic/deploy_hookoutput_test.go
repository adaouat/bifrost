package atomic_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/adaouat/forge/exec/exectest"
	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

// TestDeploy_JSONMode_RoutesHookOutputAsEvents proves that in JSON mode the
// deployer emits hook stdout as hook_output events instead of writing raw text
// into the event stream, so the stream stays valid NDJSON and remote operators
// can see hook output over SSH. A fake runner stands in for real shell execution.
func TestDeploy_JSONMode_RoutesHookOutputAsEvents(t *testing.T) {
	merged := &config.MergedConfig{
		Strategy:     "atomic",
		ReleasesRoot: t.TempDir(),
		SharedRoot:   t.TempDir(),
		Settings:     config.Settings{ReleasesToKeep: 3},
		Hooks: config.Hooks{
			PostExtract: []config.HookEntry{{Cmd: "echo running-migrations"}},
		},
	}
	opts := strategy.DeployOptions{
		Config:      merged,
		Artifact:    "../../../testdata/release.tar.gz",
		ReleaseName: "20260101000000",
		Env:         "production",
		App:         "web",
	}

	mr := exectest.NewMockRunner()
	mr.QueueResponse("running-migrations\n", "", nil) // post_extract hook stdout

	var buf bytes.Buffer
	d := atomic.New(&buf, forgeui.JSON, nil).WithRunner(mr)
	require.NoError(t, d.Deploy(context.Background(), opts))

	sawHookOutput := false
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		var ev map[string]any
		require.NoErrorf(t, json.Unmarshal([]byte(line), &ev),
			"hook output corrupted the NDJSON stream: %q", line)
		if ev["event"] == "hook_output" {
			if s, ok := ev["output"].(string); ok && strings.Contains(s, "running-migrations") {
				sawHookOutput = true
			}
		}
	}
	assert.True(t, sawHookOutput, "expected a hook_output event carrying the hook's stdout")
}

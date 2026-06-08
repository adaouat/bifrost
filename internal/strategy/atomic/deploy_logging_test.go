package atomic_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	flog "github.com/adaouat/forge/log"
	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

// TestDeploy_LogsDiagnostics asserts the atomic deployer emits operator-debug
// diagnostics (step boundaries + key facts) at Debug, and stays silent when the
// level is raised to Warn — proving the --verbose gating.
func TestDeploy_LogsDiagnostics(t *testing.T) {
	tests := []struct {
		name       string
		level      slog.Level
		wantLogged bool
	}{
		{"debug surfaces diagnostics", slog.LevelDebug, true},
		{"warn stays silent", slog.LevelWarn, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
				Env:         "production",
				App:         "web",
			}

			var buf bytes.Buffer
			logger := flog.New(&buf, tc.level)
			d := atomic.New(io.Discard, forgeui.Human, nil).WithLogger(logger)
			require.NoError(t, d.Deploy(context.Background(), opts))

			out := buf.String()
			if tc.wantLogged {
				assert.Contains(t, out, "artifact extracted")
				assert.Contains(t, out, "step completed")
				assert.Contains(t, out, "current_symlink")
			} else {
				assert.NotContains(t, out, "step completed")
				assert.NotContains(t, out, "artifact extracted")
			}
		})
	}
}

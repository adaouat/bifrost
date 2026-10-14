package atomic_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	forgeui "github.com/adaouat/forge/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

// writeLargeTarGz creates a .tar.gz at path containing one file of size bytes
// of random (incompressible) content, so the compressed artifact size is
// close to size — large enough to require many chunked reads during extraction.
func writeLargeTarGz(t *testing.T, path string, size int) {
	t.Helper()
	content := make([]byte, size)
	_, err := rand.Read(content)
	require.NoError(t, err)

	f, err := os.Create(path)
	require.NoError(t, err)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name: "big-file.bin",
		Mode: 0o644,
		Size: int64(len(content)),
	}))
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, f.Close())
}

// TestDeploy_ExtractProgress_ThrottledByPercentage proves extraction progress
// events are throttled to roughly once per 1% of the artifact's size instead
// of once per ~32 KB read chunk — a large artifact would otherwise produce
// hundreds of progress events purely to redraw one progress indicator.
func TestDeploy_ExtractProgress_ThrottledByPercentage(t *testing.T) {
	const artifactSize = 10 * 1024 * 1024 // 10 MiB: 1% threshold (~100 KB) spans several 32 KB chunks
	artifactPath := filepath.Join(t.TempDir(), "big.tar.gz")
	writeLargeTarGz(t, artifactPath, artifactSize)

	merged := &config.MergedConfig{
		Strategy:     "atomic",
		ReleasesRoot: t.TempDir(),
		SharedRoot:   t.TempDir(),
		Settings:     config.Settings{ReleasesToKeep: 3},
	}
	opts := atomic.DeployOptions{
		Config:      merged,
		Artifact:    artifactPath,
		ReleaseName: "20260101000000",
		Env:         "production",
		App:         "web",
	}

	var buf bytes.Buffer
	d := atomic.New(&buf, forgeui.JSON, nil)
	require.NoError(t, d.Deploy(context.Background(), opts))

	progressEvents := 0
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		var ev map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &ev))
		if ev["event"] == "progress" {
			progressEvents++
		}
	}

	assert.Greater(t, progressEvents, 0, "extraction of a 10 MiB artifact must report some progress")
	// The 1% threshold alone bounds emissions to roughly 100 regardless of the
	// underlying chunk size; unthrottled (one event per ~32 KB chunk) would be
	// in the hundreds for a 10 MiB artifact, so this also catches a regression
	// back to per-chunk emission.
	assert.LessOrEqual(t, progressEvents, 110, "progress events must be throttled to roughly once per 1%% of the artifact size, got %d", progressEvents)
}

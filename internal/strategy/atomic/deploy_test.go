package atomic_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/mholt/archives"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

func TestExtract_ArtifactNotFound(t *testing.T) {
	err := atomic.Extract(context.Background(), "/nonexistent/artifact.tar.gz", "/tmp/dest", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist), "expected os.ErrNotExist, got: %v", err)
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	err := atomic.Extract(context.Background(), "../../../testdata/bifrost-full.yml", "/tmp/dest", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, archives.NoMatch)
}

func TestExtract_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := atomic.Extract(ctx, "../../../testdata/release.tar.gz", t.TempDir(), nil)
	require.Error(t, err, "extract must abort on a cancelled context")
	assert.ErrorIs(t, err, context.Canceled)
}

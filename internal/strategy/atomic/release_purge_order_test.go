package atomic

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedReleaseDirs creates a release directory for each entry under root, with
// its modification time set base-age in the past — oldest entries listed
// first age the longest, matching real deploy creation order.
func seedReleaseDirs(t *testing.T, root string, ages []struct {
	name string
	age  time.Duration
}) {
	t.Helper()
	base := time.Now()
	for _, r := range ages {
		dir := filepath.Join(root, r.name)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		mtime := base.Add(-r.age)
		require.NoError(t, os.Chtimes(dir, mtime, mtime))
	}
}

// TestPurge_OrdersByModTimeNotName proves Purge uses directory creation order
// (mtime) rather than lexicographic name order to decide what's oldest.
// --release-name accepts arbitrary values (e.g. "v2.1.3"), so a release
// created later can sort lexicographically "before" one created earlier:
// here "v2.10.0" < "v2.11.0" < "v2.9.0" as strings, the opposite of creation
// order for the first and last entries.
func TestPurge_OrdersByModTimeNotName(t *testing.T) {
	root := t.TempDir()
	seedReleaseDirs(t, root, []struct {
		name string
		age  time.Duration
	}{
		{"v2.9.0", 2 * time.Hour},  // oldest
		{"v2.10.0", 1 * time.Hour}, // middle
		{"v2.11.0", 0},             // newest — active
	})
	require.NoError(t, os.Symlink(filepath.Join(root, "v2.11.0"), filepath.Join(root, "current")))

	// keepN=2: one slot for the active release, one for the newest old one.
	require.NoError(t, Purge(root, 2))

	_, err := os.Stat(filepath.Join(root, "v2.9.0"))
	assert.True(t, os.IsNotExist(err), "the oldest release (by creation time) must be purged, not the lexicographically-first name")

	_, err = os.Stat(filepath.Join(root, "v2.10.0"))
	assert.NoError(t, err, "the newer old release must survive")
}

// TestPurgeNames_RemovesExactlyGivenNames proves PurgeNames deletes exactly
// the given release directories with no scan or sort of its own — the
// deployer uses it to apply PurgePlan's already-computed candidate list
// without re-scanning the releases directory PurgePlan was just computed from.
func TestPurgeNames_RemovesExactlyGivenNames(t *testing.T) {
	root := t.TempDir()
	seedReleaseDirs(t, root, []struct {
		name string
		age  time.Duration
	}{
		{"r1", 2 * time.Hour},
		{"r2", 1 * time.Hour},
		{"r3", 0},
	})

	require.NoError(t, PurgeNames(root, []string{"r1"}))

	_, err := os.Stat(filepath.Join(root, "r1"))
	assert.True(t, os.IsNotExist(err), "r1 must be removed")

	for _, name := range []string{"r2", "r3"} {
		_, err := os.Stat(filepath.Join(root, name))
		assert.NoError(t, err, "%s must survive", name)
	}
}

// TestPurgePlan_OrdersByModTimeNotName is TestPurge_OrdersByModTimeNotName's
// dry-run counterpart: the planned candidate list must match what Purge would
// actually remove.
func TestPurgePlan_OrdersByModTimeNotName(t *testing.T) {
	root := t.TempDir()
	seedReleaseDirs(t, root, []struct {
		name string
		age  time.Duration
	}{
		{"v2.9.0", 2 * time.Hour},
		{"v2.10.0", 1 * time.Hour},
	})

	// newRelease "v2.11.0" doesn't exist on disk yet — PurgePlan is called
	// before the new release directory is created.
	candidates, err := PurgePlan(root, "v2.11.0", 2)
	require.NoError(t, err)
	assert.Equal(t, []string{"v2.9.0"}, candidates, "the oldest existing release must be the purge candidate, not the lexicographically-first name")
}

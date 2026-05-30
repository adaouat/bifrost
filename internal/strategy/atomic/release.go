package atomic

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// PlanReleaseName returns the path that CreateReleaseDir would use for the given
// name (or a timestamp if name is empty), without creating anything on disk.
func PlanReleaseName(releasesRoot, name string) string {
	if name == "" {
		name = time.Now().UTC().Format("20060102-150405")
	}
	return filepath.Join(releasesRoot, name)
}

// CreateReleaseDir creates a new release directory under releasesRoot.
// If name is empty a UTC timestamp is used (YYYYMMDD-HHMMSS).
// Returns the absolute path to the new directory.
func CreateReleaseDir(releasesRoot, name string) (string, error) {
	if name == "" {
		name = time.Now().UTC().Format("20060102-150405")
	}
	dir := filepath.Join(releasesRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating release dir %s: %w", name, err)
	}
	return dir, nil
}

// SetCurrent atomically updates {releasesRoot}/current to point at releaseDir.
// Uses a tmp symlink + rename to make the switch atomic.
func SetCurrent(releasesRoot, releaseDir string) error {
	current := filepath.Join(releasesRoot, "current")
	tmp := current + ".tmp"

	_ = os.Remove(tmp)
	if err := os.Symlink(releaseDir, tmp); err != nil {
		return fmt.Errorf("creating temp symlink: %w", err)
	}
	if err := os.Rename(tmp, current); err != nil {
		return fmt.Errorf("activating current symlink: %w", err)
	}
	return nil
}

// Purge removes old release directories under releasesRoot, keeping the keepN
// most recent. Directories are ranked by name — timestamp names sort correctly.
// The "current" symlink is always skipped.
func Purge(releasesRoot string, keepN int) error {
	entries, err := os.ReadDir(releasesRoot)
	if err != nil {
		return fmt.Errorf("reading releases dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.Name() == "current" {
			continue
		}
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}

	sort.Strings(names) // lexicographic = chronological for timestamp names

	if len(names) <= keepN {
		return nil
	}

	for _, name := range names[:len(names)-keepN] {
		if err := os.RemoveAll(filepath.Join(releasesRoot, name)); err != nil {
			return fmt.Errorf("purging release %s: %w", name, err)
		}
	}
	return nil
}

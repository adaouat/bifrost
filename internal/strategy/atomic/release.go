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
// most recent. The active release (target of the "current" symlink) is never removed.
func Purge(releasesRoot string, keepN int) error {
	entries, err := os.ReadDir(releasesRoot)
	if err != nil {
		return fmt.Errorf("reading releases dir: %w", err)
	}

	// Protect the active release regardless of its sort order.
	var activeRelease string
	if target, err := os.Readlink(filepath.Join(releasesRoot, "current")); err == nil {
		activeRelease = filepath.Base(target)
	}

	var names []string
	for _, e := range entries {
		if e.Name() == "current" || e.Name() == activeRelease {
			continue
		}
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}

	sort.Strings(names)

	for _, name := range purgeFromList(names, keepN) {
		if err := os.RemoveAll(filepath.Join(releasesRoot, name)); err != nil {
			return fmt.Errorf("purging release %s: %w", name, err)
		}
	}
	return nil
}

// PurgePlan returns the release names that would be removed by a purge keeping keepN releases.
// newRelease is protected (it will become the active release) and is never a candidate.
// Returns nil if releasesRoot does not exist (no releases yet).
func PurgePlan(releasesRoot, newRelease string, keepN int) ([]string, error) {
	entries, err := os.ReadDir(releasesRoot)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading releases dir: %w", err)
	}

	var existing []string
	for _, e := range entries {
		if e.Name() == "current" || e.Name() == newRelease {
			continue
		}
		if e.IsDir() {
			existing = append(existing, e.Name())
		}
	}

	sort.Strings(existing)
	return purgeFromList(existing, keepN), nil
}

// purgeFromList returns the existing releases to remove when keepN total slots are available
// and one slot is already reserved for the active/new release.
func purgeFromList(existing []string, keepN int) []string {
	keepOld := keepN - 1
	if keepOld < 0 {
		keepOld = 0
	}
	return purgeCandidates(existing, keepOld)
}

// purgeCandidates returns the entries from an ascending-sorted names slice that
// would be removed when keeping keepN.
func purgeCandidates(names []string, keepN int) []string {
	if len(names) <= keepN {
		return nil
	}
	return names[:len(names)-keepN]
}

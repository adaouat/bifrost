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

	names, err := releaseNamesByModTime(entries, activeRelease)
	if err != nil {
		return fmt.Errorf("reading releases dir: %w", err)
	}

	return PurgeNames(releasesRoot, purgeFromList(names, keepN))
}

// PurgeNames removes exactly the given release directories under
// releasesRoot — no scan or sort of its own. A caller that already knows
// which releases to remove (e.g. from PurgePlan) should call this directly
// instead of Purge, which would otherwise repeat the scan the caller just did.
func PurgeNames(releasesRoot string, names []string) error {
	for _, name := range names {
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

	existing, err := releaseNamesByModTime(entries, newRelease)
	if err != nil {
		return nil, fmt.Errorf("reading releases dir: %w", err)
	}
	return purgeFromList(existing, keepN), nil
}

// releaseNamesByModTime returns release directory names from entries, oldest
// first, excluding "current" and excludeName. --release-name accepts
// arbitrary values (e.g. "v2.1.3"), so sorting lexicographically would not
// reliably reflect creation order the way sorting by directory modification
// time does — release directories are created once and never renamed.
func releaseNamesByModTime(entries []os.DirEntry, excludeName string) ([]string, error) {
	type releaseDir struct {
		name    string
		modTime time.Time
	}
	var dirs []releaseDir
	for _, e := range entries {
		if e.Name() == "current" || e.Name() == excludeName || !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, fmt.Errorf("stat release %s: %w", e.Name(), err)
		}
		dirs = append(dirs, releaseDir{name: e.Name(), modTime: info.ModTime()})
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].modTime.Before(dirs[j].modTime) })

	names := make([]string, len(dirs))
	for i, d := range dirs {
		names[i] = d.name
	}
	return names, nil
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
// would be removed when keeping keepN. Callers that protect one slot for the
// active or new release should pass keepN-1; see purgeFromList.
func purgeCandidates(names []string, keepN int) []string {
	if len(names) <= keepN {
		return nil
	}
	return names[:len(names)-keepN]
}

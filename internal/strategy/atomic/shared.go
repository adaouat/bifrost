package atomic

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkShared links shared directories and files into releaseDir per spec 04.
// Each entry in dirs/files is a path relative to both sharedRoot and releaseDir.
func LinkShared(dirs []string, files []string, releaseDir, sharedRoot string) error {
	for _, rel := range dirs {
		if err := linkOne(rel, releaseDir, sharedRoot, true); err != nil {
			return err
		}
	}
	for _, rel := range files {
		if err := linkOne(rel, releaseDir, sharedRoot, false); err != nil {
			return err
		}
	}
	return nil
}

// linkOne applies the spec 04 linking algorithm for a single path entry.
func linkOne(rel, releaseDir, sharedRoot string, isDir bool) error {
	target := filepath.Join(sharedRoot, rel)
	link := filepath.Join(releaseDir, rel)

	if _, err := os.Lstat(target); os.IsNotExist(err) {
		if _, lerr := os.Lstat(link); lerr == nil {
			// Release shipped its own copy — move it to seed the shared location.
			if merr := os.MkdirAll(filepath.Dir(target), 0o755); merr != nil {
				return fmt.Errorf("preparing shared target %s: %w", rel, merr)
			}
			if merr := os.Rename(link, target); merr != nil {
				return fmt.Errorf("seeding shared %s: %w", rel, merr)
			}
		} else if isDir {
			if merr := os.MkdirAll(target, 0o755); merr != nil {
				return fmt.Errorf("creating shared dir %s: %w", rel, merr)
			}
		} else {
			if merr := os.MkdirAll(filepath.Dir(target), 0o755); merr != nil {
				return fmt.Errorf("creating parent for shared file %s: %w", rel, merr)
			}
			f, merr := os.Create(target)
			if merr != nil {
				return fmt.Errorf("creating shared file %s: %w", rel, merr)
			}
			if cerr := f.Close(); cerr != nil {
				return fmt.Errorf("closing shared file %s: %w", rel, cerr)
			}
		}
	} else if err != nil {
		return fmt.Errorf("checking shared target %s: %w", rel, err)
	}

	if err := os.RemoveAll(link); err != nil {
		return fmt.Errorf("removing link %s: %w", rel, err)
	}

	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return fmt.Errorf("creating parent for link %s: %w", rel, err)
	}

	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("symlinking %s: %w", rel, err)
	}

	return nil
}

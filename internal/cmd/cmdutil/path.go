package cmdutil

import (
	"os"
	"strings"
)

var statFile = os.Stat

// StatFile calls the current stat function. Callers outside the package use this
// so the injectable statFile variable remains unexported.
func StatFile(path string) (os.FileInfo, error) {
	return statFile(path)
}

// SetStatFile replaces the stat function (for testing). Pass nil to restore the default.
// Tests that call this must not run in parallel — this is a package-level global.
func SetStatFile(fn func(string) (os.FileInfo, error)) {
	if fn == nil {
		statFile = os.Stat
	} else {
		statFile = fn
	}
}

// ResolvePath returns the config file path to use for reading.
// Priority: explicit arg → BIFROST_FILE env var → .config/bifrost.yml → .bifrost.yml
func ResolvePath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if env := strings.TrimSpace(os.Getenv("BIFROST_FILE")); env != "" {
		return env
	}
	if _, err := statFile(".config/bifrost.yml"); err == nil {
		return ".config/bifrost.yml"
	}
	return ".bifrost.yml"
}

// ResolveInitDest returns the write destination for config init when no --config flag is set.
// Priority: BIFROST_FILE env var → InitDest (checks .config/ dir)
func ResolveInitDest() string {
	if env := strings.TrimSpace(os.Getenv("BIFROST_FILE")); env != "" {
		return env
	}
	return InitDest()
}

// InitDest returns the path to write a new config file.
// Checks for the .config/ directory (not the file) so it works before the file exists.
func InitDest() string {
	if _, err := statFile(".config"); err == nil {
		return ".config/bifrost.yml"
	}
	return ".bifrost.yml"
}

package cmdutil

import (
	"os"
	"strings"

	forgeconfig "github.com/adaouat/forge/config"
)

var statFile = os.Stat

// StatFile calls the current stat function. config init uses it to check whether
// the target file already exists; the injectable statFile keeps that testable.
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

func resolver() forgeconfig.Resolver { return forgeconfig.Resolver{App: "bifrost"} }

// ResolvePath returns the config file path to use for reading, via forge's
// resolver: explicit arg → BIFROST_FILE → .config/bifrost.yml → .bifrost.yml.
func ResolvePath(explicit string) string {
	path, _ := resolver().Resolve(explicit)
	return path
}

// ResolveInitDest returns the write destination for config init when no --config
// flag is set: BIFROST_FILE env var → InitDest.
func ResolveInitDest() string {
	if env := strings.TrimSpace(os.Getenv("BIFROST_FILE")); env != "" {
		return env
	}
	return InitDest()
}

// InitDest returns the path to write a new config file: .config/bifrost.yml if a
// .config/ directory exists, else .bifrost.yml.
func InitDest() string {
	return resolver().InitDest()
}

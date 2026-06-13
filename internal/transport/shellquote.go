package transport

import "strings"

// ShellQuote wraps s in single quotes for safe interpolation into a remote
// /bin/sh command string, escaping any embedded single quotes. Agent commands
// are assembled as shell strings and run via session.Run, so every value that
// originates from config, flags, or filenames must pass through here.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Package bifrost holds build-time assets embedded into the bifrost binary.
package bifrost

import _ "embed"

// Changelog is the project changelog, embedded so `whatsnew` can fall back to it offline.
//
//go:embed CHANGELOG.md
var Changelog string

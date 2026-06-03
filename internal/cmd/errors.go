package cmd

import "github.com/adaouat/bifrost/internal/cmderr"

// ExitError is an alias for cmderr.ExitError, kept for backward compatibility.
type ExitError = cmderr.ExitError

// Exit codes re-exported for package-cmd-local use (ADR-0003).
const (
	Usage   = cmderr.Usage
	Config  = cmderr.Config
	Runtime = cmderr.Runtime
)

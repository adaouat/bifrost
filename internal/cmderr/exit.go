// Package cmderr re-exports forge's ExitError and the generic exit-code
// vocabulary (ADR-0003) under bifrost's historic names, so existing
// &cmderr.ExitError{Code, Message} call sites keep working.
package cmderr

import forgeexit "github.com/adaouat/forge/exitcode"

// ExitError is an error that requests a specific process exit code.
// main resolves it via forge's exitcode.Resolve.
type ExitError = forgeexit.ExitError

// Generic exit codes shared across the adaouat CLI family (ADR-0003).
const (
	Usage   = forgeexit.Usage   // bad flags/args
	Config  = forgeexit.Config  // invalid config / validation failure
	Runtime = forgeexit.Runtime // external command, network, or IO failure
)

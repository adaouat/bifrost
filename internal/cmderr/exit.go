// Package cmderr re-exports forge's ExitError under bifrost's historic name, so
// existing &cmderr.ExitError{Code, Message} call sites keep working.
package cmderr

import forgeexit "github.com/adaouat/forge/exitcode"

// ExitError is an error that requests a specific process exit code.
// main resolves it via forge's exitcode.Resolve.
type ExitError = forgeexit.ExitError

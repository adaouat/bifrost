package cmd

// ExitError is an error that requests a specific process exit code.
// main.go checks for this type before falling back to exit code 1.
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string { return e.Message }

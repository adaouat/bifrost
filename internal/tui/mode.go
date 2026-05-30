package tui

import "sync/atomic"

// outputMode holds the current output mode set via SetOutputMode.
// 0 = "human" (default), 1 = "plain", 2 = "json"
var outputModeVal atomic.Int32

const (
	modeHuman int32 = 0
	modePlain int32 = 1
	modeJSON  int32 = 2
)

// SetOutputMode sets the package-level output mode. Valid values: "human", "plain", "json".
// Called once from root.PersistentPreRunE after flag validation.
func SetOutputMode(mode string) {
	switch mode {
	case "plain":
		outputModeVal.Store(modePlain)
	case "json":
		outputModeVal.Store(modeJSON)
	default:
		outputModeVal.Store(modeHuman)
	}
}

// IsHumanMode reports whether the current output mode is "human".
// Spinners, progress bars, and colors are only shown in human mode on a real TTY.
func IsHumanMode() bool {
	return outputModeVal.Load() == modeHuman
}

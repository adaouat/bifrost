package tui_test

import (
	"io"
	"testing"

	"github.com/adaouat/bifrost/internal/tui"
)

func TestNewProgressBar_AcceptsTitle(t *testing.T) {
	// Verifies the API accepts a title string; rendering is TTY-dependent
	// so we just confirm the function runs without panic on non-TTY.
	update, done := tui.NewProgressBar(100, "Extracting artifact", io.Discard)
	update(50)
	done()
}

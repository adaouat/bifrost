package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"charm.land/bubbles/v2/progress"
	"charm.land/huh/v2/spinner"
	"github.com/charmbracelet/x/term"
)

// IsTTY reports whether stdout is connected to a real terminal.
func IsTTY() bool {
	return term.IsTerminal(os.Stdout.Fd())
}

// RunWithSpinner runs fn inside a spinner when in human mode on a real TTY; otherwise calls fn directly.
func RunWithSpinner(ctx context.Context, title string, fn func(context.Context) error) error {
	if !IsHumanMode() || !IsTTY() {
		return fn(ctx)
	}
	return spinner.New().
		Title(title).
		Context(ctx).
		ActionWithErr(fn).
		Run()
}

// NewProgressBar returns an update func and a done func for rendering a titled progress bar.
// total is the expected byte count used to compute fill percentage.
// Only renders in human mode when stdout is a real TTY.
func NewProgressBar(total int64, title string, out io.Writer) (update func(n int64), done func()) {
	model := progress.New(progress.WithDefaultBlend(), progress.WithWidth(50))
	var written atomic.Int64
	active := IsHumanMode() && IsTTY()

	prefix := fmt.Sprintf("  %s   ", title)

	update = func(n int64) {
		written.Add(n)
		if !active || total <= 0 {
			return
		}
		pct := float64(written.Load()) / float64(total)
		if pct > 1 {
			pct = 1
		}
		_, _ = fmt.Fprintf(out, "\r%s%s", prefix, model.ViewAs(pct))
	}

	done = func() {
		if active {
			_, _ = fmt.Fprintf(out, "\r%s%s\n", prefix, model.ViewAs(1.0))
		}
	}

	return update, done
}

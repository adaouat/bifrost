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

// RunWithSpinner runs fn inside a spinner when stdout is a TTY; otherwise calls fn directly.
func RunWithSpinner(ctx context.Context, title string, fn func(context.Context) error) error {
	if !IsTTY() {
		return fn(ctx)
	}
	return spinner.New().
		Title(title).
		Context(ctx).
		ActionWithErr(fn).
		Run()
}

// NewProgressBar returns an update func and a done func for rendering a terminal progress bar.
// total is the expected byte count used to compute fill percentage.
// Only renders when stdout is a real TTY.
func NewProgressBar(total int64, out io.Writer) (update func(n int64), done func()) {
	model := progress.New(progress.WithDefaultBlend(), progress.WithWidth(60))
	var written atomic.Int64
	tty := IsTTY()

	update = func(n int64) {
		written.Add(n)
		if !tty || total <= 0 {
			return
		}
		pct := float64(written.Load()) / float64(total)
		if pct > 1 {
			pct = 1
		}
		_, _ = fmt.Fprintf(out, "\r%s", model.ViewAs(pct))
	}

	done = func() {
		if tty {
			_, _ = fmt.Fprintf(out, "\r%s\n", model.ViewAs(1.0))
		}
	}

	return update, done
}

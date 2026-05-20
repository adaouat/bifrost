# ADR-0003: Use Charmbracelet suite for TUI

**Status:** Accepted

## Context

Beyond basic output, the tool needs:

- **Spinner** — during artifact extraction and hook execution (seconds to minutes)
- **Progress bar** — for archive extraction with byte-level feedback
- **Interactive select** — choose a release from a list
- **Confirm prompt** — "re-enable the current release?" / "run with sudo?"
- **Multi-step wizard** — guided `release enable` flow when no `--release` flag is given
- **Styled output** — step completion markers, error blocks, tables

## Decision

Use the **Charmbracelet** suite:

| Package | Role |
|---|---|
| `github.com/charmbracelet/huh` | Interactive forms: `Select`, `Confirm`, `Input`, multi-group wizards |
| `github.com/charmbracelet/huh/spinner` | Spinner for long-running operations |
| `github.com/charmbracelet/bubbles/progress` | Progress bar with byte-count rendering |
| `charm.land/lipgloss/v2` | All terminal styling: colors, borders, padding, layout |
| `github.com/charmbracelet/log` | Structured, styled log output |

fang (ADR-0002) already pulls in lipgloss v2 — no extra dep.

## Rationale

**huh** is the decisive factor. It provides:

- Declarative form API: `huh.NewForm(huh.NewGroup(fields...)).Run()`
- Built-in themes (Charm, Dracula, Catppuccin) that integrate with lipgloss
- Accessible mode for screen readers
- Field-level validation

No equivalent library exists in the Rust ecosystem. Building the same interactivity with
Ratatui would require significant custom widget code.

**huh spinner** integrates with huh forms and accepts an `Action` func, making it trivial
to wrap any blocking operation:

```go
_ = spinner.New().Title("Extracting artifact...").Action(extract).Run()
```

**lipgloss v2** is a significant revision of the original lipgloss with a cleaner color
profile API that adapts to the terminal's capabilities (true color, 256, ANSI, no color).
fang already targets v2, so this is the natural choice.

### Why not Ratatui (Rust)

Already addressed in ADR-0001. The Charmbracelet ecosystem is the primary reason Go was
chosen over Rust.

### Why not bubbletea for the full TUI

Bubble Tea (the Elm-inspired TUI framework) is used internally by huh and bubbles. We use
it indirectly through those libraries. Directly writing Bubble Tea models for spinners and
forms would be more code for no benefit — huh already does it.

## Consequences

- All interactive prompts have a consistent look governed by a single lipgloss color scheme.
- The `--output plain` and `--output json` modes must bypass huh/spinner and write to stdout
  directly, since TUI components assume an interactive terminal.
- huh forms gracefully degrade when stdin is not a TTY (they error early with a clear
  message rather than hanging).

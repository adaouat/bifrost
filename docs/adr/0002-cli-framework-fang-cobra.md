# ADR-0002: Use fang + cobra as CLI framework

**Status:** Accepted

## Context

The tool has multiple subcommands (`config`, `artifact`, `release list`, `release enable`,
`release rollback`) with per-command flags, global flags, and auto-generated help text.
It also needs shell completions and a `--version` flag.

## Decision

Use **cobra** for command structure and **fang** as the execution wrapper.

## Rationale

**cobra** (`github.com/spf13/cobra`) is the de-facto standard Go CLI framework. It powers
`kubectl`, `gh`, `docker`, and hundreds of other tools. It handles:

- Nested subcommands with flag inheritance
- Persistent (global) flags
- Auto-generated `--help` output
- Shell completion scripts (bash, zsh, fish, PowerShell)
- `RunE` pattern for clean error propagation

**fang** (`github.com/charmbracelet/fang`) wraps a cobra root command and adds, for free:

- Styled help pages (lipgloss-rendered, respects color profiles)
- Styled error output (header + detail block, not a raw stack trace)
- Automatic `--version` / `-v` flag with commit SHA support
- Hidden `man` subcommand for manpage generation
- `completion` subcommand for shell completions
- Silent usage on error (help is not dumped on every wrong flag)

The entry point becomes a single call:

```go
fang.Execute(ctx, rootCmd, fang.WithVersion(version), fang.WithCommit(commit))
```

### Why not `urfave/cli` or `kong`

- `urfave/cli`: no styled output, weaker subcommand nesting.
- `kong`: excellent struct-tag-based API, but no fang equivalent for styled output.

Fang is marked **experimental** by Charmbracelet. The risk is accepted because:

1. It wraps cobra — worst case, fang is dropped and cobra runs directly with no other change.
2. The Charmbracelet team actively maintains it alongside lipgloss/huh.

## Consequences

- Command definitions use the standard cobra pattern; no lock-in beyond the entry point.
- Help and error output are styled consistently with the rest of the TUI without extra code.
- fang's `WithColorSchemeFunc` lets us override the theme to match our lipgloss palette.
- The `man` and `completion` subcommands are available automatically.

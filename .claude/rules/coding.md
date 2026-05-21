# Coding rules

## Error handling

- Always wrap errors with context: `fmt.Errorf("context: %w", err)`.
- Use `errors.Is` / `errors.As` for inspection, never string matching.
- Commands use `RunE` not `Run` — return errors to fang's styled error handler.
- Never call `os.Exit` below the `internal/cmd/` layer.
- Never panic in library code — return an error.

## Code quality

- No comments unless the WHY is non-obvious (a hidden constraint, a subtle invariant,
  a workaround for a specific bug). Never describe WHAT the code does.
- No multi-line comment blocks or multi-paragraph docstrings.
- No features, abstractions, or refactoring beyond what the current task requires.
- No error handling for scenarios that cannot happen (trust internal guarantees).
- Only validate at system boundaries: user input, external APIs, config files.
- No backwards-compatibility shims for code that does not exist yet.
- Three similar lines is better than a premature abstraction.

## Architecture

- Deploy logic lives in `internal/strategy/atomic/` — not in `internal/cmd/`.
- Hook execution lives in `internal/hooks/` — shared across all strategies.
- No strategy interface in v0 — it will be introduced in v1 when a second strategy exists.
- TUI components (spinners, progress, forms) only activate in `human` mode on a real TTY.
- Non-TTY detection is the responsibility of `internal/tui/`, not individual commands.

## Config

- Default config path: `./.bifrost.yml`. Always overridable with `--config <path>`.
- `strategy: atomic` is the default and only valid value in v0.

## Hook execution

- All hook commands run via `sh -c "<cmd>"` — never direct `exec` with split arguments.
- This supports `&&`, `|`, redirects, and arguments with spaces.
- `exec.Command` in the hook runner must be injectable for unit testing (mockable).

## Output modes

- `--output` flag: `human` (default), `json`, `plain`.
- `human`: colored TUI, spinners, progress bars — only on a real TTY.
- `json`: one JSON object per event to stdout, no TUI.
- `plain`: no colors, no spinners — CI-friendly.
- Check the output mode before rendering any TUI component.

## Template variables in hooks

Go `text/template` syntax:

```
{{ .Directories.Working }}    # current release directory
{{ .Directories.Current }}    # path to current symlink
{{ .Directories.Releases }}   # releases root
{{ .Directories.Shared }}     # shared root
{{ .Variables.key }}          # user-defined variables
{{ .Settings.ReleasesToKeep }}
{{ .Env.DB_HOST }}            # OS environment variables
```

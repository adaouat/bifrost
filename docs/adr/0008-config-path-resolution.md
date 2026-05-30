# ADR-0008 — Config File Path Resolution

## Status

Accepted

## Context

Bifrost needs to locate its configuration file at startup. Multiple callers (deploy, config
show/check/init, release commands) each duplicated the same two-step lookup logic
(`.config/bifrost.yml` → `.bifrost.yml`), diverging slightly over time.

Additionally, CI/CD environments often inject configuration via environment variables rather
than CLI flags (flags require wrapper scripts; env vars are native to most CI platforms).

## Decision

Centralise all config-path logic in `internal/cmd/cmdutil` and extend the priority chain
with a `BIFROST_FILE` environment variable.

### Priority chain

| Priority | Source | When used |
|---|---|---|
| 1 | `--config <path>` CLI flag | Always wins when explicitly provided |
| 2 | `BIFROST_FILE` environment variable | Useful in CI/CD where env vars are easier to inject than flags; whitespace-only value is treated as unset |
| 3 | `.config/bifrost.yml` (file exists) | XDG-style config directory; preferred when present |
| 4 | `.bifrost.yml` | Default dotfile fallback |

### Write destination for `config init`

The init command writes a new file, so step 3 cannot check whether the file exists
(it is being created). Instead, it checks whether the `.config/` **directory** exists and
writes to `.config/bifrost.yml` if it does, otherwise to `.bifrost.yml`.

### API

```go
// ResolvePath returns the file to load. Callers pass the --config flag value (empty string if unset).
func ResolvePath(explicit string) string

// ResolveInitDest returns the write destination when --config is not set.
// Priority: BIFROST_FILE → InitDest()
func ResolveInitDest() string

// InitDest checks for the .config/ directory and returns the appropriate write path.
func InitDest() string
```

## Consequences

- One canonical path resolution implementation instead of four copies.
- `BIFROST_FILE=/etc/app/bifrost.yml bifrost deploy …` works without flags.
- Whitespace-only `BIFROST_FILE` is treated as unset; callers need not trim.
- Setting `BIFROST_FILE` unexpectedly (e.g. a leftover CI env var) silently shadows the
  project config — operators should audit their CI environment if deployments pick up the
  wrong config file.

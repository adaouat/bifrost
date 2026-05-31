@.claude/rules/workflow.md
@.claude/rules/testing.md
@.claude/rules/coding.md
@.claude/rules/claude.md

# CLAUDE.md — Bifrost

Atomic deployment CLI, rewritten in Go from the Dart v1 (`../deployer`).
Named after the Norse/Marvel bridge that connects worlds — see [ADR-0005](docs/adr/0005-tool-name-bifrost.md).

## What this tool does

Manages versioned application deployments on a remote server:

1. Extracts an artifact (`.tar.gz`, `.zip`) into a timestamped release directory
2. Links shared directories and files via symlinks (persisted across deployments)
3. Switches the active release via a `current` symlink (atomic OS operation)
4. Runs ordered lifecycle hooks (pre/post enable) with template variable support
5. Purges stale releases, keeping the N most recent

Directory layout on the target server:

```
{releases_root}/
├── 20260101-120000/   ← old release
├── 20260520-141500/   ← current release
│   ├── var/log -> {shared_root}/var/log
│   └── .env    -> {shared_root}/.env
└── current -> 20260520-141500

{shared_root}/
├── var/log/
└── .env
```

## Docs

- [`docs/specs/`](docs/specs/) — full technical specifications (commands, config schema, hooks)
- [`docs/adr/`](docs/adr/) — architecture decision records
- [`docs/tasks/`](docs/tasks/) — version roadmap (v0–v5) and implementation milestones (M0–M10)

## Tech stack

| Tool | Role |
|---|---|
| **Go** (via mise) | Implementation language |
| **cobra** | CLI subcommand structure |
| **fang** | cobra wrapper: styled help/errors, `--version`, completions |
| **huh** | Interactive forms, selects, confirms, wizard flows |
| **huh/spinner** | Spinner for long-running steps |
| **bubbles/progress** | Progress bar for artifact extraction |
| **lipgloss v2** | Terminal styling |
| **charmbracelet/log** | Structured log output |
| **yaml.v3** | YAML config parsing |
| **text/template** | Hook command templating |
| **mholt/archives** | In-process archive extraction (tar.gz, zip) |
| **testcontainers-go** | Linux containers for integration tests |
| **goreleaser** | Cross-platform release builds and GitHub release automation |
| **git-cliff** | Changelog and release notes generation from conventional commits |

## Project layout

```
cmd/bifrost/main.go           # Entry point: fang.Execute
internal/
  cmd/                        # cobra command definitions
    cmdutil/                  # Path resolution helpers (ResolvePath, InitDest)
    root.go                   # Global flags: --config, --output, --dry-run, --verbose
    config/
      config.go               # `config` parent command
      show.go                 # `config show`
      check.go                # `config check`
      init.go                 # `config init`
    deploy.go                 # `deploy` command
    release/
      release.go              # `release` parent
      list.go                 # `release list`
      activate.go             # `release activate`
      rollback.go             # `release rollback`
      init.go                 # `release init`
  cmderr/                     # ExitError — shared exit-code error type
  config/
    schema.go                 # Go structs matching the YAML schema
    loader.go                 # YAML loading + strict validation
    merge.go                  # 3-level merge (global < env < app)
  strategy/
    atomic/                   # Atomic deployment strategy (v0)
      deploy.go               # Full deploy flow
      release.go              # Release directory creation, current symlink, purge
      shared.go               # Shared dir/file linking
  hooks/                      # Hook execution — shared across all strategies
    runner.go                 # sh -c, sudo, template rendering, priority sort
  tui/
    styles.go                 # lipgloss color scheme and style constants
    deploy.go                 # Deploy output: header panel, step/detail/summary lines
    events.go                 # JSONEmitter — newline-delimited JSON event writer
    header.go                 # ASCII art and version template (root help)
    mode.go                   # Output mode state (human / plain / json)
    progress.go               # IsTTY, RunWithSpinner, NewProgressBar
  testutil/                   # Shared test helpers (container setup, binary builder)
testdata/                     # Read-only test fixtures (YAML configs, small archives)
```

## Tooling (mise)

All tooling is managed by mise. Run `mise install` after cloning.

```bash
mise run build              # Compile to ./bifrost
mise run test               # Run tests
mise run lint:check         # Check all linters (hk check)
mise run lint:fix           # Auto-fix all linters (hk fix)
mise run lint:go:check      # Check Go code (golangci-lint)
mise run lint:go:fix        # Fix Go code (golangci-lint --fix)
mise run run -- <args>      # Run the CLI in dev mode
```

Integration tests (requires Docker):

```bash
go test -tags integration ./...
```

For targeted lint fixes: `hk fix -S <linter>` (e.g. `hk fix -S golangci-lint`, `hk fix -S yamlfmt`).

Go, golangci-lint, and git-cliff are installed via mise (see `.config/mise/config.toml`).

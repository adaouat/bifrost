# CLAUDE.md — Bifrost

Atomic deployment CLI, rewritten in Go from the Dart v1 (`../deployer`).
Named after the Norse/Marvel bridge that connects worlds — see [ADR-0005](docs/adr/0005-tool-name-bifrost.md).

## What this tool does

Manages versioned application deployments on a remote server:

1. Extracts an artifact (`.tar.gz`, `.zip`) into a timestamped release directory
2. Links shared directories and files via symlinks (persisted across deployments)
3. Switches the active release via a `current` symlink
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
- [`docs/adr/`](docs/adr/) — architecture decision records (language, framework, TUI, templates)
- [`docs/tasks/`](docs/tasks/) — development roadmap

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
| **goreleaser** | Cross-platform release builds and GitHub release automation |
| **git-cliff** | Changelog and release notes generation from conventional commits |

## Project layout

```
cmd/bifrost/main.go           # Entry point: fang.Execute
internal/
  cmd/                        # cobra command definitions
    root.go                   # Global flags: --config, --output, --dry-run
    config.go                 # `config` command
    artifact.go               # `artifact` command
    release/
      release.go              # `release` parent
      list.go                 # `release list`
      enable.go               # `release enable`
      rollback.go             # `release rollback`
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
    styles.go                 # lipgloss color scheme
    output.go                 # Step/error/table output helpers
    progress.go               # Spinner + progress bar wrappers
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

For targeted lint fixes use `hk fix -S <linter>` (e.g. `hk fix -S golangci-lint`, `hk fix -S yamlfmt`).

Go, golangci-lint, and git-cliff are installed via mise (see `.config/mise/config.toml`).

## Conventions

### Commits

Conventional commits are enforced by `hk` + `cocogitto`. Valid types:
`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `style`, `perf`, `ci`, `build`

Example: `feat(artifact): add --release-name flag`

### Config file

Default config path: `./.bifrost.yml`. Always overridable with `--config <path>`.

### Hook execution

All hook commands run via `sh -c "<cmd>"` (with or without `sudo`), not via direct
`exec`. This supports shell operators (`&&`, `|`, redirects) and arguments with spaces.
This is a deliberate fix over v1's `.split(' ')` approach.

### Template variables in hooks

Go `text/template` syntax (not Liquid as in v1):

```
{{ .Directories.Working }}   # Current release directory
{{ .Directories.Current }}   # Path to `current` symlink
{{ .Directories.Releases }}  # Releases root
{{ .Directories.Shared }}    # Shared root
{{ .Variables.key }}         # User-defined variables
{{ .Settings.ReleasesToKeep }}
{{ .Env.DB_HOST }}           # OS environment variables
```

### Output modes

`--output` global flag: `human` (default, colored TUI), `json`, `plain`.
TUI components (spinners, progress, forms) only activate in `human` mode on a real TTY.

### TDD

Write tests before implementation. Every code task starts with a failing test. Implementation
is written only to make the test pass, nothing more.

### Testing constraints

Two hard rules — never break them:

1. **No filesystem writes on the local host.** No `os.MkdirAll`, `os.Create`,
   `os.Symlink`, `os.Remove`, no `t.TempDir()`. Writing a compiled binary to `/tmp` is
   the only exception.
2. **No shell execution on the local host.** `sh -c` from hook execution must never run
   on the local machine.

**Two test categories** (see ADR-0007):

| Category | Command | Docker | Scope |
|---|---|---|---|
| Unit | `go test ./...` | No | Pure logic only — config, merge, template rendering |
| Integration | `go test -tags integration ./...` | Yes | Filesystem + shell — runs inside Linux containers via testcontainers-go |

All container test helpers live in `internal/testutil/`. Fixture files (YAML configs,
small archives) live in `testdata/` — read-only, never written during test runs.

`TestMain` cross-compiles the binary once (`GOOS=linux GOARCH=amd64`) to `/tmp` and
shares it across all container tests.

### Error handling

Return errors up the call stack using `fmt.Errorf("context: %w", err)`.
Commands use `RunE` and return errors to fang's styled error handler.
Never call `os.Exit` below the command layer.

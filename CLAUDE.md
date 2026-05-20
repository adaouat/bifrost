# CLAUDE.md — Deployer v2

Capistrano-style deployment CLI, rewritten in Go from the Dart v1 (`../deployer`).

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

## Project layout

```
cmd/deployer/main.go          # Entry point: fang.Execute
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
  deploy/
    artifact.go               # Archive extraction + deploy flow
    release.go                # Symlink management (current pointer)
    shared.go                 # Shared dir/file linking
    hooks.go                  # Hook execution: sh -c, sudo, templates
  tui/
    styles.go                 # lipgloss color scheme
    output.go                 # Step/error/table output helpers
    progress.go               # Spinner + progress bar wrappers
```

## Tooling (mise)

All tooling is managed by mise. Run `mise install` after cloning.

```bash
mise run build          # Compile to ./deployer
mise run test           # Run tests
mise run lint:check     # Check linting (hk check)
mise run lint:fix       # Auto-fix linting (hk fix)
mise run run -- <args>  # Run the CLI in dev mode
```

Go and golangci-lint are installed via mise (see `.config/mise/config.toml`).

## Conventions

### Commits

Conventional commits are enforced by `hk` + `cocogitto`. Valid types:
`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `style`, `perf`, `ci`, `build`

Example: `feat(artifact): add --release-name flag`

### Config file

Default config path: `./.deployer.yml`. Always overridable with `--config <path>`.

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

### Error handling

Return errors up the call stack using `fmt.Errorf("context: %w", err)`.
Commands use `RunE` and return errors to fang's styled error handler.
Never call `os.Exit` below the command layer.

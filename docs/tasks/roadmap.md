# Development Roadmap

Milestones in implementation order. Each builds on the previous.

## Working process

Each task follows a two-step commit flow:

1. **Implement** — do the work (TDD: failing test first, then implementation).
2. **Done** — mark the task `[x]` in the roadmap, commit implementation + roadmap update together.

Task status markers:

| Marker | Meaning |
|---|---|
| `[ ]` | Not started |
| `[x]` | Done |

One task at a time. The roadmap always reflects the current state of the branch.

### TDD

Write tests before implementation. Every code task starts with a failing test that defines
the expected behavior. Implementation is done only to make the test pass.

### Lint and formatting

Run `hk check` before committing. Fix all issues before the commit lands.
Use `hk fix -S <linter>` for targeted auto-fixes (e.g. `hk fix -S golangci-lint`,
`hk fix -S yamlfmt`). Use `hk fix` to fix everything at once.

---

## M0 — Foundation

Project skeleton, tooling, and release pipeline. No deployment logic yet.

### Tooling

- [x] Add Go, golangci-lint, git-cliff to mise config (`.config/mise/config.toml`)
- [x] Add Go lint and `go mod tidy` steps to hk config (`.config/hk/config.pkl`)
- [x] Add mise tasks: `build`, `test`, `run`, `lint:go:check`, `lint:go:fix`

### Go skeleton

- [x] `go mod init` — module path `github.com/adaouat/bifrost`
- [x] `internal/testutil/` — container test helpers (testcontainers-go, `TestMain` binary builder)
- [x] `testdata/` — read-only fixture files (YAML configs, small `.tar.gz` archives)
- [x] `cmd/bifrost/main.go` — `fang.Execute` entry point
- [x] `internal/cmd/root.go` — root command, global flags (`--config`, `--output`, `--dry-run`)
- [x] `internal/tui/styles.go` — lipgloss color scheme
- [x] Stub commands: `config`, `deploy`, `release list`, `release activate`, `release rollback`

### CI/CD

- [x] `cliff.toml` — git-cliff config at `.config/cliff.toml` (conventional commits, emoji groups, GitHub remote)
- [x] `.goreleaser.yml` — linux amd64/arm64, macOS amd64/arm64, windows amd64/arm64 (best effort); checksums; git-cliff changelog via `--config .config/cliff.toml`
- [x] `.github/workflows/ci.yml` — on push/PR: lint → test → build
- [x] `.github/workflows/release.yml` — on tag: goreleaser full release

Deliverable: `bifrost --help` renders styled help, `bifrost --version` works, CI passes on push, release pipeline produces artifacts.

---

## M1 — Configuration

YAML loading, validation, and 3-level merge. No filesystem operations yet.

- [x] `internal/config/schema.go` — Go structs for the full YAML schema
- [x] `internal/config/loader.go` — strict YAML parsing with clear error messages
- [x] `internal/config/merge.go` — global < env < app merge (maps deep, lists concat+sort, scalars override)
- [x] `config` command — print full merged config as JSON
- [x] `config --environment --application` — validate required fields, report all errors

Deliverable: `bifrost config` works on a real `.bifrost.yml`.

---

## M2 — Core Deployment

Artifact extraction and symlink management. No hooks yet.

- [x] `internal/strategy/atomic/deploy.go` — archive extraction (tar.gz, zip) via `mholt/archives`
- [x] `internal/strategy/atomic/shared.go` — shared dir/file linking algorithm (spec 04)
- [x] `internal/strategy/atomic/release.go` — release directory creation, `current` symlink update, purge
- [x] `deploy` command — full deploy flow (steps 1–11 from spec 03), no hooks
- [x] `internal/tui/progress.go` — spinner + progress bar wrappers
- [x] Wire progress bar to extraction byte stream

Deliverable: `bifrost deploy` deploys a real archive end-to-end.

---

## M3 — Hook System

Hook execution engine and `release enable`.

- [x] `internal/hooks/runner.go` — `sh -c` execution, sudo, template rendering, priority sort
- [x] Hook `cmd_dir` support
- [x] Hook `allow_fail` support
- [x] Hook `interactive` support (huh confirm prompt)
- [x] Wire hooks into `artifact` command
- [x] `release enable` command — full flow with hooks
- [x] `release list` command

Deliverable: full deploy with hooks, `bifrost release enable` and `bifrost release list` work.

---

## M4 — Release Management

Rollback, release naming, init command.

- [x] `release rollback` command
- [x] `--release-name` flag on `artifact`
- [x] `init` command (create roots, validate config)

Deliverable: full release management workflow complete.

---

## M5 — Config Command Restructure

Promote `config` from a single command into a subcommand group. Remove the
top-level `init` command (no replacement at root level).

- [x] `config show [--env --app]` — replaces current `config`; prints full merged
  config, or app-scoped config when `--env`/`--app` are given
- [x] `config check --env --app` — replaces `config --env --app` validation path;
  validates required fields, reports all errors, exits 2 on failure
- [x] `config init` — scaffolds a default `.bifrost.yml` with inline comments
  covering every field; overwrites only with explicit `--force`

Deliverable: `bifrost config` shows a help screen with three subcommands;
`bifrost config init` produces a ready-to-edit config file.

---

## M5.1 — Config Path Resolution

Consolidate config-path logic into a shared `internal/cmd/cmdutil` package and
extend the priority chain with a `BIFROST_FILE` environment variable.

- [x] `internal/cmd/cmdutil/path.go` — `ResolvePath(explicit string) string` and
  `InitDest() string`; priority: explicit → `BIFROST_FILE` env var → disk discovery
- [x] Update all callers (`configPath`, `resolveConfigPath`, `releaseConfigPath`,
  `initPath`) to delegate to `cmdutil`
- [x] Update `config init` to honour `BIFROST_FILE` for write destination
- [x] Update spec 02 and add ADR-0008 documenting the four-step priority chain

Deliverable: `BIFROST_FILE=/etc/bifrost.yml bifrost deploy …` works; config path
logic lives in one place.

---

## M6 — TUI Polish

Interactive flows, output modes, dry run.

- [x] `release enable` interactive selector (huh select, no `--release` needed)
- [x] `--output json` event stream (spec 06)
- [x] `--output plain` (no colors/spinners)
- [x] `--dry-run` mode for `artifact` and `release enable`
- [x] Non-TTY detection — disable TUI components automatically
- [x] `NO_COLOR` support

Deliverable: all output modes and interactive flows work.

---

## M7 — Quality

Test coverage and hardening. CI/CD is already live from M0; this milestone
fills in the test gaps that weren't covered incrementally.

- [x] Unit tests for config merge logic
- [x] Unit tests for shared resource linking algorithm
- [x] Unit tests for hook template rendering
- [x] Integration tests for `deploy` command (real filesystem, temp dir)
- [x] Integration tests for `release activate` / `release rollback`

---

## M8 — Spec 06 Completion

Close the remaining gaps between the current output and spec 06.

### Human mode — `deploy`

- [x] Header panel: bordered box showing environment › application and release name
- [ ] Per-step `✔` lines: config loaded, release dir created, artifact extracted (with duration), hooks (with count), shared dirs linked (with count), shared files linked (with count), current symlink updated, releases purged (with kept count)
- [ ] Final summary line: `Deployed in Xs  →  <release>`

### Human mode — `release list`

- [ ] Header line: `Releases for <env> › <app>  (<n> total)`
- [ ] Current release shown with `← current` suffix instead of `*` prefix

### JSON mode — `deploy`

- [ ] `{"event":"error","step":"...","message":"...","exit_code":3}` on failure
- [ ] `start`/`done` events for link, current-symlink, and purge steps

### Dry run — `deploy`

- [ ] `Would purge` line lists the actual release names that would be removed
- [ ] Hook lines with `sudo: true` show `(sudo)` at the end

---

## Bugs fixed vs v1

| Bug | Status | Fixed in |
|---|---|---|
| BUG-1: `cmd.split(' ')` breaks args with spaces | Fixed | M3 (`sh -c`) |
| BUG-2: sudo + shell operators broken | Fixed | M3 (`sh -c`) |
| BUG-3: extraction stderr hidden on failure | Fixed | M2 (always shown) |

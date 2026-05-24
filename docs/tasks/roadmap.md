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
- [ ] `internal/strategy/atomic/release.go` — release directory creation, `current` symlink update, purge
- [ ] `deploy` command — full deploy flow (steps 1–11 from spec 03), no hooks
- [ ] `internal/tui/progress.go` — spinner + progress bar wrappers
- [ ] Wire progress bar to extraction byte stream

Deliverable: `bifrost deploy` deploys a real archive end-to-end.

---

## M3 — Hook System

Hook execution engine and `release enable`.

- [ ] `internal/hooks/runner.go` — `sh -c` execution, sudo, template rendering, priority sort
- [ ] Hook `cmd_dir` support
- [ ] Hook `allow_fail` support
- [ ] Hook `interactive` support (huh confirm prompt)
- [ ] Wire hooks into `artifact` command
- [ ] `release enable` command — full flow with hooks
- [ ] `release list` command

Deliverable: full deploy with hooks, `bifrost release enable` and `bifrost release list` work.

---

## M4 — Release Management

Rollback, release naming, init command.

- [ ] `release rollback` command
- [ ] `--release-name` flag on `artifact`
- [ ] `init` command (create roots, validate config)

Deliverable: full release management workflow complete.

---

## M5 — TUI Polish

Interactive flows, output modes, dry run.

- [ ] `release enable` interactive selector (huh select, no `--release` needed)
- [ ] `--output json` event stream (spec 06)
- [ ] `--output plain` (no colors/spinners)
- [ ] `--dry-run` mode for `artifact` and `release enable`
- [ ] Non-TTY detection — disable TUI components automatically
- [ ] `NO_COLOR` support

Deliverable: all output modes and interactive flows work.

---

## M6 — Quality

Test coverage and hardening. CI/CD is already live from M0; this milestone
fills in the test gaps that weren't covered incrementally.

- [ ] Unit tests for config merge logic
- [ ] Unit tests for shared resource linking algorithm
- [ ] Unit tests for hook template rendering
- [ ] Integration tests for `artifact` command (real filesystem, temp dir)
- [ ] Integration tests for `release enable` / `release rollback`

---

## Bugs fixed vs v1

| Bug | Status | Fixed in |
|---|---|---|
| BUG-1: `cmd.split(' ')` breaks args with spaces | Fixed | M3 (`sh -c`) |
| BUG-2: sudo + shell operators broken | Fixed | M3 (`sh -c`) |
| BUG-3: extraction stderr hidden on failure | Fixed | M2 (always shown) |

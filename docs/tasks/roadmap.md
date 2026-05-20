# Development Roadmap

Milestones in implementation order. Each builds on the previous.

## Working process

Each task follows a three-step commit flow:

1. **Start** — mark the task `[~]` in the roadmap, commit the roadmap alone.
2. **Implement** — do the work.
3. **Done** — mark the task `[x]` in the roadmap, commit implementation + roadmap update together.

Task status markers:

| Marker | Meaning |
|---|---|
| `[ ]` | Not started |
| `[~]` | In progress |
| `[x]` | Done |

One task at a time. The roadmap always reflects the current state of the branch.

---

## M0 — Foundation

Project skeleton and tooling. No deployment logic yet.

- [ ] Add Go to mise config (`.config/mise/config.toml`)
- [ ] Add golangci-lint to mise config
- [ ] Add Go lint steps to hk config (`.config/hk/config.pkl`)
- [ ] Add mise tasks: `build`, `test`, `run`, `lint:go:check`, `lint:go:fix`
- [ ] `go mod init` — module path `github.com/bchatard/bifrost`
- [ ] `cmd/bifrost/main.go` — `fang.Execute` entry point
- [ ] `internal/cmd/root.go` — root command, global flags (`--config`, `--output`, `--dry-run`)
- [ ] Stub commands: `config`, `artifact`, `release list`, `release enable`, `release rollback`
- [ ] `internal/tui/styles.go` — lipgloss color scheme

Deliverable: `bifrost --help` renders styled help, `bifrost --version` works.

---

## M1 — Configuration

YAML loading, validation, and 3-level merge. No filesystem operations yet.

- [ ] `internal/config/schema.go` — Go structs for the full YAML schema
- [ ] `internal/config/loader.go` — strict YAML parsing with clear error messages
- [ ] `internal/config/merge.go` — global < env < app merge (maps deep, lists concat+sort, scalars override)
- [ ] `config` command — print full merged config as JSON
- [ ] `config --environment --application` — validate required fields, report all errors

Deliverable: `bifrost config` works on a real `.deployer.yml`.

---

## M2 — Core Deployment

Artifact extraction and symlink management. No hooks yet.

- [ ] `internal/deploy/artifact.go` — archive extraction (tar.gz, zip) via `mholt/archives`
- [ ] `internal/deploy/shared.go` — shared dir/file linking algorithm (spec 04)
- [ ] `internal/deploy/release.go` — release directory creation, `current` symlink update, purge
- [ ] `artifact` command — full deploy flow (steps 1–11 from spec 03), no hooks
- [ ] `internal/tui/progress.go` — spinner + progress bar wrappers
- [ ] Wire progress bar to extraction byte stream

Deliverable: `bifrost artifact` deploys a real archive end-to-end.

---

## M3 — Hook System

Hook execution engine and `release enable`.

- [ ] `internal/deploy/hooks.go` — `sh -c` execution, sudo, template rendering, priority sort
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

Tests and CI.

- [ ] Unit tests for config merge logic
- [ ] Unit tests for shared resource linking algorithm
- [ ] Unit tests for hook template rendering
- [ ] Integration tests for `artifact` command (real filesystem, temp dir)
- [ ] Integration tests for `release enable` / `release rollback`
- [ ] CI pipeline (lint, test, build for linux-amd64 and linux-arm64)
- [ ] Release pipeline (goreleaser or cog bump + manual artifacts)

---

## Bugs fixed vs v1

| Bug | Status | Fixed in |
|---|---|---|
| BUG-1: `cmd.split(' ')` breaks args with spaces | Fixed | M3 (`sh -c`) |
| BUG-2: sudo + shell operators broken | Fixed | M3 (`sh -c`) |
| BUG-3: extraction stderr hidden on failure | Fixed | M2 (always shown) |
